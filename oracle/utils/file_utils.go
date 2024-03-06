package utils

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

func LoadJson[TReturn any](path string) (*TReturn, error) {
	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open %v. error: %v\n", path, err)
		return nil, err
	}

	defer f.Close()

	var value TReturn
	decoder := json.NewDecoder(f)
	err = decoder.Decode(&value)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to decode %v. error: %v\n", path, err)
		return nil, err
	}

	return &value, nil
}
