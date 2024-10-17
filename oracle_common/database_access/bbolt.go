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
	TExpectedDbTx core.BaseExpectedDBTx,
] struct {
	DB *bbolt.DB
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx, TExpectedDbTx]) Init(
	filePath string, additionalBuckets [][]byte) error {
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

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx, TExpectedDbTx]) Close() error {
	return bd.DB.Close()
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx, TExpectedDbTx]) GetUnprocessedTxs(
	chainID string, priority uint8, threshold int,
) ([]*TTx, error) {
	var result []*TTx

	err := bd.DB.View(func(tx *bbolt.Tx) error {
		cursor := tx.Bucket(UnprocessedTxsBucket).Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var chainTx *TTx

			if err := json.Unmarshal(v, &chainTx); err != nil {
				return err
			}

			cChainTx, ok := any(*chainTx).(core.BaseTx)
			if !ok {
				return fmt.Errorf("could not convert tx")
			}

			if cChainTx.GetOriginChainID() == chainID && cChainTx.GetPriority() == priority {
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

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx, TExpectedDbTx]) GetAllUnprocessedTxs(
	chainID string, threshold int,
) ([]*TTx, error) {
	var result []*TTx

	err := bd.DB.View(func(tx *bbolt.Tx) error {
		cursor := tx.Bucket(UnprocessedTxsBucket).Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var chainTx *TTx

			if err := json.Unmarshal(v, &chainTx); err != nil {
				return err
			}

			cChainTx, ok := any(*chainTx).(core.BaseTx)
			if !ok {
				return fmt.Errorf("could not convert tx")
			}

			if cChainTx.GetOriginChainID() == chainID {
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

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx, TExpectedDbTx]) ClearUnprocessedTxs(chainID string) error {
	return bd.DB.Update(func(tx *bbolt.Tx) error {
		cursor := tx.Bucket(UnprocessedTxsBucket).Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var unprocessedTx *TTx

			if err := json.Unmarshal(v, &unprocessedTx); err != nil {
				return err
			}

			cUnprocessedTx, ok := any(*unprocessedTx).(core.BaseTx)
			if !ok {
				return fmt.Errorf("could not convert tx")
			}

			if cUnprocessedTx.GetOriginChainID() == chainID {
				if err := cursor.Bucket().Delete(
					cUnprocessedTx.ToUnprocessedTxKey()); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx, TExpectedDbTx]) AddTxs(
	processedTxs []*TProcessedTx, unprocessedTxs []*TTx,
) error {
	return bd.DB.Update(func(tx *bbolt.Tx) error {
		processedBucket, unprocessedBucket := tx.Bucket(ProcessedTxsBucket), tx.Bucket(UnprocessedTxsBucket)

		for _, processedTx := range processedTxs {
			bytes, err := json.Marshal(processedTx)
			if err != nil {
				return fmt.Errorf("could not marshal processed tx: %w", err)
			}

			cProcessedTx, ok := any(*processedTx).(core.BaseProcessedTx)
			if !ok {
				return fmt.Errorf("could not convert tx")
			}

			if err = processedBucket.Put(cProcessedTx.Key(), bytes); err != nil {
				return fmt.Errorf("processed tx write error: %w", err)
			}
		}

		for _, unprocessedTx := range unprocessedTxs {
			bytes, err := json.Marshal(unprocessedTx)
			if err != nil {
				return fmt.Errorf("could not marshal unprocessed tx: %w", err)
			}

			cUnprocessedTx, ok := any(*unprocessedTx).(core.BaseTx)
			if !ok {
				return fmt.Errorf("could not convert tx")
			}

			if err = unprocessedBucket.Put(cUnprocessedTx.ToUnprocessedTxKey(), bytes); err != nil {
				return fmt.Errorf("unprocessed tx write error: %w", err)
			}
		}

		return nil
	})
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx, TExpectedDbTx]) GetProcessedTx(
	chainID string, txKey []byte,
) (result *TProcessedTx, err error) {
	err = bd.DB.View(func(tx *bbolt.Tx) error {
		if data := tx.Bucket(ProcessedTxsBucket).Get(txKey); len(data) > 0 {
			return json.Unmarshal(data, &result)
		}

		return nil
	})

	return result, err
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx, TExpectedDbTx]) MarkTxs(
	expectedInvalid, expectedProcessed []*TExpectedTx, allProcessed []*TProcessedTx,
	additionalCallback func(tx *bbolt.Tx) error,
) error {
	return bd.DB.Update(func(tx *bbolt.Tx) error {
		if err := bd.markExpectedTxsAsInvalid(tx, expectedInvalid); err != nil {
			return err
		}

		if err := bd.markExpectedTxsAsProcessed(tx, expectedProcessed); err != nil {
			return err
		}

		if err := bd.markUnprocessedTxsAsProcessed(tx, allProcessed); err != nil {
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

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx, TExpectedDbTx]) AddExpectedTxs(
	expectedTxs []*TExpectedTx) error {
	return bd.DB.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(ExpectedTxsBucket)

		for _, expectedTx := range expectedTxs {
			cExpectedTx, ok := any(*expectedTx).(core.BaseExpectedTx)
			if !ok {
				return fmt.Errorf("could not convert tx")
			}

			key := cExpectedTx.Key()

			if data := bucket.Get(key); len(data) == 0 {
				expectedDBTx := cExpectedTx.NewExpectedDBTx()

				bytes, err := json.Marshal(expectedDBTx)
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

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx, TExpectedDbTx]) GetExpectedTxs(
	chainID string, priority uint8, threshold int,
) ([]*TExpectedTx, error) {
	var result []*TExpectedTx

	err := bd.DB.View(func(tx *bbolt.Tx) error {
		cursor := tx.Bucket(ExpectedTxsBucket).Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var expectedTx *TExpectedDbTx

			if err := json.Unmarshal(v, &expectedTx); err != nil {
				return err
			}

			cExpectedTx, ok := any(*expectedTx).(core.BaseExpectedDBTx)
			if !ok {
				return fmt.Errorf("could not convert tx")
			}

			if cExpectedTx.GetChainID() == chainID && cExpectedTx.GetPriority() == priority &&
				!cExpectedTx.GetIsProcessed() && !cExpectedTx.GetIsInvalid() {
				innerExpectedTx, ok := cExpectedTx.GetInnerTx().(TExpectedTx)
				if !ok {
					return fmt.Errorf("could not get inner expected tx from db expected tx")
				}

				result = append(result, &innerExpectedTx)
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

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx, TExpectedDbTx]) GetAllExpectedTxs(
	chainID string, threshold int) ([]*TExpectedTx, error) {
	var result []*TExpectedTx

	err := bd.DB.View(func(tx *bbolt.Tx) error {
		cursor := tx.Bucket(ExpectedTxsBucket).Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var expectedTx *TExpectedDbTx

			if err := json.Unmarshal(v, &expectedTx); err != nil {
				return err
			}

			cExpectedTx, ok := any(*expectedTx).(core.BaseExpectedDBTx)
			if !ok {
				return fmt.Errorf("could not convert tx")
			}

			if cExpectedTx.GetChainID() == chainID &&
				!cExpectedTx.GetIsProcessed() && !cExpectedTx.GetIsInvalid() {
				innerExpectedTx, ok := cExpectedTx.GetInnerTx().(TExpectedTx)
				if !ok {
					return fmt.Errorf("could not get inner expected tx from db expected tx")
				}

				result = append(result, &innerExpectedTx)
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

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx, TExpectedDbTx]) ClearExpectedTxs(chainID string) error {
	return bd.DB.Update(func(tx *bbolt.Tx) error {
		cursor := tx.Bucket(ExpectedTxsBucket).Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var expectedTx *TExpectedDbTx

			if err := json.Unmarshal(v, &expectedTx); err != nil {
				return err
			}

			cExpectedTx, ok := any(*expectedTx).(core.BaseExpectedDBTx)
			if !ok {
				return fmt.Errorf("could not convert tx")
			}

			if cExpectedTx.GetChainID() == chainID && !cExpectedTx.GetIsInvalid() && !cExpectedTx.GetIsProcessed() {
				if err := cursor.Bucket().Delete(cExpectedTx.Key()); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx, TExpectedDbTx]) markUnprocessedTxsAsProcessed(
	tx *bbolt.Tx, processedTxs []*TProcessedTx) error {
	processedBucket, unprocessedBucket := tx.Bucket(ProcessedTxsBucket), tx.Bucket(UnprocessedTxsBucket)

	for _, processedTx := range processedTxs {
		bytes, err := json.Marshal(processedTx)
		if err != nil {
			return fmt.Errorf("could not marshal processed tx: %w", err)
		}

		cProcessedTx, ok := any(*processedTx).(core.BaseProcessedTx)
		if !ok {
			return fmt.Errorf("could not convert tx")
		}

		if err = processedBucket.Put(cProcessedTx.Key(), bytes); err != nil {
			return fmt.Errorf("processed tx write error: %w", err)
		}

		if err := unprocessedBucket.Delete(cProcessedTx.ToUnprocessedTxKey()); err != nil {
			return fmt.Errorf("could not remove from unprocessed txs: %w", err)
		}
	}

	return nil
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx, TExpectedDbTx]) markExpectedTxsAsProcessed(
	tx *bbolt.Tx, expectedTxs []*TExpectedTx) error {
	return bd.markExpectedTxs(tx, expectedTxs, func(dbExpectedTx core.BaseExpectedDBTx) core.BaseExpectedDBTx {
		return dbExpectedTx.SetProcessed()
	})
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx, TExpectedDbTx]) markExpectedTxsAsInvalid(
	tx *bbolt.Tx, expectedTxs []*TExpectedTx) error {
	return bd.markExpectedTxs(tx, expectedTxs, func(dbExpectedTx core.BaseExpectedDBTx) core.BaseExpectedDBTx {
		return dbExpectedTx.SetInvalid()
	})
}

func (bd *BBoltDBBase[TTx, TProcessedTx, TExpectedTx, TExpectedDbTx]) markExpectedTxs(
	tx *bbolt.Tx, expectedTxs []*TExpectedTx, markFunc func(dbExpectedTx core.BaseExpectedDBTx) core.BaseExpectedDBTx,
) error {
	bucket := tx.Bucket(ExpectedTxsBucket)

	for _, expectedTx := range expectedTxs {
		cExpectedTx, ok := any(*expectedTx).(core.BaseExpectedTx)
		if !ok {
			return fmt.Errorf("could not convert tx")
		}

		key := cExpectedTx.Key()

		if data := bucket.Get(key); len(data) > 0 {
			var (
				dbExpectedTx *TExpectedDbTx
				ok           bool
			)

			if err := json.Unmarshal(data, &dbExpectedTx); err != nil {
				return err
			}

			cDBExpectedTx, ok := any(*dbExpectedTx).(core.BaseExpectedDBTx)
			if !ok {
				return fmt.Errorf("could not set db expected tx to processed")
			}

			*dbExpectedTx, ok = markFunc(cDBExpectedTx).(TExpectedDbTx)
			if !ok {
				return fmt.Errorf("could not set db expected tx to processed")
			}

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
