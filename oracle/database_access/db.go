package database_access

import (
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
)

func NewDatabase(filePath string) (core.Database, error) {
	db := &BoltDatabase{}
	if err := db.Init(filePath); err != nil {
		return nil, err
	}

	return db, nil
}
