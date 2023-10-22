package main

import (
	"path/filepath"
)

var confFilePath string

func setConfPath() {

	if confFilePath, err = filepath.Abs(appPath); err != nil {
		Log.Error("Unable to set path for configuration file: %v", err)
	}

}
