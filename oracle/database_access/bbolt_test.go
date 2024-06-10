package databaseaccess

import (
	"os"
	"path"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
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
			{OriginChainID: "prime"},
			{OriginChainID: "vector"},
		})
		require.NoError(t, err)
	})

	t.Run("GetUnprocessedTxs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		expectedTxs := []*core.CardanoTx{
			{OriginChainID: "prime", Priority: 0},
			{OriginChainID: "vector", Priority: 1},
		}

		err = db.AddUnprocessedTxs(expectedTxs)
		require.NoError(t, err)

		txs, err := db.GetUnprocessedTxs("prime", 0, 0)
		require.NoError(t, err)
		require.NotNil(t, txs)
		require.Len(t, txs, 1)
		require.Equal(t, expectedTxs[0], txs[0])

		txs, err = db.GetUnprocessedTxs("vector", 1, 1)
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
			{OriginChainID: "prime"},
			{OriginChainID: "vector"},
		}

		err = db.AddUnprocessedTxs(expectedTxs)
		require.NoError(t, err)

		txs, err := db.GetAllUnprocessedTxs("prime", 0)
		require.NoError(t, err)
		require.NotNil(t, txs)
		require.Len(t, txs, 1)
		require.Equal(t, expectedTxs[0], txs[0])

		txs, err = db.GetAllUnprocessedTxs("vector", 1)
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
			{OriginChainID: "prime"},
			{OriginChainID: "vector"},
		}

		err = db.AddUnprocessedTxs(expectedTxs)
		require.NoError(t, err)

		err = db.ClearUnprocessedTxs("prime")
		require.NoError(t, err)

		txs, err := db.GetAllUnprocessedTxs("prime", 0)
		require.NoError(t, err)
		require.Nil(t, txs)

		err = db.ClearUnprocessedTxs("vector")
		require.NoError(t, err)

		txs, err = db.GetAllUnprocessedTxs("vector", 0)
		require.NoError(t, err)
		require.Nil(t, txs)
	})

	t.Run("MarkUnprocessedTxsAsProcessed", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		expectedTxs := []*core.CardanoTx{
			{OriginChainID: "prime"},
			{OriginChainID: "vector"},
		}

		err = db.AddUnprocessedTxs(expectedTxs)
		require.NoError(t, err)

		var expectedProcessedTxs []*core.ProcessedCardanoTx

		txs, err := db.GetAllUnprocessedTxs("prime", 0)
		require.NoError(t, err)
		require.NotNil(t, txs)

		for _, tx := range txs {
			expectedProcessedTxs = append(expectedProcessedTxs, tx.ToProcessedCardanoTx(false))
		}

		txs, err = db.GetAllUnprocessedTxs("vector", 0)
		require.NoError(t, err)
		require.NotNil(t, txs)

		for _, tx := range txs {
			expectedProcessedTxs = append(expectedProcessedTxs, tx.ToProcessedCardanoTx(false))
		}

		err = db.MarkUnprocessedTxsAsProcessed([]*core.ProcessedCardanoTx{})
		require.NoError(t, err)

		err = db.MarkUnprocessedTxsAsProcessed(expectedProcessedTxs)
		require.NoError(t, err)

		txs, err = db.GetAllUnprocessedTxs("prime", 0)
		require.NoError(t, err)
		require.Nil(t, txs)

		txs, err = db.GetAllUnprocessedTxs("vector", 0)
		require.NoError(t, err)
		require.Nil(t, txs)
	})

	t.Run("GetProcessedTx", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		expectedTxs := []*core.CardanoTx{
			{OriginChainID: "prime"},
			{OriginChainID: "vector"},
		}

		err = db.AddUnprocessedTxs(expectedTxs)
		require.NoError(t, err)

		var expectedProcessedTxs []*core.ProcessedCardanoTx

		txs, err := db.GetAllUnprocessedTxs("prime", 0)
		require.NoError(t, err)
		require.NotNil(t, txs)

		for _, tx := range txs {
			expectedProcessedTxs = append(expectedProcessedTxs, tx.ToProcessedCardanoTx(false))
		}

		txs, err = db.GetAllUnprocessedTxs("vector", 0)
		require.NoError(t, err)
		require.NotNil(t, txs)

		for _, tx := range txs {
			expectedProcessedTxs = append(expectedProcessedTxs, tx.ToProcessedCardanoTx(false))
		}

		err = db.MarkUnprocessedTxsAsProcessed(expectedProcessedTxs)
		require.NoError(t, err)

		tx, err := db.GetProcessedTx("", "")
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
			{ChainID: "prime"},
			{ChainID: "vector"},
		})
		require.NoError(t, err)
	})

	t.Run("GetExpectedTxs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		const (
			primeChainID  = "prime"
			vectorChainID = "vector"
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
			primeChainID  = "prime"
			vectorChainID = "vector"
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
			{ChainID: "prime"},
			{ChainID: "vector"},
		}

		err = db.AddExpectedTxs(expectedTxs)
		require.NoError(t, err)

		err = db.ClearExpectedTxs("prime")
		require.NoError(t, err)

		txs, err := db.GetAllExpectedTxs("prime", 0)
		require.NoError(t, err)
		require.Nil(t, txs)

		err = db.ClearExpectedTxs("vector")
		require.NoError(t, err)

		txs, err = db.GetAllExpectedTxs("vector", 0)
		require.NoError(t, err)
		require.Nil(t, txs)
	})

	t.Run("MarkExpectedTxsAsProcessed", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		expectedTxs := []*core.BridgeExpectedCardanoTx{
			{ChainID: "prime"},
			{ChainID: "vector"},
		}

		err = db.AddExpectedTxs(expectedTxs)
		require.NoError(t, err)

		err = db.MarkExpectedTxsAsProcessed([]*core.BridgeExpectedCardanoTx{})
		require.NoError(t, err)

		txs, err := db.GetAllExpectedTxs("prime", 0)
		require.NoError(t, err)
		require.NotNil(t, txs)

		err = db.MarkExpectedTxsAsProcessed(txs)
		require.NoError(t, err)

		txs, err = db.GetAllExpectedTxs("vector", 0)
		require.NoError(t, err)
		require.NotNil(t, txs)

		err = db.MarkExpectedTxsAsProcessed(txs)
		require.NoError(t, err)

		txs, err = db.GetAllExpectedTxs("prime", 0)
		require.NoError(t, err)
		require.Nil(t, txs)

		txs, err = db.GetAllExpectedTxs("vector", 0)
		require.NoError(t, err)
		require.Nil(t, txs)
	})

	t.Run("MarkExpectedTxsAsInvalid", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		const (
			primeChainID  = "prime"
			vectorChainID = "vector"
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
}
