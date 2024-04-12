package main

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"hash/crc32"
	"math"
	"sync"

	"golang.org/x/exp/slices"
)

func yEncEncoder(wg *sync.WaitGroup) {

	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return // Error somewhere, terminate
		default: // Default is must to avoid blocking

			chunk, ok := <-inputChunks.Chan
			if !ok {
				return
			}
			select {
			case <-ctx.Done():
				return // Error somewhere, terminate
			default: // Default is must to avoid blocking
				func() {
					defer chunksWG.Done()

					// Calculate the CRC32 checksum of the chunk
					partChecksum := crc32.ChecksumIEEE(chunk.Part.Bytes())

					// Open a byte writer
					var outputChunk bytes.Buffer

					// obfuscate yenc file name if activated
					filename := ""
					if conf.Obfuscate && conf.ObfuscateYenc {
						hasher := md5.New()
						hasher.Write([]byte(fmt.Sprintf("%s%d", chunk.Filename, chunk.PartNumber)))
						filename = fmt.Sprintf("%x", hasher.Sum(nil))
					} else {
						filename = chunk.Filename
					}

					// Write yEnc part header
					header := fmt.Sprintf("=ybegin part=%d total=%d line=%d size=%d name=%s\r\n",
						chunk.PartNumber, chunk.TotalParts, conf.LineLength, chunk.TotalSize, filename)
					_, err := outputChunk.WriteString(header)
					if err != nil {
						checkForFatalErr(fmt.Errorf("Error writing yEnc header for part %d: %v\r\n", chunk.PartNumber, err))
					}
					partHeader := fmt.Sprintf("=ypart begin=%d end=%d size=%d\r\n",
						chunk.StartByte, chunk.EndByte, chunk.PartSize)
					_, err = outputChunk.WriteString(partHeader)
					if err != nil {
						checkForFatalErr(fmt.Errorf("Error writing yEnc part header for part %d: %v\r\n", chunk.PartNumber, err))
					}

					specialChar := []byte{0x3D, 0x00, 0x0A, 0x0D, 0x7F}

					// Encode the part to yEnc format
					for j := 1; j <= len(chunk.Part.Bytes()); j++ {
						b := chunk.Part.Bytes()[j-1]

						// yEnc encode the character
						b = byte(math.Mod(float64(b+42), float64(256)))

						// Escape special yEnc characters
						if slices.Contains(specialChar, b) {
							outputChunk.WriteByte('=')
							outputChunk.WriteByte(byte(math.Mod(float64(b+64), float64(256))))
							// Escape spaces at the line start or end
						} else if b == ' ' && (math.Mod(float64(j), float64(conf.LineLength)) == 0 || math.Mod(float64(j-1), float64(conf.LineLength)) == 0) {
							outputChunk.WriteByte('=')
							outputChunk.WriteByte(byte(math.Mod(float64(b+64), float64(256))))
						} else {
							outputChunk.WriteByte(b)
						}

						// Write line break after each line
						if math.Mod(float64(j), float64(conf.LineLength)) == 0 || j == len(chunk.Part.Bytes()) {
							outputChunk.WriteString("\r\n")
						}
					}

					// Write yEnc part footer
					var partFooter string
					if chunk.PartNumber == chunk.TotalParts {
						partFooter = fmt.Sprintf("=yend size=%d part=%d pcrc32=%08X crc32=%08X\r\n", chunk.PartSize, chunk.PartNumber, partChecksum, chunk.Checksum)
					} else {
						partFooter = fmt.Sprintf("=yend size=%d part=%d pcrc32=%08X\r\n", chunk.PartSize, chunk.PartNumber, partChecksum)
					}
					_, err = outputChunk.WriteString(partFooter)
					if err != nil {
						checkForFatalErr(fmt.Errorf("Error to writing yEnc part footer for part %d: %v\r\n", chunk.PartNumber, err))
					}

					chunk.Part = outputChunk
					chunksWG.Add(1)
					outputChunks.Chan <- chunk
				}()
			}
		}
	}
}
