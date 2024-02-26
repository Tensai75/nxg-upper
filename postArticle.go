package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"regexp"
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

				var (
					segment           nzbparser.NzbSegment
					article           nntp.Article
					partType          string
					partNumber        int64
					totalDownloadSize int64
					nxgHeader         string
				)

				if filepath.Ext(chunk.Filename) == ".par2" {
					partType = "par2"
					totalDownloadSize = totalParSize
					partNumber = parPartCounter.inc()
				} else {
					partType = "data"
					totalDownloadSize = totalDataSize
					partNumber = dataPartCounter.inc()
				}

				segment.Number = int(chunk.PartNumber)
				md5Hash := GetSHA256Hash(fmt.Sprintf("%v:%v:%v", chunk.Nzb.Comment, partType, partNumber))
				segment.Id = md5Hash[:40] + "@" + md5Hash[40:61] + "." + md5Hash[61:]

				article.Header = make(map[string][]string)
				subject := ""
				poster := ""
				if conf.Obfuscate {
					subject = GetSHA256Hash(md5Hash)
					posterHash := GetSHA256Hash(subject)
					poster = posterHash[10:25] + "@" + posterHash[30:45] + "." + posterHash[50:53]
				} else {
					if conf.ObfuscateRar {
						filenameReplace := regexp.MustCompile(`^[^.]*(.*)$`)
						filename := filenameReplace.ReplaceAllString(chunk.Filename, nzb.Comment+"$1")
						subject = fmt.Sprintf("[%v/%v] \"%s\" yEnc (%v/%v)", chunk.FileNumber, chunk.TotalFiles, filename, chunk.PartNumber, chunk.TotalParts)
					} else {
						subject = fmt.Sprintf("[%v/%v] %v - \"%s\" yEnc (%v/%v)", chunk.FileNumber, chunk.TotalFiles, nzb.Comment, chunk.Filename, chunk.PartNumber, chunk.TotalParts)
					}
					poster = conf.Poster
				}

				//x-nxg header
				if nxgHeader, err = encrypt(fmt.Sprintf("%v:%v:%v:%v:%v", chunk.FileNumber, chunk.TotalFiles, chunk.Filename, partType, totalDownloadSize), nzb.Comment); err != nil {
					Log.Warn("Unable to encrypt NxG header: %v", err)
				}

				article.Header["Subject"] = append(article.Header["Subject"], subject)
				article.Header["Date"] = append(article.Header["Date"], time.Now().Format(time.RFC1123))
				article.Header["From"] = append(article.Header["From"], poster)
				article.Header["Message-ID"] = append(article.Header["Message-ID"], "<"+segment.Id+">")
				article.Header["Path"] = append(article.Header["Path"], "")
				article.Header["Newsgroups"] = append(article.Header["Groups"], strings.Join(chunk.Nzb.Files[chunk.FileNumber-1].Groups, ","))
				article.Header["X-Nxg"] = append(article.Header["X-Nxg"], nxgHeader)
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
