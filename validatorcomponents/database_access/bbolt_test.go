package databaseaccess

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"
)

func TestBoltDatabase(t *testing.T) {
	testDir, err := os.MkdirTemp("", "boltdb-test")
	require.NoError(t, err)

	defer func() {
		os.RemoveAll(testDir)
		os.Remove(testDir)
	}()

	filePath := filepath.Join(testDir, "temp_test.db")

	dbCleanup := func() {
		if _, err := os.Stat(filePath); err == nil {
			os.Remove(filePath)
		}
	}

	t.Run("Init", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)
	})

	t.Run("Init should fail", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init("")
		require.Error(t, err)
	})

	t.Run("Close", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		err = db.Close()
		require.NoError(t, err)
	})

	const primeChainID = "chainId"

	testTxHash := []byte{1, 2, 89, 188}

	t.Run("AddBridgingRequestState", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		state := common.NewBridgingRequestState(primeChainID, testTxHash, false)
		err = db.AddBridgingRequestState(state)
		require.NoError(t, err)

		err = db.AddBridgingRequestState(state)
		require.Error(t, err)
		require.ErrorContains(t, err, "trying to add a BridgingRequestState that already exists")
	})

	t.Run("GetBridgingRequestState", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		err = db.AddBridgingRequestState(common.NewBridgingRequestState(primeChainID, testTxHash, false))
		require.NoError(t, err)

		state, err := db.GetBridgingRequestState("vect", []byte{89, 8})
		require.NoError(t, err)
		require.Nil(t, state)

		state, err = db.GetBridgingRequestState(primeChainID, testTxHash)
		require.NoError(t, err)
		require.NotNil(t, state)
	})

	t.Run("UpdateBridgingRequestState", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		sourceChainID := primeChainID
		sourceTxHash := testTxHash

		err = db.UpdateBridgingRequestState(common.NewBridgingRequestState(sourceChainID, sourceTxHash, false))
		require.Error(t, err)
		require.ErrorContains(t, err, "trying to update a BridgingRequestState that does not exist")

		state := common.NewBridgingRequestState(sourceChainID, sourceTxHash, false)

		err = db.AddBridgingRequestState(state)
		require.NoError(t, err)

		state, err = db.GetBridgingRequestState(sourceChainID, sourceTxHash)
		require.NoError(t, err)
		require.NotNil(t, state)

		state.ToInvalidRequest()
		err = db.UpdateBridgingRequestState(state)
		require.NoError(t, err)

		state, err = db.GetBridgingRequestState(sourceChainID, sourceTxHash)
		require.NoError(t, err)
		require.NotNil(t, state)
		require.Equal(t, common.BridgingRequestStatusInvalidRequest, state.Status)
	})

	t.Run("SaveGetProtocolParams", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		result, expiresAt, err := db.GetProtocolParams(primeChainID)
		require.NoError(t, err)
		require.Nil(t, result)
		require.True(t, expiresAt.IsZero())

		protocolParams := []byte(`{"maxTxSize":16384}`)
		expiry := time.Now().UTC().Add(time.Hour)

		err = db.SaveProtocolParams(primeChainID, protocolParams, expiry)
		require.NoError(t, err)

		result, expiresAt, err = db.GetProtocolParams(primeChainID)
		require.NoError(t, err)
		require.Equal(t, protocolParams, result)
		require.True(t, expiry.Equal(expiresAt))

		updatedParams := []byte(`{"maxTxSize":32768}`)
		updatedExpiry := expiry.Add(time.Hour)

		err = db.SaveProtocolParams(primeChainID, updatedParams, updatedExpiry)
		require.NoError(t, err)

		result, expiresAt, err = db.GetProtocolParams(primeChainID)
		require.NoError(t, err)
		require.Equal(t, updatedParams, result)
		require.True(t, updatedExpiry.Equal(expiresAt))
	})
}

func txHashesOf(states []*common.BridgingRequestState) [][]byte {
	result := make([][]byte, len(states))
	for i, state := range states {
		result[i] = state.SourceTxHash
	}

	return result
}

func TestBoltDatabaseSyncIndex(t *testing.T) {
	const chainID = "prime"

	newDB := func(t *testing.T) (*BBoltDatabase, string) {
		t.Helper()

		filePath := filepath.Join(t.TempDir(), "sync_index_test.db")

		db := &BBoltDatabase{}
		require.NoError(t, db.Init(filePath))

		t.Cleanup(func() {
			_ = db.Close()
		})

		return db, filePath
	}

	addStates := func(t *testing.T, db *BBoltDatabase, count int) {
		t.Helper()

		for i := 0; i < count; i++ {
			state := common.NewBridgingRequestState(chainID, []byte{byte(i)}, false)
			require.NoError(t, db.AddBridgingRequestState(state))
			require.False(t, state.CreatedAt.IsZero())
		}
	}

	t.Run("add stamps a creation time and indexes in insertion order", func(t *testing.T) {
		db, _ := newDB(t)

		addStates(t, db, 3)

		state, err := db.GetBridgingRequestState(chainID, []byte{2})
		require.NoError(t, err)
		require.False(t, state.CreatedAt.IsZero())

		states, nextFrom, err := db.GetBridgingRequestStatesPage(0, 100)
		require.NoError(t, err)
		require.Equal(t, [][]byte{{0}, {1}, {2}}, txHashesOf(states))
		require.Equal(t, uint64(4), nextFrom)
	})

	t.Run("update keeps the state at its original index", func(t *testing.T) {
		db, _ := newDB(t)

		addStates(t, db, 2)

		state, err := db.GetBridgingRequestState(chainID, []byte{0})
		require.NoError(t, err)

		state.ToInvalidRequest()
		require.NoError(t, db.UpdateBridgingRequestState(state))

		states, nextFrom, err := db.GetBridgingRequestStatesPage(0, 100)
		require.NoError(t, err)
		// still two entries in the same order, the update did not append a third
		require.Equal(t, [][]byte{{0}, {1}}, txHashesOf(states))
		require.Equal(t, uint64(3), nextFrom)
		require.Equal(t, common.BridgingRequestStatusInvalidRequest, states[0].Status)
	})

	t.Run("paging walks every state exactly once", func(t *testing.T) {
		db, _ := newDB(t)

		addStates(t, db, 5)

		var (
			seen [][]byte
			from uint64
		)

		for {
			states, nextFrom, err := db.GetBridgingRequestStatesPage(from, 2)
			require.NoError(t, err)

			seen = append(seen, txHashesOf(states)...)

			if len(states) == 0 {
				require.Equal(t, from, nextFrom)

				break
			}

			require.Greater(t, nextFrom, from)
			from = nextFrom
		}

		require.Equal(t, [][]byte{{0}, {1}, {2}, {3}, {4}}, seen)

		// the cursor after a full walk is stable and returns nothing
		states, stillNextFrom, err := db.GetBridgingRequestStatesPage(from, 2)
		require.NoError(t, err)
		require.Empty(t, states)
		require.Equal(t, from, stillNextFrom)
	})

	t.Run("paging resumes from a stored cursor and picks up later inserts", func(t *testing.T) {
		db, _ := newDB(t)

		addStates(t, db, 2)

		states, nextFrom, err := db.GetBridgingRequestStatesPage(0, 10)
		require.NoError(t, err)
		require.Len(t, states, 2)

		// a tx hash that sorts before the ones already stored, so it would be missed by a cursor
		// over the primary bucket
		require.NoError(t, db.AddBridgingRequestState(
			common.NewBridgingRequestState(chainID, []byte{}, false)))

		states, _, err = db.GetBridgingRequestStatesPage(nextFrom, 10)
		require.NoError(t, err)
		require.Equal(t, [][]byte{{}}, txHashesOf(states))
	})

	t.Run("states stored before the index existed are backfilled once", func(t *testing.T) {
		db, filePath := newDB(t)

		addStates(t, db, 3)
		require.NoError(t, db.Close())

		// drop the index and strip the fields a pre sync index version never wrote, leaving the
		// database exactly as that version left it
		reopened, err := bbolt.Open(filePath, 0660, nil)
		require.NoError(t, err)
		require.NoError(t, reopened.Update(func(tx *bbolt.Tx) error {
			require.NoError(t, tx.DeleteBucket(bridgingRequestStatesSyncBucket))
			require.NoError(t, tx.DeleteBucket(syncMetaBucket))

			bucket := tx.Bucket(bridgingRequestStatesBucket)

			var keys [][]byte

			cursor := bucket.Cursor()

			for k, _ := cursor.First(); k != nil; k, _ = cursor.Next() {
				keys = append(keys, append([]byte(nil), k...))
			}

			for _, key := range keys {
				var state *common.BridgingRequestState

				require.NoError(t, json.Unmarshal(bucket.Get(key), &state))

				state.CreatedAt = time.Time{}

				bytes, err := json.Marshal(state)
				require.NoError(t, err)
				require.NoError(t, bucket.Put(key, bytes))
			}

			return nil
		}))
		require.NoError(t, reopened.Close())

		backfilled := &BBoltDatabase{}
		require.NoError(t, backfilled.Init(filePath))

		states, _, err := backfilled.GetBridgingRequestStatesPage(0, 100)
		require.NoError(t, err)
		require.Equal(t, [][]byte{{0}, {1}, {2}}, txHashesOf(states))
		// the backfill indexes legacy records without inventing data for them
		require.True(t, states[0].CreatedAt.IsZero())
		require.Nil(t, states[0].Details)

		instanceID, err := backfilled.GetSyncInstanceID()
		require.NoError(t, err)
		require.NotEmpty(t, instanceID)

		require.NoError(t, backfilled.Close())

		// a second open must not index anything twice
		again := &BBoltDatabase{}
		require.NoError(t, again.Init(filePath))

		t.Cleanup(func() {
			_ = again.Close()
		})

		states, _, err = again.GetBridgingRequestStatesPage(0, 100)
		require.NoError(t, err)
		require.Len(t, states, 3)

		sameInstanceID, err := again.GetSyncInstanceID()
		require.NoError(t, err)
		require.Equal(t, instanceID, sameInstanceID)
	})
}
