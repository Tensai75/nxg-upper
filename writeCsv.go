package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func writeCsv() {
	if conf.CsvPath == "" {
		return
	}

	var (
		folder  = filepath.Dir(conf.CsvPath)
		csvFile *os.File
		d       = conf.CsvDelimiter
	)

	if err = os.MkdirAll(folder, os.ModePerm); err != nil {
		Log.Warn("Cannot create the CSV file: %v", err)
	}
	if csvFile, err = os.OpenFile(filepath.Join(conf.CsvPath), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666); err != nil {
		Log.Warn("Cannot open the CSV file: %v", err)
	}
	defer csvFile.Close()
	if _, err = csvFile.WriteString(fmt.Sprintf("%s%s%s%s%s%s%s%s%d\n", title, d, nzb.Comment, d, conf.Password, d, conf.Groups, d, time.Now().Unix())); err != nil {
		Log.Warn("Cannot write to the CSV file: %v", err)
	}
}
