package chain

import (
	"context"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/eth"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_eth/core"
	eventTrackerStore "github.com/Ethernal-Tech/blockchain-event-tracker/store"
	eventTracker "github.com/Ethernal-Tech/blockchain-event-tracker/tracker"

	"github.com/Ethernal-Tech/ethgo"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const ethNodeURL = "https://rpc.nexus.testnet.apexfusion.org"

func TestEthChainObserver(t *testing.T) {
	testDir, err := os.MkdirTemp("", "boltdb-test")
	require.NoError(t, err)

	foldersCleanup := func() {
		os.RemoveAll(testDir)
		os.Remove(testDir)
	}

	defer foldersCleanup()

	ctx := context.Background()
	logger := hclog.NewNullLogger()

	config := &oCore.EthChainConfig{
		StartBlockNumber: uint64(1),
		ChainID:          "nexus",
		NodeURL:          ethNodeURL,
	}

	t.Run("chain observer - initOracleState error", func(t *testing.T) {
		txsReceiverMock := &core.EthTxsReceiverMock{}
		oracleDB := &core.EthTxsProcessorDBMock{}
		indexerDB := &core.EventStoreMock{}

		indexerDB.On("GetLastProcessedBlock").Return(uint64(0), errors.New("test error")).Once()

		co, err := NewEthChainObserver(ctx, config, txsReceiverMock, oracleDB, indexerDB, logger)
		require.Error(t, err)
		require.Nil(t, co)
	})

	t.Run("chain observer - NewEventTracker error", func(t *testing.T) {
		txsReceiverMock := &core.EthTxsReceiverMock{}
		oracleDB := &core.EthTxsProcessorDBMock{}
		indexerDB := &core.EventStoreMock{}
		dbBlockNumber := uint64(1)

		indexerDB.On("GetLastProcessedBlock").Return(dbBlockNumber, nil).Once()

		// this will error out new event tracker
		indexerDB.On("GetLastProcessedBlock").Return(dbBlockNumber, errors.New("test error")).Once()

		co, err := NewEthChainObserver(ctx, config, txsReceiverMock, oracleDB, indexerDB, logger)
		require.Error(t, err)
		require.Nil(t, co)
	})

	t.Run("chain observer - NewEthChainObserver OK", func(t *testing.T) {
		txsReceiverMock := &core.EthTxsReceiverMock{}
		oracleDB := &core.EthTxsProcessorDBMock{}
		indexerDB := &core.EventStoreMock{}
		dbBlockNumber := uint64(1)

		indexerDB.On("GetLastProcessedBlock").Return(dbBlockNumber, nil)

		co, err := NewEthChainObserver(ctx, config, txsReceiverMock, oracleDB, indexerDB, logger)
		require.NoError(t, err)
		require.NotNil(t, co)

		require.Equal(t, co.config, config)
		require.Equal(t, co.logger, logger)
		require.Equal(t, co.ctx, ctx)
		require.NotNil(t, co.tracker)
	})

	t.Run("chek start stop", func(t *testing.T) {
		txsReceiverMock := &core.EthTxsReceiverMock{}
		oracleDB := &core.EthTxsProcessorDBMock{}
		indexerDB := &core.EventStoreMock{}
		dbBlockNumber := uint64(1)

		indexerDB.On("GetLastProcessedBlock").Return(dbBlockNumber, nil)

		ctx, cancelFunc := context.WithCancel(context.Background())
		defer cancelFunc()

		chainObserver, err := NewEthChainObserver(ctx, config, txsReceiverMock, oracleDB, indexerDB, logger)
		require.NoError(t, err)
		require.NotNil(t, chainObserver)

		err = chainObserver.Start()
		require.NoError(t, err)
	})

	t.Run("check newConfirmedTxs called", func(t *testing.T) {
		t.Cleanup(foldersCleanup)

		oracleDB := &core.EthTxsProcessorDBMock{}

		txsReceiverMock := &core.EthTxsReceiverMock{}
		doneCh := make(chan bool, 1)
		closed := false

		txsReceiverMock.NewUnprocessedLogFn = func(originChainId string, log *ethgo.Log) error {
			t.Helper()

			if !closed {
				close(doneCh)

				closed = true
			}

			return nil
		}

		testConfig := &oCore.EthChainConfig{
			ChainID:                 "nexus",
			NodeURL:                 ethNodeURL,
			PoolIntervalMiliseconds: 1000,
			SyncBatchSize:           10,
			NumBlockConfirmations:   1,
			StartBlockNumber:        uint64(5462655),
			RestartTrackerPullCheck: time.Second * 30,
		}

		oracleDB.On("ClearAllTxs", mock.Anything).Return(error(nil))

		ctx, cancelFunc := context.WithCancel(context.Background())
		defer cancelFunc()

		indexerDB, err := eventTrackerStore.NewBoltDBEventTrackerStore(filepath.Join(testDir, "nexus.db"))
		require.NoError(t, err)

		chainObserver, err := NewEthChainObserver(ctx, testConfig, txsReceiverMock, oracleDB, indexerDB, logger)
		require.NoError(t, err)
		require.NotNil(t, chainObserver)

		scAddress := ethgo.HexToAddress("0xc68221AD72397d85084f2D5C7089e4e9487c118c")

		eventSigs, err := eth.GetNexusEventSignatures()
		require.NoError(t, err)

		trackerConfig := &eventTracker.EventTrackerConfig{
			RPCEndpoint:            testConfig.NodeURL,
			PollInterval:           testConfig.PoolIntervalMiliseconds * time.Millisecond,
			SyncBatchSize:          testConfig.SyncBatchSize,
			NumBlockConfirmations:  testConfig.NumBlockConfirmations,
			NumOfBlocksToReconcile: uint64(0),
			EventSubscriber: &confirmedEventHandler{
				ChainID:     testConfig.ChainID,
				TxsReceiver: txsReceiverMock,
				Logger:      logger,
			},
			Logger: logger,
			LogFilter: map[ethgo.Address][]ethgo.Hash{
				scAddress: eventSigs,
			},
		}

		chainObserver.tracker, err = eventTracker.NewEventTracker(trackerConfig, indexerDB, testConfig.StartBlockNumber)
		require.NoError(t, err)

		err = chainObserver.Start()
		require.NoError(t, err)

		timer := time.NewTimer(100 * time.Second)
		defer timer.Stop()

		select {
		case <-timer.C:
			t.Fatal("timeout")
		case <-doneCh:
		}
	})
}

func Test_InitOracleState(t *testing.T) {
	blockNumber := uint64(100)
	chainID := "nexus"

	dbMock := &core.EventStoreMock{}
	oracleDBMock := &core.EthTxsProcessorDBMock{}

	t.Run("error", func(t *testing.T) {
		dbMock.On("GetLastProcessedBlock").Return(uint64(0), errors.New("test error")).Once()

		require.ErrorContains(t, initOracleState(dbMock, oracleDBMock, blockNumber, chainID, hclog.NewNullLogger()),
			"could not retrieve latest block point while initializing utxos")
	})

	t.Run("updated db with bigger blockNumber", func(t *testing.T) {
		newBlockNumber := uint64(2)
		dbBlockNumber := uint64(1)

		dbMock.On("GetLastProcessedBlock").Return(dbBlockNumber, nil).Once()
		dbMock.On("InsertLastProcessedBlock", newBlockNumber).Return(nil).Once()

		oracleDBMock.On("ClearAllTxs", chainID).Return(nil).Once()

		require.NoError(t, initOracleState(dbMock, oracleDBMock, newBlockNumber, chainID, hclog.NewNullLogger()))
	})

	t.Run("skip update - db has bigger blockNumber", func(t *testing.T) {
		newBlockNumber := uint64(1)
		dbBlockNumber := uint64(2)

		dbMock.On("GetLastProcessedBlock").Return(dbBlockNumber, nil).Once()
		dbMock.On("InsertLastProcessedBlock", newBlockNumber).Return(errors.New("test error")).Once()

		require.NoError(t, initOracleState(dbMock, oracleDBMock, newBlockNumber, chainID, hclog.NewNullLogger()))
	})

	t.Run("skip update - db has the same blockNumber", func(t *testing.T) {
		newBlockNumber := uint64(1)
		dbBlockNumber := uint64(1)

		dbMock.On("GetLastProcessedBlock").Return(dbBlockNumber, nil).Once()
		dbMock.On("InsertLastProcessedBlock", newBlockNumber).Return(errors.New("test error")).Once()

		require.NoError(t, initOracleState(dbMock, oracleDBMock, newBlockNumber, chainID, hclog.NewNullLogger()))
	})

	t.Run("ClearAllTxs errors", func(t *testing.T) {
		newBlockNumber := uint64(2)
		dbBlockNumber := uint64(1)

		dbMock.On("GetLastProcessedBlock").Return(dbBlockNumber, nil).Once()

		oracleDBMock.On("ClearAllTxs", chainID).Return(errors.New("test error")).Once()

		require.Error(t, initOracleState(dbMock, oracleDBMock, newBlockNumber, chainID, hclog.NewNullLogger()))
	})

	t.Run("InsertLastProcessedBlock errors", func(t *testing.T) {
		newBlockNumber := uint64(2)
		dbBlockNumber := uint64(1)

		dbMock.On("GetLastProcessedBlock").Return(dbBlockNumber, nil).Once()
		dbMock.On("InsertLastProcessedBlock", newBlockNumber).Return(errors.New("test error")).Once()

		oracleDBMock.On("ClearAllTxs", chainID).Return(nil).Once()

		require.Error(t, initOracleState(dbMock, oracleDBMock, newBlockNumber, chainID, hclog.NewNullLogger()))
	})
}

func Test_AddLog(t *testing.T) {
	txsReceiverMock := &core.EthTxsReceiverMock{}
	mockEventHandler := &confirmedEventHandler{
		TxsReceiver: txsReceiverMock,
		ChainID:     "nexus",
		Logger:      hclog.NewNullLogger(),
	}

	mockLog := &ethgo.Log{}

	t.Run("log processed successfully", func(t *testing.T) {
		txsReceiverMock.On("NewUnprocessedLog", "nexus", mockLog).Return(nil).Once()

		require.NoError(t, mockEventHandler.AddLog(big.NewInt(1), mockLog))
	})

	t.Run("NewUnprocessedLog errors", func(t *testing.T) {
		txsReceiverMock.On("NewUnprocessedLog", "nexus", mockLog).Return(errors.New("test error")).Once()

		require.Error(t, mockEventHandler.AddLog(big.NewInt(1), mockLog))
	})
}

func Test_LoadTrackerConfig(t *testing.T) {
	logger := hclog.NewNullLogger()
	txsReceiverMock := &core.EthTxsReceiverMock{}

	scAddress := ethgo.HexToAddress("0x00")

	eventSigs, err := eth.GetNexusEventSignatures()
	require.NoError(t, err)

	logFilter := map[ethgo.Address][]ethgo.Hash{
		scAddress: eventSigs,
	}

	expectedEventTrackerConfig := &eventTracker.EventTrackerConfig{
		RPCEndpoint:            "",
		PollInterval:           0,
		SyncBatchSize:          0,
		NumBlockConfirmations:  0,
		NumOfBlocksToReconcile: uint64(0),
		EventSubscriber: &confirmedEventHandler{
			ChainID:     "",
			TxsReceiver: txsReceiverMock,
			Logger:      logger,
		},
		Logger:    logger,
		LogFilter: logFilter,
	}

	config := &oCore.EthChainConfig{}

	t.Run("loadTrackerConfigs successful", func(t *testing.T) {
		require.Equal(t, expectedEventTrackerConfig, loadTrackerConfigs(config, txsReceiverMock, logger))
	})
}

func Test_executeIsTrackerAlive(t *testing.T) {
	indexerDB := &core.EventStoreMock{}
	ctx, cancelFunc := context.WithCancel(context.Background())

	defer cancelFunc()

	trackerConfig := &eventTracker.EventTrackerConfig{
		RPCEndpoint:     ethNodeURL,
		EventSubscriber: confirmedEventHandler{},
	}

	indexerDB.On("GetLastProcessedBlock").Return(uint64(0), nil).Once()

	tracker, err := eventTracker.NewEventTracker(trackerConfig, indexerDB, 0)
	require.NoError(t, err)

	co := &EthChainObserverImpl{
		indexerDB:     indexerDB,
		ctx:           ctx,
		trackerConfig: trackerConfig,
		trackerState:  ethChainObserverStateCreated,
		tracker:       tracker,
		logger:        hclog.NewNullLogger(),
	}

	t.Run("everything is normal", func(t *testing.T) {
		indexerDB.On("GetLastProcessedBlock").Return(uint64(1), nil).Once()

		co.executeIsTrackerAlive()

		require.Equal(t, ethChainObserverStateCreated, co.trackerState)
	})

	t.Run("restart required", func(t *testing.T) {
		indexerDB.On("GetLastProcessedBlock").Return(uint64(2), nil).Twice()

		co.lastBlock = 2

		require.NoError(t, co.tracker.Start())
		co.executeIsTrackerAlive()

		require.Equal(t, ethChainObserverStateCreated, co.trackerState)
	})

	t.Run("already closed", func(t *testing.T) {
		require.NoError(t, co.Dispose())

		co.executeIsTrackerAlive()

		require.Equal(t, ethChainObserverStateFinished, co.trackerState)
	})

	require.NoError(t, co.Dispose())

	indexerDB.AssertExpectations(t)
}
