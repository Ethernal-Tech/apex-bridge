package oracle

import (
	"context"
	"errors"
	"fmt"
	"path"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/eth_oracle/bridge"
	eth_chain "github.com/Ethernal-Tech/apex-bridge/eth_oracle/chain"
	"github.com/Ethernal-Tech/apex-bridge/eth_oracle/core"
	databaseaccess "github.com/Ethernal-Tech/apex-bridge/eth_oracle/database_access"
	"github.com/Ethernal-Tech/apex-bridge/eth_oracle/processor"
	failedtxprocessors "github.com/Ethernal-Tech/apex-bridge/eth_oracle/processor/failed_tx_processors"
	txprocessors "github.com/Ethernal-Tech/apex-bridge/eth_oracle/processor/tx_processors"
	oracleCore "github.com/Ethernal-Tech/apex-bridge/oracle/core"
	eventTrackerStore "github.com/Ethernal-Tech/blockchain-event-tracker/store"
	"github.com/hashicorp/go-hclog"
)

const (
	MainComponentName = "eth_oracle"
)

type OracleImpl struct {
	ctx                context.Context
	appConfig          *oracleCore.AppConfig
	ethTxsProcessor    core.EthTxsProcessor
	expectedTxsFetcher oracleCore.ExpectedTxsFetcher
	ethChainObservers  []core.EthChainObserver
	db                 core.Database
	logger             hclog.Logger
}

var _ core.Oracle = (*OracleImpl)(nil)

func NewEthOracle(
	ctx context.Context,
	appConfig *oracleCore.AppConfig,
	oracleBridgeSC eth.IOracleBridgeSmartContract,
	bridgeSubmitter oracleCore.BridgeSubmitter,
	indexerDbs map[string]eventTrackerStore.EventTrackerStore,
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater,
	logger hclog.Logger,
) (*OracleImpl, error) {
	db, err := databaseaccess.NewDatabase(path.Join(appConfig.Settings.DbsPath, MainComponentName+".db"))
	if err != nil {
		return nil, fmt.Errorf("failed to open eth oracle database: %w", err)
	}

	bridgeDataFetcher := bridge.NewEthBridgeDataFetcher(
		ctx, oracleBridgeSC, logger.Named("eth_bridge_data_fetcher"))

	expectedTxsFetcher := bridge.NewExpectedTxsFetcher(
		ctx, bridgeDataFetcher, appConfig, db, logger.Named("eth_expected_txs_fetcher"))

	txProcessors := []core.EthTxProcessor{
		txprocessors.NewEthBatchExecutedProcessor(logger),
		txprocessors.NewEthBridgingRequestedProcessor(logger),
		// tx_processors.NewRefundExecutedProcessor(logger),
	}

	failedTxProcessors := []core.EthTxFailedProcessor{
		failedtxprocessors.NewEthBatchExecutionFailedProcessor(logger),
		// failed_tx_processors.NewRefundExecutionFailedProcessor(logger),
	}

	ethTxsProcessor := processor.NewEthTxsProcessor(
		ctx, appConfig, db, txProcessors, failedTxProcessors, bridgeSubmitter,
		indexerDbs, bridgingRequestStateUpdater, logger.Named("eth_txs_processor"))

	ethChainObservers := make([]core.EthChainObserver, 0, len(appConfig.EthChains))

	for _, ethChainConfig := range appConfig.EthChains {
		indexerDB := indexerDbs[ethChainConfig.ChainID]

		eco, err := eth_chain.NewEthChainObserver(
			ctx, ethChainConfig, ethTxsProcessor, db, indexerDB,
			logger.Named("eth_chain_observer_"+ethChainConfig.ChainID))
		if err != nil {
			return nil, fmt.Errorf("failed to create eth chain observer for `%s`: %w", ethChainConfig.ChainID, err)
		}

		ethChainObservers = append(ethChainObservers, eco)
	}

	return &OracleImpl{
		ctx:                ctx,
		appConfig:          appConfig,
		ethTxsProcessor:    ethTxsProcessor,
		expectedTxsFetcher: expectedTxsFetcher,
		ethChainObservers:  ethChainObservers,
		db:                 db,
		logger:             logger,
	}, nil
}

func (o *OracleImpl) Start() error {
	o.logger.Debug("Starting EthOracle")

	go o.ethTxsProcessor.Start()
	go o.expectedTxsFetcher.Start()

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

	err := o.db.Close()
	if err != nil {
		o.logger.Error("Failed to close eth_oracle db", "err", err)
		errs = append(errs, fmt.Errorf("failed to close eth_oracle db. err %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors while disposing eth_oracle. errors: %w", errors.Join(errs...))
	}

	return nil
}
