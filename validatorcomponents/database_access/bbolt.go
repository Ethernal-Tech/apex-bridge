package databaseaccess

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"go.etcd.io/bbolt"
)

type BBoltDatabase struct {
	db *bbolt.DB
}

var (
	bridgingRequestStatesBucket = []byte("BridgingRequestStates")
	protocolParamsBucket        = []byte("ProtocolParams")
)

var _ core.Database = (*BBoltDatabase)(nil)

func (bd *BBoltDatabase) Init(filePath string) error {
	db, err := bbolt.Open(filePath, 0660, nil)
	if err != nil {
		return fmt.Errorf("could not open db: %w", err)
	}

	bd.db = db

	return db.Update(func(tx *bbolt.Tx) error {
		for _, bn := range [][]byte{bridgingRequestStatesBucket, protocolParamsBucket} {
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
func (bd *BBoltDatabase) AddBridgingRequestState(state *common.BridgingRequestState) error {
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
func (bd *BBoltDatabase) UpdateBridgingRequestState(state *common.BridgingRequestState) error {
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

type protocolParamsRecord struct {
	ProtocolParams []byte    `json:"protocolParams"`
	ExpiresAt      time.Time `json:"expiresAt"`
}

func (bd *BBoltDatabase) SaveProtocolParams(chainID string, protocolParams []byte, expiresAt time.Time) error {
	return bd.db.Update(func(tx *bbolt.Tx) error {
		bytes, err := json.Marshal(protocolParamsRecord{
			ProtocolParams: protocolParams,
			ExpiresAt:      expiresAt,
		})
		if err != nil {
			return fmt.Errorf("could not marshal protocol params: %w", err)
		}

		if err := tx.Bucket(protocolParamsBucket).Put([]byte(chainID), bytes); err != nil {
			return fmt.Errorf("protocol params write error: %w", err)
		}

		return nil
	})
}

func (bd *BBoltDatabase) GetProtocolParams(chainID string) ([]byte, time.Time, error) {
	var record protocolParamsRecord

	err := bd.db.View(func(tx *bbolt.Tx) error {
		if data := tx.Bucket(protocolParamsBucket).Get([]byte(chainID)); len(data) > 0 {
			return json.Unmarshal(data, &record)
		}

		return nil
	})
	if err != nil {
		return nil, time.Time{}, err
	}

	return record.ProtocolParams, record.ExpiresAt, nil
}

// GetBridgingRequestState implements core.Database.
func (bd *BBoltDatabase) GetBridgingRequestState(
	sourceChainID string, sourceTxHash []byte,
) (
	result *common.BridgingRequestState, err error,
) {
	err = bd.db.View(func(tx *bbolt.Tx) error {
		data := tx.Bucket(bridgingRequestStatesBucket).Get(common.ToBridgingRequestStateDBKey(sourceChainID, sourceTxHash))
		if len(data) > 0 {
			return json.Unmarshal(data, &result)
		}

		return nil
	})

	return result, err
}
