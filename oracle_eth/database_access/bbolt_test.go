package databaseaccess

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	cDatabaseaccess "github.com/Ethernal-Tech/apex-bridge/oracle_common/database_access"
	"github.com/Ethernal-Tech/apex-bridge/oracle_eth/core"
	"github.com/Ethernal-Tech/ethgo"
	"github.com/stretchr/testify/require"
)

func TestBoltDatabase(t *testing.T) {
	chainIDConverter := common.NewTestChainIDConverter()

	appConfig := &cCore.AppConfig{
		EthChains: map[string]*cCore.EthChainConfig{
			common.ChainIDStrNexus: {},
		},
		ChainIDConverter: chainIDConverter,
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

		typeRegister := cCore.NewTypeRegisterWithChains(appConfig, nil, reflect.TypeOf(core.EthTx{}))

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

		tx := &core.EthTx{OriginChainID: common.ChainIDStrNexus, Hash: ethgo.HexToHash("ff99aa")}

		db, err := createDB(filePath)
		require.NoError(t, err)

		require.NoError(t, db.AddTxs(nil, []*core.EthTx{tx}))

		listTxs, err := db.GetUnprocessedTxs(tx.OriginChainID, 0, 2)

		require.NoError(t, err)
		require.Len(t, listTxs, 1)
		require.Equal(t, tx, listTxs[0])
	})

	t.Run("AddProcessedTxs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		tx := &core.ProcessedEthTx{OriginChainID: common.ChainIDStrNexus, Hash: ethgo.HexToHash("ff99aa")}

		db, err := createDB(filePath)
		require.NoError(t, err)

		require.NoError(t, db.AddTxs([]*core.ProcessedEthTx{}, nil))

		require.NoError(t, db.AddTxs([]*core.ProcessedEthTx{tx}, nil))

		resTx, err := db.GetProcessedTx(cCore.DBTxID{ChainID: tx.OriginChainID, DBKey: tx.Hash[:]})

		require.NoError(t, err)
		require.Equal(t, tx, resTx)
	})

	t.Run("GetUnprocessedTxs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db, err := createDB(filePath)
		require.NoError(t, err)

		expectedTxs := []*core.EthTx{
			{OriginChainID: common.ChainIDStrNexus, Priority: 0},
		}

		err = db.AddTxs(nil, expectedTxs)
		require.NoError(t, err)

		txs, err := db.GetUnprocessedTxs(common.ChainIDStrNexus, 0, 0)
		require.NoError(t, err)
		require.NotNil(t, txs)
		require.Len(t, txs, 1)
		require.Equal(t, expectedTxs[0], txs[0])
	})

	t.Run("GetAllUnprocessedTxs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db, err := createDB(filePath)
		require.NoError(t, err)

		expectedTxs := []*core.EthTx{
			{OriginChainID: common.ChainIDStrNexus},
		}

		err = db.AddTxs(nil, expectedTxs)
		require.NoError(t, err)

		txs, err := db.GetAllUnprocessedTxs(common.ChainIDStrNexus, 0)
		require.NoError(t, err)
		require.NotNil(t, txs)
		require.Len(t, txs, 1)
		require.Equal(t, expectedTxs[0], txs[0])
	})

	t.Run("ClearAllTxs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db, err := createDB(filePath)
		require.NoError(t, err)

		unprocessedTxs := []*core.EthTx{
			{OriginChainID: common.ChainIDStrNexus},
		}

		expectedTxs := []*core.BridgeExpectedEthTx{
			{ChainID: common.ChainIDStrNexus},
		}

		err = db.AddTxs(nil, unprocessedTxs)
		require.NoError(t, err)

		err = db.AddExpectedTxs(expectedTxs)
		require.NoError(t, err)

		err = db.ClearAllTxs(common.ChainIDStrNexus)
		require.NoError(t, err)

		txs, err := db.GetAllUnprocessedTxs(common.ChainIDStrNexus, 0)
		require.NoError(t, err)
		require.Nil(t, txs)

		exTxs, err := db.GetAllExpectedTxs(common.ChainIDStrNexus, 0)
		require.NoError(t, err)
		require.Nil(t, exTxs)
	})

	t.Run("UpdateTxs - UpdateUnprocessed", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db, err := createDB(filePath)
		require.NoError(t, err)

		expectedTxs := []*core.EthTx{
			{OriginChainID: common.ChainIDStrNexus},
		}

		err = db.AddTxs(nil, expectedTxs)
		require.NoError(t, err)

		txs, err := db.GetAllUnprocessedTxs(common.ChainIDStrNexus, 0)
		require.NoError(t, err)
		require.NotNil(t, txs)

		for _, tx := range txs {
			tx.SubmitTryCount++
		}

		err = db.UpdateTxs(&core.EthUpdateTxsData{UpdateUnprocessed: txs}, chainIDConverter)
		require.NoError(t, err)

		txs, err = db.GetAllUnprocessedTxs(common.ChainIDStrNexus, 0)
		require.NoError(t, err)
		require.Len(t, txs, 1)
		require.Equal(t, uint32(1), txs[0].SubmitTryCount)
	})

	t.Run("GetPendingTxs and UpdateTxs - MoveUnprocessedToPending", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db, err := createDB(filePath)
		require.NoError(t, err)

		expectedTxs := []*core.EthTx{
			{
				Priority:      1,
				OriginChainID: common.ChainIDStrNexus,
				BlockNumber:   1,
				Hash:          ethgo.Hash{1, 2},
			},
		}

		err = db.AddTxs(nil, expectedTxs)
		require.NoError(t, err)

		txs, err := db.GetAllUnprocessedTxs(common.ChainIDStrNexus, 0)
		require.NoError(t, err)
		require.NotNil(t, txs)

		_, err = db.GetPendingTx(cCore.DBTxID{ChainID: expectedTxs[0].OriginChainID, DBKey: expectedTxs[0].Hash[:]})
		require.Error(t, err)
		require.ErrorContains(t, err, "couldn't get pending tx for entityID")

		err = db.UpdateTxs(&core.EthUpdateTxsData{MoveUnprocessedToPending: txs}, chainIDConverter)
		require.NoError(t, err)

		txs, err = db.GetAllUnprocessedTxs(common.ChainIDStrNexus, 0)
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

		expectedTxs := []*core.EthTx{
			{OriginChainID: common.ChainIDStrNexus},
		}

		err = db.AddTxs(nil, expectedTxs)
		require.NoError(t, err)

		var expectedProcessedTxs []*core.ProcessedEthTx

		txs, err := db.GetAllUnprocessedTxs(common.ChainIDStrNexus, 0)
		require.NoError(t, err)
		require.NotNil(t, txs)

		for _, tx := range txs {
			expectedProcessedTxs = append(expectedProcessedTxs, tx.ToProcessedEthTx(false))
		}

		err = db.UpdateTxs(&core.EthUpdateTxsData{MoveUnprocessedToProcessed: []*core.ProcessedEthTx{}}, chainIDConverter)
		require.NoError(t, err)

		err = db.UpdateTxs(&core.EthUpdateTxsData{MoveUnprocessedToProcessed: expectedProcessedTxs}, chainIDConverter)
		require.NoError(t, err)

		txs, err = db.GetAllUnprocessedTxs(common.ChainIDStrNexus, 0)
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

		expectedTxs := []*core.EthTx{
			{
				Priority:      1,
				OriginChainID: common.ChainIDStrNexus,
				BlockNumber:   1,
				Hash:          ethgo.Hash{1, 2},
			},
		}

		err = db.AddTxs(nil, expectedTxs)
		require.NoError(t, err)

		err = db.UpdateTxs(&core.EthUpdateTxsData{MoveUnprocessedToPending: expectedTxs}, chainIDConverter)
		require.NoError(t, err)

		iExpectedTxs := make([]cCore.BaseTx, len(expectedTxs))
		for i, tx := range expectedTxs {
			iExpectedTxs[i] = tx
		}

		err = db.UpdateTxs(&core.EthUpdateTxsData{MovePendingToUnprocessed: iExpectedTxs}, chainIDConverter)
		require.NoError(t, err)

		_, err = db.GetPendingTx(cCore.DBTxID{ChainID: expectedTxs[0].OriginChainID, DBKey: expectedTxs[0].Hash[:]})
		require.Error(t, err)
		require.ErrorContains(t, err, "couldn't get pending tx for entityID")

		txs, err := db.GetAllUnprocessedTxs(common.ChainIDStrNexus, 0)
		require.NoError(t, err)
		require.Len(t, txs, 1)
	})

	t.Run("UpdateTxs - MovePendingToProcessed", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db, err := createDB(filePath)
		require.NoError(t, err)

		expectedTxs := []*core.EthTx{
			{
				Priority:      1,
				OriginChainID: common.ChainIDStrNexus,
				BlockNumber:   1,
				Hash:          ethgo.Hash{1, 2},
			},
		}

		err = db.AddTxs(nil, expectedTxs)
		require.NoError(t, err)

		err = db.UpdateTxs(&core.EthUpdateTxsData{MoveUnprocessedToPending: expectedTxs}, chainIDConverter)
		require.NoError(t, err)

		processedTxs := make([]*core.ProcessedEthTx, len(expectedTxs))
		for i, tx := range expectedTxs {
			processedTxs[i] = tx.ToProcessedEthTx(false)
		}

		iProcessedTxs := make([]cCore.BaseProcessedTx, len(processedTxs))
		for i, tx := range processedTxs {
			iProcessedTxs[i] = tx
		}

		err = db.UpdateTxs(&core.EthUpdateTxsData{MovePendingToProcessed: iProcessedTxs}, chainIDConverter)
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

		expectedTxs := []*core.EthTx{
			{OriginChainID: common.ChainIDStrNexus},
		}

		err = db.AddTxs(nil, expectedTxs)
		require.NoError(t, err)

		var expectedProcessedTxs []*core.ProcessedEthTx

		txs, err := db.GetAllUnprocessedTxs(common.ChainIDStrNexus, 0)
		require.NoError(t, err)
		require.NotNil(t, txs)

		for _, tx := range txs {
			expectedProcessedTxs = append(expectedProcessedTxs, tx.ToProcessedEthTx(false))
		}

		err = db.UpdateTxs(&core.EthUpdateTxsData{MoveUnprocessedToProcessed: expectedProcessedTxs}, chainIDConverter)
		require.NoError(t, err)

		_, err = db.GetProcessedTx(cCore.DBTxID{ChainID: "", DBKey: []byte{}})
		require.Error(t, err)
		require.ErrorContains(t, err, "unsupported chain")

		tx, err := db.GetProcessedTx(cCore.DBTxID{ChainID: common.ChainIDStrNexus, DBKey: []byte{}})
		require.NoError(t, err)
		require.Nil(t, tx)

		tx, err = db.GetProcessedTx(cCore.DBTxID{ChainID: expectedProcessedTxs[0].OriginChainID, DBKey: expectedProcessedTxs[0].Hash[:]})
		require.NoError(t, err)
		require.NotNil(t, tx)
		require.Equal(t, expectedProcessedTxs[0], tx)
	})

	t.Run("AddExpectedTxs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db, err := createDB(filePath)
		require.NoError(t, err)

		err = db.AddExpectedTxs(nil)
		require.NoError(t, err)

		err = db.AddExpectedTxs([]*core.BridgeExpectedEthTx{})
		require.NoError(t, err)

		err = db.AddExpectedTxs([]*core.BridgeExpectedEthTx{
			{ChainID: common.ChainIDStrNexus},
		})
		require.NoError(t, err)
	})

	t.Run("GetExpectedTxs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db, err := createDB(filePath)
		require.NoError(t, err)

		expectedTxs := []*core.BridgeExpectedEthTx{
			{ChainID: common.ChainIDStrNexus, Priority: 0},
		}

		err = db.AddExpectedTxs(expectedTxs)
		require.NoError(t, err)

		txs, err := db.GetExpectedTxs(common.ChainIDStrNexus, 0, 0)
		require.NoError(t, err)
		require.NotNil(t, txs)
		require.Len(t, txs, 1)
		require.Equal(t, expectedTxs[0], txs[0])
	})

	t.Run("GetAllExpectedTxs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db, err := createDB(filePath)
		require.NoError(t, err)

		expectedTxs := []*core.BridgeExpectedEthTx{
			{ChainID: common.ChainIDStrNexus},
		}

		err = db.AddExpectedTxs(expectedTxs)
		require.NoError(t, err)

		txs, err := db.GetAllExpectedTxs(common.ChainIDStrNexus, 0)
		require.NoError(t, err)
		require.NotNil(t, txs)
		require.Len(t, txs, 1)
		require.Equal(t, expectedTxs[0], txs[0])
	})

	//nolint:dupl
	t.Run("MarkAndMoveExpectedTxsAsProcessed", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db, err := createDB(filePath)
		require.NoError(t, err)

		expectedTxs := []*core.BridgeExpectedEthTx{
			{ChainID: common.ChainIDStrNexus},
		}

		err = db.AddExpectedTxs(expectedTxs)
		require.NoError(t, err)

		err = db.UpdateTxs(&core.EthUpdateTxsData{ExpectedProcessed: []*core.BridgeExpectedEthTx{}}, chainIDConverter)
		require.NoError(t, err)

		txs, err := db.GetAllExpectedTxs(common.ChainIDStrNexus, 0)
		require.NoError(t, err)
		require.NotNil(t, txs)

		err = db.UpdateTxs(&core.EthUpdateTxsData{ExpectedProcessed: txs}, chainIDConverter)
		require.NoError(t, err)

		txs, err = db.GetAllExpectedTxs(common.ChainIDStrNexus, 0)
		require.NoError(t, err)
		require.Nil(t, txs)
	})

	//nolint:dupl
	t.Run("MarkAndMoveExpectedTxsAsInvalid", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db, err := createDB(filePath)
		require.NoError(t, err)

		expectedTxs := []*core.BridgeExpectedEthTx{
			{ChainID: common.ChainIDStrNexus},
		}

		err = db.AddExpectedTxs(expectedTxs)
		require.NoError(t, err)

		err = db.UpdateTxs(&core.EthUpdateTxsData{ExpectedInvalid: []*core.BridgeExpectedEthTx{}}, chainIDConverter)
		require.NoError(t, err)

		txs, err := db.GetAllExpectedTxs(common.ChainIDStrNexus, 0)
		require.NoError(t, err)
		require.NotNil(t, txs)

		err = db.UpdateTxs(&core.EthUpdateTxsData{ExpectedInvalid: txs}, chainIDConverter)
		require.NoError(t, err)

		txs, err = db.GetAllExpectedTxs(common.ChainIDStrNexus, 0)
		require.NoError(t, err)
		require.Nil(t, txs)
	})

	t.Run("MoveProcessedExpectedTxs - no processed expected tx", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db, err := createDB(filePath)
		require.NoError(t, err)

		expectedTxs := []*core.BridgeExpectedEthTx{
			{
				ChainID: common.ChainIDStrNexus,
			},
		}

		err = db.AddExpectedTxs(expectedTxs)
		require.NoError(t, err)

		txs, err := db.GetAllExpectedTxs(common.ChainIDStrNexus, 0)
		require.NoError(t, err)
		require.NotNil(t, txs)

		err = db.MoveProcessedExpectedTxs(common.ChainIDStrNexus)
		require.NoError(t, err)

		txs, err = db.GetAllExpectedTxs(common.ChainIDStrNexus, 0)
		require.NoError(t, err)
		require.NotNil(t, txs)
		require.Len(t, txs, 1)
	})

	t.Run("MoveProcessedExpectedTxs - invalid and processed txs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db, err := createDB(filePath)
		require.NoError(t, err)

		expectedTxs := []*core.BridgeExpectedEthTx{
			{
				ChainID:   common.ChainIDStrNexus,
				IsInvalid: true,
			},
			{
				ChainID:     common.ChainIDStrNexus,
				IsProcessed: true,
			},
		}

		err = db.AddExpectedTxs(expectedTxs)
		require.NoError(t, err)

		txs, err := db.GetAllExpectedTxs(common.ChainIDStrNexus, 0)
		require.NoError(t, err)
		require.NotNil(t, txs)

		err = db.MoveProcessedExpectedTxs(common.ChainIDStrNexus)
		require.NoError(t, err)

		txs, err = db.GetAllExpectedTxs(common.ChainIDStrNexus, 0)
		require.NoError(t, err)
		require.Nil(t, txs)
	})

	t.Run("MoveProcessedExpectedTxs - all txs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db, err := createDB(filePath)
		require.NoError(t, err)

		expectedTxs := []*core.BridgeExpectedEthTx{
			{
				ChainID: common.ChainIDStrNexus,
			},
			{
				ChainID:   common.ChainIDStrNexus,
				IsInvalid: true,
			},
			{
				ChainID:     common.ChainIDStrNexus,
				IsProcessed: true,
			},
		}

		err = db.AddExpectedTxs(expectedTxs)
		require.NoError(t, err)

		txs, err := db.GetAllExpectedTxs(common.ChainIDStrNexus, 0)
		require.NoError(t, err)
		require.NotNil(t, txs)

		err = db.MoveProcessedExpectedTxs(common.ChainIDStrNexus)
		require.NoError(t, err)

		txs, err = db.GetAllExpectedTxs(common.ChainIDStrNexus, 0)
		require.NoError(t, err)
		require.NotNil(t, txs)
		require.Len(t, txs, 1)
	})
}
