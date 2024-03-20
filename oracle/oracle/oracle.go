package oracle

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/oracle/bridge"
	"github.com/Ethernal-Tech/apex-bridge/oracle/chain"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle/database_access"
	"github.com/Ethernal-Tech/apex-bridge/oracle/processor"
	"github.com/Ethernal-Tech/apex-bridge/oracle/processor/failed_tx_processors"
	"github.com/Ethernal-Tech/apex-bridge/oracle/processor/tx_processors"
	"github.com/Ethernal-Tech/apex-bridge/oracle/utils"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
	"github.com/hashicorp/go-hclog"
)

const (
	MainComponentName = "oracle"
)

var (
	errBlockSyncerFatal = errors.New("block syncer fatal error")
)

type OracleImpl struct {
	appConfig             *core.AppConfig
	cardanoTxsProcessor   core.CardanoTxsProcessor
	cardanoChainObservers map[string]core.CardanoChainObserver
	db                    core.Database
	bridgeDataFetcher     *bridge.BridgeDataFetcherImpl
	claimsSubmitter       core.ClaimsSubmitter
	logger                hclog.Logger

	errorCh chan error
	closeCh chan bool
}

var _ core.Oracle = (*OracleImpl)(nil)

func NewOracle(appConfig *core.AppConfig, initialUtxos *core.InitialUtxos) *OracleImpl {

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

	if err := utils.CreateDirectoryIfNotExists(appConfig.Settings.DbsPath); err != nil {
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

	bridgeDataFetcher := bridge.NewBridgeDataFetcher(appConfig, db, logger.Named("bridge_data_fetcher"))
	claimsSubmitter := bridge.NewClaimsSubmitter(appConfig, logger.Named("claims_submitter"))

	var txProcessors []core.CardanoTxProcessor
	txProcessors = append(txProcessors, tx_processors.NewBatchExecutedProcessor())
	txProcessors = append(txProcessors, tx_processors.NewBridgingRequestedProcessor())
	// txProcessors = append(txProcessors, tx_processors.NewRefundExecutedProcessor())

	var failedTxProcessors []core.CardanoTxFailedProcessor
	failedTxProcessors = append(failedTxProcessors, failed_tx_processors.NewBatchExecutionFailedProcessor())
	// failedTxProcessors = append(failedTxProcessors, failed_tx_processors.NewRefundExecutionFailedProcessor())

	getCardanoChainObserverDb := func(chainId string) indexer.Database {
		cco := oracle.cardanoChainObservers[chainId]
		if cco != nil {
			return cco.GetDb()
		}

		return nil
	}

	cardanoTxsProcessor := processor.NewCardanoTxsProcessor(appConfig, db, txProcessors, failedTxProcessors, claimsSubmitter, getCardanoChainObserverDb, logger.Named("cardano_txs_processor"))

	var cardanoChainObservers map[string]core.CardanoChainObserver = make(map[string]core.CardanoChainObserver)

	for _, cardanoChainConfig := range appConfig.CardanoChains {
		initialUtxosForChain := (*initialUtxos)[cardanoChainConfig.ChainId]
		cco := chain.NewCardanoChainObserver(appConfig.Settings, cardanoChainConfig, initialUtxosForChain, cardanoTxsProcessor, db, bridgeDataFetcher)
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
	oracle.bridgeDataFetcher = bridgeDataFetcher
	oracle.claimsSubmitter = claimsSubmitter
	oracle.db = db
	oracle.logger = logger
	oracle.closeCh = make(chan bool, 1)

	return oracle
}

func (o *OracleImpl) Start() error {
	o.logger.Debug("Starting Oracle")

	go o.cardanoTxsProcessor.Start()
	go o.bridgeDataFetcher.Start()

	for _, co := range o.cardanoChainObservers {
		err := co.Start()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start cardano chain observer: %v. error: %v\n", co.GetConfig().ChainId, err)
			o.logger.Error("Failed to start cardano chain observer", "err", err)
			return err
		}
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

	o.cardanoTxsProcessor.Stop()
	o.bridgeDataFetcher.Stop()
	o.claimsSubmitter.Dispose()
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

	select {
	case errorOrigin := <-agg:
		fmt.Fprintf(os.Stderr, "%v cardano chain observer critical error: %v\n", errorOrigin.origin, errorOrigin.err)
		o.logger.Error("Cardano chain observer critical error", "origin", errorOrigin.origin, "err", errorOrigin.err)
		o.errorCh <- errorOrigin.err
	case <-o.closeCh:
	}
	o.logger.Debug("Exiting oracle error handler")
}
