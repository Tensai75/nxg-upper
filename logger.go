package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/acarl005/stripansi"
)

type Logger struct {
	Error func(string, ...interface{})
	Warn  func(string, ...interface{})
	Info  func(string, ...interface{})
	Succ  func(string, ...interface{})
	Debug func(string, ...interface{})
}

// global error logger variables
var (
	logFile *os.File
	logger  *log.Logger
	Log     = Logger{
		Error: logError,
		Warn:  logWarn,
		Info:  logInfo,
		Succ:  logSuccess,
		Debug: logDebug,
	}
)

// error logger function
func logError(logText string, vars ...interface{}) {
	logEntry("error", logText, vars...)
}

func logWarn(logText string, vars ...interface{}) {
	logEntry("warn", logText, vars...)
}

func logInfo(logText string, vars ...interface{}) {
	logEntry("info", logText, vars...)
}

func logSuccess(logText string, vars ...interface{}) {
	logEntry("success", logText, vars...)
}

func logDebug(logText string, vars ...interface{}) {
	logEntry("debug", logText, vars...)
}

func logEntry(logType string, logText string, vars ...interface{}) {

	// log error
	if logType == "error" {
		if logger != nil {
			logger.Printf("ERROR:   %s\n", stripansi.Strip(strings.Trim(fmt.Sprintf(logText, vars...), " \r\n")))
		}
		fmt.Fprintf(os.Stderr, "ERROR:   %s\n", strings.Trim(fmt.Sprintf(logText, vars...), " \r\n"))
	}

	// log warn
	if logType == "warn" {
		if logger != nil {
			logger.Printf("WARNING: %s\n", stripansi.Strip(strings.Trim(fmt.Sprintf(logText, vars...), " \r\n")))
		}
		if conf.Verbose > 1 {
			fmt.Printf("WARNING: %s\n", strings.Trim(fmt.Sprintf(logText, vars...), " \r\n"))
		}
	}

	// log info
	if logType == "info" {
		if logger != nil {
			logger.Printf("INFO:    %s\n", stripansi.Strip(strings.Trim(fmt.Sprintf(logText, vars...), " \r\n")))
		}
		if conf.Verbose > 0 {
			fmt.Printf("INFO:    %s\n", strings.Trim(fmt.Sprintf(logText, vars...), " \r\n"))
		}
	}

	// log success
	if logType == "success" {
		if logger != nil {
			logger.Printf("SUCCESS: %s\n", stripansi.Strip(strings.Trim(fmt.Sprintf(logText, vars...), " \r\n")))
		}
		if conf.Verbose > 0 {
			fmt.Printf("SUCCESS: %s\n", strings.Trim(fmt.Sprintf(logText, vars...), " \r\n"))
		}
	}

	// log debug
	if logType == "debug" && conf.Debug {
		if logger != nil {
			logger.Printf("DEBUG:   %s\n", stripansi.Strip(strings.Trim(fmt.Sprintf(logText, vars...), " \r\n")))
		}
		if conf.Verbose > 2 {
			fmt.Printf("DEBUG:   %s\n", strings.Trim(fmt.Sprintf(logText, vars...), " \r\n"))
		}
	}

}

func initLogger(path string) {
	var err error
	if err = os.MkdirAll(path, os.ModePerm); err != nil {
		checkForFatalErr(fmt.Errorf("Fatal error while opening log file '%s': %s\n", path, err))
	}
	if logFile, err = os.OpenFile(filepath.Join(path, logFileName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666); err == nil {
		logger = log.New(logFile, "", log.Ldate|log.Ltime)
		Log.Info("%s %s", appName, appVersion)
	} else {
		checkForFatalErr(fmt.Errorf("Fatal error while opening log file '%s': %s", path, err))
	}
}

func logClose() {
	// clean up
	if logFile != nil {
		logFile.Close()
	}
}
