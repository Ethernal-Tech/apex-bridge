package database_access

import (
	"os"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/stretchr/testify/require"
)

func TestBoltDatabase(t *testing.T) {
	const filePath = "temp_test.db"

	dbCleanup := func() {
		if _, err := os.Stat(filePath); err == nil {
			os.Remove(filePath)
		}
	}

	t.Cleanup(dbCleanup)

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
			{OriginChainId: "prime"},
			{OriginChainId: "vector"},
		})
		require.NoError(t, err)
	})

	t.Run("GetUnprocessedTxs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		expectedTxs := []*core.CardanoTx{
			{OriginChainId: "prime"},
			{OriginChainId: "vector"},
		}

		err = db.AddUnprocessedTxs(expectedTxs)
		require.NoError(t, err)

		txs, err := db.GetUnprocessedTxs("prime", 0)
		require.NoError(t, err)
		require.NotNil(t, txs)
		require.Len(t, txs, 1)
		require.Equal(t, expectedTxs[0], txs[0])

		txs, err = db.GetUnprocessedTxs("vector", 1)
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
			{OriginChainId: "prime"},
			{OriginChainId: "vector"},
		}

		err = db.AddUnprocessedTxs(expectedTxs)
		require.NoError(t, err)

		err = db.ClearUnprocessedTxs("prime")
		require.NoError(t, err)

		txs, err := db.GetUnprocessedTxs("prime", 0)
		require.NoError(t, err)
		require.Nil(t, txs)

		err = db.ClearUnprocessedTxs("vector")
		require.NoError(t, err)

		txs, err = db.GetUnprocessedTxs("vector", 0)
		require.NoError(t, err)
		require.Nil(t, txs)
	})

	t.Run("MarkUnprocessedTxsAsProcessed", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		expectedTxs := []*core.CardanoTx{
			{OriginChainId: "prime"},
			{OriginChainId: "vector"},
		}

		err = db.AddUnprocessedTxs(expectedTxs)
		require.NoError(t, err)

		var expectedProcessedTxs []*core.ProcessedCardanoTx

		txs, err := db.GetUnprocessedTxs("prime", 0)
		require.NoError(t, err)
		require.NotNil(t, txs)

		for _, tx := range txs {
			expectedProcessedTxs = append(expectedProcessedTxs, tx.ToProcessedCardanoTx(false))
		}

		txs, err = db.GetUnprocessedTxs("vector", 0)
		require.NoError(t, err)
		require.NotNil(t, txs)

		for _, tx := range txs {
			expectedProcessedTxs = append(expectedProcessedTxs, tx.ToProcessedCardanoTx(false))
		}

		err = db.MarkUnprocessedTxsAsProcessed([]*core.ProcessedCardanoTx{})
		require.NoError(t, err)

		err = db.MarkUnprocessedTxsAsProcessed(expectedProcessedTxs)
		require.NoError(t, err)

		txs, err = db.GetUnprocessedTxs("prime", 0)
		require.NoError(t, err)
		require.Nil(t, txs)

		txs, err = db.GetUnprocessedTxs("vector", 0)
		require.NoError(t, err)
		require.Nil(t, txs)
	})

	t.Run("GetProcessedTx", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		expectedTxs := []*core.CardanoTx{
			{OriginChainId: "prime"},
			{OriginChainId: "vector"},
		}

		err = db.AddUnprocessedTxs(expectedTxs)
		require.NoError(t, err)

		var expectedProcessedTxs []*core.ProcessedCardanoTx

		txs, err := db.GetUnprocessedTxs("prime", 0)
		require.NoError(t, err)
		require.NotNil(t, txs)

		for _, tx := range txs {
			expectedProcessedTxs = append(expectedProcessedTxs, tx.ToProcessedCardanoTx(false))
		}

		txs, err = db.GetUnprocessedTxs("vector", 0)
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

		tx, err = db.GetProcessedTx(expectedProcessedTxs[0].OriginChainId, expectedProcessedTxs[0].Hash)
		require.NoError(t, err)
		require.NotNil(t, tx)
		require.Equal(t, expectedProcessedTxs[0], tx)

		tx, err = db.GetProcessedTx(expectedProcessedTxs[1].OriginChainId, expectedProcessedTxs[1].Hash)
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
			{ChainId: "prime"},
			{ChainId: "vector"},
		})
		require.NoError(t, err)
	})

	t.Run("GetExpectedTxs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		expectedTxs := []*core.BridgeExpectedCardanoTx{
			{ChainId: "prime"},
			{ChainId: "vector"},
		}

		err = db.AddExpectedTxs(expectedTxs)
		require.NoError(t, err)

		txs, err := db.GetExpectedTxs("prime", 0)
		require.NoError(t, err)
		require.NotNil(t, txs)
		require.Len(t, txs, 1)
		require.Equal(t, expectedTxs[0], txs[0])

		txs, err = db.GetExpectedTxs("vector", 1)
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
			{ChainId: "prime"},
			{ChainId: "vector"},
		}

		err = db.AddExpectedTxs(expectedTxs)
		require.NoError(t, err)

		err = db.ClearExpectedTxs("prime")
		require.NoError(t, err)

		txs, err := db.GetExpectedTxs("prime", 0)
		require.NoError(t, err)
		require.Nil(t, txs)

		err = db.ClearExpectedTxs("vector")
		require.NoError(t, err)

		txs, err = db.GetExpectedTxs("vector", 0)
		require.NoError(t, err)
		require.Nil(t, txs)
	})

	t.Run("MarkExpectedTxsAsProcessed", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		expectedTxs := []*core.BridgeExpectedCardanoTx{
			{ChainId: "prime"},
			{ChainId: "vector"},
		}

		err = db.AddExpectedTxs(expectedTxs)
		require.NoError(t, err)

		err = db.MarkExpectedTxsAsProcessed([]*core.BridgeExpectedCardanoTx{})
		require.NoError(t, err)

		txs, err := db.GetExpectedTxs("prime", 0)
		require.NoError(t, err)
		require.NotNil(t, txs)

		err = db.MarkExpectedTxsAsProcessed(txs)
		require.NoError(t, err)

		txs, err = db.GetExpectedTxs("vector", 0)
		require.NoError(t, err)
		require.NotNil(t, txs)

		err = db.MarkExpectedTxsAsProcessed(txs)
		require.NoError(t, err)

		txs, err = db.GetExpectedTxs("prime", 0)
		require.NoError(t, err)
		require.Nil(t, txs)

		txs, err = db.GetExpectedTxs("vector", 0)
		require.NoError(t, err)
		require.Nil(t, txs)
	})

	t.Run("MarkExpectedTxsAsInvalid", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db := &BBoltDatabase{}
		err := db.Init(filePath)
		require.NoError(t, err)

		expectedTxs := []*core.BridgeExpectedCardanoTx{
			{ChainId: "prime"},
			{ChainId: "vector"},
		}

		err = db.AddExpectedTxs(expectedTxs)
		require.NoError(t, err)

		err = db.MarkExpectedTxsAsInvalid([]*core.BridgeExpectedCardanoTx{})
		require.NoError(t, err)

		txs, err := db.GetExpectedTxs("prime", 0)
		require.NoError(t, err)
		require.NotNil(t, txs)

		err = db.MarkExpectedTxsAsInvalid(txs)
		require.NoError(t, err)

		txs, err = db.GetExpectedTxs("vector", 0)
		require.NoError(t, err)
		require.NotNil(t, txs)

		err = db.MarkExpectedTxsAsInvalid(txs)
		require.NoError(t, err)

		txs, err = db.GetExpectedTxs("prime", 0)
		require.NoError(t, err)
		require.Nil(t, txs)

		txs, err = db.GetExpectedTxs("vector", 0)
		require.NoError(t, err)
		require.Nil(t, txs)
	})
}
