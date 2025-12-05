package chain

import (
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

	defer func() {
		os.RemoveAll(testDir)
		os.Remove(testDir)
	}()

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

		co, err := NewEthChainObserver(config, txsReceiverMock, oracleDB, indexerDB, logger)
		require.Error(t, err)
		require.Nil(t, co)
	})

	t.Run("chain observer - NewEventTracker error", func(t *testing.T) {
		txsReceiverMock := &core.EthTxsReceiverMock{}
		oracleDB := &core.EthTxsProcessorDBMock{}
		indexerDB := &core.EventStoreMock{}

		indexerDB.On("GetLastProcessedBlock").Return(uint64(0), errors.New("test error")).Once()

		co, err := NewEthChainObserver(config, txsReceiverMock, oracleDB, indexerDB, logger)
		require.Error(t, err)
		require.Nil(t, co)
	})

	t.Run("chain observer - NewEthChainObserver OK", func(t *testing.T) {
		txsReceiverMock := &core.EthTxsReceiverMock{}
		oracleDB := &core.EthTxsProcessorDBMock{}
		indexerDB := &core.EventStoreMock{}
		dbBlockNumber := uint64(1)

		indexerDB.On("GetLastProcessedBlock").Return(dbBlockNumber, nil)

		co, err := NewEthChainObserver(config, txsReceiverMock, oracleDB, indexerDB, logger)
		require.NoError(t, err)
		require.NotNil(t, co)

		require.Equal(t, co.config, config)
		require.Equal(t, co.logger, logger)
	})

	t.Run("chek start stop", func(t *testing.T) {
		txsReceiverMock := &core.EthTxsReceiverMock{}
		oracleDB := &core.EthTxsProcessorDBMock{}
		indexerDB := &core.EventStoreMock{}
		dbBlockNumber := uint64(1)

		indexerDB.On("GetLastProcessedBlock").Return(dbBlockNumber, nil)

		chainObserver, err := NewEthChainObserver(config, txsReceiverMock, oracleDB, indexerDB, logger)
		require.NoError(t, err)
		require.NotNil(t, chainObserver)

		err = chainObserver.Start()
		require.NoError(t, err)
	})

	// wTODO: uncomment when new smart contracts are deployed on the testnet
	// t.Run("check newConfirmedTxs called", func(t *testing.T) {
	// 	oracleDB := &core.EthTxsProcessorDBMock{}

	// 	txsReceiverMock := &core.EthTxsReceiverMock{}
	// 	doneCh := make(chan bool, 1)
	// 	closed := false

	// 	txsReceiverMock.NewUnprocessedLogFn = func(originChainId string, log *ethgo.Log) error {
	// 		t.Helper()

	// 		if !closed {
	// 			close(doneCh)

	// 			closed = true
	// 		}

	// 		return nil
	// 	}

	// 	testConfig := &oCore.EthChainConfig{
	// 		ChainID:                 "nexus",
	// 		NodeURL:                 ethNodeURL,
	// 		PoolIntervalMiliseconds: 1000,
	// 		SyncBatchSize:           10,
	// 		NumBlockConfirmations:   1,
	// 		StartBlockNumber:        uint64(5462655),
	// 		RestartTrackerPullCheck: time.Second * 30,
	// 		BridgingAddresses: oCore.EthBridgingAddresses{
	// 			BridgingAddress: "0xc68221AD72397d85084f2D5C7089e4e9487c118c",
	// 		},
	// 	}

	// 	oracleDB.On("ClearAllTxs", mock.Anything).Return(error(nil))

	// 	indexerDB, err := eventTrackerStore.NewBoltDBEventTrackerStore(filepath.Join(testDir, "nexus.db"))
	// 	require.NoError(t, err)

	// 	chainObserver, err := NewEthChainObserver(testConfig, txsReceiverMock, oracleDB, indexerDB, logger)
	// 	require.NoError(t, err)
	// 	require.NotNil(t, chainObserver)

	// 	require.NoError(t, chainObserver.Start())

	// 	select {
	// 	case <-time.After(100 * time.Second):
	// 		t.Fatal("timeout")
	// 	case <-doneCh:
	// 	}
	// })
}

func TestInitOracleState(t *testing.T) {
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

func TestEthChainObserver_AddLog(t *testing.T) {
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

	eventSigs, err := eth.GetGatewayEventSignatures()
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

func TestEthChainObserver_ExecuteIsTrackerAlive(t *testing.T) {
	indexerDB := &core.EventStoreMock{}

	co := &EthChainObserverImpl{
		indexerDB:   indexerDB,
		txsReceiver: &core.EthTxsReceiverMock{},
		logger:      hclog.NewNullLogger(),
	}

	t.Run("everything is normal", func(t *testing.T) {
		indexerDB.On("GetLastProcessedBlock").Return(uint64(1), nil).Once()

		require.True(t, co.updateIsTrackerAlive())
		require.Equal(t, uint64(1), co.lastBlock)
	})

	t.Run("restart required", func(t *testing.T) {
		indexerDB.On("GetLastProcessedBlock").Return(uint64(2), nil).Once()

		co.lastBlock = 2

		require.False(t, co.updateIsTrackerAlive())
		require.Equal(t, uint64(2), co.lastBlock)
	})

	indexerDB.AssertExpectations(t)
}

func TestEthChainObserver_Dispose(t *testing.T) {
	testDir, err := os.MkdirTemp("", "boltdb-test")
	require.NoError(t, err)

	defer func() {
		os.RemoveAll(testDir)
		os.Remove(testDir)
	}()

	logger := hclog.NewNullLogger()
	oracleDB := &core.EthTxsProcessorDBMock{}
	txsReceiverMock := &core.EthTxsReceiverMock{}
	testConfig := &oCore.EthChainConfig{
		ChainID:                 "nexus",
		NodeURL:                 ethNodeURL,
		PoolIntervalMiliseconds: 1000,
		SyncBatchSize:           10,
		NumBlockConfirmations:   1,
		StartBlockNumber:        uint64(5462655),
		RestartTrackerPullCheck: time.Second * 30,
		BridgingAddresses: oCore.EthBridgingAddresses{
			BridgingAddress: "0xc68221AD72397d85084f2D5C7089e4e9487c118c",
		},
	}

	oracleDB.On("ClearAllTxs", mock.Anything).Return(error(nil))

	indexerDB, err := eventTrackerStore.NewBoltDBEventTrackerStore(filepath.Join(testDir, "nexus.db"))
	require.NoError(t, err)

	chainObserver, err := NewEthChainObserver(testConfig, txsReceiverMock, oracleDB, indexerDB, logger)
	require.NoError(t, err)
	require.NotNil(t, chainObserver)

	require.NoError(t, chainObserver.Start())
	require.NoError(t, chainObserver.Dispose())
}
