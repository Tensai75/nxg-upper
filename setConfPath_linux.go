package main

import (
	"path/filepath"
)

var confFilePath string

func setConfPath() {
	var err error
	if confFilePath, err = filepath.Abs(filepath.Join(homePath, ".config")); err != nil {
		Log.Error("Unable to set path for configuration file: %v", err)
	}

}
