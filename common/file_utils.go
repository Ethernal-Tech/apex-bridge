package common

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func CreateDirectoryIfNotExists(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		// If the directory doesn't exist, create it
		return os.MkdirAll(dirPath, os.ModePerm)
	}

	return nil
}

func RemoveDirOrFilePathIfExists(dirOrFilePath string) (err error) {
	if _, err = os.Stat(dirOrFilePath); err == nil {
		os.RemoveAll(dirOrFilePath)
	}

	return err
}

func LoadJson[TReturn any](path string) (*TReturn, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open %v. error: %v", path, err)
	}

	defer f.Close()

	var value TReturn
	decoder := json.NewDecoder(f)
	err = decoder.Decode(&value)
	if err != nil {
		return nil, fmt.Errorf("failed to decode %v. error: %v", path, err)
	}

	return &value, nil
}

// Loads config from defined path or from root
// Prefix defined as: (prefix)_config.json
func LoadConfig[TReturn any](configPath string, configPrefix string) (*TReturn, error) {
	var (
		config *TReturn
		err    error
	)

	if configPath == "" {
		ex, err := os.Executable()
		if err != nil {
			return nil, err
		}

		if strings.TrimSpace(configPrefix) != "" {
			configPath = path.Join(filepath.Dir(ex), strings.Join([]string{configPrefix, "config.json"}, "_"))
		} else {
			configPath = path.Join(filepath.Dir(ex), "config.json")
		}

	}

	config, err = LoadJson[TReturn](configPath)
	if err != nil {
		return nil, err
	}

	return config, nil
}
