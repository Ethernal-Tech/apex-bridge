package databaseaccess

import (
	"fmt"
	"path/filepath"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_eth/core"
)

func NewDatabase(filePath string) (core.Database, error) {
	if err := common.CreateDirectoryIfNotExists(filepath.Dir(filePath), 0770); err != nil {
		return nil, fmt.Errorf("failed to create directory for oracle_eth database: %w", err)
	}

	db := &BBoltDatabase{}
	if err := db.Init(filePath); err != nil {
		return nil, err
	}

	return db, nil
}
