package main

import (
	"fmt"
	"sync"
	"time"
)

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
							chunksWG.Add(1)
							articles.Chan <- article
						} else {
							Log.Warn("After %d retries unable to post message no %v of file \"%s\": %v", conf.Retries, article.Segment.Number, article.Nzb.Files[article.FileNo-1].Filename, err)
							saveArticle(article)
						}
					} else {
						if conf.HeaderCheck {
							if err := headerCheck(conn, &article); err != nil {
								Log.Debug("Header check failed for message no %v of file \"%s\": ", article.Segment.Number, article.Nzb.Files[article.FileNo-1].Filename)
								article.Retries++
								if article.Retries <= conf.Retries {
									chunksWG.Add(1)
									articles.Chan <- article
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
					chunksWG.Done()
				}
			}
		}
	}
}
