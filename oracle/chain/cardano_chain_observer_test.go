package chain

import (
	"context"
	"errors"
	"os"
	"path"
	"testing"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	indexerDb "github.com/Ethernal-Tech/cardano-infrastructure/indexer/db"
	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCardanoChainObserver(t *testing.T) {
	testDir, err := os.MkdirTemp("", "boltdb-test")
	require.NoError(t, err)

	foldersCleanup := func() {
		os.RemoveAll(testDir)
		os.Remove(testDir)
	}

	defer foldersCleanup()

	settings := core.AppSettings{
		DbsPath: testDir,
		Logger: logger.LoggerConfig{
			LogFilePath: testDir,
		},
	}

	chainConfig := &core.CardanoChainConfig{
		ChainID:                common.ChainIDStrPrime,
		NetworkAddress:         "backbone.cardano-mainnet.iohk.io:3001",
		NetworkMagic:           764824073,
		StartBlockHash:         "335ac2d90bc37906c1264cfdc5769a652293cf64fa42b0c74d323473938b8ff1",
		StartSlot:              127933773,
		ConfirmationBlockCount: 10,
		OtherAddressesOfInterest: []string{
			"addr1v9kganeshgdqyhwnyn9stxxgl7r4y2ejfyqjn88n7ncapvs4sugsd",
		},
	}

	txsProcessorMock := &core.CardanoTxsProcessorMock{}
	txsProcessorMock.On("NewUnprocessedTxs", mock.Anything, mock.Anything).Return(error(nil))

	initDB := func(t *testing.T) indexer.Database {
		t.Helper()

		require.NoError(t, common.CreateDirectoryIfNotExists(settings.DbsPath, 0750))

		indexerDB, err := indexerDb.NewDatabaseInit("", path.Join(settings.DbsPath, chainConfig.ChainID+".db"))
		require.NoError(t, err)

		return indexerDB
	}

	t.Run("check ErrorCh", func(t *testing.T) {
		t.Cleanup(foldersCleanup)

		bridgeDataFetcher := &core.BridgeDataFetcherMock{}
		bridgeDataFetcher.On("FetchLatestBlockPoint", mock.Anything).Return(&indexer.BlockPoint{}, error(nil))

		db := &core.CardanoTxsProcessorDBMock{}
		db.On("ClearUnprocessedTxs", mock.Anything).Return(error(nil))
		db.On("ClearExpectedTxs", mock.Anything).Return(error(nil))

		indexerDB := initDB(t)

		chainObserver, err := NewCardanoChainObserver(context.Background(), chainConfig, txsProcessorMock, db, indexerDB, bridgeDataFetcher, hclog.NewNullLogger())
		require.NoError(t, err)
		require.NotNil(t, chainObserver)

		defer chainObserver.Dispose()

		errChan := chainObserver.ErrorCh()
		require.NotNil(t, errChan)
	})

	t.Run("check GetConfig", func(t *testing.T) {
		t.Cleanup(foldersCleanup)

		bridgeDataFetcher := &core.BridgeDataFetcherMock{}
		bridgeDataFetcher.On("FetchLatestBlockPoint", mock.Anything).Return(&indexer.BlockPoint{}, error(nil))

		db := &core.CardanoTxsProcessorDBMock{}
		db.On("ClearUnprocessedTxs", mock.Anything).Return(error(nil))
		db.On("ClearExpectedTxs", mock.Anything).Return(error(nil))

		indexerDB := initDB(t)

		chainObserver, err := NewCardanoChainObserver(context.Background(), chainConfig, txsProcessorMock, db, indexerDB, bridgeDataFetcher, hclog.NewNullLogger())
		require.NoError(t, err)
		require.NotNil(t, chainObserver)

		config := chainObserver.GetConfig()
		require.NotNil(t, config)
		require.Equal(t, chainConfig, config)
	})

	t.Run("check start stop", func(t *testing.T) {
		t.Cleanup(foldersCleanup)

		bridgeDataFetcher := &core.BridgeDataFetcherMock{}
		bridgeDataFetcher.On("FetchLatestBlockPoint", mock.Anything).Return(&indexer.BlockPoint{}, error(nil))

		db := &core.CardanoTxsProcessorDBMock{}
		db.On("ClearUnprocessedTxs", mock.Anything).Return(error(nil))
		db.On("ClearExpectedTxs", mock.Anything).Return(error(nil))

		indexerDB := initDB(t)

		ctx, cancelFunc := context.WithCancel(context.Background())
		defer cancelFunc()

		chainObserver, err := NewCardanoChainObserver(ctx, chainConfig, txsProcessorMock, db, indexerDB, bridgeDataFetcher, hclog.NewNullLogger())
		require.NoError(t, err)
		require.NotNil(t, chainObserver)

		err = chainObserver.Start()
		require.NoError(t, err)
	})

	t.Run("check newConfirmedTxs called", func(t *testing.T) {
		t.Cleanup(foldersCleanup)

		bridgeDataFetcher := &core.BridgeDataFetcherMock{}
		bridgeDataFetcher.On("FetchLatestBlockPoint", mock.Anything).Return(&indexer.BlockPoint{}, error(nil))

		db := &core.CardanoTxsProcessorDBMock{}
		db.On("ClearUnprocessedTxs", mock.Anything).Return(error(nil))
		db.On("ClearExpectedTxs", mock.Anything).Return(error(nil))

		indexerDB := initDB(t)

		ctx, cancelFunc := context.WithCancel(context.Background())
		defer cancelFunc()

		chainObserver, err := NewCardanoChainObserver(ctx, chainConfig, txsProcessorMock, db, indexerDB, bridgeDataFetcher, hclog.NewNullLogger())
		require.NoError(t, err)
		require.NotNil(t, chainObserver)

		doneCh := make(chan bool, 1)
		closed := false

		txsProcessorMock.NewUnprocessedTxsFn = func(originChainId string, txs []*indexer.Tx) error {
			t.Helper()

			if !closed {
				close(doneCh)

				closed = true
			}

			return nil
		}

		err = chainObserver.Start()
		require.NoError(t, err)

		timer := time.NewTimer(60 * time.Second)
		defer timer.Stop()

		select {
		case <-timer.C:
			t.Fatal("timeout")
		case <-doneCh:
		}
	})
}

func Test_initOracleState(t *testing.T) {
	dbMock := &indexer.DatabaseMock{}
	dbWriterMock := &indexer.DBTransactionWriterMock{}
	utxos := []*indexer.TxInputOutput{
		{
			Input:  indexer.TxInput{Hash: indexer.NewHashFromHexString("0x1")},
			Output: indexer.TxOutput{Amount: 100},
		},
	}
	blockSlot := uint64(100)
	oracleDbMock := &core.CardanoTxsProcessorDBMock{}
	blockHash := "0xA1"
	chainID := "prime"

	t.Run("error", func(t *testing.T) {
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{}, errors.New("test error")).Once()

		require.Error(t, initOracleState(dbMock, oracleDbMock, blockHash, blockSlot, utxos, chainID, hclog.NewNullLogger()))
	})

	t.Run("empty hash", func(t *testing.T) {
		require.NoError(t, initOracleState(dbMock, oracleDbMock, "", blockSlot, utxos, chainID, hclog.NewNullLogger()))
	})

	t.Run("updated in db smaller slot", func(t *testing.T) {
		dbMock.On("GetLatestBlockPoint").Return((*indexer.BlockPoint)(nil), error(nil)).Once()
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{
			BlockSlot: uint64(50),
		}, error(nil)).Once()

		dbMock.On("OpenTx").Return(dbWriterMock).Twice()
		dbWriterMock.On("DeleteAllTxOutputsPhysically").Return(dbMock).Twice()
		dbWriterMock.On("AddTxOutputs", utxos).Return(dbMock).Twice()
		dbWriterMock.On("SetLatestBlockPoint", &indexer.BlockPoint{
			BlockSlot: blockSlot,
			BlockHash: indexer.NewHashFromHexString(blockHash),
		}).Return(dbMock).Twice()
		dbWriterMock.On("Execute").Return(error(nil)).Twice()
		oracleDbMock.On("ClearUnprocessedTxs", chainID).Return(error(nil)).Twice()
		oracleDbMock.On("ClearExpectedTxs", chainID).Return(error(nil)).Twice()

		// empty
		require.NoError(t, initOracleState(dbMock, oracleDbMock, blockHash, blockSlot, utxos, chainID, hclog.NewNullLogger()))
		// smaller
		require.NoError(t, initOracleState(dbMock, oracleDbMock, blockHash, blockSlot, utxos, chainID, hclog.NewNullLogger()))
	})

	t.Run("skipping - has equal slot", func(t *testing.T) {
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{
			BlockSlot: uint64(100),
		}, error(nil)).Once()

		require.NoError(t, initOracleState(dbMock, oracleDbMock, blockHash, blockSlot, utxos, chainID, hclog.NewNullLogger()))
	})

	t.Run("skipping - has greater slot", func(t *testing.T) {
		dbMock.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{
			BlockSlot: uint64(150),
		}, error(nil)).Once()

		require.NoError(t, initOracleState(dbMock, oracleDbMock, blockHash, blockSlot, utxos, chainID, hclog.NewNullLogger()))
	})

	t.Run("ClearUnprocessedTxs error", func(t *testing.T) {
		dbMock.On("GetLatestBlockPoint").Return((*indexer.BlockPoint)(nil), error(nil)).Once()
		oracleDbMock.On("ClearUnprocessedTxs", chainID).Return(errors.New("test error")).Once()

		require.Error(t, initOracleState(dbMock, oracleDbMock, blockHash, blockSlot, utxos, chainID, hclog.NewNullLogger()))
	})

	t.Run("ClearExpectedTxs error", func(t *testing.T) {
		dbMock.On("GetLatestBlockPoint").Return((*indexer.BlockPoint)(nil), error(nil)).Once()
		oracleDbMock.On("ClearUnprocessedTxs", chainID).Return(error(nil)).Once()
		oracleDbMock.On("ClearExpectedTxs", chainID).Return(errors.New("test error")).Once()

		require.Error(t, initOracleState(dbMock, oracleDbMock, blockHash, blockSlot, utxos, chainID, hclog.NewNullLogger()))
	})
}
