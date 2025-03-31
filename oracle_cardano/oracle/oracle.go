package oracle

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/bridge"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/chain"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	databaseaccess "github.com/Ethernal-Tech/apex-bridge/oracle_cardano/database_access"
	failedtxprocessors "github.com/Ethernal-Tech/apex-bridge/oracle_cardano/processor/tx_processors/failed"
	successtxprocessors "github.com/Ethernal-Tech/apex-bridge/oracle_cardano/processor/tx_processors/success"
	cardanotxsprocessor "github.com/Ethernal-Tech/apex-bridge/oracle_cardano/processor/txs_processor"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	txsprocessor "github.com/Ethernal-Tech/apex-bridge/oracle_common/processor/txs_processor"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/hashicorp/go-hclog"
	"go.etcd.io/bbolt"
)

var (
	errBlockSyncerFatal = errors.New("block syncer fatal error")
)

type OracleImpl struct {
	ctx                      context.Context
	appConfig                *cCore.AppConfig
	cardanoTxsProcessor      cCore.TxsProcessor
	cardanoChainObservers    []core.CardanoChainObserver
	db                       core.Database
	expectedTxsFetcher       cCore.ExpectedTxsFetcher
	confirmedBlockSubmitters []cCore.ConfirmedBlocksSubmitter
	chainInfos               map[string]*chain.CardanoChainInfo
	logger                   hclog.Logger

	errorCh chan error
}

var _ core.Oracle = (*OracleImpl)(nil)

func NewCardanoOracle(
	ctx context.Context,
	boltDB *bbolt.DB,
	typeRegister common.TypeRegister,
	appConfig *cCore.AppConfig,
	oracleBridgeSC eth.IOracleBridgeSmartContract,
	bridgeSubmitter core.BridgeSubmitter,
	indexerDbs map[string]indexer.Database,
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater,
	logger hclog.Logger,
) (*OracleImpl, error) {
	db := &databaseaccess.BBoltDatabase{}
	db.Init(boltDB, appConfig, typeRegister)

	bridgeDataFetcher := bridge.NewCardanoBridgeDataFetcher(ctx, oracleBridgeSC, logger.Named("bridge_data_fetcher"))

	expectedTxsFetcher := bridge.NewExpectedTxsFetcher(
		ctx, bridgeDataFetcher, appConfig, db, logger.Named("expected_txs_fetcher"))

	chainInfos := make(map[string]*chain.CardanoChainInfo, len(appConfig.CardanoChains))

	for _, cc := range appConfig.CardanoChains {
		info := chain.NewCardanoChainInfo(cc)

		if err := info.Populate(ctx); err != nil {
			return nil, err
		}

		chainInfos[cc.ChainID] = info
	}

	refundRequestProcessor := successtxprocessors.NewRefundRequestProcessor(logger, chainInfos)
	txProcessors := cardanotxsprocessor.NewTxProcessorsCollection(
		[]core.CardanoTxSuccessProcessor{
			successtxprocessors.NewBatchExecutedProcessor(logger),
			successtxprocessors.NewBridgingRequestedProcessor(refundRequestProcessor, logger),
			successtxprocessors.NewHotWalletIncrementProcessor(logger),
			refundRequestProcessor,
		},
		[]core.CardanoTxFailedProcessor{
			failedtxprocessors.NewBatchExecutionFailedProcessor(logger),
		},
	)

	txsProcessorLogger := logger.Named("cardano_txs_processor")

	cardanoTxsReceiver := cardanotxsprocessor.NewCardanoTxsReceiverImpl(
		appConfig, db, txProcessors, bridgingRequestStateUpdater, txsProcessorLogger,
	)

	cardanoStateProcessor := cardanotxsprocessor.NewCardanoStateProcessor(
		ctx, appConfig, db, txProcessors,
		indexerDbs, txsProcessorLogger,
	)

	cardanoTxsProcessor := txsprocessor.NewTxsProcessorImpl(
		ctx, appConfig, cardanoStateProcessor, bridgeDataFetcher, bridgeSubmitter,
		bridgingRequestStateUpdater, txsProcessorLogger)

	cardanoChainObservers := make([]core.CardanoChainObserver, 0, len(appConfig.CardanoChains))
	confirmedBlockSubmitters := make([]cCore.ConfirmedBlocksSubmitter, 0, len(appConfig.CardanoChains))

	for _, cardanoChainConfig := range appConfig.CardanoChains {
		indexerDB := indexerDbs[cardanoChainConfig.ChainID]

		cbs, err := bridge.NewConfirmedBlocksSubmitter(
			ctx, bridgeSubmitter, appConfig, indexerDB, cardanoChainConfig.ChainID, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create cardano block submitter for `%s`: %w", cardanoChainConfig.ChainID, err)
		}

		confirmedBlockSubmitters = append(confirmedBlockSubmitters, cbs)

		cco, err := chain.NewCardanoChainObserver(
			ctx, cardanoChainConfig, cardanoTxsReceiver, db, indexerDB,
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
		chainInfos:               chainInfos,
		db:                       db,
		logger:                   logger,
	}, nil
}

func (o *OracleImpl) Start() error {
	o.logger.Debug("Starting CardanoOracle")

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

	o.logger.Debug("Started CardanoOracle")

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

	if len(errs) > 0 {
		return fmt.Errorf("errors while disposing oracle_cardano. errors: %w", errors.Join(errs...))
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
	o.logger.Debug("Exiting oracle_cardano error handler")
}
