package common

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func CreateDirectoryIfNotExists(dirPath string, perm fs.FileMode) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		// If the directory doesn't exist, create it
		return os.MkdirAll(dirPath, perm)
	}

	return nil
}

func RemoveDirOrFilePathIfExists(dirOrFilePath string) (err error) {
	if _, err = os.Stat(dirOrFilePath); err == nil {
		os.RemoveAll(dirOrFilePath)
	}

	return err
}

func LoadJSON[TReturn any](path string) (*TReturn, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open %v. error: %w", path, err)
	}

	defer f.Close()

	var value TReturn

	if err = json.NewDecoder(f).Decode(&value); err != nil {
		return nil, fmt.Errorf("failed to decode %v. error: %w", path, err)
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

		if prfx := strings.TrimSpace(configPrefix); prfx != "" {
			configPath = path.Join(filepath.Dir(ex), strings.Join([]string{prfx, "config.json"}, "_"))
		} else {
			configPath = path.Join(filepath.Dir(ex), "config.json")
		}
	}

	config, err = LoadJSON[TReturn](configPath)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func SaveJSON[TObj any](path string, obj TObj, pretty bool) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create %v. error: %w", path, err)
	}

	defer f.Close()

	encoder := json.NewEncoder(f)
	if pretty {
		encoder.SetIndent("", "    ")
	}

	if err = encoder.Encode(&obj); err != nil {
		return fmt.Errorf("failed to encode %v. error: %w", path, err)
	}

	return nil
}
