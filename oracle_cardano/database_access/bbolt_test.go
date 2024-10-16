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

	t.Run("ClearUnprocessedTxs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		require.NoError(t, db.Init(filePath))

		expectedTxs := []*core.CardanoTx{
			{OriginChainID: common.ChainIDStrPrime},
			{OriginChainID: common.ChainIDStrVector},
		}

		err = db.AddTxs(nil, expectedTxs)
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

		err = db.MarkTxs(nil, nil, []*core.ProcessedCardanoTx{})
		require.NoError(t, err)

		err = db.MarkTxs(nil, nil, expectedProcessedTxs)
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

		err = db.MarkTxs(nil, nil, expectedProcessedTxs)
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

	t.Run("ClearExpectedTxs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		require.NoError(t, db.Init(filePath))

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
		require.NoError(t, db.Init(filePath))

		expectedTxs := []*core.BridgeExpectedCardanoTx{
			{ChainID: common.ChainIDStrPrime},
			{ChainID: common.ChainIDStrVector},
		}

		err = db.AddExpectedTxs(expectedTxs)
		require.NoError(t, err)

		err = db.MarkTxs(nil, []*core.BridgeExpectedCardanoTx{}, nil)
		require.NoError(t, err)

		txs, err := db.GetAllExpectedTxs(common.ChainIDStrPrime, 0)
		require.NoError(t, err)
		require.NotNil(t, txs)

		err = db.MarkTxs(nil, txs, nil)
		require.NoError(t, err)

		txs, err = db.GetAllExpectedTxs(common.ChainIDStrVector, 0)
		require.NoError(t, err)
		require.NotNil(t, txs)

		err = db.MarkTxs(nil, txs, nil)
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

		err = db.MarkTxs([]*core.BridgeExpectedCardanoTx{}, nil, nil)
		require.NoError(t, err)

		txs, err := db.GetAllExpectedTxs(primeChainID, 0)
		require.NoError(t, err)
		require.NotNil(t, txs)

		err = db.MarkTxs(txs, nil, nil)
		require.NoError(t, err)

		txs, err = db.GetAllExpectedTxs(vectorChainID, 0)
		require.NoError(t, err)
		require.NotNil(t, txs)

		err = db.MarkTxs(txs, nil, nil)
		require.NoError(t, err)

		txs, err = db.GetAllExpectedTxs(primeChainID, 0)
		require.NoError(t, err)
		require.Nil(t, txs)

		txs, err = db.GetAllExpectedTxs(vectorChainID, 0)
		require.NoError(t, err)
		require.Nil(t, txs)
	})
}
