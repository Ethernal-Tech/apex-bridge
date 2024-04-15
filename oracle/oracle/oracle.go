package oracle

import (
	"errors"
	"fmt"
	"os"
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
	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
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
	cardanoChainObservers    map[string]core.CardanoChainObserver
	db                       core.Database
	expectedTxsFetcher       core.ExpectedTxsFetcher
	bridgeDataFetcher        *bridge.BridgeDataFetcherImpl
	bridgeSubmitter          core.BridgeSubmitter
	confirmedBlockSubmitters map[string]core.ConfirmedBlocksSubmitter
	logger                   hclog.Logger

	errorCh chan error
	closeCh chan bool
}

var _ core.Oracle = (*OracleImpl)(nil)

func NewOracle(appConfig *core.AppConfig) *OracleImpl {

	oracle := &OracleImpl{}

	logger, err := logger.NewLogger(logger.LoggerConfig{
		LogLevel:      hclog.Level(appConfig.Settings.LogLevel),
		JSONLogFormat: false,
		AppendFile:    true,
		LogFilePath:   appConfig.Settings.LogsPath + MainComponentName,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return nil
	}

	if err := common.CreateDirectoryIfNotExists(appConfig.Settings.DbsPath); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		logger.Error("Create directory failed", "err", err)
		return nil
	}

	db, err := database_access.NewDatabase(appConfig.Settings.DbsPath + MainComponentName + ".db")
	if db == nil || err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		logger.Error("Open database failed", "err", err)
		return nil
	}

	wallet, err := ethtxhelper.NewEthTxWalletFromSecretManager(appConfig.Bridge.SecretsManager)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while creating wallet for bridge: %v\n", err)
		logger.Error("error while creating wallet for bridge", "err", err)
		return nil
	}

	bridgeSC := eth.NewOracleBridgeSmartContract(appConfig.Bridge.NodeUrl, appConfig.Bridge.SmartContractAddress)
	bridgeSCWithWallet, err := eth.NewOracleBridgeSmartContractWithWallet(
		appConfig.Bridge.NodeUrl, appConfig.Bridge.SmartContractAddress, wallet)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		logger.Error("Failed to create OracleBridgeSmartContractWithWallet", "err", err)
		return nil
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

	var indexerDbs map[string]indexer.Database = make(map[string]indexer.Database)
	for _, cardanoChainConfig := range appConfig.CardanoChains {
		indexerDb, err := indexerDb.NewDatabaseInit("", appConfig.Settings.DbsPath+cardanoChainConfig.ChainId+".db")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			logger.Error("Failed to open indexer database failed", "chainId", cardanoChainConfig.ChainId, "err", err)
			return nil
		}

		indexerDbs[cardanoChainConfig.ChainId] = indexerDb
	}

	cardanoTxsProcessor := processor.NewCardanoTxsProcessor(appConfig, db, txProcessors, failedTxProcessors, bridgeSubmitter, indexerDbs, logger.Named("cardano_txs_processor"))

	var cardanoChainObservers map[string]core.CardanoChainObserver = make(map[string]core.CardanoChainObserver)
	var confirmedBlockSubmitters map[string]core.ConfirmedBlocksSubmitter = make(map[string]core.ConfirmedBlocksSubmitter)

	for _, cardanoChainConfig := range appConfig.CardanoChains {
		indexerDb := indexerDbs[cardanoChainConfig.ChainId]
		cbs, err := bridge.NewConfirmedBlocksSubmitter(bridgeSubmitter, appConfig, db, indexerDb, cardanoChainConfig.ChainId, logger.Named("confirmed_blocks_submitter_"+cardanoChainConfig.ChainId))
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create cardano block submitter for chain: %v error: %v\n", cardanoChainConfig.ChainId, err)
			logger.Error("failed to create cardano block submitter for chain", "chainId", cardanoChainConfig.ChainId, "error:", err)
			return nil
		}

		confirmedBlockSubmitters[cardanoChainConfig.ChainId] = cbs

		cco := chain.NewCardanoChainObserver(appConfig.Settings, cardanoChainConfig, cardanoTxsProcessor, db, indexerDb, bridgeDataFetcher)
		if cco == nil {
			fmt.Fprintf(os.Stderr, "failed to create cardano chain observer for chain: %v\n", cardanoChainConfig.ChainId)
			logger.Error("failed to create cardano chain observer for chain", "chainId", cardanoChainConfig.ChainId)
			return nil
		}

		cardanoChainObservers[cardanoChainConfig.ChainId] = cco
	}

	oracle.appConfig = appConfig
	oracle.cardanoTxsProcessor = cardanoTxsProcessor
	oracle.cardanoChainObservers = cardanoChainObservers
	oracle.indexerDbs = indexerDbs
	oracle.expectedTxsFetcher = expectedTxsFetcher
	oracle.bridgeDataFetcher = bridgeDataFetcher
	oracle.bridgeSubmitter = bridgeSubmitter
	oracle.confirmedBlockSubmitters = confirmedBlockSubmitters
	oracle.db = db
	oracle.logger = logger
	oracle.closeCh = make(chan bool, 1)

	return oracle
}

func (o *OracleImpl) Start() error {
	o.logger.Debug("Starting Oracle")

	go o.cardanoTxsProcessor.Start()
	go o.expectedTxsFetcher.Start()

	for _, co := range o.cardanoChainObservers {
		err := co.Start()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start cardano chain observer: %v. error: %v\n", co.GetConfig().ChainId, err)
			o.logger.Error("Failed to start cardano chain observer", "err", err)
			return err
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
			fmt.Fprintf(os.Stderr, "Failed to stop cardano chain observer: %v. error: %v\n", co.GetConfig().ChainId, err)
			o.logger.Error("Failed to stop cardano chain observer", "err", err)
		}
	}

	for _, cbs := range o.confirmedBlockSubmitters {
		err := cbs.Dispose()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to stop block submitter. error: %v\n", err)
			o.logger.Error("Failed to stop block submitter cardano", "err", err)
			return err
		}
	}

	for _, indexerDb := range o.indexerDbs {
		err := indexerDb.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
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
		fmt.Fprintf(os.Stderr, "%v cardano chain observer critical error: %v\n", errorOrigin.origin, errorOrigin.err)
		o.logger.Error("Cardano chain observer critical error", "origin", errorOrigin.origin, "err", errorOrigin.err)
		o.errorCh <- errorOrigin.err
	case <-o.closeCh:
	}
	o.logger.Debug("Exiting oracle error handler")
}
