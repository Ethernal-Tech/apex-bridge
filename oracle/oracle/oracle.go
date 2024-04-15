package oracle

import (
	"errors"
	"fmt"
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
	errBlockSyncerFatal              = errors.New("block syncer fatal error")
	errConfirmedBlocksSubmitterFatal = errors.New("confirmed blocks submitter fatal error")
)

type OracleImpl struct {
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
	closeCh chan bool
}

var _ core.Oracle = (*OracleImpl)(nil)

func NewOracle(appConfig *core.AppConfig, logger hclog.Logger) (*OracleImpl, error) {
	if err := common.CreateDirectoryIfNotExists(appConfig.Settings.DbsPath, 0660); err != nil {
		return nil, fmt.Errorf("failed to create directory for oracle database: %w", err)
	}

	db, err := database_access.NewDatabase(appConfig.Settings.DbsPath + MainComponentName + ".db")
	if err != nil {
		return nil, fmt.Errorf("failed to open oracle database: %w", err)
	}

	wallet, err := ethtxhelper.NewEthTxWalletFromSecretManager(appConfig.Bridge.SecretsManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create oracle wallet: %w", err)
	}

	bridgeSC := eth.NewOracleBridgeSmartContract(appConfig.Bridge.NodeUrl, appConfig.Bridge.SmartContractAddress)
	bridgeSCWithWallet, err := eth.NewOracleBridgeSmartContractWithWallet(
		appConfig.Bridge.NodeUrl, appConfig.Bridge.SmartContractAddress, wallet)
	if err != nil {
		return nil, fmt.Errorf("failed to create oracle bridge smart contract: %w", err)
	}

	bridgeDataFetcher := bridge.NewBridgeDataFetcher(bridgeSC, logger.Named("bridge_data_fetcher"))
	bridgeSubmitter := bridge.NewBridgeSubmitter(bridgeSCWithWallet, logger.Named("bridge_submitter"))

	expectedTxsFetcher := bridge.NewExpectedTxsFetcher(bridgeDataFetcher, appConfig, db, logger.Named("expected_txs_fetcher"))

	var txProcessors []core.CardanoTxProcessor
	txProcessors = append(txProcessors, tx_processors.NewBatchExecutedProcessor())
	txProcessors = append(txProcessors, tx_processors.NewBridgingRequestedProcessor())
	// txProcessors = append(txProcessors, tx_processors.NewRefundExecutedProcessor())

	var failedTxProcessors []core.CardanoTxFailedProcessor
	failedTxProcessors = append(failedTxProcessors, failed_tx_processors.NewBatchExecutionFailedProcessor())
	// failedTxProcessors = append(failedTxProcessors, failed_tx_processors.NewRefundExecutionFailedProcessor())

	indexerDbs := make(map[string]indexer.Database, len(appConfig.CardanoChains))
	for _, cardanoChainConfig := range appConfig.CardanoChains {
		indexerDb, err := indexerDb.NewDatabaseInit("",
			path.Join(appConfig.Settings.DbsPath, cardanoChainConfig.ChainId+".db"))
		if err != nil {
			return nil, fmt.Errorf("failed to open oracle indexer db for `%s`: %w", cardanoChainConfig.ChainId, err)
		}

		indexerDbs[cardanoChainConfig.ChainId] = indexerDb
	}

	cardanoTxsProcessor := processor.NewCardanoTxsProcessor(appConfig, db, txProcessors, failedTxProcessors, bridgeSubmitter, indexerDbs, logger.Named("cardano_txs_processor"))

	cardanoChainObservers := make([]core.CardanoChainObserver, 0, len(appConfig.CardanoChains))
	confirmedBlockSubmitters := make([]core.ConfirmedBlocksSubmitter, 0, len(appConfig.CardanoChains))

	for _, cardanoChainConfig := range appConfig.CardanoChains {
		indexerDb := indexerDbs[cardanoChainConfig.ChainId]
		cbs, err := bridge.NewConfirmedBlocksSubmitter(
			bridgeSubmitter, appConfig, db, indexerDb, cardanoChainConfig.ChainId, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create cardano block submitter for `%s`: %w", cardanoChainConfig.ChainId, err)
		}

		confirmedBlockSubmitters = append(confirmedBlockSubmitters, cbs)

		cco, err := chain.NewCardanoChainObserver(
			cardanoChainConfig, cardanoTxsProcessor, db, indexerDb, bridgeDataFetcher, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create cardano chain observer for `%s`: %w", cardanoChainConfig.ChainId, err)
		}

		cardanoChainObservers = append(cardanoChainObservers, cco)
	}

	return &OracleImpl{
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
		closeCh:                  make(chan bool, 1),
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

func (o *OracleImpl) Stop() error {
	o.logger.Debug("Stopping Oracle")

	for _, co := range o.cardanoChainObservers {
		err := co.Stop()
		if err != nil {
			o.logger.Error("Failed to stop cardano chain observer", "chain", co.GetConfig().ChainId, "err", err)
		}
	}

	for _, cbs := range o.confirmedBlockSubmitters {
		err := cbs.Dispose()
		if err != nil {
			o.logger.Error("Failed to stop block submitter cardano", "err", err)
			return err
		}
	}

	for _, indexerDb := range o.indexerDbs {
		err := indexerDb.Close()
		if err != nil {
			o.logger.Error("Failed to close indexer db", "err", err)
		}
	}

	o.cardanoTxsProcessor.Stop()
	o.expectedTxsFetcher.Stop()
	o.bridgeSubmitter.Dispose()
	o.bridgeDataFetcher.Dispose()
	o.db.Close()

	close(o.errorCh)
	close(o.closeCh)

	o.logger.Debug("Stopped Oracle")

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
		go func(errChan <-chan error, closeChan <-chan bool, origin string) {
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
				case <-closeChan:
					break outsideloop
				}
			}
			o.logger.Debug("Exiting error handler", "origin", origin)
		}(co.ErrorCh(), o.closeCh, co.GetConfig().ChainId)
	}

	for _, cbs := range o.confirmedBlockSubmitters {
		go func(errChan <-chan error, closeChan <-chan bool, origin string) {
		outsideloop:
			for {
				select {
				case err := <-errChan:
					if err != nil {
						o.logger.Error("chain confirmed block submitter error", "origin", origin, "err", err)
						if strings.Contains(err.Error(), errConfirmedBlocksSubmitterFatal.Error()) {
							agg <- ErrorOrigin{
								err:    err,
								origin: origin,
							}
							break outsideloop
						}
					}
				case <-closeChan:
					break outsideloop
				}
			}
			o.logger.Debug("Exiting error handler", "origin", origin)
		}(cbs.ErrorCh(), o.closeCh, cbs.GetChainId())
	}

	select {
	case errorOrigin := <-agg:
		o.logger.Error("Cardano chain observer critical error", "origin", errorOrigin.origin, "err", errorOrigin.err)
		o.errorCh <- errorOrigin.err
	case <-o.closeCh:
	}
	o.logger.Debug("Exiting oracle error handler")
}
