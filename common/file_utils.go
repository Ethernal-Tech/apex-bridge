package common

import (
	"encoding/json"
	"fmt"
	"os"
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
