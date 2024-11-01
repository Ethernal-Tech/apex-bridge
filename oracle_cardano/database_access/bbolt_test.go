package databaseaccess

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	cDatabaseaccess "github.com/Ethernal-Tech/apex-bridge/oracle_common/database_access"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/stretchr/testify/require"
)

func TestBoltDatabase(t *testing.T) {
	appConfig := &cCore.AppConfig{
		CardanoChains: map[string]*cCore.CardanoChainConfig{
			common.ChainIDStrPrime:  {},
			common.ChainIDStrVector: {},
		},
	}

	appConfig.FillOut()

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

	createDB := func(dbFilePath string) (*BBoltDatabase, error) {
		boltDB, err := cDatabaseaccess.NewDatabase(dbFilePath, appConfig)
		if err != nil {
			return nil, err
		}

		typeRegister := &cCore.TxTypeRegister{}
		typeRegister.SetTTxTypes(appConfig, reflect.TypeOf(core.CardanoTx{}), nil)

		oracleDB := &BBoltDatabase{}
		oracleDB.Init(boltDB, appConfig, typeRegister)

		return oracleDB, nil
	}

	t.Run("Init", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		_, err := createDB(filePath)
		require.NoError(t, err)
	})

	t.Run("Init should fail", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		_, err := createDB("")
		require.Error(t, err)
	})

	t.Run("Close", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db, err := createDB(filePath)
		require.NoError(t, err)

		err = db.DB.Close()
		require.NoError(t, err)
	})

	t.Run("AddUnprocessedTxs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		tx := &core.CardanoTx{OriginChainID: common.ChainIDStrPrime}

		db, err := createDB(filePath)
		require.NoError(t, err)

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

		db, err := createDB(filePath)
		require.NoError(t, err)

		require.NoError(t, db.AddTxs([]*core.ProcessedCardanoTx{}, nil))

		require.NoError(t, db.AddTxs([]*core.ProcessedCardanoTx{tx}, nil))

		resTx, err := db.GetProcessedTx(cCore.DBTxID{ChainID: tx.OriginChainID, DBKey: tx.Hash[:]})

		require.NoError(t, err)
		require.Equal(t, tx, resTx)
	})

	t.Run("GetUnprocessedTxs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db, err := createDB(filePath)
		require.NoError(t, err)

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

		db, err := createDB(filePath)
		require.NoError(t, err)

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

		db, err := createDB(filePath)
		require.NoError(t, err)

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

		db, err := createDB(filePath)
		require.NoError(t, err)

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

		db, err := createDB(filePath)
		require.NoError(t, err)

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

		_, err = db.GetPendingTx(cCore.DBTxID{ChainID: expectedTxs[0].OriginChainID, DBKey: expectedTxs[0].Hash[:]})
		require.Error(t, err)
		require.ErrorContains(t, err, "couldn't get pending tx for entityID")

		err = db.UpdateTxs(&core.CardanoUpdateTxsData{MoveUnprocessedToPending: txs})
		require.NoError(t, err)

		txs, err = db.GetAllUnprocessedTxs(common.ChainIDStrPrime, 0)
		require.NoError(t, err)
		require.Len(t, txs, 0)

		pendingTx, err := db.GetPendingTx(cCore.DBTxID{ChainID: expectedTxs[0].OriginChainID, DBKey: expectedTxs[0].Hash[:]})
		require.NoError(t, err)
		require.NotNil(t, pendingTx)
	})

	t.Run("UpdateTxs - MoveUnprocessedToProcessed", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db, err := createDB(filePath)
		require.NoError(t, err)

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
			pTx, err := db.GetProcessedTx(cCore.DBTxID{ChainID: tx.OriginChainID, DBKey: tx.Hash[:]})
			require.NoError(t, err)
			require.NotNil(t, pTx)
		}
	})

	t.Run("UpdateTxs - MovePendingToUnprocessed", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db, err := createDB(filePath)
		require.NoError(t, err)

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

		iExpectedTxs := make([]cCore.BaseTx, len(expectedTxs))
		for i, tx := range expectedTxs {
			iExpectedTxs[i] = tx
		}

		err = db.UpdateTxs(&core.CardanoUpdateTxsData{MovePendingToUnprocessed: iExpectedTxs})
		require.NoError(t, err)

		_, err = db.GetPendingTx(cCore.DBTxID{ChainID: expectedTxs[0].OriginChainID, DBKey: expectedTxs[0].Hash[:]})
		require.Error(t, err)
		require.ErrorContains(t, err, "couldn't get pending tx for entityID")

		txs, err := db.GetAllUnprocessedTxs(common.ChainIDStrPrime, 0)
		require.NoError(t, err)
		require.Len(t, txs, 1)
	})

	t.Run("UpdateTxs - MovePendingToProcessed", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db, err := createDB(filePath)
		require.NoError(t, err)

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

		iProcessedTxs := make([]cCore.BaseProcessedTx, len(processedTxs))
		for i, tx := range processedTxs {
			iProcessedTxs[i] = tx
		}

		err = db.UpdateTxs(&core.CardanoUpdateTxsData{MovePendingToProcessed: iProcessedTxs})
		require.NoError(t, err)

		_, err = db.GetPendingTx(cCore.DBTxID{ChainID: expectedTxs[0].OriginChainID, DBKey: expectedTxs[0].Hash[:]})
		require.Error(t, err)
		require.ErrorContains(t, err, "couldn't get pending tx for entityID")

		for _, tx := range processedTxs {
			pTx, err := db.GetProcessedTx(cCore.DBTxID{ChainID: tx.OriginChainID, DBKey: tx.Hash[:]})
			require.NoError(t, err)
			require.NotNil(t, pTx)
		}
	})

	t.Run("GetProcessedTx", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db, err := createDB(filePath)
		require.NoError(t, err)

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

		_, err = db.GetProcessedTx(cCore.DBTxID{ChainID: "", DBKey: []byte{}})
		require.Error(t, err)
		require.ErrorContains(t, err, "unsupported chain")

		tx, err := db.GetProcessedTx(cCore.DBTxID{ChainID: common.ChainIDStrPrime, DBKey: []byte{}})
		require.NoError(t, err)
		require.Nil(t, tx)

		tx, err = db.GetProcessedTx(cCore.DBTxID{ChainID: expectedProcessedTxs[0].OriginChainID, DBKey: expectedProcessedTxs[0].Hash[:]})
		require.NoError(t, err)
		require.NotNil(t, tx)
		require.Equal(t, expectedProcessedTxs[0], tx)

		tx, err = db.GetProcessedTx(cCore.DBTxID{ChainID: expectedProcessedTxs[1].OriginChainID, DBKey: expectedProcessedTxs[1].Hash[:]})
		require.NoError(t, err)
		require.NotNil(t, tx)
		require.Equal(t, expectedProcessedTxs[1], tx)
	})

	t.Run("AddExpectedTxs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db, err := createDB(filePath)
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

		db, err := createDB(filePath)
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

		db, err := createDB(filePath)
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

	t.Run("MarkExpectedTxsAsProcessed", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db, err := createDB(filePath)
		require.NoError(t, err)

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

		db, err := createDB(filePath)
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
