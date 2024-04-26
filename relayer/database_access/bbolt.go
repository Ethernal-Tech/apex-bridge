package databaseaccess

import (
	"fmt"
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/relayer/core"
	"go.etcd.io/bbolt"
)

var (
	submittedBatchIDBucket = []byte("submittedBatchId")
)

type BBoltDatabase struct {
	db *bbolt.DB
}

var _ core.Database = (*BBoltDatabase)(nil)

func (bd *BBoltDatabase) Init(filePath string) error {
	db, err := bbolt.Open(filePath, 0660, nil)
	if err != nil {
		return fmt.Errorf("could not open db: %w", err)
	}

	bd.db = db

	return db.Update(func(tx *bbolt.Tx) error {
		for _, bn := range [][]byte{submittedBatchIDBucket} {
			_, err := tx.CreateBucketIfNotExists(bn)
			if err != nil {
				return fmt.Errorf("could not bucket: %s, err: %w", string(bn), err)
			}
		}

		return nil
	})
}

func (bd *BBoltDatabase) Close() error {
	return bd.db.Close()
}

func (bd *BBoltDatabase) AddLastSubmittedBatchID(chainID string, batchID *big.Int) error {
	return bd.db.Update(func(tx *bbolt.Tx) error {
		bytes, err := batchID.MarshalText()
		if err != nil {
			return fmt.Errorf("could not marshal batch ID: %w", err)
		}

		if err := tx.Bucket(submittedBatchIDBucket).Put([]byte(chainID), bytes); err != nil {
			return fmt.Errorf("last submitted batch ID write error: %w", err)
		}

		return nil
	})
}

func (bd *BBoltDatabase) GetLastSubmittedBatchID(chainID string) (*big.Int, error) {
	var result *big.Int

	err := bd.db.View(func(tx *bbolt.Tx) error {
		bytes := tx.Bucket(submittedBatchIDBucket).Get([]byte(chainID))
		if bytes == nil {
			return nil
		}

		result = new(big.Int)
		if err := result.UnmarshalText(bytes); err != nil {
			return fmt.Errorf("could not unmarshal last submitted batch ID: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}
