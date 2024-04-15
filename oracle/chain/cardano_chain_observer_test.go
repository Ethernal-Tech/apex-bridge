package chain

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	indexerDb "github.com/Ethernal-Tech/cardano-infrastructure/indexer/db"
	"github.com/hashicorp/go-hclog"
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
		DbsPath:  testDir,
		LogsPath: testDir,
		LogLevel: 2,
	}

	foldersCleanup := func() {
		common.RemoveDirOrFilePathIfExists(settings.DbsPath)
		common.RemoveDirOrFilePathIfExists(settings.LogsPath)
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
		bridgeDataFetcher.On("FetchLatestBlockPoint").Return(&indexer.BlockPoint{}, nil)
		db := &core.CardanoTxsProcessorDbMock{}
		db.On("ClearUnprocessedTxs").Return(nil)
		db.On("ClearExpectedTxs").Return(nil)

		indexerDb := initDb(t)

		chainObserver, err := NewCardanoChainObserver(chainConfig, &core.CardanoTxsProcessorMock{}, db, indexerDb, bridgeDataFetcher, hclog.NewNullLogger())
		require.NoError(t, err)
		require.NotNil(t, chainObserver)

		errChan := chainObserver.ErrorCh()
		require.NotNil(t, errChan)
	})

	t.Run("check GetConfig", func(t *testing.T) {
		t.Cleanup(foldersCleanup)

		bridgeDataFetcher := &core.BridgeDataFetcherMock{}
		bridgeDataFetcher.On("FetchLatestBlockPoint").Return(&indexer.BlockPoint{}, nil)
		db := &core.CardanoTxsProcessorDbMock{}
		db.On("ClearUnprocessedTxs").Return(nil)
		db.On("ClearExpectedTxs").Return(nil)

		indexerDb := initDb(t)

		chainObserver, err := NewCardanoChainObserver(chainConfig, &core.CardanoTxsProcessorMock{}, db, indexerDb, bridgeDataFetcher, hclog.NewNullLogger())
		require.NoError(t, err)
		require.NotNil(t, chainObserver)

		config := chainObserver.GetConfig()
		require.NotNil(t, config)
		require.Equal(t, chainConfig, config)
	})

	t.Run("check start stop", func(t *testing.T) {
		t.Cleanup(foldersCleanup)

		bridgeDataFetcher := &core.BridgeDataFetcherMock{}
		bridgeDataFetcher.On("FetchLatestBlockPoint").Return(&indexer.BlockPoint{}, nil)
		db := &core.CardanoTxsProcessorDbMock{}
		db.On("ClearUnprocessedTxs").Return(nil)
		db.On("ClearExpectedTxs").Return(nil)

		indexerDb := initDb(t)

		chainObserver, err := NewCardanoChainObserver(chainConfig, &core.CardanoTxsProcessorMock{}, db, indexerDb, bridgeDataFetcher, hclog.NewNullLogger())
		require.NoError(t, err)
		require.NotNil(t, chainObserver)

		err = chainObserver.Start()
		require.NoError(t, err)

		err = chainObserver.Stop()
		require.NoError(t, err)
	})

	t.Run("check newConfirmedTxs called", func(t *testing.T) {
		t.Cleanup(foldersCleanup)

		txsProcessor := &core.CardanoTxsProcessorMock{}
		txsProcessor.On("NewUnprocessedTxs").Return()

		bridgeDataFetcher := &core.BridgeDataFetcherMock{}
		bridgeDataFetcher.On("FetchLatestBlockPoint").Return(&indexer.BlockPoint{}, nil)
		db := &core.CardanoTxsProcessorDbMock{}
		db.On("ClearUnprocessedTxs").Return(nil)
		db.On("ClearExpectedTxs").Return(nil)

		indexerDb := initDb(t)

		chainObserver, err := NewCardanoChainObserver(chainConfig, txsProcessor, db, indexerDb, bridgeDataFetcher, hclog.NewNullLogger())
		require.NoError(t, err)
		require.NotNil(t, chainObserver)

		doneCh := make(chan bool, 1)

		txsProcessor.NewUnprocessedTxsFn = func(originChainId string, txs []*indexer.Tx) error {
			t.Helper()
			close(doneCh)
			return nil
		}

		err = chainObserver.Start()
		require.NoError(t, err)

		defer chainObserver.Stop()

		timer := time.NewTimer(60 * time.Second)
		defer timer.Stop()

		select {
		case <-timer.C:
			t.Fail()
		case <-doneCh:
		}
	})
}
