package databaseaccess

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
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
		require.NoError(t, db.Init(filePath))
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
		require.NoError(t, db.Init(filePath))

		err = db.Close()
		require.NoError(t, err)
	})

	t.Run("AddUnprocessedTxs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		tx := &core.CardanoTx{OriginChainID: common.ChainIDStrPrime}

		db := &BBoltDatabase{}
		require.NoError(t, db.Init(filePath))

		require.NoError(t, db.AddTxs(nil, []*core.CardanoTx{}))

		require.NoError(t, db.AddTxs(nil, []*core.CardanoTx{tx, {OriginChainID: common.ChainIDStrVector}}))

		listTxs, err := db.GetUnprocessedTxs(tx.OriginChainID, 0, 2)

		require.NoError(t, err)
		require.Len(t, listTxs, 1)
		require.Equal(t, tx, listTxs[0])
	})

	t.Run("AddProcessedTxs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		tx := &core.ProcessedCardanoTx{OriginChainID: common.ChainIDStrPrime}

		db := &BBoltDatabase{}
		require.NoError(t, db.Init(filePath))

		require.NoError(t, db.AddTxs([]*core.ProcessedCardanoTx{}, nil))

		require.NoError(t, db.AddTxs([]*core.ProcessedCardanoTx{tx}, nil))

		resTx, err := db.GetProcessedTx(tx.OriginChainID, tx.Hash)

		require.NoError(t, err)
		require.Equal(t, tx, resTx)
	})

	t.Run("GetUnprocessedTxs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		require.NoError(t, db.Init(filePath))

		expectedTxs := []*core.CardanoTx{
			{OriginChainID: common.ChainIDStrPrime, Priority: 0},
			{OriginChainID: common.ChainIDStrVector, Priority: 1},
		}

		err = db.AddTxs(nil, expectedTxs)
		require.NoError(t, err)

		txs, err := db.GetUnprocessedTxs(common.ChainIDStrPrime, 0, 0)
		require.NoError(t, err)
		require.NotNil(t, txs)
		require.Len(t, txs, 1)
		require.Equal(t, expectedTxs[0], txs[0])

		txs, err = db.GetUnprocessedTxs(common.ChainIDStrVector, 1, 1)
		require.NoError(t, err)
		require.NotNil(t, txs)
		require.Len(t, txs, 1)
		require.Equal(t, expectedTxs[1], txs[0])
	})

	t.Run("GetAllUnprocessedTxs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		require.NoError(t, db.Init(filePath))

		expectedTxs := []*core.CardanoTx{
			{OriginChainID: common.ChainIDStrPrime},
			{OriginChainID: common.ChainIDStrVector},
		}

		err = db.AddTxs(nil, expectedTxs)
		require.NoError(t, err)

		txs, err := db.GetAllUnprocessedTxs(common.ChainIDStrPrime, 0)
		require.NoError(t, err)
		require.NotNil(t, txs)
		require.Len(t, txs, 1)
		require.Equal(t, expectedTxs[0], txs[0])

		txs, err = db.GetAllUnprocessedTxs(common.ChainIDStrVector, 1)
		require.NoError(t, err)
		require.NotNil(t, txs)
		require.Len(t, txs, 1)
		require.Equal(t, expectedTxs[1], txs[0])
	})

	t.Run("ClearAllTxs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		require.NoError(t, db.Init(filePath))

		unprocessedTxs := []*core.CardanoTx{
			{OriginChainID: common.ChainIDStrPrime},
			{OriginChainID: common.ChainIDStrVector},
		}

		expectedTxs := []*core.BridgeExpectedCardanoTx{
			{ChainID: common.ChainIDStrPrime},
			{ChainID: common.ChainIDStrVector},
		}

		err = db.AddTxs(nil, unprocessedTxs)
		require.NoError(t, err)

		err = db.AddExpectedTxs(expectedTxs)
		require.NoError(t, err)

		err = db.ClearAllTxs(common.ChainIDStrPrime)
		require.NoError(t, err)

		txs, err := db.GetAllUnprocessedTxs(common.ChainIDStrPrime, 0)
		require.NoError(t, err)
		require.Nil(t, txs)

		exTxs, err := db.GetAllExpectedTxs(common.ChainIDStrPrime, 0)
		require.NoError(t, err)
		require.Nil(t, exTxs)

		err = db.ClearAllTxs(common.ChainIDStrVector)
		require.NoError(t, err)

		txs, err = db.GetAllUnprocessedTxs(common.ChainIDStrVector, 0)
		require.NoError(t, err)
		require.Nil(t, txs)

		exTxs, err = db.GetAllExpectedTxs(common.ChainIDStrVector, 0)
		require.NoError(t, err)
		require.Nil(t, exTxs)
	})

	t.Run("UpdateTxs - UpdateUnprocessed", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		require.NoError(t, db.Init(filePath))

		expectedTxs := []*core.CardanoTx{
			{OriginChainID: common.ChainIDStrPrime},
		}

		err = db.AddTxs(nil, expectedTxs)
		require.NoError(t, err)

		txs, err := db.GetAllUnprocessedTxs(common.ChainIDStrPrime, 0)
		require.NoError(t, err)
		require.NotNil(t, txs)

		for _, tx := range txs {
			tx.TryCount++
		}

		err = db.UpdateTxs(&core.CardanoUpdateTxsData{UpdateUnprocessed: txs})
		require.NoError(t, err)

		txs, err = db.GetAllUnprocessedTxs(common.ChainIDStrPrime, 0)
		require.NoError(t, err)
		require.Len(t, txs, 1)
		require.Equal(t, uint32(1), txs[0].TryCount)
	})

	t.Run("GetPendingTxs and UpdateTxs - MoveUnprocessedToPending", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		require.NoError(t, db.Init(filePath))

		expectedTxs := []*core.CardanoTx{
			{
				Priority:      1,
				OriginChainID: common.ChainIDStrPrime,
				Tx: indexer.Tx{
					BlockSlot: 1,
					Hash:      indexer.Hash{1, 2},
				},
			},
		}

		err = db.AddTxs(nil, expectedTxs)
		require.NoError(t, err)

		txs, err := db.GetAllUnprocessedTxs(common.ChainIDStrPrime, 0)
		require.NoError(t, err)
		require.NotNil(t, txs)

		keys := make([][]byte, len(expectedTxs))
		for i, tx := range expectedTxs {
			keys[i] = tx.ToUnprocessedTxKey()
		}

		_, err = db.GetPendingTxs(keys)
		require.Error(t, err)
		require.ErrorContains(t, err, "couldn't get pending tx for key")

		err = db.UpdateTxs(&core.CardanoUpdateTxsData{MoveUnprocessedToPending: txs})
		require.NoError(t, err)

		txs, err = db.GetAllUnprocessedTxs(common.ChainIDStrPrime, 0)
		require.NoError(t, err)
		require.Len(t, txs, 0)

		pendingTxs, err := db.GetPendingTxs(keys)
		require.NoError(t, err)
		require.Len(t, pendingTxs, 1)
	})

	t.Run("UpdateTxs - MoveUnprocessedToProcessed", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		require.NoError(t, db.Init(filePath))

		expectedTxs := []*core.CardanoTx{
			{OriginChainID: common.ChainIDStrPrime},
			{OriginChainID: common.ChainIDStrVector},
		}

		err = db.AddTxs(nil, expectedTxs)
		require.NoError(t, err)

		var expectedProcessedTxs []*core.ProcessedCardanoTx

		txs, err := db.GetAllUnprocessedTxs(common.ChainIDStrPrime, 0)
		require.NoError(t, err)
		require.NotNil(t, txs)

		for _, tx := range txs {
			expectedProcessedTxs = append(expectedProcessedTxs, tx.ToProcessedCardanoTx(false))
		}

		txs, err = db.GetAllUnprocessedTxs(common.ChainIDStrVector, 0)
		require.NoError(t, err)
		require.NotNil(t, txs)

		for _, tx := range txs {
			expectedProcessedTxs = append(expectedProcessedTxs, tx.ToProcessedCardanoTx(false))
		}

		err = db.UpdateTxs(&core.CardanoUpdateTxsData{MoveUnprocessedToProcessed: []*core.ProcessedCardanoTx{}})
		require.NoError(t, err)

		err = db.UpdateTxs(&core.CardanoUpdateTxsData{MoveUnprocessedToProcessed: expectedProcessedTxs})
		require.NoError(t, err)

		txs, err = db.GetAllUnprocessedTxs(common.ChainIDStrPrime, 0)
		require.NoError(t, err)
		require.Nil(t, txs)

		txs, err = db.GetAllUnprocessedTxs(common.ChainIDStrVector, 0)
		require.NoError(t, err)
		require.Nil(t, txs)

		for _, tx := range expectedProcessedTxs {
			pTx, err := db.GetProcessedTx(tx.OriginChainID, tx.Hash)
			require.NoError(t, err)
			require.NotNil(t, pTx)
		}
	})

	t.Run("UpdateTxs - MovePendingToUnprocessed", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		require.NoError(t, db.Init(filePath))

		expectedTxs := []*core.CardanoTx{
			{
				Priority:      1,
				OriginChainID: common.ChainIDStrPrime,
				Tx: indexer.Tx{
					BlockSlot: 1,
					Hash:      indexer.Hash{1, 2},
				},
			},
		}

		err = db.AddTxs(nil, expectedTxs)
		require.NoError(t, err)

		err = db.UpdateTxs(&core.CardanoUpdateTxsData{MoveUnprocessedToPending: expectedTxs})
		require.NoError(t, err)

		err = db.UpdateTxs(&core.CardanoUpdateTxsData{MovePendingToUnprocessed: expectedTxs})
		require.NoError(t, err)

		keys := make([][]byte, len(expectedTxs))
		for i, tx := range expectedTxs {
			keys[i] = tx.ToUnprocessedTxKey()
		}

		_, err := db.GetPendingTxs(keys)
		require.Error(t, err)
		require.ErrorContains(t, err, "couldn't get pending tx for key")

		txs, err := db.GetAllUnprocessedTxs(common.ChainIDStrPrime, 0)
		require.NoError(t, err)
		require.Len(t, txs, 1)
	})

	t.Run("UpdateTxs - MovePendingToProcessed", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		require.NoError(t, db.Init(filePath))

		expectedTxs := []*core.CardanoTx{
			{
				Priority:      1,
				OriginChainID: common.ChainIDStrPrime,
				Tx: indexer.Tx{
					BlockSlot: 1,
					Hash:      indexer.Hash{1, 2},
				},
			},
		}

		err = db.AddTxs(nil, expectedTxs)
		require.NoError(t, err)

		err = db.UpdateTxs(&core.CardanoUpdateTxsData{MoveUnprocessedToPending: expectedTxs})
		require.NoError(t, err)

		processedTxs := make([]*core.ProcessedCardanoTx, len(expectedTxs))
		for i, tx := range expectedTxs {
			processedTxs[i] = tx.ToProcessedCardanoTx(false)
		}

		err = db.UpdateTxs(&core.CardanoUpdateTxsData{MovePendingToProcessed: processedTxs})
		require.NoError(t, err)

		keys := make([][]byte, len(expectedTxs))
		for i, tx := range expectedTxs {
			keys[i] = tx.ToUnprocessedTxKey()
		}

		_, err := db.GetPendingTxs(keys)
		require.Error(t, err)
		require.ErrorContains(t, err, "couldn't get pending tx for key")

		for _, tx := range processedTxs {
			pTx, err := db.GetProcessedTx(tx.OriginChainID, tx.Hash)
			require.NoError(t, err)
			require.NotNil(t, pTx)
		}
	})

	t.Run("GetProcessedTx", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		require.NoError(t, db.Init(filePath))

		expectedTxs := []*core.CardanoTx{
			{OriginChainID: common.ChainIDStrPrime},
			{OriginChainID: common.ChainIDStrVector},
		}

		err = db.AddTxs(nil, expectedTxs)
		require.NoError(t, err)

		var expectedProcessedTxs []*core.ProcessedCardanoTx

		txs, err := db.GetAllUnprocessedTxs(common.ChainIDStrPrime, 0)
		require.NoError(t, err)
		require.NotNil(t, txs)

		for _, tx := range txs {
			expectedProcessedTxs = append(expectedProcessedTxs, tx.ToProcessedCardanoTx(false))
		}

		txs, err = db.GetAllUnprocessedTxs(common.ChainIDStrVector, 0)
		require.NoError(t, err)
		require.NotNil(t, txs)

		for _, tx := range txs {
			expectedProcessedTxs = append(expectedProcessedTxs, tx.ToProcessedCardanoTx(false))
		}

		err = db.UpdateTxs(&core.CardanoUpdateTxsData{MoveUnprocessedToProcessed: expectedProcessedTxs})
		require.NoError(t, err)

		tx, err := db.GetProcessedTx("", indexer.Hash{})
		require.NoError(t, err)
		require.Nil(t, tx)

		tx, err = db.GetProcessedTx(expectedProcessedTxs[0].OriginChainID, expectedProcessedTxs[0].Hash)
		require.NoError(t, err)
		require.NotNil(t, tx)
		require.Equal(t, expectedProcessedTxs[0], tx)

		tx, err = db.GetProcessedTx(expectedProcessedTxs[1].OriginChainID, expectedProcessedTxs[1].Hash)
		require.NoError(t, err)
		require.NotNil(t, tx)
		require.Equal(t, expectedProcessedTxs[1], tx)
	})

	t.Run("AddExpectedTxs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		require.NoError(t, db.Init(filePath))

		err = db.AddExpectedTxs(nil)
		require.NoError(t, err)

		err = db.AddExpectedTxs([]*core.BridgeExpectedCardanoTx{})
		require.NoError(t, err)

		err = db.AddExpectedTxs([]*core.BridgeExpectedCardanoTx{
			{ChainID: common.ChainIDStrPrime},
			{ChainID: common.ChainIDStrVector},
		})
		require.NoError(t, err)
	})

	t.Run("GetExpectedTxs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		const (
			primeChainID  = common.ChainIDStrPrime
			vectorChainID = common.ChainIDStrVector
		)

		db := &BBoltDatabase{}
		require.NoError(t, db.Init(filePath))

		expectedTxs := []*core.BridgeExpectedCardanoTx{
			{ChainID: primeChainID, Priority: 0},
			{ChainID: vectorChainID, Priority: 1},
		}

		err = db.AddExpectedTxs(expectedTxs)
		require.NoError(t, err)

		txs, err := db.GetExpectedTxs(primeChainID, 0, 0)
		require.NoError(t, err)
		require.NotNil(t, txs)
		require.Len(t, txs, 1)
		require.Equal(t, expectedTxs[0], txs[0])

		txs, err = db.GetExpectedTxs(vectorChainID, 1, 1)
		require.NoError(t, err)
		require.NotNil(t, txs)
		require.Len(t, txs, 1)
		require.Equal(t, expectedTxs[1], txs[0])
	})

	t.Run("GetAllExpectedTxs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		const (
			primeChainID  = common.ChainIDStrPrime
			vectorChainID = common.ChainIDStrVector
		)

		db := &BBoltDatabase{}
		require.NoError(t, db.Init(filePath))

		expectedTxs := []*core.BridgeExpectedCardanoTx{
			{ChainID: primeChainID},
			{ChainID: vectorChainID},
		}

		err = db.AddExpectedTxs(expectedTxs)
		require.NoError(t, err)

		txs, err := db.GetAllExpectedTxs(primeChainID, 0)
		require.NoError(t, err)
		require.NotNil(t, txs)
		require.Len(t, txs, 1)
		require.Equal(t, expectedTxs[0], txs[0])

		txs, err = db.GetAllExpectedTxs(vectorChainID, 1)
		require.NoError(t, err)
		require.NotNil(t, txs)
		require.Len(t, txs, 1)
		require.Equal(t, expectedTxs[1], txs[0])
	})

	t.Run("MarkExpectedTxsAsProcessed", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		require.NoError(t, db.Init(filePath))

		expectedTxs := []*core.BridgeExpectedCardanoTx{
			{ChainID: common.ChainIDStrPrime},
			{ChainID: common.ChainIDStrVector},
		}

		err = db.AddExpectedTxs(expectedTxs)
		require.NoError(t, err)

		err = db.UpdateTxs(&core.CardanoUpdateTxsData{ExpectedProcessed: []*core.BridgeExpectedCardanoTx{}})
		require.NoError(t, err)

		txs, err := db.GetAllExpectedTxs(common.ChainIDStrPrime, 0)
		require.NoError(t, err)
		require.NotNil(t, txs)

		err = db.UpdateTxs(&core.CardanoUpdateTxsData{ExpectedProcessed: txs})
		require.NoError(t, err)

		txs, err = db.GetAllExpectedTxs(common.ChainIDStrVector, 0)
		require.NoError(t, err)
		require.NotNil(t, txs)

		err = db.UpdateTxs(&core.CardanoUpdateTxsData{ExpectedProcessed: txs})
		require.NoError(t, err)

		txs, err = db.GetAllExpectedTxs(common.ChainIDStrPrime, 0)
		require.NoError(t, err)
		require.Nil(t, txs)

		txs, err = db.GetAllExpectedTxs(common.ChainIDStrVector, 0)
		require.NoError(t, err)
		require.Nil(t, txs)
	})

	t.Run("MarkExpectedTxsAsInvalid", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		require.NoError(t, db.Init(filePath))

		const (
			primeChainID  = common.ChainIDStrPrime
			vectorChainID = common.ChainIDStrVector
		)

		expectedTxs := []*core.BridgeExpectedCardanoTx{
			{ChainID: primeChainID},
			{ChainID: vectorChainID},
		}

		err = db.AddExpectedTxs(expectedTxs)
		require.NoError(t, err)

		err = db.UpdateTxs(&core.CardanoUpdateTxsData{ExpectedInvalid: []*core.BridgeExpectedCardanoTx{}})
		require.NoError(t, err)

		txs, err := db.GetAllExpectedTxs(primeChainID, 0)
		require.NoError(t, err)
		require.NotNil(t, txs)

		err = db.UpdateTxs(&core.CardanoUpdateTxsData{ExpectedInvalid: txs})
		require.NoError(t, err)

		txs, err = db.GetAllExpectedTxs(vectorChainID, 0)
		require.NoError(t, err)
		require.NotNil(t, txs)

		err = db.UpdateTxs(&core.CardanoUpdateTxsData{ExpectedInvalid: txs})
		require.NoError(t, err)

		txs, err = db.GetAllExpectedTxs(primeChainID, 0)
		require.NoError(t, err)
		require.Nil(t, txs)

		txs, err = db.GetAllExpectedTxs(vectorChainID, 0)
		require.NoError(t, err)
		require.Nil(t, txs)
	})
}
