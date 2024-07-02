package main

import (
	"encoding/base64"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	progressbar "github.com/schollz/progressbar/v3"
)

type AverageBitrate struct {
	sync.Mutex
	startTime    time.Time
	endTime      time.Time
	bytes        int64
	setStartTime sync.Once
}

func (a *AverageBitrate) start() {
	a.setStartTime.Do(func() {
		a.startTime = time.Now()
		a.endTime = time.Now()
	})
}

func (a *AverageBitrate) calc(size int64) string {
	a.Lock()
	defer a.Unlock()
	a.endTime = time.Now()
	a.bytes += size
	return humanRate(float64(a.bytes) / float64(time.Duration.Seconds(a.endTime.Sub(a.startTime))))
}

func (a *AverageBitrate) get() string {
	a.Lock()
	defer a.Unlock()
	return humanRate(float64(a.bytes) / float64(time.Duration.Seconds(a.endTime.Sub(a.startTime))))
}

var averageBitrate AverageBitrate

func folderPoster(path string) error {

	var (
		parBlockSize uint64
		inputFiles   []string
		err          error
	)

	if totalDataSize, err = calculateFolderSize(path, ""); err != nil {
		return err
	}
	Log.Debug("Total data size: %v Bytes", totalDataSize)

	if conf.MakeRar {
		tempPath = filepath.Join(conf.TempPath, shortHeader)
		if err = os.MkdirAll(tempPath, os.ModePerm); err != nil {
			return err
		}
		Log.Debug("Temp path set to: %v", tempPath)
		if err = rarFolder(path, tempPath, totalDataSize); err != nil {
			return err
		}
	} else {
		tempPath = path
	}

	if conf.MakePar2 {
		parBlockSize = calculateParBlockSize(totalDataSize)
		Log.Debug("Par block size: %v Bytes", parBlockSize)
		if err = parFolder(tempPath, parBlockSize); err != nil {
			return err
		}
		if totalParSize, err = calculateFolderSize(tempPath, ".par2"); err != nil {
			return err
		}
		Log.Debug("Total par size: %v Bytes", totalParSize)
	}

	Log.Debug("Upload path: %v", tempPath)

	if totalDataParts, err = calculateFolderTotalDataParts(tempPath); err != nil {
		return err
	}
	Log.Debug("Total data parts: %v", totalDataParts)
	if totalParParts, err = calculateFolderTotalParParts(tempPath); err != nil {
		return err
	}
	Log.Debug("Total par2 parts: %v", totalParParts)

	// create base64 encoded header
	partBytes := []byte(fmt.Sprintf(":%d:%d", totalDataParts, totalParParts))
	fullHeaderBytes := []byte(shortHeader)
	copy(fullHeaderBytes[len(fullHeaderBytes)-len(partBytes):], partBytes)
	Log.Debug("Unencoded header: %v", string(fullHeaderBytes))
	nzb.Comment = base64.StdEncoding.EncodeToString(fullHeaderBytes)
	Log.Debug("Encoded header: %v", nzb.Comment)

	// launche the go-routines
	for i := uint32(1); i <= conf.Connections*2; i++ {
		go yEncEncoder()
	}
	for i := uint32(1); i <= conf.Connections*2; i++ {
		go articlePoster()
	}
	for i := uint32(1); i <= conf.Connections*2; i++ {
		go poster()
	}
	for i := uint32(1); i <= conf.HeaderCheckConns*2; i++ {
		go headerCheck()
	}

	Log.Info("Starting upload")

	// progress bar
	if conf.Verbose > 0 {
		// the total upload size is an approximation with an estimated overhead of 4% (yEnc and headers)
		totalUploadSize := int((totalDataParts + totalParParts) * conf.ArticleSize * 104 / 100)
		uploadProgressBar = progressbar.NewOptions(totalUploadSize,
			progressbar.OptionShowBytes(true),
			progressbar.OptionSetDescription("INFO:    Uploading    "),
			progressbar.OptionSetRenderBlankState(true),
			progressbar.OptionThrottle(time.Millisecond*100),
			progressbar.OptionShowElapsedTimeOnFinish(),
			progressbar.OptionOnCompletion(newline),
		)
	}

	if err = filepath.WalkDir(tempPath, func(filePath string, dir fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !dir.IsDir() {
			inputFiles = append(inputFiles, filePath)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("error while walking path \"%v\": %v", path, err)
	}

	for i, filePath := range inputFiles {
		select {
		case <-ctx.Done():
			return nil // Error somewhere, terminate
		default: // Default is must to avoid blocking
			filesChan <- File{
				path:       filePath,
				fileNo:     uint64(i + 1),
				totalFiles: uint64(len(inputFiles)),
				groups:     conf.GroupsArray,
				poster:     conf.Poster,
				nzb:        &nzb,
			}
		}
	}

	close(filesChan)
	return nil

}

func calculateFolderSize(path string, extension string) (uint64, error) {
	var size uint64
	if err := filepath.Walk(path, func(file string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if extension == "" || filepath.Ext(file) == extension {
				size += uint64(info.Size())
			}
		}
		return nil
	}); err != nil {
		return 0, err
	}
	return size, nil
}

func calculateFolderTotalDataParts(path string) (uint64, error) {
	var totalParts uint64
	if err := filepath.Walk(path, func(file string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(file) != ".par2" {
			size := uint64(info.Size())
			fileParts := size / conf.ArticleSize
			if size%conf.ArticleSize != 0 {
				fileParts++
			}
			totalParts += fileParts
		}
		return nil
	}); err != nil {
		return 0, err
	}
	return totalParts, nil
}

func calculateFolderTotalParParts(path string) (uint64, error) {
	var totalParts uint64
	if err := filepath.Walk(path, func(file string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(file) == ".par2" {
			size := uint64(info.Size())
			fileParts := size / conf.ArticleSize
			if size%conf.ArticleSize != 0 {
				fileParts++
			}
			totalParts += fileParts
		}
		return nil
	}); err != nil {
		return 0, err
	}
	return totalParts, nil
}

func newline() { fmt.Println() }

func humanRate(b float64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%.1f B/s", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB/s",
		float64(b)/float64(div), "KMGTPE"[exp])
}
