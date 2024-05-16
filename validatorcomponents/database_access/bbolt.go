package databaseaccess

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
	submittedBatchIDBucket      = []byte("SubmittedBatchId")
)

var _ core.Database = (*BBoltDatabase)(nil)

func (bd *BBoltDatabase) Init(filePath string) error {
	db, err := bbolt.Open(filePath, 0660, nil)
	if err != nil {
		return fmt.Errorf("could not open db: %w", err)
	}

	bd.db = db

	return db.Update(func(tx *bbolt.Tx) error {
		for _, bn := range [][]byte{bridgingRequestStatesBucket, submittedBatchIDBucket} {
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

// AddBridgingRequestState implements core.Database.
func (bd *BBoltDatabase) AddBridgingRequestState(state *core.BridgingRequestState) error {
	return bd.db.Update(func(tx *bbolt.Tx) error {
		if len(tx.Bucket(bridgingRequestStatesBucket).Get(state.ToDBKey())) > 0 {
			return fmt.Errorf("trying to add a BridgingRequestState that already exists")
		}

		bytes, err := json.Marshal(state)
		if err != nil {
			return fmt.Errorf("could not marshal BridgingRequestState: %w", err)
		}

		if err = tx.Bucket(bridgingRequestStatesBucket).Put(state.ToDBKey(), bytes); err != nil {
			return fmt.Errorf("BridgingRequestState write error: %w", err)
		}

		return nil
	})
}

// UpdateBridgingRequestState implements core.Database.
func (bd *BBoltDatabase) UpdateBridgingRequestState(state *core.BridgingRequestState) error {
	return bd.db.Update(func(tx *bbolt.Tx) error {
		if len(tx.Bucket(bridgingRequestStatesBucket).Get(state.ToDBKey())) == 0 {
			return fmt.Errorf("trying to update a BridgingRequestState that does not exist")
		}

		bytes, err := json.Marshal(state)
		if err != nil {
			return fmt.Errorf("could not marshal BridgingRequestState: %w", err)
		}

		if err = tx.Bucket(bridgingRequestStatesBucket).Put(state.ToDBKey(), bytes); err != nil {
			return fmt.Errorf("BridgingRequestState write error: %w", err)
		}

		return nil
	})
}

// GetBridgingRequestState implements core.Database.
func (bd *BBoltDatabase) GetBridgingRequestState(
	sourceChainID string, sourceTxHash string,
) (
	result *core.BridgingRequestState, err error,
) {
	err = bd.db.View(func(tx *bbolt.Tx) error {
		data := tx.Bucket(bridgingRequestStatesBucket).Get(core.ToBridgingRequestStateDBKey(sourceChainID, sourceTxHash))
		if len(data) > 0 {
			return json.Unmarshal(data, &result)
		}

		return nil
	})

	return result, err
}

// GetBridgingRequestStatesByBatchID implements core.Database.
func (bd *BBoltDatabase) GetBridgingRequestStatesByBatchID(
	destinationChainID string, batchID uint64,
) (
	[]*core.BridgingRequestState, error,
) {
	var result []*core.BridgingRequestState

	err := bd.db.View(func(tx *bbolt.Tx) error {
		cursor := tx.Bucket(bridgingRequestStatesBucket).Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var state *core.BridgingRequestState

			if err := json.Unmarshal(v, &state); err != nil {
				return err
			}

			if state.BatchID == batchID && state.DestinationChainID == destinationChainID {
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

// AddLastSubmittedBatchID implements core.Database.
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

// GetLastSubmittedBatchID implements core.Database.
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
