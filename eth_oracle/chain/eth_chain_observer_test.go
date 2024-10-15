package chain

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/eth_oracle/core"
	oracleCore "github.com/Ethernal-Tech/apex-bridge/oracle/core"
	eventTrackerStore "github.com/Ethernal-Tech/blockchain-event-tracker/store"
	eventTracker "github.com/Ethernal-Tech/blockchain-event-tracker/tracker"

	"github.com/Ethernal-Tech/ethgo"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

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

	config := &oracleCore.EthChainConfig{
		StartBlockNumber: uint64(1),
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

		testConfig := &oracleCore.EthChainConfig{
			ChainID:                 "nexus",
			NodeURL:                 "https://lb.drpc.org/ogrpc?network=ethereum&dkey=Asw61vWBgEgGiV0n4Kq-zfq1GZCgZgMR75uYyp-Zw4Id",
			PoolIntervalMiliseconds: 1 * time.Second,
			SyncBatchSize:           10,
			NumBlockConfirmations:   1,
			StartBlockNumber:        uint64(20588829),
		}

		oracleDB.On("ClearUnprocessedTxs", mock.Anything).Return(error(nil))
		oracleDB.On("ClearExpectedTxs", mock.Anything).Return(error(nil))

		ctx, cancelFunc := context.WithCancel(context.Background())
		defer cancelFunc()

		indexerDB, err := eventTrackerStore.NewBoltDBEventTrackerStore(filepath.Join(testDir, "nexus.db"))
		require.NoError(t, err)

		chainObserver, err := NewEthChainObserver(ctx, testConfig, txsReceiverMock, oracleDB, indexerDB, logger)
		require.NoError(t, err)
		require.NotNil(t, chainObserver)

		scAddress := ethgo.HexToAddress("0x8A796072784aaD48Bf321fBF98725Fb825E3e567")
		eventSigs := []ethgo.Hash{
			ethgo.HexToHash("0xc3d58168c5ae7397731d063d5bbf3d657854427343f4c083240f7aacaa2d0f62"),
		}

		trackerConfig := &eventTracker.EventTrackerConfig{
			RPCEndpoint:            testConfig.NodeURL,
			PollInterval:           1 * time.Second,
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

		oracleDBMock.On("ClearUnprocessedTxs", chainID).Return(nil).Once()
		oracleDBMock.On("ClearExpectedTxs", chainID).Return(nil).Once()

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

	t.Run("ClearUnprocessedTxs errors", func(t *testing.T) {
		newBlockNumber := uint64(2)
		dbBlockNumber := uint64(1)

		dbMock.On("GetLastProcessedBlock").Return(dbBlockNumber, nil).Once()

		oracleDBMock.On("ClearUnprocessedTxs", chainID).Return(errors.New("test error")).Once()

		require.Error(t, initOracleState(dbMock, oracleDBMock, newBlockNumber, chainID, hclog.NewNullLogger()))
	})

	t.Run("ClearExpectedTxs errors", func(t *testing.T) {
		newBlockNumber := uint64(2)
		dbBlockNumber := uint64(1)

		dbMock.On("GetLastProcessedBlock").Return(dbBlockNumber, nil).Once()

		oracleDBMock.On("ClearUnprocessedTxs", chainID).Return(nil).Once()
		oracleDBMock.On("ClearExpectedTxs", chainID).Return(errors.New("test error")).Once()

		require.Error(t, initOracleState(dbMock, oracleDBMock, newBlockNumber, chainID, hclog.NewNullLogger()))
	})

	t.Run("InsertLastProcessedBlock errors", func(t *testing.T) {
		newBlockNumber := uint64(2)
		dbBlockNumber := uint64(1)

		dbMock.On("GetLastProcessedBlock").Return(dbBlockNumber, nil).Once()
		dbMock.On("InsertLastProcessedBlock", newBlockNumber).Return(errors.New("test error")).Once()

		oracleDBMock.On("ClearUnprocessedTxs", chainID).Return(nil).Once()
		oracleDBMock.On("ClearExpectedTxs", chainID).Return(nil).Once()

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

		require.NoError(t, mockEventHandler.AddLog(mockLog))
	})

	t.Run("NewUnprocessedLog errors", func(t *testing.T) {
		txsReceiverMock.On("NewUnprocessedLog", "nexus", mockLog).Return(errors.New("test error")).Once()

		require.Error(t, mockEventHandler.AddLog(mockLog))
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

	config := &oracleCore.EthChainConfig{}

	t.Run("loadTrackerConfigs successful", func(t *testing.T) {
		require.Equal(t, expectedEventTrackerConfig, loadTrackerConfigs(config, txsReceiverMock, logger))
	})
}
