package main

import "fmt"

func saveArticle(article Article) {

	select {
	case <-ctx.Done():
		return // Error somewhere, terminate
	default: // Default is must to avoid blocking

		// todo: save faild articles
		failed := failedMessages.inc()
		if failed > totalParParts*80/100 {
			checkForFatalErr(fmt.Errorf("Posting of %d messages faild. Recovery of upload not possible!", failed))
		}
	}

	return
}
