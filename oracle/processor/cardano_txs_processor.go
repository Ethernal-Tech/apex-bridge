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

type CardanoTxsProcessorImpl struct {
	appConfig       *core.AppConfig
	db              core.CardanoTxsProcessorDb
	txProcessors    []core.CardanoTxProcessor
	claimsSubmitter core.ClaimsSubmitter
	logger          hclog.Logger
	closeCh         chan bool
}

var _ core.CardanoTxsProcessor = (*CardanoTxsProcessorImpl)(nil)

func NewCardanoTxsProcessor(
	appConfig *core.AppConfig,
	db core.CardanoTxsProcessorDb,
	txProcessors []core.CardanoTxProcessor,
	claimsSubmitter core.ClaimsSubmitter,
	logger hclog.Logger,
) *CardanoTxsProcessorImpl {

	return &CardanoTxsProcessorImpl{
		appConfig:       appConfig,
		db:              db,
		txProcessors:    txProcessors,
		claimsSubmitter: claimsSubmitter,
		logger:          logger,
		closeCh:         make(chan bool, 1),
	}
}

func (bp *CardanoTxsProcessorImpl) NewUnprocessedTxs(originChainId string, txs []*indexer.Tx) error {
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

func (bp *CardanoTxsProcessorImpl) Start() error {
	bp.logger.Debug("Starting CardanoTxsProcessor")
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

func (bp *CardanoTxsProcessorImpl) Stop() error {
	bp.closeCh <- true
	bp.logger.Debug("Stopping CardanoTxsProcessor")
	return nil
}

func (bp *CardanoTxsProcessorImpl) checkUnprocessedTxs() {
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
	var invalidTxHashes []string
	bridgeClaims := &core.BridgeClaims{}

	for _, unprocessedTx := range unprocessedTxs {
		var txProcessed = false
		for _, txProcessor := range bp.txProcessors {
			relevant, err := txProcessor.IsTxRelevant(unprocessedTx, bp.appConfig)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to check if tx is relevant. error: %v\n", err)
				bp.logger.Error("Failed to check if tx is relevant", "tx", unprocessedTx, "err", err)
				continue
			}

			if relevant {
				err := txProcessor.ValidateAndAddClaim(bridgeClaims, unprocessedTx, bp.appConfig)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to ValidateAndAddClaim. error: %v\n", err)
					bp.logger.Error("Failed to ValidateAndAddClaim", "tx", unprocessedTx, "err", err)
					continue
				}

				processedTxs = append(processedTxs, unprocessedTx)
				txProcessed = true
				break
			}
		}

		if !txProcessed {
			// transfer an unprocessed tx to invalid txs bucket, to keep as history
			invalidTxHashes = append(invalidTxHashes, unprocessedTx.Hash)
			// and mark it as processed to prevent it from being fetched again as unprocessed
			processedTxs = append(processedTxs, unprocessedTx)
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

	if len(invalidTxHashes) > 0 {
		bp.logger.Debug("Saving invalid txs", "txs", invalidTxHashes)
		err := bp.db.AddInvalidTxHashes(invalidTxHashes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save invalid txs. error: %v\n", err)
			bp.logger.Error("Failed to save invalid txs", "err", err)
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
