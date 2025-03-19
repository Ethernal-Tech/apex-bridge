package databaseaccess

import (
	"encoding/json"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"go.etcd.io/bbolt"
)

type BBoltDatabase struct {
	db *bbolt.DB
}

var (
	bridgingRequestStatesBucket = []byte("BridgingRequestStates")
)

var _ core.Database = (*BBoltDatabase)(nil)

func (bd *BBoltDatabase) Init(filePath string) error {
	db, err := bbolt.Open(filePath, 0660, nil)
	if err != nil {
		return fmt.Errorf("could not open db: %w", err)
	}

	bd.db = db

	return db.Update(func(tx *bbolt.Tx) error {
		for _, bn := range [][]byte{bridgingRequestStatesBucket} {
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
	sourceChainID string, sourceTxHash common.Hash,
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
