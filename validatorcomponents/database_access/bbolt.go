package database_access

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"go.etcd.io/bbolt"
)

type BBoltDatabase struct {
	db *bbolt.DB
}

var (
	bridgingRequestStatesBucket = []byte("BridgingRequestStates")
	submittedBatchIdBucket      = []byte("SubmittedBatchId")
)

var _ core.Database = (*BBoltDatabase)(nil)

func (bd *BBoltDatabase) Init(filePath string) error {
	db, err := bbolt.Open(filePath, 0660, nil)
	if err != nil {
		return fmt.Errorf("could not open db: %v", err)
	}

	bd.db = db

	return db.Update(func(tx *bbolt.Tx) error {
		for _, bn := range [][]byte{bridgingRequestStatesBucket, submittedBatchIdBucket} {
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

// AddBridgingRequestState implements core.Database.
func (bd *BBoltDatabase) AddBridgingRequestState(state *core.BridgingRequestState) error {
	return bd.db.Update(func(tx *bbolt.Tx) error {
		if len(tx.Bucket(bridgingRequestStatesBucket).Get(state.ToDbKey())) > 0 {
			return fmt.Errorf("trying to add a BridgingRequestState that already exists")
		}

		bytes, err := json.Marshal(state)
		if err != nil {
			return fmt.Errorf("could not marshal BridgingRequestState: %v", err)
		}

		if err = tx.Bucket(bridgingRequestStatesBucket).Put(state.ToDbKey(), bytes); err != nil {
			return fmt.Errorf("BridgingRequestState write error: %v", err)
		}

		return nil
	})
}

// UpdateBridgingRequestState implements core.Database.
func (bd *BBoltDatabase) UpdateBridgingRequestState(state *core.BridgingRequestState) error {
	return bd.db.Update(func(tx *bbolt.Tx) error {
		if len(tx.Bucket(bridgingRequestStatesBucket).Get(state.ToDbKey())) == 0 {
			return fmt.Errorf("trying to update a BridgingRequestState that does not exist")
		}

		bytes, err := json.Marshal(state)
		if err != nil {
			return fmt.Errorf("could not marshal BridgingRequestState: %v", err)
		}

		if err = tx.Bucket(bridgingRequestStatesBucket).Put(state.ToDbKey(), bytes); err != nil {
			return fmt.Errorf("BridgingRequestState write error: %v", err)
		}

		return nil
	})
}

// GetBridgingRequestState implements core.Database.
func (bd *BBoltDatabase) GetBridgingRequestState(sourceChainId string, sourceTxHash string) (result *core.BridgingRequestState, err error) {
	err = bd.db.View(func(tx *bbolt.Tx) error {
		if data := tx.Bucket(bridgingRequestStatesBucket).Get(core.ToBridgingRequestStateDbKey(sourceChainId, sourceTxHash)); len(data) > 0 {
			return json.Unmarshal(data, &result)
		}

		return nil
	})

	return result, err
}

// GetBridgingRequestStatesByBatchId implements core.Database.
func (bd *BBoltDatabase) GetBridgingRequestStatesByBatchId(destinationChainId string, batchId uint64) ([]*core.BridgingRequestState, error) {
	var result []*core.BridgingRequestState

	err := bd.db.View(func(tx *bbolt.Tx) error {
		cursor := tx.Bucket(bridgingRequestStatesBucket).Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var state *core.BridgingRequestState

			if err := json.Unmarshal(v, &state); err != nil {
				return err
			}

			if state.BatchId == batchId && state.DestinationChainId == destinationChainId {
				result = append(result, state)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetUserBridgingRequestStates implements core.Database.
func (bd *BBoltDatabase) GetUserBridgingRequestStates(sourceChainId string, userAddr string) ([]*core.BridgingRequestState, error) {
	var result []*core.BridgingRequestState

	err := bd.db.View(func(tx *bbolt.Tx) error {
		cursor := tx.Bucket(bridgingRequestStatesBucket).Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var state *core.BridgingRequestState

			if err := json.Unmarshal(v, &state); err != nil {
				return err
			}

			if state.SourceChainId == sourceChainId {

				for _, inputAddr := range state.InputAddrs {
					if userAddr == inputAddr {
						result = append(result, state)
						break
					}
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

// AddLastSubmittedBatchId implements core.Database.
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

// GetLastSubmittedBatchId implements core.Database.
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
