package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

var conf Args

func loadConfig() error {

	// Set the file name of the configurations file
	viper.SetConfigName(configFileName)

	// Set the path to look for the configurations file
	viper.AddConfigPath(confFilePath)

	// Set config type to yaml
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		if strings.Contains(err.Error(), "Not Found") {
			fmt.Println("Configuration file not found. Creating configuration file...")
			defaultConfig := []byte(defaultConfig())
			if err := os.WriteFile(filepath.Join(confFilePath, configFileName), defaultConfig, 0644); err != nil {
				checkForFatalErr(fmt.Errorf("error creating configuration file: %v", err))
			} else {
				Log.Info("Configuration file \"%v\" created.", filepath.Join(confFilePath, configFileName))
				fmt.Println("Please edit default values!")
				exit(0, nil)
			}
		} else {
			checkForFatalErr(fmt.Errorf("error reading configuration file: %v", err))
		}
	}

	if err := viper.Unmarshal(&conf); err != nil {
		checkForFatalErr(fmt.Errorf("unable to decode configuration file structure: %v", err))
	}

	return nil
}
