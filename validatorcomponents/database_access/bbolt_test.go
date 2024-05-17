package databaseaccess

import (
	"math/big"
	"os"
	"path"
	"testing"

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

	filePath := path.Join(testDir, "temp_test.db")

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

	const (
		primeChainID = "chainId"
		testTxHash   = "0xtest"
	)

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

		state, err := db.GetBridgingRequestState("vect", "0xtest2")
		require.NoError(t, err)
		require.Nil(t, state)

		state, err = db.GetBridgingRequestState(primeChainID, testTxHash)
		require.NoError(t, err)
		require.NotNil(t, state)
	})

	t.Run("GetBridgingRequestStatesByBatchID", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		sourceChainID := primeChainID
		sourceTxHash := testTxHash
		destinationChainID := "vector"
		batchID := uint64(1)

		stateToAdd := core.NewBridgingRequestState(sourceChainID, sourceTxHash)
		require.NoError(t, stateToAdd.ToSubmittedToBridge(destinationChainID))
		require.NoError(t, stateToAdd.ToIncludedInBatch(batchID))

		err = db.AddBridgingRequestState(stateToAdd)
		require.NoError(t, err)

		states, err := db.GetBridgingRequestStatesByBatchID(destinationChainID, 2)
		require.NoError(t, err)
		require.Nil(t, states)

		states, err = db.GetBridgingRequestStatesByBatchID(destinationChainID, batchID)
		require.NoError(t, err)
		require.NotNil(t, states)
		require.Len(t, states, 1)
		require.Equal(t, sourceChainID, states[0].SourceChainID)
		require.Equal(t, sourceTxHash, states[0].SourceTxHash)
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

		require.NoError(t, state.ToInvalidRequest())
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
