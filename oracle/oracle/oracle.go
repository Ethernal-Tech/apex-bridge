package oracle

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/Ethernal-Tech/apex-bridge/oracle/bridge"
	"github.com/Ethernal-Tech/apex-bridge/oracle/chain"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle/database_access"
	"github.com/Ethernal-Tech/apex-bridge/oracle/processor"
	"github.com/Ethernal-Tech/apex-bridge/oracle/processor/failed_tx_processors"
	"github.com/Ethernal-Tech/apex-bridge/oracle/processor/tx_processors"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	indexerDb "github.com/Ethernal-Tech/cardano-infrastructure/indexer/db"
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
	indexerDbs               map[string]indexer.Database
	cardanoChainObservers    []core.CardanoChainObserver
	db                       core.Database
	expectedTxsFetcher       core.ExpectedTxsFetcher
	bridgeDataFetcher        *bridge.BridgeDataFetcherImpl
	bridgeSubmitter          core.BridgeSubmitter
	confirmedBlockSubmitters []core.ConfirmedBlocksSubmitter
	logger                   hclog.Logger

	errorCh chan error
}

var _ core.Oracle = (*OracleImpl)(nil)

func NewOracle(
	ctx context.Context,
	appConfig *core.AppConfig,
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater,
	logger hclog.Logger,
) (*OracleImpl, error) {
	db, err := database_access.NewDatabase(path.Join(appConfig.Settings.DbsPath, MainComponentName+".db"))
	if err != nil {
		return nil, fmt.Errorf("failed to open oracle database: %w", err)
	}

	wallet, err := ethtxhelper.NewEthTxWalletFromSecretManagerConfig(appConfig.Bridge.SecretsManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create blade wallet for oracle: %w", err)
	}

	bridgeSC := eth.NewOracleBridgeSmartContract(appConfig.Bridge.NodeUrl, appConfig.Bridge.SmartContractAddress)
	bridgeSCWithWallet, err := eth.NewOracleBridgeSmartContractWithWallet(
		appConfig.Bridge.NodeUrl, appConfig.Bridge.SmartContractAddress, wallet)
	if err != nil {
		return nil, fmt.Errorf("failed to create oracle bridge smart contract: %w", err)
	}

	bridgeDataFetcher := bridge.NewBridgeDataFetcher(ctx, bridgeSC, logger.Named("bridge_data_fetcher"))
	bridgeSubmitter := bridge.NewBridgeSubmitter(ctx, bridgeSCWithWallet, logger.Named("bridge_submitter"))

	expectedTxsFetcher := bridge.NewExpectedTxsFetcher(ctx, bridgeDataFetcher, appConfig, db, logger.Named("expected_txs_fetcher"))

	txProcessors := []core.CardanoTxProcessor{
		tx_processors.NewBatchExecutedProcessor(),
		tx_processors.NewBridgingRequestedProcessor(),
		// tx_processors.NewRefundExecutedProcessor(),
	}

	failedTxProcessors := []core.CardanoTxFailedProcessor{
		failed_tx_processors.NewBatchExecutionFailedProcessor(),
		// failed_tx_processors.NewRefundExecutionFailedProcessor(),
	}

	indexerDbs := make(map[string]indexer.Database, len(appConfig.CardanoChains))
	for _, cardanoChainConfig := range appConfig.CardanoChains {
		indexerDb, err := indexerDb.NewDatabaseInit("",
			path.Join(appConfig.Settings.DbsPath, cardanoChainConfig.ChainId+".db"))
		if err != nil {
			return nil, fmt.Errorf("failed to open oracle indexer db for `%s`: %w", cardanoChainConfig.ChainId, err)
		}

		indexerDbs[cardanoChainConfig.ChainId] = indexerDb
	}

	cardanoTxsProcessor := processor.NewCardanoTxsProcessor(ctx, appConfig, db, txProcessors, failedTxProcessors, bridgeSubmitter, indexerDbs, bridgingRequestStateUpdater, logger.Named("cardano_txs_processor"))

	cardanoChainObservers := make([]core.CardanoChainObserver, 0, len(appConfig.CardanoChains))
	confirmedBlockSubmitters := make([]core.ConfirmedBlocksSubmitter, 0, len(appConfig.CardanoChains))

	for _, cardanoChainConfig := range appConfig.CardanoChains {
		indexerDb := indexerDbs[cardanoChainConfig.ChainId]
		cbs, err := bridge.NewConfirmedBlocksSubmitter(
			ctx, bridgeSubmitter, appConfig, db, indexerDb, cardanoChainConfig.ChainId, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create cardano block submitter for `%s`: %w", cardanoChainConfig.ChainId, err)
		}

		confirmedBlockSubmitters = append(confirmedBlockSubmitters, cbs)

		cco, err := chain.NewCardanoChainObserver(
			ctx, cardanoChainConfig, cardanoTxsProcessor, db, indexerDb, bridgeDataFetcher, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create cardano chain observer for `%s`: %w", cardanoChainConfig.ChainId, err)
		}

		cardanoChainObservers = append(cardanoChainObservers, cco)
	}

	return &OracleImpl{
		ctx:                      ctx,
		appConfig:                appConfig,
		cardanoTxsProcessor:      cardanoTxsProcessor,
		cardanoChainObservers:    cardanoChainObservers,
		indexerDbs:               indexerDbs,
		expectedTxsFetcher:       expectedTxsFetcher,
		bridgeDataFetcher:        bridgeDataFetcher,
		bridgeSubmitter:          bridgeSubmitter,
		confirmedBlockSubmitters: confirmedBlockSubmitters,
		db:                       db,
		logger:                   logger,
	}, nil
}

func (o *OracleImpl) Start() error {
	o.logger.Debug("Starting Oracle")

	go o.cardanoTxsProcessor.Start()
	go o.expectedTxsFetcher.Start()

	for _, co := range o.cardanoChainObservers {
		err := co.Start()
		if err != nil {
			return fmt.Errorf("failed to start observer for %s: %w", co.GetConfig().ChainId, err)
		}
	}

	for _, cbs := range o.confirmedBlockSubmitters {
		cbs.StartSubmit()
	}

	o.errorCh = make(chan error, 1)
	go o.errorHandler()

	o.logger.Debug("Started Oracle")

	return nil
}

func (o *OracleImpl) Dispose() error {
	errs := make([]error, 0)
	for _, indexerDb := range o.indexerDbs {
		err := indexerDb.Close()
		if err != nil {
			o.logger.Error("Failed to close indexer db", "err", err)
			errs = append(errs, fmt.Errorf("failed to close indexer db. err %w", err))
		}
	}

	for _, cco := range o.cardanoChainObservers {
		err := cco.Dispose()
		if err != nil {
			o.logger.Error("error while disposing cardano chain observer", "chainId", cco.GetConfig().ChainId, "err", err)
			errs = append(errs, fmt.Errorf("error while disposing cardano chain observer. chainId: %v, err: %w",
				cco.GetConfig().ChainId, err))
		}
	}

	err := o.db.Close()
	if err != nil {
		o.logger.Error("Failed to close oracle db", "err", err)
		errs = append(errs, fmt.Errorf("failed to close oracle db. err %w", err))
	}

	close(o.errorCh)

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
	defer close(agg)

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
		}(co.ErrorCh(), co.GetConfig().ChainId)
	}

	select {
	case errorOrigin := <-agg:
		o.logger.Error("Cardano chain observer critical error", "origin", errorOrigin.origin, "err", errorOrigin.err)
		o.errorCh <- errorOrigin.err
	case <-o.ctx.Done():
	}
	o.logger.Debug("Exiting oracle error handler")
}
