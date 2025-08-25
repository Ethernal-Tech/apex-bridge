package oracle

import (
	"context"
	"errors"
	"fmt"

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
	"github.com/Ethernal-Tech/apex-bridge/validatorobserver"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/hashicorp/go-hclog"
	"go.etcd.io/bbolt"
)

type OracleImpl struct {
	ctx                      context.Context
	appConfig                *cCore.AppConfig
	cardanoTxsProcessor      cCore.TxsProcessor
	cardanoChainObservers    []core.CardanoChainObserver
	db                       core.Database
	expectedTxsFetcher       cCore.ExpectedTxsFetcher
	confirmedBlockSubmitters []cCore.ConfirmedBlocksSubmitter
	validatorSetObserver     validatorobserver.IValidatorSetObserver
	logger                   hclog.Logger
	indexers                 map[string]*indexer.BlockIndexer
}

var _ core.Oracle = (*OracleImpl)(nil)

func NewCardanoOracle(
	ctx context.Context,
	boltDB *bbolt.DB,
	typeRegister common.TypeRegister,
	appConfig *cCore.AppConfig,
	oracleBridgeSC eth.IOracleBridgeSmartContract,
	bridgeSubmitter cCore.BridgeSubmitter,
	indexerDbs map[string]indexer.Database,
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater,
	validatorSetObserver validatorobserver.IValidatorSetObserver,
	logger hclog.Logger,
) (*OracleImpl, error) {
	db := &databaseaccess.BBoltDatabase{}
	db.Init(boltDB, appConfig, typeRegister)

	bridgeDataFetcher := bridge.NewCardanoBridgeDataFetcher(ctx, oracleBridgeSC, logger.Named("bridge_data_fetcher"))

	expectedTxsFetcher := bridge.NewExpectedTxsFetcher(
		ctx, bridgeDataFetcher, appConfig, db, logger.Named("expected_txs_fetcher"))

	txProcessors := cardanotxsprocessor.NewTxProcessorsCollection(
		[]core.CardanoTxSuccessProcessor{
			successtxprocessors.NewBatchExecutedProcessor(logger),
			successtxprocessors.NewBridgingRequestedProcessor(logger),
			successtxprocessors.NewHotWalletIncrementProcessor(logger),
			// tx_processors.NewRefundExecutedProcessor(logger),
		},
		[]core.CardanoTxFailedProcessor{
			failedtxprocessors.NewBatchExecutionFailedProcessor(logger),
			// failed_tx_processors.NewRefundExecutionFailedProcessor(logger),
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
		bridgingRequestStateUpdater, validatorSetObserver, txsProcessorLogger)

	cardanoChainObservers := make([]core.CardanoChainObserver, 0, len(appConfig.CardanoChains))
	confirmedBlockSubmitters := make([]cCore.ConfirmedBlocksSubmitter, 0, len(appConfig.CardanoChains))

	indexers := make(map[string]*indexer.BlockIndexer, len(appConfig.CardanoChains))

	for _, cardanoChainConfig := range appConfig.CardanoChains {
		indexerDB := indexerDbs[cardanoChainConfig.ChainID]

		cbs, err := bridge.NewConfirmedBlocksSubmitter(
			bridgeSubmitter, appConfig, db, indexerDB, cardanoChainConfig.ChainID, logger)
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

		indexers[cardanoChainConfig.ChainID] = cco.GetIndexer()

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
		validatorSetObserver:     validatorSetObserver,
		logger:                   logger,
		indexers:                 indexers,
	}, nil
}

func (o *OracleImpl) GetIndexers() map[string]*indexer.BlockIndexer {
	return o.indexers
}

func (o *OracleImpl) Start() error {
	o.logger.Debug("Starting CardanoOracle")

	go o.cardanoTxsProcessor.Start()
	go o.expectedTxsFetcher.Start()

	for _, cbs := range o.confirmedBlockSubmitters {
		cbs.Start(o.ctx)
	}

	for _, co := range o.cardanoChainObservers {
		err := co.Start()
		if err != nil {
			return fmt.Errorf("failed to start observer for %s: %w", co.GetConfig().ChainID, err)
		}
	}

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
