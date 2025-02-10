package databaseaccess

import (
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/stretchr/testify/require"
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

	testTxHash := common.Hash{1, 2, 89, 188}

	t.Run("AddBridgingRequestState", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		state := core.NewBridgingRequestState(primeChainID, testTxHash)
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

		err = db.AddBridgingRequestState(core.NewBridgingRequestState(primeChainID, testTxHash))
		require.NoError(t, err)

		state, err := db.GetBridgingRequestState("vect", common.Hash{89, 8})
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

		err = db.UpdateBridgingRequestState(core.NewBridgingRequestState(sourceChainID, sourceTxHash))
		require.Error(t, err)
		require.ErrorContains(t, err, "trying to update a BridgingRequestState that does not exist")

		state := core.NewBridgingRequestState(sourceChainID, sourceTxHash)

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
		require.Equal(t, core.BridgingRequestStatusInvalidRequest, state.Status)
	})

	t.Run("AddLastSubmittedBatchID", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		batchID := big.NewInt(1)

		err = db.AddLastSubmittedBatchID(primeChainID, batchID)
		require.NoError(t, err)
	})

	t.Run("GetLastSubmittedBatchID", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		batchIDToInsert := big.NewInt(1)

		batchID, err := db.GetLastSubmittedBatchID(primeChainID)
		require.NoError(t, err)
		require.Nil(t, batchID)

		err = db.AddLastSubmittedBatchID(primeChainID, batchIDToInsert)
		require.NoError(t, err)

		batchID, err = db.GetLastSubmittedBatchID(primeChainID)
		require.NoError(t, err)
		require.NotNil(t, batchID)
		require.Equal(t, batchIDToInsert, batchID)
	})
}
