package databaseaccess

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
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

	t.Run("AddUnprocessedTxs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		err = db.AddUnprocessedTxs(nil)
		require.NoError(t, err)

		err = db.AddUnprocessedTxs([]*core.CardanoTx{})
		require.NoError(t, err)

		err = db.AddUnprocessedTxs([]*core.CardanoTx{
			{OriginChainID: common.ChainIDStrPrime},
			{OriginChainID: common.ChainIDStrVector},
		})
		require.NoError(t, err)
	})

	t.Run("GetUnprocessedTxs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		expectedTxs := []*core.CardanoTx{
			{OriginChainID: common.ChainIDStrPrime, Priority: 0},
			{OriginChainID: common.ChainIDStrVector, Priority: 1},
		}

		err = db.AddUnprocessedTxs(expectedTxs)
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
		err := db.Init(filePath)
		require.NoError(t, err)

		expectedTxs := []*core.CardanoTx{
			{OriginChainID: common.ChainIDStrPrime},
			{OriginChainID: common.ChainIDStrVector},
		}

		err = db.AddUnprocessedTxs(expectedTxs)
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

	t.Run("ClearUnprocessedTxs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		expectedTxs := []*core.CardanoTx{
			{OriginChainID: common.ChainIDStrPrime},
			{OriginChainID: common.ChainIDStrVector},
		}

		err = db.AddUnprocessedTxs(expectedTxs)
		require.NoError(t, err)

		err = db.ClearUnprocessedTxs(common.ChainIDStrPrime)
		require.NoError(t, err)

		txs, err := db.GetAllUnprocessedTxs(common.ChainIDStrPrime, 0)
		require.NoError(t, err)
		require.Nil(t, txs)

		err = db.ClearUnprocessedTxs(common.ChainIDStrVector)
		require.NoError(t, err)

		txs, err = db.GetAllUnprocessedTxs(common.ChainIDStrVector, 0)
		require.NoError(t, err)
		require.Nil(t, txs)
	})

	t.Run("MarkUnprocessedTxsAsProcessed", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		expectedTxs := []*core.CardanoTx{
			{OriginChainID: common.ChainIDStrPrime},
			{OriginChainID: common.ChainIDStrVector},
		}

		err = db.AddUnprocessedTxs(expectedTxs)
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

		err = db.MarkUnprocessedTxsAsProcessed([]*core.ProcessedCardanoTx{})
		require.NoError(t, err)

		err = db.MarkUnprocessedTxsAsProcessed(expectedProcessedTxs)
		require.NoError(t, err)

		txs, err = db.GetAllUnprocessedTxs(common.ChainIDStrPrime, 0)
		require.NoError(t, err)
		require.Nil(t, txs)

		txs, err = db.GetAllUnprocessedTxs(common.ChainIDStrVector, 0)
		require.NoError(t, err)
		require.Nil(t, txs)
	})

	t.Run("GetProcessedTx", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		expectedTxs := []*core.CardanoTx{
			{OriginChainID: common.ChainIDStrPrime},
			{OriginChainID: common.ChainIDStrVector},
		}

		err = db.AddUnprocessedTxs(expectedTxs)
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

		err = db.MarkUnprocessedTxsAsProcessed(expectedProcessedTxs)
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
		err := db.Init(filePath)
		require.NoError(t, err)

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
		err := db.Init(filePath)
		require.NoError(t, err)

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
		err := db.Init(filePath)
		require.NoError(t, err)

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

	t.Run("ClearExpectedTxs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		expectedTxs := []*core.BridgeExpectedCardanoTx{
			{ChainID: common.ChainIDStrPrime},
			{ChainID: common.ChainIDStrVector},
		}

		err = db.AddExpectedTxs(expectedTxs)
		require.NoError(t, err)

		err = db.ClearExpectedTxs(common.ChainIDStrPrime)
		require.NoError(t, err)

		txs, err := db.GetAllExpectedTxs(common.ChainIDStrPrime, 0)
		require.NoError(t, err)
		require.Nil(t, txs)

		err = db.ClearExpectedTxs(common.ChainIDStrVector)
		require.NoError(t, err)

		txs, err = db.GetAllExpectedTxs(common.ChainIDStrVector, 0)
		require.NoError(t, err)
		require.Nil(t, txs)
	})

	t.Run("MarkExpectedTxsAsProcessed", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		expectedTxs := []*core.BridgeExpectedCardanoTx{
			{ChainID: common.ChainIDStrPrime},
			{ChainID: common.ChainIDStrVector},
		}

		err = db.AddExpectedTxs(expectedTxs)
		require.NoError(t, err)

		err = db.MarkExpectedTxsAsProcessed([]*core.BridgeExpectedCardanoTx{})
		require.NoError(t, err)

		txs, err := db.GetAllExpectedTxs(common.ChainIDStrPrime, 0)
		require.NoError(t, err)
		require.NotNil(t, txs)

		err = db.MarkExpectedTxsAsProcessed(txs)
		require.NoError(t, err)

		txs, err = db.GetAllExpectedTxs(common.ChainIDStrVector, 0)
		require.NoError(t, err)
		require.NotNil(t, txs)

		err = db.MarkExpectedTxsAsProcessed(txs)
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
		err := db.Init(filePath)
		require.NoError(t, err)

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

		err = db.MarkExpectedTxsAsInvalid([]*core.BridgeExpectedCardanoTx{})
		require.NoError(t, err)

		txs, err := db.GetAllExpectedTxs(primeChainID, 0)
		require.NoError(t, err)
		require.NotNil(t, txs)

		err = db.MarkExpectedTxsAsInvalid(txs)
		require.NoError(t, err)

		txs, err = db.GetAllExpectedTxs(vectorChainID, 0)
		require.NoError(t, err)
		require.NotNil(t, txs)

		err = db.MarkExpectedTxsAsInvalid(txs)
		require.NoError(t, err)

		txs, err = db.GetAllExpectedTxs(primeChainID, 0)
		require.NoError(t, err)
		require.Nil(t, txs)

		txs, err = db.GetAllExpectedTxs(vectorChainID, 0)
		require.NoError(t, err)
		require.Nil(t, txs)
	})

	t.Run("AddChainBalance", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		err = db.AddChainBalance(&core.ChainBalance{})
		require.NoError(t, err)

		err = db.AddChainBalance(&core.ChainBalance{
			ChainID: "prime",
			Height:  uint64(1000),
			Amount:  "19999999999",
		})
		require.NoError(t, err)
	})

	t.Run("GetChainBalance", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		const (
			primeChainID  = common.ChainIDStrPrime
			vectorChainID = common.ChainIDStrVector
		)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		primeBalanceUpdate := &core.ChainBalance{
			ChainID: primeChainID,
			Height:  1,
			Amount:  "1",
		}

		vectorBalanceUpdate := &core.ChainBalance{
			ChainID: vectorChainID,
			Height:  2,
			Amount:  "2",
		}

		err = db.AddChainBalance(primeBalanceUpdate)
		require.NoError(t, err)

		err = db.AddChainBalance(vectorBalanceUpdate)
		require.NoError(t, err)

		balance, err := db.GetChainBalance(primeChainID, 1)
		require.NoError(t, err)
		require.Equal(t, balance, primeBalanceUpdate)

		balance, err = db.GetChainBalance(vectorChainID, 2)
		require.NoError(t, err)
		require.Equal(t, balance, vectorBalanceUpdate)
	})

	t.Run("GetAllChainBalances", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		const (
			primeChainID = common.ChainIDStrPrime
		)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		primeBalanceUpdate := &core.ChainBalance{
			ChainID: primeChainID,
			Height:  1,
			Amount:  "1",
		}

		primeBalanceUpdate2 := &core.ChainBalance{
			ChainID: primeChainID,
			Height:  2,
			Amount:  "2",
		}

		err = db.AddChainBalance(primeBalanceUpdate)
		require.NoError(t, err)

		err = db.AddChainBalance(primeBalanceUpdate2)
		require.NoError(t, err)

		balances, err := db.GetAllChainBalances(primeChainID, 0)
		require.NoError(t, err)
		require.Len(t, balances, 2)
		require.Equal(t, balances[0], primeBalanceUpdate)
		require.Equal(t, balances[1], primeBalanceUpdate2)
	})

	t.Run("GetLastChainBalances", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		const (
			primeChainID = common.ChainIDStrPrime
		)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		primeBalanceUpdate := &core.ChainBalance{
			ChainID: primeChainID,
			Height:  1,
			Amount:  "1",
		}

		primeBalanceUpdate2 := &core.ChainBalance{
			ChainID: primeChainID,
			Height:  2,
			Amount:  "2",
		}

		err = db.AddChainBalance(primeBalanceUpdate)
		require.NoError(t, err)

		err = db.AddChainBalance(primeBalanceUpdate2)
		require.NoError(t, err)

		balances, err := db.GetLastChainBalances(primeChainID, 0)
		require.NoError(t, err)
		require.Len(t, balances, 2)
		require.Equal(t, balances[1], primeBalanceUpdate)
		require.Equal(t, balances[0], primeBalanceUpdate2)
	})
}
