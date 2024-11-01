package oracle

import (
	"context"
	"errors"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	txsprocessor "github.com/Ethernal-Tech/apex-bridge/oracle_common/processor/txs_processor"
	"github.com/Ethernal-Tech/apex-bridge/oracle_eth/bridge"
	eth_chain "github.com/Ethernal-Tech/apex-bridge/oracle_eth/chain"
	"github.com/Ethernal-Tech/apex-bridge/oracle_eth/core"
	databaseaccess "github.com/Ethernal-Tech/apex-bridge/oracle_eth/database_access"
	failedtxprocessors "github.com/Ethernal-Tech/apex-bridge/oracle_eth/processor/tx_processors/failed"
	successtxprocessors "github.com/Ethernal-Tech/apex-bridge/oracle_eth/processor/tx_processors/success"
	ethtxsprocessor "github.com/Ethernal-Tech/apex-bridge/oracle_eth/processor/txs_processor"
	eventTrackerStore "github.com/Ethernal-Tech/blockchain-event-tracker/store"
	"github.com/hashicorp/go-hclog"
	"go.etcd.io/bbolt"
)

const (
	MainComponentName = "oracle_eth"
)

type OracleImpl struct {
	ctx                      context.Context
	appConfig                *oCore.AppConfig
	ethTxsProcessor          oCore.TxsProcessor
	expectedTxsFetcher       oCore.ExpectedTxsFetcher
	ethChainObservers        []core.EthChainObserver
	confirmedBlockSubmitters []core.EthConfirmedBlocksSubmitter
	db                       core.Database
	logger                   hclog.Logger
}

var _ core.Oracle = (*OracleImpl)(nil)

func NewEthOracle(
	ctx context.Context,
	boltDB *bbolt.DB,
	typeRegister common.TypeRegister,
	appConfig *oCore.AppConfig,
	oracleBridgeSC eth.IOracleBridgeSmartContract,
	bridgeSubmitter core.BridgeSubmitter,
	indexerDbs map[string]eventTrackerStore.EventTrackerStore,
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater,
	logger hclog.Logger,
) (*OracleImpl, error) {
	db := &databaseaccess.BBoltDatabase{}
	db.Init(boltDB, appConfig, typeRegister)

	bridgeDataFetcher := bridge.NewEthBridgeDataFetcher(
		ctx, oracleBridgeSC, logger.Named("eth_bridge_data_fetcher"))

	expectedTxsFetcher := bridge.NewExpectedTxsFetcher(
		ctx, bridgeDataFetcher, appConfig, db, logger.Named("eth_expected_txs_fetcher"))

	txProcessors := ethtxsprocessor.NewTxProcessorsCollection(
		[]core.EthTxSuccessProcessor{
			successtxprocessors.NewEthBatchExecutedProcessor(logger),
			successtxprocessors.NewEthBridgingRequestedProcessor(logger),
			successtxprocessors.NewHotWalletIncrementProcessor(logger),
			// tx_processors.NewRefundExecutedProcessor(logger),
		},
		[]core.EthTxFailedProcessor{
			failedtxprocessors.NewEthBatchExecutionFailedProcessor(logger),
			// failed_tx_processors.NewRefundExecutionFailedProcessor(logger),
		},
	)

	txsProcessorLogger := logger.Named("eth_txs_processor")

	ethTxsReceiver := ethtxsprocessor.NewEthTxsReceiverImpl(
		appConfig, db, txProcessors, bridgingRequestStateUpdater, txsProcessorLogger)

	ethStateProcessor := ethtxsprocessor.NewEthStateProcessor(
		ctx, appConfig, db, txProcessors,
		indexerDbs, txsProcessorLogger,
	)

	ethTxsProcessor := txsprocessor.NewTxsProcessorImpl(
		ctx, appConfig, ethStateProcessor, bridgeSubmitter,
		bridgingRequestStateUpdater, txsProcessorLogger)

	ethChainObservers := make([]core.EthChainObserver, 0, len(appConfig.EthChains))
	confirmedBlockSubmitters := make([]core.EthConfirmedBlocksSubmitter, 0, len(appConfig.EthChains))

	for _, ethChainConfig := range appConfig.EthChains {
		indexerDB := indexerDbs[ethChainConfig.ChainID]

		cbs, err := bridge.NewConfirmedBlocksSubmitter(
			ctx, bridgeSubmitter, appConfig, indexerDB, ethChainConfig.ChainID, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create evm block submitter for `%s`: %w", ethChainConfig.ChainID, err)
		}

		confirmedBlockSubmitters = append(confirmedBlockSubmitters, cbs)

		eco, err := eth_chain.NewEthChainObserver(
			ctx, ethChainConfig, ethTxsReceiver, db, indexerDB,
			logger.Named("eth_chain_observer_"+ethChainConfig.ChainID))
		if err != nil {
			return nil, fmt.Errorf("failed to create eth chain observer for `%s`: %w", ethChainConfig.ChainID, err)
		}

		ethChainObservers = append(ethChainObservers, eco)
	}

	return &OracleImpl{
		ctx:                      ctx,
		appConfig:                appConfig,
		ethTxsProcessor:          ethTxsProcessor,
		expectedTxsFetcher:       expectedTxsFetcher,
		ethChainObservers:        ethChainObservers,
		confirmedBlockSubmitters: confirmedBlockSubmitters,
		db:                       db,
		logger:                   logger,
	}, nil
}

func (o *OracleImpl) Start() error {
	o.logger.Debug("Starting EthOracle")

	go o.ethTxsProcessor.Start()
	go o.expectedTxsFetcher.Start()

	for _, cbs := range o.confirmedBlockSubmitters {
		cbs.StartSubmit()
	}

	for _, eco := range o.ethChainObservers {
		err := eco.Start()
		if err != nil {
			return fmt.Errorf("failed to start eth observer for %s: %w", eco.GetConfig().ChainID, err)
		}
	}

	o.logger.Debug("Started EthOracle")

	return nil
}

func (o *OracleImpl) Dispose() error {
	errs := make([]error, 0)

	for _, eco := range o.ethChainObservers {
		err := eco.Dispose()
		if err != nil {
			o.logger.Error("error while disposing eth chain observer", "chainId", eco.GetConfig().ChainID, "err", err)
			errs = append(errs, fmt.Errorf("error while disposing eth chain observer. chainId: %v, err: %w",
				eco.GetConfig().ChainID, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors while disposing oracle_eth. errors: %w", errors.Join(errs...))
	}

	return nil
}
