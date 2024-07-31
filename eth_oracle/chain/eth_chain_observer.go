package chain

import (
	"context"
	"fmt"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	ethOracleCore "github.com/Ethernal-Tech/apex-bridge/eth_oracle/core"
	oracleCore "github.com/Ethernal-Tech/apex-bridge/oracle/core"
	eventTrackerStore "github.com/Ethernal-Tech/blockchain-event-tracker/store"
	eventTracker "github.com/Ethernal-Tech/blockchain-event-tracker/tracker"
	"github.com/Ethernal-Tech/ethgo"
	"github.com/hashicorp/go-hclog"
)

type EthChainObserverImpl struct {
	ctx     context.Context
	config  *oracleCore.EthChainConfig
	tracker *eventTracker.EventTracker
	logger  hclog.Logger
}

var _ ethOracleCore.EthChainObserver = (*EthChainObserverImpl)(nil)

func NewEthChainObserver(
	ctx context.Context,
	config *oracleCore.EthChainConfig,
	txsProcessor ethOracleCore.EthTxsProcessor,
	oracleDB ethOracleCore.EthTxsProcessorDB,
	indexerDB eventTrackerStore.EventTrackerStore,
	logger hclog.Logger,
) (*EthChainObserverImpl, error) {
	trackerConfig := loadTrackerConfigs(config, txsProcessor, logger)

	err := initOracleState(indexerDB, oracleDB, config.StartBlockNumber, config.ChainID, logger)
	if err != nil {
		return nil, err
	}

	ethTracker, err := eventTracker.NewEventTracker(trackerConfig, indexerDB, config.StartBlockNumber)
	if err != nil {
		return nil, err
	}

	return &EthChainObserverImpl{
		ctx:     ctx,
		logger:  logger,
		config:  config,
		tracker: ethTracker,
	}, nil
}

func (co *EthChainObserverImpl) Start() error {
	if err := co.tracker.Start(); err != nil {
		return err
	}

	return nil
}

func (co *EthChainObserverImpl) Dispose() error {
	co.tracker.Close()

	return nil
}

func (co *EthChainObserverImpl) GetConfig() *oracleCore.EthChainConfig {
	return co.config
}

func loadTrackerConfigs(config *oracleCore.EthChainConfig, txsProcessor ethOracleCore.EthTxsProcessor,
	logger hclog.Logger,
) *eventTracker.EventTrackerConfig {
	bridgingAddress := config.BridgingAddresses.BridgingAddress
	scAddress := ethgo.HexToAddress(bridgingAddress)

	events := []string{"Deposit", "Withdraw"}

	eventSigs, err := getEventSignatures(events)
	if err != nil {
		logger.Error("failed to get event signatures", err)

		return nil
	}

	logFilter := make(map[ethgo.Address][]ethgo.Hash)
	logFilter[scAddress] = append(logFilter[scAddress], eventSigs...)

	return &eventTracker.EventTrackerConfig{
		RPCEndpoint:            config.NodeURL,
		PollInterval:           config.PoolIntervalMiliseconds * time.Millisecond,
		SyncBatchSize:          config.SyncBatchSize,
		NumBlockConfirmations:  config.NumBlockConfirmations,
		NumOfBlocksToReconcile: uint64(0),
		EventSubscriber: &confirmedEventHandler{
			ChainID:      config.ChainID,
			TxsProcessor: txsProcessor,
			Logger:       logger,
		},
		Logger:    logger,
		LogFilter: logFilter,
	}
}

type confirmedEventHandler struct {
	TxsProcessor ethOracleCore.EthTxsProcessor
	ChainID      string
	Logger       hclog.Logger
}

func (handler confirmedEventHandler) AddLog(log *ethgo.Log) error {
	handler.Logger.Info("Confirmed Event Handler invoked",
		"block hash", log.BlockHash, "block number", log.BlockNumber, "tx hash", log.TransactionHash)

	err := handler.TxsProcessor.NewUnprocessedLog(handler.ChainID, log)
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

	if err := oracleDB.ClearUnprocessedTxs(chainID); err != nil {
		return err
	}

	if err := oracleDB.ClearExpectedTxs(chainID); err != nil {
		return err
	}

	return db.InsertLastProcessedBlock(blockNumber)
}

func getEventSignatures(events []string) ([]ethgo.Hash, error) {
	abi, err := contractbinding.GatewayMetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	hashes := make([]ethgo.Hash, len(events))
	for i, ev := range events {
		hashes[i] = ethgo.Hash(abi.Events[ev].ID)
	}

	return hashes, nil
}
