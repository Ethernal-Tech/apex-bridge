package oracle

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/oracle/bridge"
	"github.com/Ethernal-Tech/apex-bridge/oracle/chain"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	databaseaccess "github.com/Ethernal-Tech/apex-bridge/oracle/database_access"
	"github.com/Ethernal-Tech/apex-bridge/oracle/processor"
	failedtxprocessors "github.com/Ethernal-Tech/apex-bridge/oracle/processor/failed_tx_processors"
	txprocessors "github.com/Ethernal-Tech/apex-bridge/oracle/processor/tx_processors"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/hashicorp/go-hclog"
)

const (
	MainComponentName = "oracle"
)

var (
	errBlockSyncerFatal = errors.New("block syncer fatal error")
)

type OracleImpl struct {
	ctx                      context.Context
	appConfig                *core.AppConfig
	cardanoTxsProcessor      core.CardanoTxsProcessor
	cardanoChainObservers    []core.CardanoChainObserver
	db                       core.Database
	expectedTxsFetcher       core.ExpectedTxsFetcher
	confirmedBlockSubmitters []core.ConfirmedBlocksSubmitter
	logger                   hclog.Logger

	errorCh chan error
}

var _ core.Oracle = (*OracleImpl)(nil)

func NewOracle(
	ctx context.Context,
	appConfig *core.AppConfig,
	oracleBridgeSC eth.IOracleBridgeSmartContract,
	bridgeSubmitter core.BridgeSubmitter,
	indexerDbs map[string]indexer.Database,
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater,
	logger hclog.Logger,
) (*OracleImpl, error) {
	db, err := databaseaccess.NewDatabase(filepath.Join(appConfig.Settings.DbsPath, MainComponentName+".db"))
	if err != nil {
		return nil, fmt.Errorf("failed to open oracle database: %w", err)
	}

	bridgeDataFetcher := bridge.NewBridgeDataFetcher(ctx, oracleBridgeSC, logger.Named("bridge_data_fetcher"))

	expectedTxsFetcher := bridge.NewExpectedTxsFetcher(
		ctx, bridgeDataFetcher, appConfig, db, logger.Named("expected_txs_fetcher"))

	txProcessors := []core.CardanoTxProcessor{
		txprocessors.NewBatchExecutedProcessor(logger),
		txprocessors.NewBridgingRequestedProcessor(logger),
		// tx_processors.NewRefundExecutedProcessor(logger),
	}

	failedTxProcessors := []core.CardanoTxFailedProcessor{
		failedtxprocessors.NewBatchExecutionFailedProcessor(logger),
		// failed_tx_processors.NewRefundExecutionFailedProcessor(logger),
	}

	cardanoTxsProcessor := processor.NewCardanoTxsProcessor(
		ctx, appConfig, db, txProcessors, failedTxProcessors, bridgeSubmitter,
		indexerDbs, bridgingRequestStateUpdater, logger.Named("cardano_txs_processor"))

	chainBalanceFetcher := chain.NewCardanoBalanceTracking(ctx, appConfig, db, logger)

	cardanoChainObservers := make([]core.CardanoChainObserver, 0, len(appConfig.CardanoChains))
	confirmedBlockSubmitters := make([]core.ConfirmedBlocksSubmitter, 0, len(appConfig.CardanoChains))

	for _, cardanoChainConfig := range appConfig.CardanoChains {
		indexerDB := indexerDbs[cardanoChainConfig.ChainID]

		cbs, err := bridge.NewConfirmedBlocksSubmitter(
			ctx, bridgeSubmitter, appConfig, db, indexerDB, cardanoChainConfig.ChainID, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create cardano block submitter for `%s`: %w", cardanoChainConfig.ChainID, err)
		}

		confirmedBlockSubmitters = append(confirmedBlockSubmitters, cbs)

		cco, err := chain.NewCardanoChainObserver(
			ctx, cardanoChainConfig, cardanoTxsProcessor, chainBalanceFetcher, db, indexerDB,
			logger.Named("cardano_chain_observer_"+cardanoChainConfig.ChainID))
		if err != nil {
			return nil, fmt.Errorf("failed to create cardano chain observer for `%s`: %w", cardanoChainConfig.ChainID, err)
		}

		cardanoChainObservers = append(cardanoChainObservers, cco)
	}

	return &OracleImpl{
		ctx:                      ctx,
		appConfig:                appConfig,
		cardanoTxsProcessor:      cardanoTxsProcessor,
		cardanoChainObservers:    cardanoChainObservers,
		expectedTxsFetcher:       expectedTxsFetcher,
		confirmedBlockSubmitters: confirmedBlockSubmitters,
		db:                       db,
		logger:                   logger,
	}, nil
}

func (o *OracleImpl) Start() error {
	o.logger.Debug("Starting Oracle")

	go o.cardanoTxsProcessor.Start()
	go o.expectedTxsFetcher.Start()

	for _, cbs := range o.confirmedBlockSubmitters {
		cbs.StartSubmit()
	}

	for _, co := range o.cardanoChainObservers {
		err := co.Start()
		if err != nil {
			return fmt.Errorf("failed to start observer for %s: %w", co.GetConfig().ChainID, err)
		}
	}

	o.errorCh = make(chan error, 1)
	go o.errorHandler()

	o.logger.Debug("Started Oracle")

	return nil
}

func (o *OracleImpl) Dispose() error {
	errs := make([]error, 0)

	for _, cco := range o.cardanoChainObservers {
		err := cco.Dispose()
		if err != nil {
			o.logger.Error("error while disposing cardano chain observer", "chainId", cco.GetConfig().ChainID, "err", err)
			errs = append(errs, fmt.Errorf("error while disposing cardano chain observer. chainId: %v, err: %w",
				cco.GetConfig().ChainID, err))
		}
	}

	err := o.db.Close()
	if err != nil {
		o.logger.Error("Failed to close oracle db", "err", err)
		errs = append(errs, fmt.Errorf("failed to close oracle db. err %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors while disposing oracle. errors: %w", errors.Join(errs...))
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

	for _, co := range o.cardanoChainObservers {
		go func(errChan <-chan error, origin string) {
		outsideloop:
			for {
				select {
				case err := <-errChan:
					if err != nil {
						o.logger.Error("chain observer error", "origin", origin, "err", err)
						if strings.Contains(err.Error(), errBlockSyncerFatal.Error()) {
							agg <- ErrorOrigin{
								err:    err,
								origin: origin,
							}

							break outsideloop
						}
					}
				case <-o.ctx.Done():
					break outsideloop
				}
			}
			o.logger.Debug("Exiting error handler", "origin", origin)
		}(co.ErrorCh(), co.GetConfig().ChainID)
	}

	select {
	case errorOrigin := <-agg:
		o.logger.Error("Cardano chain observer critical error", "origin", errorOrigin.origin, "err", errorOrigin.err)
		o.errorCh <- errorOrigin.err
	case <-o.ctx.Done():
	}
	o.logger.Debug("Exiting oracle error handler")
}
