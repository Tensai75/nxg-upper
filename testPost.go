package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Tensai75/nntp"
)

var (
	testPostCounter Counter
)

func (c *safeConn) testPost(article *nntp.Article) error {
	// for posting error testing
	counter := testPostCounter.inc()
	if counter%150 == 0 {
		// return fmt.Errorf("Test Error")
	}

	return writeArticle(testPath, article)
}

func post(conn *safeConn, article *nntp.Article) error {
	if conf.Test != "" {
		return conn.testPost(article)
	}
	return conn.Post(article)
}

func writeArticle(path string, article *nntp.Article) error {

	outputFile, err := os.Create(filepath.Join(path, fmt.Sprintf("%v.txt", strings.TrimFunc(article.Header["Message-ID"][0], func(r rune) bool {
		if string(r) == "<" || string(r) == ">" {
			return true
		}
		return false
	}))))
	defer outputFile.Close()
	if err != nil {
		return err
	}
	for output := range article.Header {
		text := output + ": " + article.Header[output][0] + "\r\n"
		if _, err = outputFile.Write([]byte(text)); err != nil {
			return err
		}
	}
	bytes := make([]byte, conf.ArticleSize)
	var readBytes int
	for {
		if readBytes, err = article.Body.Read(bytes); err != nil {
			if err.Error() == "EOF" {
				break
			}
		}
		if _, err = outputFile.Write(bytes[:readBytes]); err != nil {
			return err
		}
	}

	savedMessages.inc()
	return nil

}
