package databaseaccess

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	cDatabaseaccess "github.com/Ethernal-Tech/apex-bridge/oracle_common/database_access"
	"github.com/Ethernal-Tech/apex-bridge/oracle_solana/core"
	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/require"
)

func TestBoltDatabase(t *testing.T) {
	chainIDConverter := common.NewTestChainIDConverter()

	appConfig := &cCore.AppConfig{
		SolanaChains: map[string]*cCore.SolanaChainConfig{
			common.ChainIDStrSolana: {},
		},
		ChainIDConverter: chainIDConverter,
	}

	appConfig.FillOut()

	testDir, err := os.MkdirTemp("", "boltdb-sol-test")
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

		typeRegister := common.NewTypeRegister()
		typeRegister.SetType(common.ChainIDStrSolana, reflect.TypeOf(core.SolanaTx{}))

		oracleDB := &BBoltDatabase{}
		oracleDB.Init(boltDB, appConfig, typeRegister)

		return oracleDB, nil
	}

	t.Run("Init", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db, err := createDB(filePath)
		require.NoError(t, err)
		require.NotNil(t, db)
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

	t.Run("AddUnprocessedTxs and GetUnprocessedTxs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db, err := createDB(filePath)
		require.NoError(t, err)

		sig1 := solana.Signature{1, 2, 3}
		tx1 := &core.SolanaTx{
			OriginChainID: common.ChainIDStrSolana,
			Priority:      0,
			SlotNumber:    100,
			TxSignature:   sig1,
		}

		err = db.AddTxs(nil, []*core.SolanaTx{tx1})
		require.NoError(t, err)

		txs, err := db.GetUnprocessedTxs(common.ChainIDStrSolana, 0, 0)
		require.NoError(t, err)
		require.Len(t, txs, 1)
		require.Equal(t, sig1, txs[0].TxSignature)
	})

	t.Run("GetAllUnprocessedTxs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db, err := createDB(filePath)
		require.NoError(t, err)

		sig1 := solana.Signature{1, 2, 3}
		sig2 := solana.Signature{4, 5, 6}

		err = db.AddTxs(nil, []*core.SolanaTx{
			{OriginChainID: common.ChainIDStrSolana, Priority: 0, SlotNumber: 100, TxSignature: sig1},
			{OriginChainID: common.ChainIDStrSolana, Priority: 1, SlotNumber: 200, TxSignature: sig2},
		})
		require.NoError(t, err)

		txs, err := db.GetAllUnprocessedTxs(common.ChainIDStrSolana, 0)
		require.NoError(t, err)
		require.Len(t, txs, 2)
	})

	t.Run("AddProcessedTxs and GetProcessedTx", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db, err := createDB(filePath)
		require.NoError(t, err)

		sig := solana.Signature{10, 20, 30}
		processedTx := &core.ProcessedSolanaTx{
			OriginChainID: common.ChainIDStrSolana,
			TxSignature:   sig,
			SlotNumber:    100,
			IsInvalid:     false,
		}

		err = db.AddTxs([]*core.ProcessedSolanaTx{processedTx}, nil)
		require.NoError(t, err)

		retrieved, err := db.GetProcessedTx(cCore.DBTxID{
			ChainID: common.ChainIDStrSolana,
			DBKey:   sig[:],
		})
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		require.Equal(t, sig, retrieved.TxSignature)
	})

	t.Run("ClearAllTxs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db, err := createDB(filePath)
		require.NoError(t, err)

		err = db.AddTxs(nil, []*core.SolanaTx{
			{OriginChainID: common.ChainIDStrSolana, Priority: 0, SlotNumber: 100, TxSignature: solana.Signature{1}},
		})
		require.NoError(t, err)

		txs, err := db.GetAllUnprocessedTxs(common.ChainIDStrSolana, 0)
		require.NoError(t, err)
		require.Len(t, txs, 1)

		err = db.ClearAllTxs(common.ChainIDStrSolana)
		require.NoError(t, err)

		txs, err = db.GetAllUnprocessedTxs(common.ChainIDStrSolana, 0)
		require.NoError(t, err)
		require.Empty(t, txs)
	})

	t.Run("AddExpectedTxs and GetExpectedTxs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db, err := createDB(filePath)
		require.NoError(t, err)

		expectedTx := &core.BridgeExpectedSolanaTx{
			ChainID:  common.ChainIDStrSolana,
			Hash:     [32]byte{1, 2, 3},
			TTL:      100,
			Priority: 0,
			Metadata: []byte("test_metadata"),
		}

		err = db.AddExpectedTxs([]*core.BridgeExpectedSolanaTx{expectedTx})
		require.NoError(t, err)

		txs, err := db.GetExpectedTxs(common.ChainIDStrSolana, 0, 0)
		require.NoError(t, err)
		require.Len(t, txs, 1)
		require.Equal(t, expectedTx.Hash, txs[0].Hash)
	})

	t.Run("GetAllExpectedTxs", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db, err := createDB(filePath)
		require.NoError(t, err)

		err = db.AddExpectedTxs([]*core.BridgeExpectedSolanaTx{
			{ChainID: common.ChainIDStrSolana, Hash: [32]byte{1, 2, 3}, TTL: 100, Priority: 0},
			{ChainID: common.ChainIDStrSolana, Hash: [32]byte{4, 5, 6}, TTL: 200, Priority: 0},
		})
		require.NoError(t, err)

		txs, err := db.GetAllExpectedTxs(common.ChainIDStrSolana, 0)
		require.NoError(t, err)
		require.Len(t, txs, 2)
	})

	t.Run("UpdateTxs MoveUnprocessedToProcessed", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db, err := createDB(filePath)
		require.NoError(t, err)

		sig := solana.Signature{11, 22, 33}
		tx := &core.SolanaTx{
			OriginChainID: common.ChainIDStrSolana,
			Priority:      0,
			SlotNumber:    100,
			TxSignature:   sig,
		}

		err = db.AddTxs(nil, []*core.SolanaTx{tx})
		require.NoError(t, err)

		updateData := &core.SolanaUpdateTxsData{
			MoveUnprocessedToProcessed: []*core.ProcessedSolanaTx{
				tx.ToProcessedSolanaTx(false),
			},
		}

		err = db.UpdateTxs(updateData, chainIDConverter)
		require.NoError(t, err)

		unprocessedTxs, err := db.GetAllUnprocessedTxs(common.ChainIDStrSolana, 0)
		require.NoError(t, err)
		require.Empty(t, unprocessedTxs)

		processedTx, err := db.GetProcessedTx(cCore.DBTxID{
			ChainID: common.ChainIDStrSolana,
			DBKey:   sig[:],
		})
		require.NoError(t, err)
		require.NotNil(t, processedTx)
	})

	t.Run("UpdateTxs MoveUnprocessedToPending", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db, err := createDB(filePath)
		require.NoError(t, err)

		sig := solana.Signature{44, 55, 66}
		tx := &core.SolanaTx{
			OriginChainID: common.ChainIDStrSolana,
			Priority:      0,
			SlotNumber:    100,
			TxSignature:   sig,
		}

		err = db.AddTxs(nil, []*core.SolanaTx{tx})
		require.NoError(t, err)

		updateData := &core.SolanaUpdateTxsData{
			MoveUnprocessedToPending: []*core.SolanaTx{tx},
		}

		err = db.UpdateTxs(updateData, chainIDConverter)
		require.NoError(t, err)

		unprocessedTxs, err := db.GetAllUnprocessedTxs(common.ChainIDStrSolana, 0)
		require.NoError(t, err)
		require.Empty(t, unprocessedTxs)

		pendingTx, err := db.GetPendingTx(cCore.DBTxID{
			ChainID: common.ChainIDStrSolana,
			DBKey:   sig[:],
		})
		require.NoError(t, err)
		require.NotNil(t, pendingTx)
	})

	t.Run("GetUnprocessedBatchEvents", func(t *testing.T) {
		t.Cleanup(dbCleanup)

		db, err := createDB(filePath)
		require.NoError(t, err)

		events, err := db.GetUnprocessedBatchEvents(common.ChainIDStrSolana)
		require.NoError(t, err)
		require.Empty(t, events)
	})
}
