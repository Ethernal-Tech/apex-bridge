package utils

import (
	"os"
)

func CreateDirectoryIfNotExists(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		// If the directory doesn't exist, create it
		return os.MkdirAll(dirPath, os.ModePerm)
	}

	return nil
}
