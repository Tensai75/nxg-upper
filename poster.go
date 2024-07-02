package main

import (
	"time"
)

func poster() {
	for {
		select {
		case <-ctx.Done():
			return // Error somewhere, terminate
		default: // Default is must to avoid blocking
			article, ok := <-articlesChan
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
				if err := post(&articleToPost); err != nil {
					Log.Debug("Error posting message no %v of file \"%s\": %v", article.Segment.Number, article.Nzb.Files[article.FileNo-1].Filename, err)
					article.Retries++
					if article.Retries <= conf.Retries {
						chunksWG.Add(1)
						articlesChan <- article
					} else {
						Log.Warn("After %d retries unable to post message no %v of file \"%s\": %v", conf.Retries, article.Segment.Number, article.Nzb.Files[article.FileNo-1].Filename, err)
						saveArticle(article)
					}
				} else {
					averageBitrate.calc(int64(article.Segment.Bytes))
					if conf.Verbose > 0 {
						uploadProgressBar.Add(article.Segment.Bytes)
					}
					if conf.HeaderCheck {
						article.PostTime = time.Now()
						chunksWG.Add(1)
						headerCheckChan <- article
					} else {
						finalizePost(article)
					}
				}
				chunksWG.Done()
			}
		}
	}
}

func headerCheck() {

	headerCheckWaitTime := time.Duration(conf.HeaderCheckDelay) * time.Second

	for {
		select {
		case <-ctx.Done():
			return // Error somewhere, terminate
		default: // Default is must to avoid blocking
			article, ok := <-headerCheckChan
			if !ok {
				return
			}
			select {
			case <-ctx.Done():
				return // Error somewhere, terminate
			default: // Default is must to avoid blocking
				waitTime := time.Until(article.PostTime.Add(headerCheckWaitTime))
				time.Sleep(waitTime)
				if conf.Test != "" {
					// for posting error testing
					counter := testPostCounter.inc()
					if counter%100 == 0 {
						chunksWG.Add(1)
						go headerCheckError(article)
					} else {
						finalizePost(article)
					}
				} else {
					if conn, err := headerCheckPool.Get(ctx); err != nil {
						checkForFatalErr(err)
					} else {
						if _, _, err := conn.Stat(article.Article.Header["Message-ID"][0]); err != nil {
							headerCheckPool.Put(conn)
							chunksWG.Add(1)
							go headerCheckError(article)
						} else {
							headerCheckPool.Put(conn)
							finalizePost(article)
						}
					}
				}
				chunksWG.Done()
			}
		}
	}
}

func headerCheckError(article Article) {
	Log.Debug("Header check failed for message no %v of file \"%s\": ", article.Segment.Number, article.Nzb.Files[article.FileNo-1].Filename)
	article.Retries++
	if article.Retries <= conf.Retries {
		articlesChan <- article
	} else {
		Log.Warn("After %d retries unable to post message no %v of file \"%s\": Header check failed", conf.Retries, article.Segment.Number, article.Nzb.Files[article.FileNo-1].Filename)
		saveArticle(article)
		chunksWG.Done()
	}
}

func finalizePost(article Article) {
	if article.Retries > 0 {
		Log.Debug("Posting of message no %v of file \"%s\" successful after %v retries", article.Segment.Number, article.Nzb.Files[article.FileNo-1].Filename, article.Retries)
	}
	article.Nzb.Files[article.FileNo-1].Date = int(time.Now().Unix())
	article.Nzb.Files[article.FileNo-1].Segments = append(article.Nzb.Files[article.FileNo-1].Segments, *article.Segment)
	postedMessages.inc()
}
