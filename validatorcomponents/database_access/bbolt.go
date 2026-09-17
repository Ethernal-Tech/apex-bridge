package databaseaccess

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
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
	// bridgingRequestStatesSyncBucket is an insertion ordered index over bridgingRequestStatesBucket:
	// big endian sequence number -> bridging request state key. The primary bucket is keyed by tx
	// hash, so it cannot be walked in a resumable order; this one can.
	bridgingRequestStatesSyncBucket = []byte("BridgingRequestStatesSync")
	syncMetaBucket                  = []byte("BridgingRequestStatesSyncMeta")
)

var (
	syncMetaInstanceIDKey = []byte("instanceId")
	syncMetaBackfilledKey = []byte("backfilled")
)

func uint64ToBEBytes(value uint64) []byte {
	result := make([]byte, 8) //nolint:mnd
	binary.BigEndian.PutUint64(result, value)

	return result
}

var _ core.Database = (*BBoltDatabase)(nil)

func (bd *BBoltDatabase) Init(filePath string) error {
	db, err := bbolt.Open(filePath, 0660, nil)
	if err != nil {
		return fmt.Errorf("could not open db: %w", err)
	}

	bd.db = db

	return db.Update(func(tx *bbolt.Tx) error {
		for _, bn := range [][]byte{
			bridgingRequestStatesBucket, protocolParamsBucket,
			bridgingRequestStatesSyncBucket, syncMetaBucket,
		} {
			_, err := tx.CreateBucketIfNotExists(bn)
			if err != nil {
				return fmt.Errorf("could not bucket: %s, err: %w", string(bn), err)
			}
		}

		return initSyncIndex(tx)
	})
}

// initSyncIndex indexes every bridging request state stored before the index existed. It runs once,
// the first time a database is opened by a version that has the index, which is what lets
// AddBridgingRequestState be the only other place that ever writes to it.
func initSyncIndex(tx *bbolt.Tx) error {
	metaBucket := tx.Bucket(syncMetaBucket)

	if len(metaBucket.Get(syncMetaInstanceIDKey)) == 0 {
		instanceID := make([]byte, 16) //nolint:mnd
		if _, err := rand.Read(instanceID); err != nil {
			return fmt.Errorf("could not generate sync instance id: %w", err)
		}

		if err := metaBucket.Put(syncMetaInstanceIDKey, []byte(hex.EncodeToString(instanceID))); err != nil {
			return fmt.Errorf("sync instance id write error: %w", err)
		}
	}

	if len(metaBucket.Get(syncMetaBackfilledKey)) > 0 {
		return nil
	}

	cursor := tx.Bucket(bridgingRequestStatesBucket).Cursor()

	for k, _ := cursor.First(); k != nil; k, _ = cursor.Next() {
		if err := addToSyncIndex(tx, k); err != nil {
			return err
		}
	}

	return metaBucket.Put(syncMetaBackfilledKey, []byte{1})
}

// addToSyncIndex appends a bridging request state key to the insertion ordered index.
func addToSyncIndex(tx *bbolt.Tx, stateKey []byte) error {
	syncBucket := tx.Bucket(bridgingRequestStatesSyncBucket)

	syncIndex, err := syncBucket.NextSequence()
	if err != nil {
		return fmt.Errorf("could not get next sync index: %w", err)
	}

	if err := syncBucket.Put(uint64ToBEBytes(syncIndex), stateKey); err != nil {
		return fmt.Errorf("sync index write error: %w", err)
	}

	return nil
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

		if state.CreatedAt.IsZero() {
			state.CreatedAt = time.Now().UTC()
		}

		if err := addToSyncIndex(tx, state.ToDBKey()); err != nil {
			return err
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

		// no sync index write here: the state is already in the primary bucket, so it was either
		// added by AddBridgingRequestState or covered by the initSyncIndex backfill

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

// GetBridgingRequestStatesPage implements core.Database. It returns up to limit states starting at
// fromSyncIndex, in insertion order, along with the sync index to resume from on the next call.
func (bd *BBoltDatabase) GetBridgingRequestStatesPage(
	fromSyncIndex uint64, limit int,
) (
	result []*common.BridgingRequestState, nextSyncIndex uint64, err error,
) {
	result = make([]*common.BridgingRequestState, 0, limit)
	nextSyncIndex = fromSyncIndex

	err = bd.db.View(func(tx *bbolt.Tx) error {
		statesBucket := tx.Bucket(bridgingRequestStatesBucket)
		cursor := tx.Bucket(bridgingRequestStatesSyncBucket).Cursor()

		for k, v := cursor.Seek(uint64ToBEBytes(fromSyncIndex)); k != nil; k, v = cursor.Next() {
			if len(result) == limit {
				break
			}

			nextSyncIndex = binary.BigEndian.Uint64(k) + 1

			data := statesBucket.Get(v)
			if len(data) == 0 {
				continue
			}

			var state *common.BridgingRequestState
			if err := json.Unmarshal(data, &state); err != nil {
				return fmt.Errorf("could not unmarshal BridgingRequestState: %w", err)
			}

			result = append(result, state)
		}

		return nil
	})
	if err != nil {
		return nil, fromSyncIndex, err
	}

	return result, nextSyncIndex, nil
}

// GetSyncInstanceID implements core.Database. It changes whenever the database is recreated, which
// tells a paginated reader that its stored sync index no longer means anything.
func (bd *BBoltDatabase) GetSyncInstanceID() (result string, err error) {
	err = bd.db.View(func(tx *bbolt.Tx) error {
		result = string(tx.Bucket(syncMetaBucket).Get(syncMetaInstanceIDKey))

		return nil
	})

	return result, err
}
