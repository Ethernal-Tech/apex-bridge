package database_access

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"go.etcd.io/bbolt"
)

type BBoltDatabase struct {
	db *bbolt.DB
}

var (
	unprocessedTxsBucket = []byte("UnprocessedTxs")
	processedTxsBucket   = []byte("ProcessedTxs")
	expectedTxsBucket    = []byte("ExpectedTxs")
)

var _ core.Database = (*BBoltDatabase)(nil)

func (bd *BBoltDatabase) Init(filePath string) error {
	db, err := bbolt.Open(filePath, 0660, nil)
	if err != nil {
		return fmt.Errorf("could not open db: %v", err)
	}

	bd.db = db

	return db.Update(func(tx *bbolt.Tx) error {
		for _, bn := range [][]byte{unprocessedTxsBucket, processedTxsBucket, expectedTxsBucket} {
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

func (bd *BBoltDatabase) AddUnprocessedTxs(unprocessedTxs []*core.CardanoTx) error {
	return bd.db.Update(func(tx *bbolt.Tx) error {
		for _, unprocessedTx := range unprocessedTxs {
			bytes, err := json.Marshal(unprocessedTx)
			if err != nil {
				return fmt.Errorf("could not marshal unprocessed tx: %v", err)
			}

			if err = tx.Bucket(unprocessedTxsBucket).Put([]byte(unprocessedTx.ToUnprocessedTxKey()), bytes); err != nil {
				return fmt.Errorf("unprocessed tx write error: %v", err)
			}
		}

		return nil
	})
}

func (bd *BBoltDatabase) GetUnprocessedTxs(chainId string, threshold int) ([]*core.CardanoTx, error) {
	var result []*core.CardanoTx

	err := bd.db.View(func(tx *bbolt.Tx) error {
		cursor := tx.Bucket(unprocessedTxsBucket).Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var cardanoTx *core.CardanoTx

			if err := json.Unmarshal(v, &cardanoTx); err != nil {
				return err
			}

			if cardanoTx.OriginChainId == chainId {
				result = append(result, cardanoTx)
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

func (bd *BBoltDatabase) ClearUnprocessedTxs(chainId string) error {
	return bd.db.Update(func(tx *bbolt.Tx) error {
		cursor := tx.Bucket(unprocessedTxsBucket).Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var unprocessedTx *core.CardanoTx

			if err := json.Unmarshal(v, &unprocessedTx); err != nil {
				return err
			}

			if strings.Compare(unprocessedTx.OriginChainId, chainId) == 0 {
				if err := cursor.Bucket().Delete([]byte(unprocessedTx.ToUnprocessedTxKey())); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (bd *BBoltDatabase) MarkUnprocessedTxsAsProcessed(processedTxs []*core.ProcessedCardanoTx) error {
	return bd.db.Update(func(tx *bbolt.Tx) error {
		for _, processedTx := range processedTxs {
			bytes, err := json.Marshal(processedTx)
			if err != nil {
				return fmt.Errorf("could not marshal processed tx: %v", err)
			}

			if err = tx.Bucket(processedTxsBucket).Put(processedTx.Key(), bytes); err != nil {
				return fmt.Errorf("processed tx write error: %v", err)
			}

			if err := tx.Bucket(unprocessedTxsBucket).Delete([]byte(processedTx.ToUnprocessedTxKey())); err != nil {
				return fmt.Errorf("could not remove from unprocessed txs: %v", err)
			}
		}

		return nil
	})
}

func (bd *BBoltDatabase) GetProcessedTx(chainId string, txHash string) (result *core.ProcessedCardanoTx, err error) {
	err = bd.db.View(func(tx *bbolt.Tx) error {
		if data := tx.Bucket(processedTxsBucket).Get([]byte(core.ToCardanoTxKey(chainId, txHash))); len(data) > 0 {
			return json.Unmarshal(data, &result)
		}

		return nil
	})

	return result, err
}

func (bd *BBoltDatabase) AddExpectedTxs(expectedTxs []*core.BridgeExpectedCardanoTx) error {
	return bd.db.Update(func(tx *bbolt.Tx) error {
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

func (bd *BBoltDatabase) GetExpectedTxs(chainId string, threshold int) ([]*core.BridgeExpectedCardanoTx, error) {
	var result []*core.BridgeExpectedCardanoTx

	err := bd.db.View(func(tx *bbolt.Tx) error {
		cursor := tx.Bucket(expectedTxsBucket).Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var expectedTx *core.BridgeExpectedCardanoDbTx

			if err := json.Unmarshal(v, &expectedTx); err != nil {
				return err
			}

			if expectedTx.ChainId == chainId && !expectedTx.IsProcessed && !expectedTx.IsInvalid {
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

func (bd *BBoltDatabase) ClearExpectedTxs(chainId string) error {
	return bd.db.Update(func(tx *bbolt.Tx) error {
		cursor := tx.Bucket(expectedTxsBucket).Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var expectedTx *core.BridgeExpectedCardanoDbTx

			if err := json.Unmarshal(v, &expectedTx); err != nil {
				return err
			}

			if strings.Compare(expectedTx.ChainId, chainId) == 0 {
				if err := cursor.Bucket().Delete(expectedTx.Key()); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (bd *BBoltDatabase) MarkExpectedTxsAsProcessed(expectedTxs []*core.BridgeExpectedCardanoTx) error {
	return bd.db.Update(func(tx *bbolt.Tx) error {
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

func (bd *BBoltDatabase) MarkExpectedTxsAsInvalid(expectedTxs []*core.BridgeExpectedCardanoTx) error {
	return bd.db.Update(func(tx *bbolt.Tx) error {
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
