package chain

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/eth"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	ethOracleCore "github.com/Ethernal-Tech/apex-bridge/oracle_eth/core"
	eventTrackerStore "github.com/Ethernal-Tech/blockchain-event-tracker/store"
	eventTracker "github.com/Ethernal-Tech/blockchain-event-tracker/tracker"
	"github.com/Ethernal-Tech/ethgo"
	"github.com/hashicorp/go-hclog"
)

type ethChainObserverState byte

const (
	ethChainObserverStateDisposed ethChainObserverState = iota
	ethChainObserverStateCreated
	ethChainObserverStateFinished
)

type EthChainObserverImpl struct {
	ctx           context.Context
	config        *oCore.EthChainConfig
	tracker       *eventTracker.EventTracker
	indexerDB     eventTrackerStore.EventTrackerStore
	lastBlockLock sync.Mutex
	lastBlock     uint64
	trackerState  ethChainObserverState
	trackerConfig *eventTracker.EventTrackerConfig
	logger        hclog.Logger
}

var _ ethOracleCore.EthChainObserver = (*EthChainObserverImpl)(nil)

func NewEthChainObserver(
	ctx context.Context,
	config *oCore.EthChainConfig,
	txsReceiver ethOracleCore.EthTxsReceiver,
	oracleDB ethOracleCore.EthTxsProcessorDB,
	indexerDB eventTrackerStore.EventTrackerStore,
	logger hclog.Logger,
) (*EthChainObserverImpl, error) {
	trackerConfig := loadTrackerConfigs(config, txsReceiver, logger)

	err := initOracleState(indexerDB, oracleDB, config.StartBlockNumber, config.ChainID, logger)
	if err != nil {
		return nil, err
	}

	ethTracker, err := eventTracker.NewEventTracker(trackerConfig, indexerDB, config.StartBlockNumber)
	if err != nil {
		return nil, err
	}

	return &EthChainObserverImpl{
		ctx:           ctx,
		config:        config,
		tracker:       ethTracker,
		indexerDB:     indexerDB,
		trackerConfig: trackerConfig,
		trackerState:  ethChainObserverStateCreated,
		logger:        logger,
	}, nil
}

func (co *EthChainObserverImpl) Start() error {
	if err := co.tracker.Start(); err != nil {
		return err
	}

	// restart tracker if it is not alive
	go func() {
		for {
			select {
			case <-co.ctx.Done():
				return
			case <-time.After(co.config.RestartTrackerPullCheck):
				co.executeIsTrackerAlive()
			}
		}
	}()

	return nil
}

func (co *EthChainObserverImpl) Dispose() error {
	co.lastBlockLock.Lock()
	defer co.lastBlockLock.Unlock()

	if co.trackerState == ethChainObserverStateCreated {
		co.tracker.Close()
	}

	co.trackerState = ethChainObserverStateFinished

	return nil
}

func (co *EthChainObserverImpl) GetConfig() *oCore.EthChainConfig {
	return co.config
}

func (co *EthChainObserverImpl) executeIsTrackerAlive() {
	co.lastBlockLock.Lock()
	defer co.lastBlockLock.Unlock()

	if co.trackerState == ethChainObserverStateFinished {
		co.logger.Debug("eth tracker is already closed")

		return
	}

	block, err := co.indexerDB.GetLastProcessedBlock()
	if err != nil {
		co.logger.Warn("failed to retrieve last processed eth block from eth tracker: %w")

		return
	}

	// everything is ok, tracker block is greater then previous saved
	if block > co.lastBlock {
		co.lastBlock = block

		return
	}

	// close only if there is tracker to close
	if co.trackerState == ethChainObserverStateCreated {
		co.tracker.Close()

		select {
		case <-co.ctx.Done():
			return
		case <-co.tracker.GetFinishClosingCh():
		}

		co.trackerState = ethChainObserverStateDisposed
	}

	ethTracker, err := eventTracker.NewEventTracker(co.trackerConfig, co.indexerDB, block)
	if err != nil {
		co.logger.Warn("failed to create new eth block tracker: %w")

		return
	}

	if err := ethTracker.Start(); err != nil {
		co.logger.Warn("failed to restart eth block tracker: %w")

		return
	}

	co.tracker = ethTracker
	co.trackerState = ethChainObserverStateCreated
}

func loadTrackerConfigs(config *oCore.EthChainConfig, txsReceiver ethOracleCore.EthTxsReceiver,
	logger hclog.Logger,
) *eventTracker.EventTrackerConfig {
	bridgingAddress := config.BridgingAddresses.BridgingAddress
	scAddress := ethgo.HexToAddress(bridgingAddress)

	eventSigs, err := eth.GetNexusEventSignatures()
	if err != nil {
		logger.Error("failed to get nexus event signatures", "err", err)

		return nil
	}

	logFilter := map[ethgo.Address][]ethgo.Hash{
		scAddress: eventSigs,
	}

	return &eventTracker.EventTrackerConfig{
		RPCEndpoint:            config.NodeURL,
		PollInterval:           config.PoolIntervalMiliseconds * time.Millisecond,
		SyncBatchSize:          config.SyncBatchSize,
		NumBlockConfirmations:  config.NumBlockConfirmations,
		NumOfBlocksToReconcile: uint64(0),
		EventSubscriber: &confirmedEventHandler{
			ChainID:     config.ChainID,
			TxsReceiver: txsReceiver,
			Logger:      logger,
		},
		Logger:    logger,
		LogFilter: logFilter,
	}
}

type confirmedEventHandler struct {
	TxsReceiver ethOracleCore.EthTxsReceiver
	ChainID     string
	Logger      hclog.Logger
}

func (handler confirmedEventHandler) AddLog(_ *big.Int, log *ethgo.Log) error {
	handler.Logger.Info("Confirmed Event Handler invoked",
		"block hash", log.BlockHash, "block number", log.BlockNumber, "tx hash", log.TransactionHash)

	err := handler.TxsReceiver.NewUnprocessedLog(handler.ChainID, log)
	if err != nil {
		handler.Logger.Error("Failed to process new log", "err", err, "log", log)

		return err
	}

	handler.Logger.Info("Log has been processed", "log", log)

	return nil
}

func initOracleState(
	db eventTrackerStore.EventTrackerStore, oracleDB ethOracleCore.EthTxsProcessorDB, blockNumber uint64,
	chainID string, logger hclog.Logger,
) error {
	currentBlockNumber, err := db.GetLastProcessedBlock()
	if err != nil {
		return fmt.Errorf("could not retrieve latest block point while initializing utxos: %w", err)
	}

	// in oracle we already have more recent block
	if currentBlockNumber >= blockNumber {
		logger.Info("Oracle database contains more recent block", "block number", currentBlockNumber)

		return nil
	}

	if err := oracleDB.ClearAllTxs(chainID); err != nil {
		return err
	}

	return db.InsertLastProcessedBlock(blockNumber)
}
