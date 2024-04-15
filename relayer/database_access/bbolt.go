package database_access

import (
	"fmt"
	"math/big"
	"path"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/relayer/core"
	"go.etcd.io/bbolt"
)

var (
	submittedBatchIdBucket = []byte("submittedBatchId")
)

type BBoltDatabase struct {
	db *bbolt.DB
}

var _ core.Database = (*BBoltDatabase)(nil)

func (bd *BBoltDatabase) Init(filePath string) error {
	if err := common.CreateDirectoryIfNotExists(path.Dir(filePath)); err != nil {
		return err
	}

	db, err := bbolt.Open(filePath, 0600, nil)
	if err != nil {
		return fmt.Errorf("could not open db: %v", err)
	}

	bd.db = db

	return db.Update(func(tx *bbolt.Tx) error {
		for _, bn := range [][]byte{submittedBatchIdBucket} {
			_, err := tx.CreateBucketIfNotExists(bn)
			if err != nil {
				return fmt.Errorf("could not bucket: %s, err: %v", string(bn), err)
			}
		}

		return nil
	})
}

func (bd *BBoltDatabase) Close() error {
	return bd.db.Close()
}

func (bd *BBoltDatabase) AddLastSubmittedBatchId(chainId string, batchId *big.Int) error {
	return bd.db.Update(func(tx *bbolt.Tx) error {
		bytes, err := batchId.MarshalText()
		if err != nil {
			return fmt.Errorf("could not marshal batch ID: %v", err)
		}

		if err := tx.Bucket(submittedBatchIdBucket).Put([]byte(chainId), bytes); err != nil {
			return fmt.Errorf("last submitted batch ID write error: %v", err)
		}

		return nil
	})
}

func (bd *BBoltDatabase) GetLastSubmittedBatchId(chainId string) (*big.Int, error) {
	var result *big.Int

	err := bd.db.View(func(tx *bbolt.Tx) error {
		bytes := tx.Bucket(submittedBatchIdBucket).Get([]byte(chainId))
		if bytes == nil {
			return nil
		}

		result = new(big.Int)
		if err := result.UnmarshalText(bytes); err != nil {
			return fmt.Errorf("could not unmarshal last submitted batch ID: %v", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}
