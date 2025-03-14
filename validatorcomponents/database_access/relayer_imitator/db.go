package databaseaccess

import (
	"fmt"
	"path/filepath"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/relayer/core"
)

func NewDatabase(pathToFile string) (core.Database, error) {
	if err := common.CreateDirectoryIfNotExists(filepath.Dir(pathToFile), 0770); err != nil {
		return nil, fmt.Errorf("failed to create directory for validator components database: %w", err)
	}

	db := &BBoltDatabase{}
	if err := db.Init(pathToFile); err != nil {
		return nil, err
	}

	return db, nil
}
