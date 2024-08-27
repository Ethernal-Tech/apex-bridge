package chain

import (
	"context"
	"errors"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/eth_oracle/core"
	oracleCore "github.com/Ethernal-Tech/apex-bridge/oracle/core"
	eventTracker "github.com/Ethernal-Tech/blockchain-event-tracker/tracker"
	"github.com/Ethernal-Tech/ethgo"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestEthChainObserver(t *testing.T) {
	ctx := context.Background()
	txsProcessorMock := &core.EthTxsProcessorMock{}
	oracleDB := &core.EthTxsProcessorDBMock{}
	indexerDB := &core.EventStoreMock{}
	logger := hclog.NewNullLogger()

	config := &oracleCore.EthChainConfig{
		StartBlockNumber: uint64(1),
	}

	t.Run("chain observer - initOracleState error", func(t *testing.T) {
		indexerDB.On("GetLastProcessedBlock").Return(uint64(0), errors.New("test error")).Once()

		co, err := NewEthChainObserver(ctx, config, txsProcessorMock, oracleDB, indexerDB, logger)
		require.Error(t, err)
		require.Nil(t, co)
	})

	t.Run("chain observer - NewEventTracker error", func(t *testing.T) {
		dbBlockNumber := uint64(1)

		indexerDB.On("GetLastProcessedBlock").Return(dbBlockNumber, nil).Once()

		// this will error out new event tracker
		indexerDB.On("GetLastProcessedBlock").Return(dbBlockNumber, errors.New("test error")).Once()

		co, err := NewEthChainObserver(ctx, config, txsProcessorMock, oracleDB, indexerDB, logger)
		require.Error(t, err)
		require.Nil(t, co)
	})

	t.Run("chain observer - NewEthChainObserver OK", func(t *testing.T) {
		dbBlockNumber := uint64(1)

		indexerDB.On("GetLastProcessedBlock").Return(dbBlockNumber, nil)

		co, err := NewEthChainObserver(ctx, config, txsProcessorMock, oracleDB, indexerDB, logger)
		require.NoError(t, err)
		require.NotNil(t, co)

		require.Equal(t, co.config, config)
		require.Equal(t, co.logger, logger)
		require.Equal(t, co.ctx, ctx)
		require.NotNil(t, co.tracker)
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
	txProcessorMock := &core.EthTxsProcessorMock{}
	mockEventHandler := &confirmedEventHandler{
		TxsProcessor: txProcessorMock,
		ChainID:      "nexus",
		Logger:       hclog.NewNullLogger(),
	}

	mockLog := &ethgo.Log{}

	t.Run("log processed successfully", func(t *testing.T) {
		txProcessorMock.On("NewUnprocessedLog", "nexus", mockLog).Return(nil).Once()

		require.NoError(t, mockEventHandler.AddLog(mockLog))
	})

	t.Run("NewUnprocessedLog errors", func(t *testing.T) {
		txProcessorMock.On("NewUnprocessedLog", "nexus", mockLog).Return(errors.New("test error")).Once()

		require.Error(t, mockEventHandler.AddLog(mockLog))
	})
}

func Test_LoadTrackerConfig(t *testing.T) {
	logger := hclog.NewNullLogger()
	txsProcessorMock := &core.EthTxsProcessorMock{}

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
			ChainID:      "",
			TxsProcessor: txsProcessorMock,
			Logger:       logger,
		},
		Logger:    logger,
		LogFilter: logFilter,
	}

	config := &oracleCore.EthChainConfig{}

	t.Run("loadTrackerConfigs successful", func(t *testing.T) {
		require.Equal(t, expectedEventTrackerConfig, loadTrackerConfigs(config, txsProcessorMock, logger))
	})
}
