package database_access

import (
	"encoding/json"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/boltdb/bolt"
)

type BoltDatabase struct {
	db *bolt.DB
}

var (
	unprocessedTxsBucket = []byte("UnprocessedTxs")
)

var _ core.Database = (*BoltDatabase)(nil)

func (bd *BoltDatabase) Init(filePath string) error {
	db, err := bolt.Open(filePath, 0600, nil)
	if err != nil {
		return fmt.Errorf("could not open db: %v", err)
	}

	bd.db = db

	return db.Update(func(tx *bolt.Tx) error {
		for _, bn := range [][]byte{unprocessedTxsBucket} {
			_, err := tx.CreateBucketIfNotExists(bn)
			if err != nil {
				return fmt.Errorf("could not bucket: %s, err: %v", string(bn), err)
			}
		}

		return nil
	})
}

func (bd *BoltDatabase) Close() error {
	return bd.db.Close()
}

func (bd *BoltDatabase) AddUnprocessedTxs(unprocessedTxs []*core.CardanoTx) error {
	return bd.db.Update(func(tx *bolt.Tx) error {
		for _, unprocessedTx := range unprocessedTxs {
			bytes, err := json.Marshal(unprocessedTx)
			if err != nil {
				return fmt.Errorf("could not marshal unprocessed tx: %v", err)
			}

			if err = tx.Bucket(unprocessedTxsBucket).Put(unprocessedTx.Key(), bytes); err != nil {
				return fmt.Errorf("unprocessed tx write error: %v", err)
			}
		}

		return nil
	})
}

func (bd *BoltDatabase) GetUnprocessedTxs(threshold int) ([]*core.CardanoTx, error) {
	var result []*core.CardanoTx

	err := bd.db.View(func(tx *bolt.Tx) error {
		cursor := tx.Bucket(unprocessedTxsBucket).Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var block *core.CardanoTx

			if err := json.Unmarshal(v, &block); err != nil {
				return err
			}

			result = append(result, block)
			if threshold > 0 && len(result) == threshold {
				break
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (bd *BoltDatabase) MarkTxsAsProcessed(processedTxs []*core.CardanoTx) error {
	return bd.db.Update(func(tx *bolt.Tx) error {
		for _, processedTx := range processedTxs {
			if err := tx.Bucket(unprocessedTxsBucket).Delete(processedTx.Key()); err != nil {
				return fmt.Errorf("could not remove from unprocessed txs: %v", err)
			}
		}

		return nil
	})
}
