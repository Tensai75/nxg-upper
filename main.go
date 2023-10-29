package main

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/Tensai75/nntp"
	"github.com/Tensai75/nzb-monkey-go/nzbparser"
	"github.com/schollz/progressbar/v3"
)

type Chunk struct {
	Filename   string
	Checksum   uint32
	FileNumber int
	TotalFiles int
	PartNumber int64
	TotalParts int64
	PartSize   int64
	TotalSize  int64
	StartByte  int64
	EndByte    int64
	Part       bytes.Buffer
	Nzb        *nzbparser.Nzb
}

type File struct {
	path       string
	fileNo     int
	totalFiles int
	groups     []string
	poster     string
	nzb        *nzbparser.Nzb
}

type Article struct {
	Segment *nzbparser.NzbSegment
	Nzb     *nzbparser.Nzb
	Article nntp.Article
	FileNo  int
	Retries int
}

type Counter struct {
	mu      sync.Mutex
	counter int64
}

func (c *Counter) inc() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counter++
	return c.counter
}

// global variables
var (
	appName        = "NxG Upper"
	appVersion     string
	logFileName    = "nxg-upper.log"
	configFileName = "nxg-upper.conf"

	appExec           string
	appPath           string
	homePath          string
	tempPath          string
	testPath          string
	postFunc          func(*nntp.Article) error
	nzb               nzbparser.Nzb
	title             string
	shortHeader       string
	totalDataParts    int64
	totalParParts     int64
	uploadProgressBar *progressbar.ProgressBar
	fatalError        sync.Once
	err               error

	// counters
	dataPartCounter   Counter
	parPartCounter    Counter
	postedMessages    Counter
	failedMessages    Counter
	savedMessages     Counter
	failedConnections Counter

	// WaitGroups
	filePosterWG    sync.WaitGroup
	yEncoderWG      sync.WaitGroup
	articlePosterWG sync.WaitGroup
	posterWG        sync.WaitGroup

	// cancel context
	ctx, cancel = context.WithCancel(context.Background())
)

func init() {
	// set path variables
	if appExec, err = os.Executable(); err != nil {
		checkForFatalErr(fmt.Errorf("Unable to determine application path"))
	}
	appPath = filepath.Dir(appExec)
	if homePath, err = os.UserHomeDir(); err != nil {
		checkForFatalErr(fmt.Errorf("Unable to determine home path"))
	}
}

func main() {
	defer cancel()
	setConfPath()
	loadConfig()
	// init logger
	if conf.LogFilePath != "" {
		initLogger(conf.LogFilePath)
		defer logClose()
	}
	parseArguments()
	checkArguments()

	// check article size
	if conf.ArticleSize > 0 && conf.ArticleSize%64 != 0 {
		conf.ArticleSize = (conf.ArticleSize / 64) * 64
		Log.Warn("Provided article size is not a multiple of 64 Bytes")
		Log.Warn("Article size set to %v bytes", conf.ArticleSize)
	} else if conf.ArticleSize == 0 {
		conf.ArticleSize = 768000
		Log.Warn("No article size provided. Article size set to %v bytes", conf.ArticleSize)
	}

	title = filepath.Base(conf.Path)
	shortHeader = randomString(21, 0, true)

	// check for test mode
	if conf.Test != "" {
		testPath = filepath.Join(conf.Test, title)
		if err = os.MkdirAll(testPath, os.ModePerm); err != nil {
			checkForFatalErr(fmt.Errorf("Cannot create test path: %v", err))
		}
	}

	// launche the go-routines
	filePosterWG.Add(1)
	go filePoster(&filePosterWG)

	if err = folderPoster(conf.Path); err != nil {
		checkForFatalErr(err)
	}

	filePosterWG.Wait()
	inputChunks.close()
	yEncoderWG.Wait()
	outputChunks.close()
	articlePosterWG.Wait()
	articles.close()
	posterWG.Wait()

	// prevent execution if fatal error occurred
	if ctx.Err() == nil {
		if conf.Verbose > 0 {
			uploadProgressBar.Finish()
		}
		Log.Info("Upload successful with average upload speed at %s", averageBitrate.get())
		// Update nzb
		nzb.Meta = make(map[string]string)
		nzb.Meta["Title"] = title
		if conf.Password != "" {
			nzb.Meta["Password"] = conf.Password
		}

		// sort the segments and files
		for _, file := range nzb.Files {
			sort.Sort(file.Segments)
		}
		sort.Sort(nzb.Files)

		// Create the nzb file
		fileName := filepath.Base(conf.Path)
		if err := os.MkdirAll(conf.NzbPath, os.ModePerm); err != nil {
			Log.Error("Cannot create the NZB file: %w", err)
		}
		if conf.Password != "" {
			fileName += fmt.Sprintf("{{%s}}", conf.Password)
		}
		outputFile, err := os.Create(filepath.Join(conf.NzbPath, fileName+".nzb"))
		if err != nil {
			Log.Error("Cannot create the NZB file: %w", err)
		}
		defer outputFile.Close()
		nzbByte, err := nzbparser.Write(&nzb)
		if err != nil {
			Log.Error("Cannot create the NZB xml structure: %w", err)
		}
		_, err = outputFile.Write(nzbByte)
		if err != nil {
			Log.Error("Cannot write the NZB file: %w", err)
		}
		writeCsv()
		exit(0, nil)
	}
	// if fatal error occurred wait for the check error function to terminate the program
	time.Sleep(math.MaxInt64)
}

func randomString(length int, salt int64, num bool) string {
	var charset string
	if num {
		charset = "0123456789abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	} else {
		charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	}
	source, _ := strconv.ParseInt((fmt.Sprintf("%v%v", salt, time.Now().UnixNano())), 10, 64)
	seededRand := rand.New(rand.NewSource(source))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset)-1)]
	}
	return string(b)
}

func checkForFatalErr(err error) {
	if err != nil {
		fatalError.Do(func() {
			if uploadProgressBar != nil {
				uploadProgressBar.Clear()
			}
			Log.Info("Fatal Error detected")
			Log.Info("Cancelling upload")
			cancel()

			// finish the filePoster go routines to close the open files
			files.empty()
			files.close()
			filePosterWG.Wait()

			exit(1, fmt.Errorf("%v", err))
		})
		return
	}
}

func exit(exitCode int, exitErr error) {
	// clean up
	if tempPath != "" {
		if conf.DelTempFolder {
			Log.Info("Deleting temporary folder \"%v\"", tempPath)
			if err = os.RemoveAll(tempPath); err != nil {
				Log.Warn("Error while deleting temporary folder: %v", err)
			}
		}
	}
	if exitCode == 0 && conf.DelInputFolder {
		Log.Info("Deleting input folder folder \"%v\"", conf.Path)
		if err = os.RemoveAll(conf.Path); err != nil {
			Log.Warn("Error while deleting input folder: %v", err)
		}
	}
	if exitCode > 0 {
		Log.Error("%v", exitErr)
	} else {
		fmt.Println(nzb.Comment, conf.Password)
	}
	os.Exit(exitCode)
}
