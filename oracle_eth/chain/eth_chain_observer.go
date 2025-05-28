package chain

import (
	"fmt"
	"math/big"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/eth"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	ethOracleCore "github.com/Ethernal-Tech/apex-bridge/oracle_eth/core"
	eventTrackerStore "github.com/Ethernal-Tech/blockchain-event-tracker/store"
	eventTracker "github.com/Ethernal-Tech/blockchain-event-tracker/tracker"
	"github.com/Ethernal-Tech/ethgo"
	"github.com/hashicorp/go-hclog"
)

type EthChainObserverImpl struct {
	config      *oCore.EthChainConfig
	indexerDB   eventTrackerStore.EventTrackerStore
	txsReceiver ethOracleCore.EthTxsReceiver
	lastBlock   uint64
	closedCh    chan struct{}
	logger      hclog.Logger
}

var _ ethOracleCore.EthChainObserver = (*EthChainObserverImpl)(nil)

func NewEthChainObserver(
	config *oCore.EthChainConfig,
	txsReceiver ethOracleCore.EthTxsReceiver,
	oracleDB ethOracleCore.EthTxsProcessorDB,
	indexerDB eventTrackerStore.EventTrackerStore,
	logger hclog.Logger,
) (*EthChainObserverImpl, error) {
	err := initOracleState(indexerDB, oracleDB, config.StartBlockNumber, config.ChainID, logger)
	if err != nil {
		return nil, err
	}

	return &EthChainObserverImpl{
		config:      config,
		indexerDB:   indexerDB,
		txsReceiver: txsReceiver,
		closedCh:    make(chan struct{}),
		logger:      logger,
	}, nil
}

func (co *EthChainObserverImpl) Start() error {
	co.logger.Debug("Starting eth chain observer", "endpoint", co.config.NodeURL)

	trackerConfig := loadTrackerConfigs(co.config, co.txsReceiver, co.logger)

	tracker, notifyClosedCh, err := newEventTrackerWrapper(trackerConfig, co.indexerDB)
	if err != nil {
		co.logger.Error("Failed to create event tracker", "error", err)
	}

	go tracker.Start()

	go func() {
		for {
			select {
			case <-co.closedCh:
				tracker.Close() // close old tracker

				return

			case <-time.After(co.config.RestartTrackerPullCheck):
				// restart tracker if it is not alive
				co.logger.Debug("Check if tracker is alive", "endpoint", trackerConfig.RPCEndpoint)

				if !co.updateIsTrackerAlive() {
					co.logger.Debug("Tracker is not alive anymore", "endpoint", trackerConfig.RPCEndpoint)

					tracker.Close() // close old tracker

					select {
					case <-co.closedCh:
					case <-notifyClosedCh:
						_ = co.Start()
					}

					return
				}
			}
		}
	}()

	return nil
}

func (co *EthChainObserverImpl) Dispose() error {
	close(co.closedCh)

	return nil
}

func (co *EthChainObserverImpl) GetConfig() *oCore.EthChainConfig {
	return co.config
}

func (co *EthChainObserverImpl) updateIsTrackerAlive() bool {
	block, err := co.indexerDB.GetLastProcessedBlock()
	if err != nil {
		co.logger.Warn("failed to retrieve last processed eth block from eth tracker: %w")

		return true
	}

	// everything is ok, tracker block is greater then previous saved
	if block > co.lastBlock {
		co.lastBlock = block // update last block number

		return true
	}

	return false
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
		StartBlockFromGenesis: config.StartBlockNumber,
		LogFilter:             logFilter,
		// add timestamp to the logger to differentiate between multiple instances
		Logger: logger.Named(time.Now().UTC().String()),
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
