package oracle

import (
	"context"
	"errors"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	oChain "github.com/Ethernal-Tech/apex-bridge/oracle_common/chain"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	txsprocessor "github.com/Ethernal-Tech/apex-bridge/oracle_common/processor/txs_processor"
	"github.com/Ethernal-Tech/apex-bridge/oracle_solana/bridge"
	sol_chain "github.com/Ethernal-Tech/apex-bridge/oracle_solana/chain"
	"github.com/Ethernal-Tech/apex-bridge/oracle_solana/core"
	databaseaccess "github.com/Ethernal-Tech/apex-bridge/oracle_solana/database_access"
	failedtxprocessors "github.com/Ethernal-Tech/apex-bridge/oracle_solana/processor/tx_processors/failed"
	successtxprocessors "github.com/Ethernal-Tech/apex-bridge/oracle_solana/processor/tx_processors/success"
	soltxprocessor "github.com/Ethernal-Tech/apex-bridge/oracle_solana/processor/txs_processor"
	store "github.com/Ethernal-Tech/solana-infrastructure/tracker/store"
	"github.com/hashicorp/go-hclog"
	"go.etcd.io/bbolt"
)

type OracleImpl struct {
	ctx                      context.Context
	appConfig                *oCore.AppConfig
	solanaTxsProcessor       oCore.TxsProcessor
	expectedTxsFetcher       oCore.ExpectedTxsFetcher
	solanaChainObservers     []core.SolanaChainObserver
	confirmedBlockSubmitters []oCore.ConfirmedBlocksSubmitter
	logger                   hclog.Logger
}

var _ core.Oracle = (*OracleImpl)(nil)

func NewSolanaOracle(
	ctx context.Context,
	boltDB *bbolt.DB,
	typeRegister common.TypeRegister,
	appConfig *oCore.AppConfig,
	cardanoChainInfos map[string]*oChain.CardanoChainInfo,
	oracleBridgeSC eth.IOracleBridgeSmartContract,
	bridgeSubmitter oCore.BridgeSubmitter,
	logger hclog.Logger,
	indexerDbs map[string]store.StorageHandler,
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater,
) (*OracleImpl, error) {
	db := &databaseaccess.BBoltDatabase{}
	db.Init(boltDB, appConfig, typeRegister)

	bridgeDataFetcher := bridge.NewSolanaBridgeDataFetcher(
		ctx, oracleBridgeSC, indexerDbs, logger.Named("solana_bridge_data_fetcher"))

	expectedTxsFetcher := bridge.NewExpectedTxsFetcher(
		ctx, bridgeDataFetcher, appConfig, db, logger.Named("solana_expected_txs_fetcher"))

	var (
		refundRequestProcessor core.SolanaTxSuccessRefundProcessor = successtxprocessors.NewRefundDisabledProcessor()

		successProcessors = []core.SolanaTxSuccessProcessor{}
	)

	if appConfig.RefundEnabled {
		refundRequestProcessor = successtxprocessors.NewRefundRequestProcessorSkyline(logger)
		successProcessors = append(successProcessors, refundRequestProcessor)
	}

	successProcessors = append(successProcessors,
		successtxprocessors.NewSolanaBatchExecutedProcessor(logger),
		successtxprocessors.NewSolanaHotWalletIncrementProcessor(logger),
		successtxprocessors.NewSolanaBridgingRequestedProcessor(refundRequestProcessor, logger, cardanoChainInfos),
	)

	failedProcessors := []core.SolanaTxFailedProcessor{
		failedtxprocessors.NewSolanaBatchExecutionFailedProcessor(logger),
	}

	txProcessors := soltxprocessor.NewTxProcessorsCollection(
		successProcessors,
		failedProcessors,
	)

	solanaTxsReceiver := soltxprocessor.NewSolTxsReceiver(
		appConfig,
		db,
		txProcessors,
		bridgingRequestStateUpdater,
		logger.Named("solana_txs_receiver"),
	)

	solanaStateProcessor := soltxprocessor.NewSolStateProcessor(
		ctx,
		appConfig,
		db,
		txProcessors,
		indexerDbs,
		logger.Named("solana_state_processor"),
	)

	solanaTxsProcessor := txsprocessor.NewTxsProcessorImpl(
		ctx,
		appConfig,
		solanaStateProcessor,
		bridgeDataFetcher,
		bridgeSubmitter,
		bridgingRequestStateUpdater,
		logger.Named("solana_txs_processor"),
	)

	solanaChainObservers := make([]core.SolanaChainObserver, 0, len(appConfig.SolanaChains))
	confirmedBlockSubmitters := make([]oCore.ConfirmedBlocksSubmitter, 0, len(appConfig.SolanaChains))

	for _, solanaChainConfig := range appConfig.SolanaChains {
		indexerDB := indexerDbs[solanaChainConfig.ChainID]

		cbs, err := bridge.NewConfirmedBlocksSubmitter(
			bridgeSubmitter, appConfig, db, indexerDB, solanaChainConfig.ChainID, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create solana block submitter for `%s`: %w", solanaChainConfig.ChainID, err)
		}

		confirmedBlockSubmitters = append(confirmedBlockSubmitters, cbs)

		solanaChainObserver, err := sol_chain.NewSolanaChainObserver(
			solanaChainConfig,
			solanaTxsReceiver,
			indexerDB,
			logger.Named("solana_chain_observer_"+solanaChainConfig.ChainID),
		)
		if err != nil {
			return nil, err
		}

		solanaChainObservers = append(solanaChainObservers, solanaChainObserver)
	}

	return &OracleImpl{
		ctx:                      ctx,
		appConfig:                appConfig,
		solanaChainObservers:     solanaChainObservers,
		confirmedBlockSubmitters: confirmedBlockSubmitters,
		logger:                   logger,
		solanaTxsProcessor:       solanaTxsProcessor,
		expectedTxsFetcher:       expectedTxsFetcher,
	}, nil
}

func (o *OracleImpl) Start() error {
	o.logger.Debug("Starting SolanaOracle")

	go o.solanaTxsProcessor.Start()
	go o.expectedTxsFetcher.Start()

	for _, cbs := range o.confirmedBlockSubmitters {
		cbs.Start(o.ctx)
	}

	for _, solanaChainObserver := range o.solanaChainObservers {
		err := solanaChainObserver.Start()
		if err != nil {
			return fmt.Errorf("failed to start solana chain observer for %s: %w", solanaChainObserver.GetConfig().ChainID, err)
		}
	}

	o.logger.Debug("Started SolanaOracle")

	return nil
}

func (o *OracleImpl) Dispose() error {
	errs := make([]error, 0)

	for _, solanaChainObserver := range o.solanaChainObservers {
		err := solanaChainObserver.Dispose()
		if err != nil {
			chainID := solanaChainObserver.GetConfig().ChainID

			o.logger.Error("error while disposing solana chain observer", "chainId", chainID, "err", err)
			errs = append(errs,
				fmt.Errorf("error while disposing solana chain observer. chainId: %v, err: %w", chainID, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors while disposing oracle_solana. errors: %w", errors.Join(errs...))
	}

	return nil
}
