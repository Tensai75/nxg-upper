package main

import (
	"bufio"
	"bytes"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"regexp"

	"github.com/Tensai75/nzbparser"
)

func filePoster() {

	for {
		select {
		case <-ctx.Done():
			return // Error somewhere, terminate
		default: // Default is must to avoid blocking
			file, ok := <-filesChan
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
					checkForFatalErr(fmt.Errorf("error opening input file: %w", err))
				}
				defer inputFile.Close()

				// Get the file name and file size
				fileInfo, err := inputFile.Stat()
				if err != nil {
					checkForFatalErr(fmt.Errorf("error getting the input file info: %w", err))
				}
				fileName := fileInfo.Name()
				totalSize := uint64(fileInfo.Size())

				// Calculate the CRC32 checksum
				crc := crc32.NewIEEE()
				if _, err := io.Copy(crc, inputFile); err != nil {
					checkForFatalErr(fmt.Errorf("error calculating the CRC32 checksum: %w", err))
				}
				checksum := crc.Sum32()

				// Reset the input file offset to start encoding from the beginning
				if _, err := inputFile.Seek(0, io.SeekStart); err != nil {
					checkForFatalErr(fmt.Errorf("error seeking the input file: %w", err))
				}

				// Calculate the number of parts
				totalParts := totalSize / conf.ArticleSize
				if totalSize%conf.ArticleSize != 0 {
					totalParts++
				}

				var partNumber uint64 = 1

				// prepare the nzbFile
				var nzbFile nzbparser.NzbFile
				nzbFile.Groups = append(nzbFile.Groups, file.groups...)
				nzbFile.Poster = file.poster
				nzbFile.Filename = fileName
				file.nzb.Files = append(file.nzb.Files, nzbFile)

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
							checkForFatalErr(fmt.Errorf("error reading input file: %w", err))
						}
						inputBuffer.Write(chunk[:readBytes])

						file.nzb.Files[file.fileNo-1].Subject = fmt.Sprintf("[%v/%v] %s - \"%s\" yEnc (%v/%v)", file.fileNo, file.totalFiles, nzb.Comment, fileName, partNumber, totalParts)

						chunksWG.Add(1)
						inputChunksChan <- Chunk{
							Filename:   fileName,
							Checksum:   checksum,
							FileNumber: file.fileNo,
							TotalFiles: file.totalFiles,
							PartNumber: partNumber,
							TotalParts: totalParts,
							PartSize:   uint64(readBytes),
							TotalSize:  totalSize,
							StartByte:  ((partNumber - 1) * conf.ArticleSize) + 1,
							EndByte:    ((partNumber - 1) * conf.ArticleSize) + uint64(readBytes),
							Part:       *inputBuffer,
							Nzb:        file.nzb,
						}
						partNumber++
					}
				}
				if conf.MakeRar && conf.ObfuscateRar {
					filenameReplace := regexp.MustCompile(`^[^.]*(.*)$`)
					filename := filenameReplace.ReplaceAllString(fileName, nzb.Comment+"$1")
					file.nzb.Files[file.fileNo-1].Subject = fmt.Sprintf("[%v/%v] \"%s\" (1/%v) yEnc", file.fileNo, file.totalFiles, filename, totalParts)
				} else {
					file.nzb.Files[file.fileNo-1].Subject = fmt.Sprintf("[%v/%v] %s - \"%s\" (1/%v) yEnc", file.fileNo, file.totalFiles, nzb.Comment, fileName, totalParts)
				}
			}
		}
	}

}
