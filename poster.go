package main

import (
	"fmt"
	"runtime/debug"
	"sync"
	"time"
)

var failedArticlesChan = make(chan Article, 0)

func init() {
	go failedArticlesHandler()
}

func failedArticlesHandler() {
	for {
		article, ok := <-failedArticlesChan
		if !ok {
			return
		}
		go func(article Article) {
			if err := TryCatch(func() { articles.Chan <- article })(); err != nil {
				Log.Debug("Error while trying to add message no %v of file \"%s\" back to the queue: %v", article.Segment.Number, article.Nzb.Files[article.FileNo-1].Filename, err)
				saveArticle(article)
			} else {
				Log.Debug("Added message no %v of file \"%s\" back to the queue", article.Segment.Number, article.Nzb.Files[article.FileNo-1].Filename)
			}
		}(article)
	}
}

func poster(wg *sync.WaitGroup, connNumber int, retries int) {

	defer wg.Done()

	select {
	case <-ctx.Done():
		return // Error somewhere, terminate
	default: // Default is must to avoid blocking

		if retries > 0 {
			Log.Debug("Connection %d waiting %d seconds to reconnect", connNumber, conf.ConnWaitTime)
			time.Sleep(time.Second * time.Duration(conf.ConnWaitTime))
		}

		conn, err := ConnectNNTP()
		if err != nil {
			retries++
			if retries > conf.ConnRetries {
				Log.Error("Connection %d failed after %d retries: %v", connNumber, retries-1, err)
				failed := failedConnections.inc()
				if failed >= int64(conf.Connections) {
					checkForFatalErr(fmt.Errorf("All connections failed"))
				}
				return
			}
			Log.Warn("Connection %d error: %v", connNumber, err)
			Log.Warn("Waiting %v seconds to re-connect", conf.ConnWaitTime)
			wg.Add(1)
			go poster(wg, connNumber, retries)
			return
		} else {
			defer conn.Close()
		}

		for {
			select {
			case <-ctx.Done():
				return // Error somewhere, terminate
			default: // Default is must to avoid blocking
				article, ok := <-articles.Chan
				if !ok {
					return
				}
				select {
				case <-ctx.Done():
					return // Error somewhere, terminate
				default: // Default is must to avoid blocking

					// start bitrate timer
					averageBitrate.start()

					articleToPost := article.Article
					if err = post(conn, &articleToPost); err != nil {
						Log.Debug("Error posting message no %v of file \"%s\": %v", article.Segment.Number, article.Nzb.Files[article.FileNo-1].Filename, err)
						article.Retries++
						if article.Retries <= conf.Retries {
							failedArticlesChan <- article
						} else {
							Log.Warn("After %d retries unable to post message no %v of file \"%s\": %v", conf.Retries, article.Segment.Number, article.Nzb.Files[article.FileNo-1].Filename, err)
							saveArticle(article)
						}
					} else {
						if conf.HeaderCheck {
							if _, err := conn.Head(article.Article.Header["Message-ID"][0]); err != nil {
								Log.Debug("Header check failed for message no %v of file \"%s\": ", article.Segment.Number, article.Nzb.Files[article.FileNo-1].Filename)
								article.Retries++
								if article.Retries <= conf.Retries {
									failedArticlesChan <- article
								} else {
									Log.Warn("After %d retries unable to post message no %v of file \"%s\": Header check failed", conf.Retries, article.Segment.Number, article.Nzb.Files[article.FileNo-1].Filename)
									saveArticle(article)
								}
							}
						}
						article.Nzb.Files[article.FileNo-1].Date = int(time.Now().Unix())
						article.Nzb.Files[article.FileNo-1].Segments = append(article.Nzb.Files[article.FileNo-1].Segments, *article.Segment)
						postedMessages.inc()
						averageBitrate.calc(int64(article.Segment.Bytes))
						if conf.Verbose > 0 {
							uploadProgressBar.Add(article.Segment.Bytes)
						}
					}
				}
			}
		}
	}
}

func TryCatch(f func()) func() error {
	return func() (err error) {
		defer func() {
			if panicInfo := recover(); panicInfo != nil {
				err = fmt.Errorf("%v, %s", panicInfo, string(debug.Stack()))
				return
			}
		}()
		f() // calling the decorated function
		return err
	}
}
