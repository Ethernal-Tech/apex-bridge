package chain

import (
	"context"
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

	defer func() {
		os.RemoveAll(testDir)
		os.Remove(testDir)
	}()

	settings := core.AppSettings{
		DbsPath: testDir,
		Logger: logger.LoggerConfig{
			LogFilePath: testDir,
		},
	}

	foldersCleanup := func() {
		os.RemoveAll(testDir)
	}

	chainConfig := &core.CardanoChainConfig{
		ChainId:                "prime",
		NetworkAddress:         "backbone.cardano-mainnet.iohk.io:3001",
		NetworkMagic:           764824073,
		StartBlockHash:         "df12b7a87cc041f238f400e3a410d1edb2f4a6fbcbedafff8fd9d507ef4947d7",
		StartSlot:              76593517,
		StartBlockNumber:       8000030,
		ConfirmationBlockCount: 10,
		OtherAddressesOfInterest: []string{
			"addr1v9kganeshgdqyhwnyn9stxxgl7r4y2ejfyqjn88n7ncapvs4sugsd",
		},
	}

	txsProcessorMock := &core.CardanoTxsProcessorMock{}
	txsProcessorMock.On("NewUnprocessedTxs", mock.Anything, mock.Anything).Return(error(nil))

	initDb := func(t *testing.T) indexer.Database {
		t.Helper()

		require.NoError(t, common.CreateDirectoryIfNotExists(settings.DbsPath, 0750))

		indexerDb, err := indexerDb.NewDatabaseInit("", path.Join(settings.DbsPath, chainConfig.ChainId+".db"))
		require.NoError(t, err)

		return indexerDb
	}

	t.Run("check ErrorCh", func(t *testing.T) {
		t.Cleanup(foldersCleanup)

		bridgeDataFetcher := &core.BridgeDataFetcherMock{}
		bridgeDataFetcher.On("FetchLatestBlockPoint", mock.Anything).Return(&indexer.BlockPoint{}, error(nil))
		db := &core.CardanoTxsProcessorDbMock{}
		db.On("ClearUnprocessedTxs", mock.Anything).Return(error(nil))
		db.On("ClearExpectedTxs", mock.Anything).Return(error(nil))

		indexerDb := initDb(t)

		chainObserver, err := NewCardanoChainObserver(context.Background(), chainConfig, txsProcessorMock, db, indexerDb, bridgeDataFetcher, hclog.NewNullLogger(), false)
		require.NoError(t, err)
		require.NotNil(t, chainObserver)

		errChan := chainObserver.ErrorCh()
		require.NotNil(t, errChan)
	})

	t.Run("check GetConfig", func(t *testing.T) {
		t.Cleanup(foldersCleanup)

		bridgeDataFetcher := &core.BridgeDataFetcherMock{}
		bridgeDataFetcher.On("FetchLatestBlockPoint", mock.Anything).Return(&indexer.BlockPoint{}, error(nil))
		db := &core.CardanoTxsProcessorDbMock{}
		db.On("ClearUnprocessedTxs", mock.Anything).Return(error(nil))
		db.On("ClearExpectedTxs", mock.Anything).Return(error(nil))

		indexerDb := initDb(t)

		chainObserver, err := NewCardanoChainObserver(context.Background(), chainConfig, txsProcessorMock, db, indexerDb, bridgeDataFetcher, hclog.NewNullLogger(), false)
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
		db := &core.CardanoTxsProcessorDbMock{}
		db.On("ClearUnprocessedTxs", mock.Anything).Return(error(nil))
		db.On("ClearExpectedTxs", mock.Anything).Return(error(nil))

		indexerDb := initDb(t)

		ctx, cancelFunc := context.WithCancel(context.Background())
		defer cancelFunc()

		chainObserver, err := NewCardanoChainObserver(ctx, chainConfig, txsProcessorMock, db, indexerDb, bridgeDataFetcher, hclog.NewNullLogger(), false)
		require.NoError(t, err)
		require.NotNil(t, chainObserver)

		err = chainObserver.Start()
		require.NoError(t, err)
	})

	t.Run("check newConfirmedTxs called", func(t *testing.T) {
		t.Cleanup(foldersCleanup)

		bridgeDataFetcher := &core.BridgeDataFetcherMock{}
		bridgeDataFetcher.On("FetchLatestBlockPoint", mock.Anything).Return(&indexer.BlockPoint{}, error(nil))
		db := &core.CardanoTxsProcessorDbMock{}
		db.On("ClearUnprocessedTxs", mock.Anything).Return(error(nil))
		db.On("ClearExpectedTxs", mock.Anything).Return(error(nil))

		indexerDb := initDb(t)

		ctx, cancelFunc := context.WithCancel(context.Background())
		defer cancelFunc()

		chainObserver, err := NewCardanoChainObserver(ctx, chainConfig, txsProcessorMock, db, indexerDb, bridgeDataFetcher, hclog.NewNullLogger(), false)
		require.NoError(t, err)
		require.NotNil(t, chainObserver)

		doneCh := make(chan bool, 1)

		txsProcessorMock.NewUnprocessedTxsFn = func(originChainId string, txs []*indexer.Tx) error {
			t.Helper()
			close(doneCh)
			return nil
		}

		err = chainObserver.Start()
		require.NoError(t, err)

		timer := time.NewTimer(60 * time.Second)
		defer timer.Stop()

		select {
		case <-timer.C:
			t.Fail()
		case <-doneCh:
		}
	})
}
