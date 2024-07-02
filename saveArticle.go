package main

import "fmt"

func saveArticle(article Article) {

	select {
	case <-ctx.Done():
		return // Error somewhere, terminate
	default: // Default is must to avoid blocking

		// todo: save faild articles
		failed := failedMessages.inc()
		if failed > totalParParts {
			checkForFatalErr(fmt.Errorf("posting of %d messages faild (recovery of upload not possible)", failed))
		}
	}

}
