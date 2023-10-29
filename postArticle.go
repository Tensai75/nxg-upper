package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Tensai75/nntp"
	"github.com/Tensai75/nzb-monkey-go/nzbparser"
)

func articlePoster(wg *sync.WaitGroup) {

	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return // Error somewhere, terminate
		default: // Default is must to avoid blocking
			chunk, ok := <-outputChunks.Chan
			if !ok {
				return
			}
			select {
			case <-ctx.Done():
				return // Error somewhere, terminate
			default: // Default is must to avoid blocking

				var segment nzbparser.NzbSegment
				var article nntp.Article
				var partType string
				var partNumber int64

				if filepath.Ext(chunk.Filename) == ".par2" {
					partType = "par2"
					partNumber = parPartCounter.inc()
				} else {
					partType = "data"
					partNumber = dataPartCounter.inc()
				}

				segment.Number = int(chunk.PartNumber)
				md5Hash := GetSHA256Hash(fmt.Sprintf("%v:%v:%v", chunk.Nzb.Comment, partType, partNumber))
				segment.Id = md5Hash[:40] + "@" + md5Hash[40:61] + "." + md5Hash[61:]

				article.Header = make(map[string][]string)
				subject := ""
				if conf.Obfuscate {
					subject = md5Hash
				} else {
					subject = fmt.Sprintf("[%v/%v] %v - \"%s\" yEnc (%v/%v)", chunk.FileNumber, chunk.TotalFiles, nzb.Comment, chunk.Filename, chunk.PartNumber, chunk.TotalParts)
				}
				article.Header["Subject"] = append(article.Header["Subject"], subject)
				article.Header["Date"] = append(article.Header["Date"], time.Now().Format(time.RFC1123))
				article.Header["From"] = append(article.Header["From"], conf.Poster)
				article.Header["Message-ID"] = append(article.Header["Message-ID"], "<"+segment.Id+">")
				article.Header["Path"] = append(article.Header["Path"], "")
				article.Header["Newsgroups"] = append(article.Header["Groups"], strings.Join(chunk.Nzb.Files[chunk.FileNumber-1].Groups, ","))
				article.Body = &chunk.Part

				segment.Bytes = chunk.Part.Len()
				for _, header := range article.Header {
					segment.Bytes += len(header[0])
				}

				articles.Chan <- Article{
					Segment: &segment,
					Nzb:     chunk.Nzb,
					Article: article,
					FileNo:  chunk.FileNumber,
				}
			}
		}
	}

}

func GetSHA256Hash(text string) string {
	hasher := sha256.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
