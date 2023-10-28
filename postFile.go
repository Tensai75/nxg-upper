package main

import (
	"bufio"
	"bytes"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"sync"

	"github.com/Tensai75/nzb-monkey-go/nzbparser"
)

func filePoster(wg *sync.WaitGroup) {

	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return // Error somewhere, terminate
		default: // Default is must to avoid blocking
			file, ok := <-files.Chan
			if !ok {
				return
			}
			select {
			case <-ctx.Done():
				return // Error somewhere, terminate
			default: // Default is must to avoid blocking

				// Open the input file
				inputFile, err := os.Open(file.path)
				if err != nil {
					checkForFatalErr(fmt.Errorf("Error opening input file: %w", err))
				}
				defer inputFile.Close()

				// Get the file name and file size
				fileInfo, err := inputFile.Stat()
				if err != nil {
					checkForFatalErr(fmt.Errorf("Error getting the input file info: %w", err))
				}
				fileName := fileInfo.Name()
				totalSize := fileInfo.Size()

				// Calculate the CRC32 checksum
				crc := crc32.NewIEEE()
				if _, err := io.Copy(crc, inputFile); err != nil {
					checkForFatalErr(fmt.Errorf("Error calculating the CRC32 checksum: %w", err))
				}
				checksum := crc.Sum32()

				// Reset the input file offset to start encoding from the beginning
				if _, err := inputFile.Seek(0, io.SeekStart); err != nil {
					checkForFatalErr(fmt.Errorf("Error seeking the input file: %w", err))
				}

				// Calculate the number of parts
				totalParts := totalSize / conf.ArticleSize
				if totalSize%conf.ArticleSize != 0 {
					totalParts++
				}

				var partNumber int64 = 1

				// prepare the nzbFile
				var nzbFile nzbparser.NzbFile
				file.nzb.Files = append(file.nzb.Files, nzbFile)
				file.nzb.Files[file.fileNo-1].Groups = append(file.nzb.Files[file.fileNo-1].Groups, file.groups...)
				file.nzb.Files[file.fileNo-1].Poster = file.poster
				file.nzb.Files[file.fileNo-1].Filename = fileName

				// Create a buffered reader and chunk
				inputReader := bufio.NewReader(inputFile)
				chunk := make([]byte, conf.ArticleSize)

				// Read the file in chunks
			FOR:
				for {
					select {
					case <-ctx.Done():
						return // Error somewhere, terminate
					default: // Default is must to avoid blocking
						inputBuffer := bytes.NewBuffer(make([]byte, 0))
						var readBytes int
						if readBytes, err = inputReader.Read(chunk); err != nil {
							if err.Error() == "EOF" {
								break FOR
							}
							checkForFatalErr(fmt.Errorf("Error reading input file: %w", err))
						}
						inputBuffer.Write(chunk[:readBytes])

						file.nzb.Files[file.fileNo-1].Subject = fmt.Sprintf("[%v/%v] %s - \"%s\" yEnc (%v/%v)", file.fileNo, file.totalFiles, nzb.Comment, fileName, partNumber, totalParts)

						inputChunks.Chan <- Chunk{
							Filename:   fileName,
							Checksum:   checksum,
							FileNumber: file.fileNo,
							TotalFiles: file.totalFiles,
							PartNumber: partNumber,
							TotalParts: totalParts,
							PartSize:   int64(readBytes),
							TotalSize:  totalSize,
							StartByte:  ((partNumber - 1) * conf.ArticleSize) + 1,
							EndByte:    ((partNumber - 1) * conf.ArticleSize) + int64(readBytes),
							Part:       *inputBuffer,
							Nzb:        file.nzb,
						}
						partNumber++
					}
				}
				file.nzb.Files[file.fileNo-1].Subject = fmt.Sprintf("[%v/%v] %v \"%s\" (1/%v) yEnc", file.fileNo, file.totalFiles, nzb.Comment, fileName, totalParts)
			}
		}
	}

}
