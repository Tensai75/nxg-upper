package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
)

func parFolder(path string, parBlockSize uint64) error {

	Log.Info("Starting par process")

	var (
		parameters     []string
		cmdReader      io.ReadCloser
		scanner        *bufio.Scanner
		parProgressBar *progressbar.ProgressBar
		err            error
	)

	// clean target folder from par2 files
	Log.Debug("Deleting par2 files")
	if err = deleteAll(filepath.Join(path, "*.par2")); err != nil {
		return err
	}

	parName := ""
	if conf.MakeRar && conf.ObfuscateRar {
		parName = shortHeader
	} else {
		parName = filepath.Base(conf.Path)
	}

	// detect par executable
	parExe := filepath.Base(conf.Par2Exe)
	parExeFileName := strings.ToLower(parExe[:len(parExe)-len(filepath.Ext(parExe))])

	// set parameters
	switch parExeFileName {
	case "par2":
		parameters = append(parameters, "create", "-l")
		parameters = append(parameters, fmt.Sprintf("-s%v", parBlockSize))
		parameters = append(parameters, fmt.Sprintf("-r%v", conf.Redundancy))
		parameters = append(parameters, fmt.Sprintf("%v", filepath.Join(path, parName)))
	case "parpar":
		parameters = append(parameters, fmt.Sprintf("-p%vB", conf.VolumeSize))
		parameters = append(parameters, fmt.Sprintf("-s%vB", parBlockSize))
		parameters = append(parameters, fmt.Sprintf("-r%v%%", conf.Redundancy))
		parameters = append(parameters, fmt.Sprintf("-o%v", filepath.Join(path, parName+".par2")))
	default:
		return fmt.Errorf("unknown par executable: %s", conf.Par2Exe)
	}

	if err = filepath.WalkDir(path, func(filePath string, dir fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !dir.IsDir() {
			parameters = append(parameters, filePath)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("error while walking path \"%v\": %v", path, err)
	}

	cmd := exec.Command(conf.Par2Exe, parameters...)
	Log.Debug("Par command: %s", cmd.String())
	if conf.Debug || conf.Verbose > 0 {
		// create a pipe for the output of the program
		switch parExeFileName {
		case "par2":
			if cmdReader, err = cmd.StdoutPipe(); err != nil {
				return err
			}
		case "parpar":
			if cmdReader, err = cmd.StderrPipe(); err != nil {
				return err
			}
		}
		scanner = bufio.NewScanner(cmdReader)
		scanner.Split(scanLines)
		go func() {
			// progress bar
			if conf.Verbose > 0 {
				parProgressBar = progressbar.NewOptions(int(100),
					progressbar.OptionSetDescription("INFO:    Creating par2"),
					progressbar.OptionSetRenderBlankState(true),
					progressbar.OptionThrottle(time.Millisecond*100),
					progressbar.OptionShowElapsedTimeOnFinish(),
					progressbar.OptionOnCompletion(newline),
				)
			}

			for scanner.Scan() {
				output := strings.Trim(scanner.Text(), " \r\n")
				if output != "" && !strings.Contains(output, "%") {
					Log.Debug("PAR: %v", output)
				}
				if conf.Verbose > 0 {
					exp := regexp.MustCompile(`(\d+)\.?\d*%`)
					if output != "" && exp.MatchString(output) {
						percentStr := exp.FindStringSubmatch(output)
						percentInt, _ := strconv.Atoi(percentStr[1])
						parProgressBar.Set(percentInt)
					}
				}
			}

		}()
	}
	if err = cmd.Run(); err != nil {
		return err
	}
	if conf.Verbose > 0 {
		parProgressBar.Finish()
	}
	Log.Info("Creation of par2 files successful")
	return nil
}

func calculateParBlockSize(folderSize uint64) uint64 {

	maxParBlocks := uint64(32768)

	if folderSize/conf.ArticleSize < maxParBlocks {
		return conf.ArticleSize
	} else {
		minParBlockSize := (folderSize / maxParBlocks) + 1
		multiplier := minParBlockSize / conf.ArticleSize
		if minParBlockSize%conf.ArticleSize != 0 {
			multiplier++
		}
		return multiplier * conf.ArticleSize
	}
}

func deleteAll(path string) (err error) {
	contents, err := filepath.Glob(path)
	Log.Debug("Files to delete: %s", contents)
	if err != nil {
		return
	}
	for _, item := range contents {
		err = os.RemoveAll(item)
		if err != nil {
			return
		}
	}
	return
}

func scanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexAny(data, "\r\n"); i >= 0 {
		if data[i] == '\n' {
			// We have a line terminated by single newline.
			return i + 1, data[0:i], nil
		}
		advance = i + 1
		if len(data) > i+1 && data[i+1] == '\n' {
			advance += 1
		}
		return advance, data[0:i], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}
