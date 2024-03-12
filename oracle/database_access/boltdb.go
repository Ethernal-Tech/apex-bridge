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
	invalidTxsBucket     = []byte("InvalidTxs")
	expectedTxsBucket    = []byte("ExpectedTxs")
)

var _ core.Database = (*BoltDatabase)(nil)

func (bd *BoltDatabase) Init(filePath string) error {
	db, err := bolt.Open(filePath, 0600, nil)
	if err != nil {
		return fmt.Errorf("could not open db: %v", err)
	}

	bd.db = db

	return db.Update(func(tx *bolt.Tx) error {
		for _, bn := range [][]byte{unprocessedTxsBucket, invalidTxsBucket, expectedTxsBucket} {
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
			var cardanoTx *core.CardanoTx

			if err := json.Unmarshal(v, &cardanoTx); err != nil {
				return err
			}

			result = append(result, cardanoTx)
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

func (bd *BoltDatabase) MarkUnprocessedTxsAsProcessed(processedTxs []*core.CardanoTx) error {
	return bd.db.Update(func(tx *bolt.Tx) error {
		for _, processedTx := range processedTxs {
			if err := tx.Bucket(unprocessedTxsBucket).Delete(processedTx.Key()); err != nil {
				return fmt.Errorf("could not remove from unprocessed txs: %v", err)
			}
		}

		return nil
	})
}

func (bd *BoltDatabase) AddInvalidTxHashes(invalidTxHashes []string) error {
	return bd.db.Update(func(tx *bolt.Tx) error {
		for _, invalidTxHash := range invalidTxHashes {
			bytes, err := json.Marshal(invalidTxHash)
			if err != nil {
				return fmt.Errorf("could not marshal invalid tx: %v", err)
			}

			if err = tx.Bucket(invalidTxsBucket).Put([]byte(invalidTxHash), bytes); err != nil {
				return fmt.Errorf("invalid tx write error: %v", err)
			}
		}

		return nil
	})
}

func (bd *BoltDatabase) AddExpectedTxs(expectedTxs []*core.BridgeExpectedCardanoTx) error {
	return bd.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(expectedTxsBucket)
		for _, expectedTx := range expectedTxs {
			if data := bucket.Get(expectedTx.Key()); len(data) == 0 {
				expectedDbTx := &core.BridgeExpectedCardanoDbTx{
					BridgeExpectedCardanoTx: *expectedTx,
					IsProcessed:             false,
					IsInvalid:               false,
				}

				bytes, err := json.Marshal(expectedDbTx)
				if err != nil {
					return fmt.Errorf("could not marshal expected tx: %v", err)
				}

				if err = bucket.Put(expectedDbTx.Key(), bytes); err != nil {
					return fmt.Errorf("expected tx write error: %v", err)
				}
			}
		}

		return nil
	})
}

func (bd *BoltDatabase) GetExpectedTxs(threshold int) ([]*core.BridgeExpectedCardanoTx, error) {
	var result []*core.BridgeExpectedCardanoTx

	err := bd.db.View(func(tx *bolt.Tx) error {
		cursor := tx.Bucket(expectedTxsBucket).Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var expectedTx *core.BridgeExpectedCardanoDbTx

			if err := json.Unmarshal(v, &expectedTx); err != nil {
				return err
			}

			if !expectedTx.IsProcessed && !expectedTx.IsInvalid {
				result = append(result, &expectedTx.BridgeExpectedCardanoTx)
				if threshold > 0 && len(result) == threshold {
					break
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (bd *BoltDatabase) MarkExpectedTxsAsProcessed(expectedTxs []*core.BridgeExpectedCardanoTx) error {
	return bd.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(expectedTxsBucket)
		for _, expectedTx := range expectedTxs {
			if data := bucket.Get(expectedTx.Key()); len(data) > 0 {
				var dbExpectedTx *core.BridgeExpectedCardanoDbTx

				if err := json.Unmarshal(data, &dbExpectedTx); err != nil {
					return err
				}

				dbExpectedTx.IsProcessed = true

				bytes, err := json.Marshal(dbExpectedTx)
				if err != nil {
					return fmt.Errorf("could not marshal db expected tx: %v", err)
				}

				if err := bucket.Put(dbExpectedTx.Key(), bytes); err != nil {
					return fmt.Errorf("db expected tx write error: %v", err)
				}
			}
		}

		return nil
	})
}

func (bd *BoltDatabase) MarkExpectedTxsAsInvalid(expectedTxs []*core.BridgeExpectedCardanoTx) error {
	return bd.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(expectedTxsBucket)
		for _, expectedTx := range expectedTxs {
			if data := bucket.Get(expectedTx.Key()); len(data) > 0 {
				var dbExpectedTx *core.BridgeExpectedCardanoDbTx

				if err := json.Unmarshal(data, &dbExpectedTx); err != nil {
					return err
				}

				dbExpectedTx.IsInvalid = true

				bytes, err := json.Marshal(dbExpectedTx)
				if err != nil {
					return fmt.Errorf("could not marshal db expected tx: %v", err)
				}

				if err := bucket.Put(dbExpectedTx.Key(), bytes); err != nil {
					return fmt.Errorf("db expected tx write error: %v", err)
				}
			}
		}

		return nil
	})
}
