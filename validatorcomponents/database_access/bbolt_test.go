package database_access

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

	t.Run("AddBridgingRequestState", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		state := core.NewBridgingRequestState("prime", "0xtest", nil)
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

		chainId := "prime"
		txHash := "0xtest"

		err = db.AddBridgingRequestState(core.NewBridgingRequestState(chainId, txHash, nil))
		require.NoError(t, err)

		state, err := db.GetBridgingRequestState("vect", "0xtest2")
		require.NoError(t, err)
		require.Nil(t, state)

		state, err = db.GetBridgingRequestState(chainId, txHash)
		require.NoError(t, err)
		require.NotNil(t, state)
	})

	t.Run("GetBridgingRequestStatesByBatchId", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		sourceChainId := "prime"
		sourceTxHash := "0xtest"
		destinationChainId := "vector"
		batchId := uint64(1)

		stateToAdd := core.NewBridgingRequestState(sourceChainId, sourceTxHash, nil)
		stateToAdd.ToSubmittedToBridge(destinationChainId)
		stateToAdd.ToIncludedInBatch(batchId)

		err = db.AddBridgingRequestState(stateToAdd)
		require.NoError(t, err)

		states, err := db.GetBridgingRequestStatesByBatchId(destinationChainId, 2)
		require.NoError(t, err)
		require.Nil(t, states)

		states, err = db.GetBridgingRequestStatesByBatchId(destinationChainId, batchId)
		require.NoError(t, err)
		require.NotNil(t, states)
		require.Len(t, states, 1)
		require.Equal(t, sourceChainId, states[0].SourceChainId)
		require.Equal(t, sourceTxHash, states[0].SourceTxHash)
	})

	t.Run("GetUserBridgingRequestStates", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		userAddr := "0xuser"
		userAddrs := []string{userAddr}
		sourceChainId := "prime"
		sourceTxHash := "0xtest"

		stateToAdd := core.NewBridgingRequestState(sourceChainId, sourceTxHash, userAddrs)

		err = db.AddBridgingRequestState(stateToAdd)
		require.NoError(t, err)

		states, err := db.GetUserBridgingRequestStates(sourceChainId, "0xtest")
		require.NoError(t, err)
		require.Nil(t, states)

		states, err = db.GetUserBridgingRequestStates(sourceChainId, userAddr)
		require.NoError(t, err)
		require.NotNil(t, states)
		require.Len(t, states, 1)
		require.Equal(t, sourceChainId, states[0].SourceChainId)
		require.Equal(t, sourceTxHash, states[0].SourceTxHash)
	})

	t.Run("UpdateBridgingRequestState", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		sourceChainId := "prime"
		sourceTxHash := "0xtest"

		err = db.UpdateBridgingRequestState(core.NewBridgingRequestState(sourceChainId, sourceTxHash, nil))
		require.Error(t, err)
		require.ErrorContains(t, err, "trying to update a BridgingRequestState that does not exist")

		state := core.NewBridgingRequestState(sourceChainId, sourceTxHash, nil)

		err = db.AddBridgingRequestState(state)
		require.NoError(t, err)

		state, err = db.GetBridgingRequestState(sourceChainId, sourceTxHash)
		require.NoError(t, err)
		require.NotNil(t, state)

		state.ToInvalidRequest()
		err = db.UpdateBridgingRequestState(state)
		require.NoError(t, err)

		state, err = db.GetBridgingRequestState(sourceChainId, sourceTxHash)
		require.NoError(t, err)
		require.NotNil(t, state)
		require.Equal(t, core.BridgingRequestStatusInvalidRequest, state.Status)
	})

	t.Run("AddLastSubmittedBatchId", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		chainId := "prime"
		batchId := big.NewInt(1)

		err = db.AddLastSubmittedBatchId(chainId, batchId)
		require.NoError(t, err)
	})

	t.Run("GetLastSubmittedBatchId", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		chainId := "prime"
		batchIdToInsert := big.NewInt(1)

		batchId, err := db.GetLastSubmittedBatchId(chainId)
		require.NoError(t, err)
		require.Nil(t, batchId)

		err = db.AddLastSubmittedBatchId(chainId, batchIdToInsert)
		require.NoError(t, err)

		batchId, err = db.GetLastSubmittedBatchId(chainId)
		require.NoError(t, err)
		require.NotNil(t, batchId)
		require.Equal(t, batchIdToInsert, batchId)
	})
}
