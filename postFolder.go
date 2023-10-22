package main

import (
	"encoding/base64"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	progressbar "github.com/schollz/progressbar/v3"
)

func folderPoster(path string) error {

	var (
		folderSize   int64
		parBlockSize int64
		inputFiles   []string
		err          error
	)

	if conf.MakeRar {
		tempPath = filepath.Join(conf.TempPath, shortHeader)
		if err = os.MkdirAll(tempPath, os.ModePerm); err != nil {
			return err
		}
		Log.Debug("Temp path set to: %v", tempPath)
		if folderSize, err = calculateFolderSize(path); err != nil {
			return err
		}
		if err = rarFolder(path, tempPath, folderSize); err != nil {
			return err
		}
	} else {
		tempPath = path
	}

	if conf.MakePar2 {
		if folderSize, err = calculateFolderSize(tempPath); err != nil {
			return err
		}
		parBlockSize = calculateParBlockSize(folderSize)
		Log.Debug("Par block size: %v Bytes", parBlockSize)
		if err = parFolder(tempPath, parBlockSize); err != nil {
			return err
		}
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
	for i := 1; i <= conf.Connections*2; i++ {
		yEncoderWG.Add(1)
		go yEncEncoder(&yEncoderWG)
	}
	for i := 1; i <= conf.Connections*2; i++ {
		articlePosterWG.Add(1)
		go articlePoster(&articlePosterWG)
	}
	for i := 1; i <= conf.Connections; i++ {
		posterWG.Add(1)
		go poster(&posterWG, i, 0)
	}

	Log.Info("Starting upload")

	// progress bar
	if conf.Verbose > 0 {
		uploadProgressBar = progressbar.NewOptions(int(totalDataParts+totalParParts),
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
		return fmt.Errorf("Error while walking path \"%v\": %v", path, err)
	}

	for i, filePath := range inputFiles {
		select {
		case <-ctx.Done():
			return nil // Error somewhere, terminate
		default: // Default is must to avoid blocking
			files.Chan <- File{
				path:       filePath,
				fileNo:     i + 1,
				totalFiles: len(inputFiles),
				groups:     conf.GroupsArray,
				poster:     conf.Poster,
				nzb:        &nzb,
			}
		}
	}

	files.close()
	return nil

}

func calculateFolderSize(path string) (int64, error) {
	var size int64
	if err = filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	}); err != nil {
		return 0, err
	}
	return size, nil
}

func calculateFolderTotalDataParts(path string) (int64, error) {
	var totalParts int64
	if err = filepath.Walk(path, func(file string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(file) != ".par2" {
			fileParts := info.Size() / conf.ArticleSize
			if info.Size()%conf.ArticleSize != 0 {
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

func calculateFolderTotalParParts(path string) (int64, error) {
	var totalParts int64
	if err = filepath.Walk(path, func(file string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(file) == ".par2" {
			fileParts := info.Size() / conf.ArticleSize
			if info.Size()%conf.ArticleSize != 0 {
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
