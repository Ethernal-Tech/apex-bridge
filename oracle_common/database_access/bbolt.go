package databaseaccess

import (
	"encoding/json"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"go.etcd.io/bbolt"
)

var (
	UnprocessedTxsBucket = []byte("UnprocessedTxs")
	ProcessedTxsBucket   = []byte("ProcessedTxs")
	ExpectedTxsBucket    = []byte("ExpectedTxs")
)

type BBoltDBBase[
	TTx core.BaseTx,
	TProcessedTx core.BaseProcessedTx,
	TExpectedTx core.BaseExpectedTx,
] struct {
	DB *bbolt.DB
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx]) Init(
	filePath string, additionalBuckets [][]byte,
) error {
	db, err := bbolt.Open(filePath, 0660, nil)
	if err != nil {
		return fmt.Errorf("could not open db: %w", err)
	}

	bd.DB = db

	allBuckets := append([][]byte{
		UnprocessedTxsBucket, ProcessedTxsBucket, ExpectedTxsBucket,
	}, additionalBuckets...)

	return db.Update(func(tx *bbolt.Tx) error {
		for _, bn := range allBuckets {
			_, err := tx.CreateBucketIfNotExists(bn)
			if err != nil {
				return fmt.Errorf("could not bucket: %s, err: %w", string(bn), err)
			}
		}

		return nil
	})
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx]) Close() error {
	return bd.DB.Close()
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx]) GetUnprocessedTxs(
	chainID string, priority uint8, threshold int,
) ([]TTx, error) {
	var result []TTx

	err := bd.DB.View(func(tx *bbolt.Tx) error {
		cursor := tx.Bucket(UnprocessedTxsBucket).Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var chainTx TTx

			if err := json.Unmarshal(v, &chainTx); err != nil {
				return err
			}

			if chainTx.GetOriginChainID() == chainID && chainTx.GetPriority() == priority {
				result = append(result, chainTx)
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

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx]) GetAllUnprocessedTxs(
	chainID string, threshold int,
) ([]TTx, error) {
	var result []TTx

	err := bd.DB.View(func(tx *bbolt.Tx) error {
		cursor := tx.Bucket(UnprocessedTxsBucket).Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var chainTx TTx

			if err := json.Unmarshal(v, &chainTx); err != nil {
				return err
			}

			if chainTx.GetOriginChainID() == chainID {
				result = append(result, chainTx)
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

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx]) ClearUnprocessedTxs(chainID string) error {
	return bd.DB.Update(func(tx *bbolt.Tx) error {
		cursor := tx.Bucket(UnprocessedTxsBucket).Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var unprocessedTx TTx

			if err := json.Unmarshal(v, &unprocessedTx); err != nil {
				return err
			}

			if unprocessedTx.GetOriginChainID() == chainID {
				err := cursor.Bucket().Delete(unprocessedTx.ToUnprocessedTxKey())
				if err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx]) AddTxs(
	processedTxs []TProcessedTx, unprocessedTxs []TTx,
) error {
	return bd.DB.Update(func(tx *bbolt.Tx) error {
		processedBucket, unprocessedBucket := tx.Bucket(ProcessedTxsBucket), tx.Bucket(UnprocessedTxsBucket)

		for _, processedTx := range processedTxs {
			bytes, err := json.Marshal(processedTx)
			if err != nil {
				return fmt.Errorf("could not marshal processed tx: %w", err)
			}

			if err = processedBucket.Put(processedTx.Key(), bytes); err != nil {
				return fmt.Errorf("processed tx write error: %w", err)
			}
		}

		for _, unprocessedTx := range unprocessedTxs {
			bytes, err := json.Marshal(unprocessedTx)
			if err != nil {
				return fmt.Errorf("could not marshal unprocessed tx: %w", err)
			}

			if err = unprocessedBucket.Put(unprocessedTx.ToUnprocessedTxKey(), bytes); err != nil {
				return fmt.Errorf("unprocessed tx write error: %w", err)
			}
		}

		return nil
	})
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx]) GetProcessedTx(
	chainID string, txKey []byte,
) (result TProcessedTx, err error) {
	err = bd.DB.View(func(tx *bbolt.Tx) error {
		if data := tx.Bucket(ProcessedTxsBucket).Get(txKey); len(data) > 0 {
			return json.Unmarshal(data, &result)
		}

		return nil
	})

	return result, err
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx]) MarkTxs(
	expectedInvalid, expectedProcessed []TExpectedTx, allProcessed []TProcessedTx,
	additionalCallback func(tx *bbolt.Tx) error,
) error {
	return bd.DB.Update(func(tx *bbolt.Tx) error {
		err := bd.markExpectedTxs(tx, expectedInvalid, func(expectedTx TExpectedTx) {
			expectedTx.SetInvalid()
		})
		if err != nil {
			return err
		}

		err = bd.markExpectedTxs(tx, expectedProcessed, func(expectedTx TExpectedTx) {
			expectedTx.SetProcessed()
		})
		if err != nil {
			return err
		}

		if err := bd.markAsProcessed(tx, allProcessed); err != nil {
			return err
		}

		if additionalCallback != nil {
			if err := additionalCallback(tx); err != nil {
				return err
			}
		}

		return nil
	})
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx]) AddExpectedTxs(expectedTxs []TExpectedTx) error {
	return bd.DB.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(ExpectedTxsBucket)

		for _, expectedTx := range expectedTxs {
			key := expectedTx.Key()

			if data := bucket.Get(key); len(data) == 0 {
				bytes, err := json.Marshal(expectedTx)
				if err != nil {
					return fmt.Errorf("could not marshal expected tx: %w", err)
				}

				if err = bucket.Put(key, bytes); err != nil {
					return fmt.Errorf("expected tx write error: %w", err)
				}
			}
		}

		return nil
	})
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx]) GetExpectedTxs(
	chainID string, priority uint8, threshold int,
) ([]TExpectedTx, error) {
	var result []TExpectedTx

	err := bd.DB.View(func(tx *bbolt.Tx) error {
		cursor := tx.Bucket(ExpectedTxsBucket).Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var expectedTx TExpectedTx

			if err := json.Unmarshal(v, &expectedTx); err != nil {
				return err
			}

			if expectedTx.GetChainID() == chainID && expectedTx.GetPriority() == priority &&
				!expectedTx.GetIsProcessed() && !expectedTx.GetIsInvalid() {
				result = append(result, expectedTx)
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

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx]) GetAllExpectedTxs(
	chainID string, threshold int,
) ([]TExpectedTx, error) {
	var result []TExpectedTx

	err := bd.DB.View(func(tx *bbolt.Tx) error {
		cursor := tx.Bucket(ExpectedTxsBucket).Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var expectedTx TExpectedTx

			if err := json.Unmarshal(v, &expectedTx); err != nil {
				return err
			}

			if expectedTx.GetChainID() == chainID && !expectedTx.GetIsProcessed() && !expectedTx.GetIsInvalid() {
				result = append(result, expectedTx)
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

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx]) ClearExpectedTxs(chainID string) error {
	return bd.DB.Update(func(tx *bbolt.Tx) error {
		cursor := tx.Bucket(ExpectedTxsBucket).Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var expectedTx TExpectedTx

			if err := json.Unmarshal(v, &expectedTx); err != nil {
				return err
			}

			if expectedTx.GetChainID() == chainID && !expectedTx.GetIsInvalid() && !expectedTx.GetIsProcessed() {
				if err := cursor.Bucket().Delete(expectedTx.Key()); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx]) markAsProcessed(
	tx *bbolt.Tx, processedTxs []TProcessedTx,
) error {
	processedBucket, unprocessedBucket := tx.Bucket(ProcessedTxsBucket), tx.Bucket(UnprocessedTxsBucket)

	for _, processedTx := range processedTxs {
		bytes, err := json.Marshal(processedTx)
		if err != nil {
			return fmt.Errorf("could not marshal processed tx: %w", err)
		}

		if err = processedBucket.Put(processedTx.Key(), bytes); err != nil {
			return fmt.Errorf("processed tx write error: %w", err)
		}

		if err := unprocessedBucket.Delete(processedTx.ToUnprocessedTxKey()); err != nil {
			return fmt.Errorf("could not remove from unprocessed txs: %w", err)
		}
	}

	return nil
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx]) markExpectedTxs(
	tx *bbolt.Tx, expectedTxs []TExpectedTx, markFunc func(expectedTx TExpectedTx),
) error {
	bucket := tx.Bucket(ExpectedTxsBucket)

	for _, expectedTx := range expectedTxs {
		key := expectedTx.Key()

		if data := bucket.Get(key); len(data) > 0 {
			var dbExpectedTx TExpectedTx

			if err := json.Unmarshal(data, &dbExpectedTx); err != nil {
				return err
			}

			markFunc(dbExpectedTx)

			bytes, err := json.Marshal(dbExpectedTx)
			if err != nil {
				return fmt.Errorf("could not marshal db expected tx: %w", err)
			}

			if err := bucket.Put(key, bytes); err != nil {
				return fmt.Errorf("db expected tx write error: %w", err)
			}
		}
	}

	return nil
}
