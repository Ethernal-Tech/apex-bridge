package processor

import (
	"fmt"
	"os"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/oracle/core"

	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/hashicorp/go-hclog"
)

const (
	CheckUnprocessedTxsTickTimeMs = 1000
)

type CardanoBlockProcessorImpl struct {
	appConfig       *core.AppConfig
	db              core.CardanoBlockProcessorDb
	txProcessors    []core.CardanoTxProcessor
	claimsSubmitter core.ClaimsSubmitter
	logger          hclog.Logger
	closeCh         chan bool
}

var _ core.CardanoBlockProcessor = (*CardanoBlockProcessorImpl)(nil)

func NewCardanoBlockProcessor(
	appConfig *core.AppConfig,
	db core.CardanoBlockProcessorDb,
	txProcessors []core.CardanoTxProcessor,
	claimsSubmitter core.ClaimsSubmitter,
	logger hclog.Logger,
) *CardanoBlockProcessorImpl {

	return &CardanoBlockProcessorImpl{
		appConfig:       appConfig,
		db:              db,
		txProcessors:    txProcessors,
		claimsSubmitter: claimsSubmitter,
		logger:          logger,
		closeCh:         make(chan bool, 1),
	}
}

func (bp *CardanoBlockProcessorImpl) NewUnprocessedTxs(originChainId string, txs []*indexer.Tx) error {
	bp.logger.Debug("NewUnprocessedTxs", "txs", txs)

	var relevantTxs []*core.CardanoTx
	for _, tx := range txs {
		cardanoTx := &core.CardanoTx{
			OriginChainId: originChainId,
			Tx:            *tx,
		}

		for _, txProcessor := range bp.txProcessors {
			relevant, err := txProcessor.IsTxRelevant(cardanoTx, bp.appConfig)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to check if tx is relevant. error: %v\n", err)
				bp.logger.Error("Failed to check if tx is relevant", "err", err)
				continue
			}

			if relevant {
				relevantTxs = append(relevantTxs, cardanoTx)
				break
			}
		}
	}

	if len(relevantTxs) > 0 {
		bp.logger.Debug("Adding relevant txs to db", "txs", relevantTxs)
		err := bp.db.AddUnprocessedTxs(relevantTxs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to add unprocessed txs. error: %v\n", err)
			bp.logger.Error("Failed to add unprocessed txs", "err", err)

			return err
		}
	}

	return nil
}

func (bp *CardanoBlockProcessorImpl) Start() error {
	bp.logger.Debug("Starting CardanoBlockProcessor")
	ticker := time.NewTicker(CheckUnprocessedTxsTickTimeMs * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			bp.checkUnprocessedTxs()
		case <-bp.closeCh:
			return nil
		}
	}
}

func (bp *CardanoBlockProcessorImpl) Stop() error {
	bp.closeCh <- true
	bp.logger.Debug("Stopping CardanoBlockProcessor")
	return nil
}

func (bp *CardanoBlockProcessorImpl) checkUnprocessedTxs() {
	bp.logger.Debug("Checking unprocessed txs")

	unprocessedTxs, err := bp.db.GetUnprocessedTxs(bp.appConfig.Settings.MaxBridgingClaimsToGroup)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get unprocessed txs. error: %v\n", err)
		bp.logger.Error("Failed to get unprocessed txs", "err", err)
		return
	}

	if len(unprocessedTxs) == 0 {
		return
	}

	var processedTxs []*core.CardanoTx
	bridgeClaims := &core.BridgeClaims{}

	for _, unprocessedTx := range unprocessedTxs {
		for _, txProcessor := range bp.txProcessors {
			relevant, err := txProcessor.IsTxRelevant(unprocessedTx, bp.appConfig)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to check if tx is relevant. error: %v\n", err)
				bp.logger.Error("Failed to check if tx is relevant", "err", err)
				continue
			}

			if relevant {
				err := txProcessor.ValidateAndAddClaim(bridgeClaims, unprocessedTx, bp.appConfig)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to ValidateAndAddClaim. error: %v\n", err)
					bp.logger.Error("Failed to ValidateAndAddClaim", "err", err)
					continue
				}

				processedTxs = append(processedTxs, unprocessedTx)
				break
			}
		}
	}

	if bridgeClaims.Any() {
		bp.logger.Debug("Submitting bridge claims", "claims", bridgeClaims)
		err := bp.claimsSubmitter.SubmitClaims(bridgeClaims)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to submit claims. error: %v\n", err)
			bp.logger.Error("Failed to submit claims", "err", err)
			return
		}
	}

	if len(processedTxs) > 0 {
		bp.logger.Debug("Marking txs as processed", "txs", processedTxs)
		err := bp.db.MarkTxsAsProcessed(processedTxs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to mark txs as processed. error: %v\n", err)
			bp.logger.Error("Failed to mark txs as processed", "err", err)
			return
		}
	}
}
