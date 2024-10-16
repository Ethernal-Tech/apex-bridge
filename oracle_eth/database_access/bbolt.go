package databaseaccess

import (
	"encoding/json"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/oracle_eth/core"
	"github.com/Ethernal-Tech/ethgo"
	"go.etcd.io/bbolt"
)

type BBoltDatabase struct {
	db *bbolt.DB
}

var (
	unprocessedTxsBucket            = []byte("UnprocessedTxs")
	processedTxsBucket              = []byte("ProcessedTxs")
	processedTxsByInnerActionBucket = []byte("ProcessedTxsByInnerAction")
	expectedTxsBucket               = []byte("ExpectedTxs")
)

var _ core.Database = (*BBoltDatabase)(nil)

func (bd *BBoltDatabase) Init(filePath string) error {
	db, err := bbolt.Open(filePath, 0660, nil)
	if err != nil {
		return fmt.Errorf("could not open db: %w", err)
	}

	bd.db = db

	return db.Update(func(tx *bbolt.Tx) error {
		for _, bn := range [][]byte{
			unprocessedTxsBucket, processedTxsBucket,
			expectedTxsBucket, processedTxsByInnerActionBucket,
		} {
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

func (bd *BBoltDatabase) AddUnprocessedTxs(unprocessedTxs []*core.EthTx) error {
	return bd.db.Update(func(tx *bbolt.Tx) error {
		for _, unprocessedTx := range unprocessedTxs {
			bytes, err := json.Marshal(unprocessedTx)
			if err != nil {
				return fmt.Errorf("could not marshal unprocessed tx: %w", err)
			}

			if err = tx.Bucket(unprocessedTxsBucket).Put(unprocessedTx.ToUnprocessedTxKey(), bytes); err != nil {
				return fmt.Errorf("unprocessed tx write error: %w", err)
			}
		}

		return nil
	})
}

func (bd *BBoltDatabase) GetUnprocessedTxs(
	chainID string, priority uint8, threshold int,
) ([]*core.EthTx, error) {
	var result []*core.EthTx

	err := bd.db.View(func(tx *bbolt.Tx) error {
		cursor := tx.Bucket(unprocessedTxsBucket).Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var ethTx *core.EthTx

			if err := json.Unmarshal(v, &ethTx); err != nil {
				return err
			}

			if ethTx.OriginChainID == chainID && ethTx.Priority == priority {
				result = append(result, ethTx)
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

func (bd *BBoltDatabase) GetAllUnprocessedTxs(
	chainID string, threshold int,
) ([]*core.EthTx, error) {
	var result []*core.EthTx

	err := bd.db.View(func(tx *bbolt.Tx) error {
		cursor := tx.Bucket(unprocessedTxsBucket).Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var ethTx *core.EthTx

			if err := json.Unmarshal(v, &ethTx); err != nil {
				return err
			}

			if ethTx.OriginChainID == chainID {
				result = append(result, ethTx)
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

func (bd *BBoltDatabase) ClearUnprocessedTxs(chainID string) error {
	return bd.db.Update(func(tx *bbolt.Tx) error {
		cursor := tx.Bucket(unprocessedTxsBucket).Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var unprocessedTx *core.EthTx

			if err := json.Unmarshal(v, &unprocessedTx); err != nil {
				return err
			}

			if unprocessedTx.OriginChainID == chainID {
				if err := cursor.Bucket().Delete(unprocessedTx.ToUnprocessedTxKey()); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (bd *BBoltDatabase) MarkUnprocessedTxsAsProcessed(processedTxs []*core.ProcessedEthTx) error {
	return bd.db.Update(func(tx *bbolt.Tx) error {
		for _, processedTx := range processedTxs {
			bytes, err := json.Marshal(processedTx)
			if err != nil {
				return fmt.Errorf("could not marshal processed tx: %w", err)
			}

			if err = tx.Bucket(processedTxsBucket).Put(processedTx.Key(), bytes); err != nil {
				return fmt.Errorf("processed tx write error: %w", err)
			}

			innerActionTxBytes, err := json.Marshal(processedTx.ToProcessedTxByInnerAction())
			if err != nil {
				return fmt.Errorf("could not marshal processed tx by inner action: %w", err)
			}

			if err = tx.Bucket(processedTxsByInnerActionBucket).Put(
				processedTx.KeyByInnerAction(), innerActionTxBytes); err != nil {
				return fmt.Errorf("processed tx by inner action write error: %w", err)
			}

			if err := tx.Bucket(unprocessedTxsBucket).Delete(processedTx.ToUnprocessedTxKey()); err != nil {
				return fmt.Errorf("could not remove from unprocessed txs: %w", err)
			}
		}

		return nil
	})
}

func (bd *BBoltDatabase) AddProcessedTxs(processedTxs []*core.ProcessedEthTx) error {
	return bd.db.Update(func(tx *bbolt.Tx) error {
		for _, processedTx := range processedTxs {
			bytes, err := json.Marshal(processedTx)
			if err != nil {
				return fmt.Errorf("could not marshal processed tx: %w", err)
			}

			if err = tx.Bucket(processedTxsBucket).Put(processedTx.Key(), bytes); err != nil {
				return fmt.Errorf("processed tx write error: %w", err)
			}
		}

		return nil
	})
}

func (bd *BBoltDatabase) GetProcessedTx(
	chainID string, txHash ethgo.Hash,
) (result *core.ProcessedEthTx, err error) {
	err = bd.db.View(func(tx *bbolt.Tx) error {
		if data := tx.Bucket(processedTxsBucket).Get(core.ToEthTxKey(chainID, txHash)); len(data) > 0 {
			return json.Unmarshal(data, &result)
		}

		return nil
	})

	return result, err
}

func (bd *BBoltDatabase) GetProcessedTxByInnerActionTxHash(
	chainID string, innerActionTxHash ethgo.Hash,
) (result *core.ProcessedEthTx, err error) {
	err = bd.db.View(func(tx *bbolt.Tx) error {
		var processedTxByInnerAction *core.ProcessedEthTxByInnerAction
		if data := tx.Bucket(processedTxsByInnerActionBucket).Get(
			core.ToEthTxKey(chainID, innerActionTxHash)); len(data) > 0 {
			if err := json.Unmarshal(data, &processedTxByInnerAction); err != nil {
				return err
			}

			if data := tx.Bucket(processedTxsBucket).Get(
				core.ToEthTxKey(chainID, processedTxByInnerAction.Hash)); len(data) > 0 {
				return json.Unmarshal(data, &result)
			}
		}

		return nil
	})

	return result, err
}

func (bd *BBoltDatabase) AddExpectedTxs(expectedTxs []*core.BridgeExpectedEthTx) error {
	return bd.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(expectedTxsBucket)

		for _, expectedTx := range expectedTxs {
			key := expectedTx.Key()

			if data := bucket.Get(key); len(data) == 0 {
				expectedDBTx := &core.BridgeExpectedEthDBTx{
					BridgeExpectedEthTx: *expectedTx,
					IsProcessed:         false,
					IsInvalid:           false,
				}

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

func (bd *BBoltDatabase) GetExpectedTxs(
	chainID string, priority uint8, threshold int,
) ([]*core.BridgeExpectedEthTx, error) {
	var result []*core.BridgeExpectedEthTx

	err := bd.db.View(func(tx *bbolt.Tx) error {
		cursor := tx.Bucket(expectedTxsBucket).Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var expectedTx *core.BridgeExpectedEthDBTx

			if err := json.Unmarshal(v, &expectedTx); err != nil {
				return err
			}

			if expectedTx.ChainID == chainID && expectedTx.Priority == priority &&
				!expectedTx.IsProcessed && !expectedTx.IsInvalid {
				result = append(result, &expectedTx.BridgeExpectedEthTx)
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

func (bd *BBoltDatabase) GetAllExpectedTxs(chainID string, threshold int) ([]*core.BridgeExpectedEthTx, error) {
	var result []*core.BridgeExpectedEthTx

	err := bd.db.View(func(tx *bbolt.Tx) error {
		cursor := tx.Bucket(expectedTxsBucket).Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var expectedTx *core.BridgeExpectedEthDBTx

			if err := json.Unmarshal(v, &expectedTx); err != nil {
				return err
			}

			if expectedTx.ChainID == chainID && !expectedTx.IsProcessed && !expectedTx.IsInvalid {
				result = append(result, &expectedTx.BridgeExpectedEthTx)
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

func (bd *BBoltDatabase) ClearExpectedTxs(chainID string) error {
	return bd.db.Update(func(tx *bbolt.Tx) error {
		cursor := tx.Bucket(expectedTxsBucket).Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var expectedTx *core.BridgeExpectedEthDBTx

			if err := json.Unmarshal(v, &expectedTx); err != nil {
				return err
			}

			if expectedTx.ChainID == chainID && !expectedTx.IsInvalid && !expectedTx.IsProcessed {
				if err := cursor.Bucket().Delete(expectedTx.Key()); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (bd *BBoltDatabase) MarkExpectedTxsAsProcessed(expectedTxs []*core.BridgeExpectedEthTx) error {
	return bd.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(expectedTxsBucket)

		for _, expectedTx := range expectedTxs {
			key := expectedTx.Key()

			if data := bucket.Get(key); len(data) > 0 {
				var dbExpectedTx *core.BridgeExpectedEthDBTx

				if err := json.Unmarshal(data, &dbExpectedTx); err != nil {
					return err
				}

				dbExpectedTx.IsProcessed = true

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
	})
}

func (bd *BBoltDatabase) MarkExpectedTxsAsInvalid(expectedTxs []*core.BridgeExpectedEthTx) error {
	return bd.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(expectedTxsBucket)

		for _, expectedTx := range expectedTxs {
			key := expectedTx.Key()

			if data := bucket.Get(key); len(data) > 0 {
				var dbExpectedTx *core.BridgeExpectedEthDBTx

				if err := json.Unmarshal(data, &dbExpectedTx); err != nil {
					return err
				}

				dbExpectedTx.IsInvalid = true

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
	})
}
