package oracle

import (
	"fmt"
	"os"

	"github.com/Ethernal-Tech/apex-bridge/oracle/bridge"
	"github.com/Ethernal-Tech/apex-bridge/oracle/chain"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle/database_access"
	"github.com/Ethernal-Tech/apex-bridge/oracle/processor"
	"github.com/Ethernal-Tech/apex-bridge/oracle/processor/tx_processors"
	"github.com/Ethernal-Tech/apex-bridge/oracle/utils"
	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
	"github.com/hashicorp/go-hclog"
)

const (
	MainComponentName = "oracle"
)

type OracleImpl struct {
	appConfig             *core.AppConfig
	cardanoTxsProcessor   core.CardanoTxsProcessor
	cardanoChainObservers []core.CardanoChainObserver
	db                    core.Database
	logger                hclog.Logger

	// TODO: implement critical error handling
	errorCh chan error
}

var _ core.Oracle = (*OracleImpl)(nil)

func NewOracle(appConfig *core.AppConfig, initialUtxos *core.InitialUtxos) *OracleImpl {
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
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		logger.Error("Open database failed", "err", err)
		return nil
	}

	var txProcessors []core.CardanoTxProcessor
	txProcessors = append(txProcessors, tx_processors.NewBatchExecutedProcessor())
	txProcessors = append(txProcessors, tx_processors.NewBridgingRequestedProcessor())
	// txProcessors = append(txProcessors, tx_processors.NewRefundExecutedProcessor())

	claimsSubmitter := bridge.NewClaimsSubmitter(logger.Named("claims_submitter"))

	cardanoTxsProcessor := processor.NewCardanoTxsProcessor(appConfig, db, txProcessors, claimsSubmitter, logger.Named("cardano_txs_processor"))

	var cardanoChainObservers []core.CardanoChainObserver

	for _, cardanoChainConfig := range appConfig.CardanoChains {
		initialUtxosForChain := (*initialUtxos)[cardanoChainConfig.ChainId]
		cardanoChainObservers = append(
			cardanoChainObservers,
			chain.NewCardanoChainObserver(appConfig.Settings, cardanoChainConfig, initialUtxosForChain, cardanoTxsProcessor),
		)
	}

	return &OracleImpl{
		appConfig:             appConfig,
		cardanoTxsProcessor:   cardanoTxsProcessor,
		cardanoChainObservers: cardanoChainObservers,
		db:                    db,
		logger:                logger,
		errorCh:               make(chan error, 1),
	}
}

func (o *OracleImpl) Start() error {
	go o.cardanoTxsProcessor.Start()

	for _, co := range o.cardanoChainObservers {
		err := co.Start()
		if err != nil {
			// TODO: handle retry start
			fmt.Fprintf(os.Stderr, "Failed to start cardano chain observer: %v. error: %v\n", co.GetConfig().ChainId, err)
			o.logger.Error("Failed to start cardano chain observer", "err", err)
		}
	}

	return nil
}

func (o *OracleImpl) Stop() error {
	for _, co := range o.cardanoChainObservers {
		err := co.Stop()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to stop cardano chain observer: %v. error: %v\n", co.GetConfig().ChainId, err)
			o.logger.Error("Failed to stop cardano chain observer", "err", err)
		}
	}

	o.cardanoTxsProcessor.Stop()
	o.db.Close()

	close(o.errorCh)

	return nil
}

func (o *OracleImpl) ErrorCh() <-chan error {
	return o.errorCh
}
