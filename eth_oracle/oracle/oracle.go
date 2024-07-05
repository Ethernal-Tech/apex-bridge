package oracle

import (
	"context"
	"errors"
	"fmt"
	"path"

	"github.com/Ethernal-Tech/apex-bridge/common"
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
	ctx               context.Context
	appConfig         *oracleCore.AppConfig
	ethTxsProcessor   core.EthTxsProcessor
	ethChainObservers []core.EthChainObserver
	db                core.Database
	logger            hclog.Logger

	errorCh chan error
}

var _ oracleCore.Oracle = (*OracleImpl)(nil)

func NewEthOracle(
	ctx context.Context,
	appConfig *oracleCore.AppConfig,
	bridgeDataFetcher oracleCore.BridgeDataFetcher,
	bridgeSubmitter oracleCore.BridgeSubmitter,
	indexerDbs map[string]eventTrackerStore.EventTrackerStore,
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater,
	logger hclog.Logger,
) (*OracleImpl, error) {
	db, err := databaseaccess.NewDatabase(path.Join(appConfig.Settings.DbsPath, MainComponentName+".db"))
	if err != nil {
		return nil, fmt.Errorf("failed to open eth oracle database: %w", err)
	}

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

	ethChainObservers := make([]core.EthChainObserver, 0, len(appConfig.CardanoChains))

	for _, ethChainConfig := range appConfig.EthChains {
		indexerDB := indexerDbs[ethChainConfig.ChainID]

		eco, err := eth_chain.NewEthChainObserver(
			ctx, ethChainConfig, ethTxsProcessor, db, indexerDB, bridgeDataFetcher,
			logger.Named("eth_chain_observer_"+ethChainConfig.ChainID))
		if err != nil {
			return nil, fmt.Errorf("failed to create eth chain observer for `%s`: %w", ethChainConfig.ChainID, err)
		}

		ethChainObservers = append(ethChainObservers, eco)
	}

	return &OracleImpl{
		ctx:               ctx,
		appConfig:         appConfig,
		ethTxsProcessor:   ethTxsProcessor,
		ethChainObservers: ethChainObservers,
		db:                db,
		logger:            logger,
	}, nil
}

func (o *OracleImpl) Start() error {
	o.logger.Debug("Starting EthOracle")

	go o.ethTxsProcessor.Start()

	o.errorCh = make(chan error, 1)
	go o.errorHandler()

	o.logger.Debug("Started EthOracle")

	return nil
}

func (o *OracleImpl) Dispose() error {
	errs := make([]error, 0)

	err := o.db.Close()
	if err != nil {
		o.logger.Error("Failed to close eth_oracle db", "err", err)
		errs = append(errs, fmt.Errorf("failed to close eth_oracle db. err %w", err))
	}

	close(o.errorCh)

	if len(errs) > 0 {
		return fmt.Errorf("errors while disposing eth_oracle. errors: %w", errors.Join(errs...))
	}

	return nil
}

func (o *OracleImpl) ErrorCh() <-chan error {
	return o.errorCh
}

type ErrorOrigin struct {
	err    error
	origin string
}

func (o *OracleImpl) errorHandler() {
	agg := make(chan ErrorOrigin)
	defer close(agg)

	select {
	case errorOrigin := <-agg:
		o.logger.Error("critical error", "origin", errorOrigin.origin, "err", errorOrigin.err)
		o.errorCh <- errorOrigin.err
	case <-o.ctx.Done():
	}
	o.logger.Debug("Exiting eth_oracle error handler")
}
