package main

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
)

func rarFolder(path string, tempPath string, folderSize int64) error {

	Log.Info("Starting rar process")

	var (
		volumeSize     int64
		parameters     []string
		cmdReader      io.ReadCloser
		scanner        *bufio.Scanner
		rarProgressBar *progressbar.ProgressBar
		err            error
	)

	if volumeSize, err = setVolumeSize(folderSize, conf.ArticleSize); err != nil {
		return err
	}
	Log.Debug("Rar volume size: %v Bytes", volumeSize)

	parameters = append(parameters, "a", "-o+", "-idcd", "-ep1", "-r")
	parameters = append(parameters, fmt.Sprintf("-m%v", conf.Compression))
	if conf.Encrypt || conf.Password != "" {
		if conf.Password == "" {
			conf.Password = randomString(30, 0, false)
		}
		parameters = append(parameters, fmt.Sprintf("-hp%v", conf.Password))
	}
	if volumeSize > 0 {
		parameters = append(parameters, fmt.Sprintf("-v%vb", volumeSize))
	}
	parameters = append(parameters, filepath.Join(tempPath, shortHeader+".rar"))
	parameters = append(parameters, filepath.Join(path, "*"))

	cmd := exec.Command(conf.RarExe, parameters...)
	Log.Debug("Rar command: %s", cmd.String())
	if conf.Debug || conf.Verbose > 0 {
		// create a pipe for the output of the program
		if cmdReader, err = cmd.StdoutPipe(); err != nil {
			return err
		}
		scanner = bufio.NewScanner(cmdReader)
		go func() {
			// progress bar
			if conf.Verbose > 0 {
				rarProgressBar = progressbar.NewOptions(int(100),
					progressbar.OptionSetDescription("INFO:    Creating rar "),
					progressbar.OptionSetRenderBlankState(true),
					progressbar.OptionThrottle(time.Millisecond*100),
					progressbar.OptionShowElapsedTimeOnFinish(),
					progressbar.OptionOnCompletion(newline),
				)
			}

			for scanner.Scan() {
				output := strings.Trim(scanner.Text(), " \r\n")
				if conf.Debug {
					if output != "" && !strings.Contains(output, "%") {
						Log.Debug("RAR: %v", output)
					}
				}
				if conf.Verbose > 0 {
					exp := regexp.MustCompile(`0*(\d+)%`)
					if output != "" && exp.MatchString(output) {
						percentStr := exp.FindStringSubmatch(output)
						percentInt, _ := strconv.Atoi(percentStr[1])
						rarProgressBar.Set(percentInt)
					}
				}
			}
		}()
	}
	if err = cmd.Run(); err != nil {
		return err
	}
	if conf.Verbose > 0 {
		rarProgressBar.Finish()
	}
	Log.Info("Creation of rar archives successful")
	return nil
}

func setVolumeSize(folderSize int64, blockSize int64) (int64, error) {

	defaultVolumeSize := calculateVolumeSize(conf.VolumeSize, blockSize)

	if conf.MakeVolumes && defaultVolumeSize < folderSize {
		if (conf.MaxVolumes == 0) || (folderSize/int64(conf.MaxVolumes) <= defaultVolumeSize) {
			return defaultVolumeSize, nil
		} else {
			volumeSize := (folderSize / int64(conf.MaxVolumes)) + 1
			return calculateVolumeSize(volumeSize, blockSize), nil
		}
	}
	return 0, nil
}

func calculateVolumeSize(volumeSize int64, blockSize int64) int64 {
	if volumeSize > 0 {
		multiplier := volumeSize / blockSize
		if volumeSize%blockSize != 0 {
			multiplier++
		}
		return multiplier * blockSize
	}
	return 0
}
