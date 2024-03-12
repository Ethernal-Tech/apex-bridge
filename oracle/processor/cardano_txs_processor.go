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
	TickTimeMs                  = 2000
	LastExpectedBlockSlotOffset = 5
)

type CardanoTxsProcessorImpl struct {
	appConfig                 *core.AppConfig
	db                        core.CardanoTxsProcessorDb
	txProcessors              []core.CardanoTxProcessor
	failedTxProcessors        []core.CardanoTxFailedProcessor
	claimsSubmitter           core.ClaimsSubmitter
	getCardanoChainObserverDb GetCardanoChainObserverDbCallback
	logger                    hclog.Logger
	closeCh                   chan bool
}

type GetCardanoChainObserverDbCallback = func(chainId string) indexer.Database

var _ core.CardanoTxsProcessor = (*CardanoTxsProcessorImpl)(nil)

func NewCardanoTxsProcessor(
	appConfig *core.AppConfig,
	db core.CardanoTxsProcessorDb,
	txProcessors []core.CardanoTxProcessor,
	failedTxProcessors []core.CardanoTxFailedProcessor,
	claimsSubmitter core.ClaimsSubmitter,
	getCardanoChainObserverDb GetCardanoChainObserverDbCallback,
	logger hclog.Logger,
) *CardanoTxsProcessorImpl {

	return &CardanoTxsProcessorImpl{
		appConfig:                 appConfig,
		db:                        db,
		txProcessors:              txProcessors,
		failedTxProcessors:        failedTxProcessors,
		claimsSubmitter:           claimsSubmitter,
		getCardanoChainObserverDb: getCardanoChainObserverDb,
		logger:                    logger,
		closeCh:                   make(chan bool, 1),
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

	timerTime := TickTimeMs * time.Millisecond
	timer := time.NewTimer(timerTime)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			bp.generateClaims()
		case <-bp.closeCh:
			return nil
		}

		timer.Reset(timerTime)
	}
}

func (bp *CardanoTxsProcessorImpl) Stop() error {
	bp.logger.Debug("Stopping CardanoTxsProcessor")
	close(bp.closeCh)
	return nil
}

func (bp *CardanoTxsProcessorImpl) generateClaims() {
	bp.logger.Debug("Generating claims")

	bridgeClaims := &core.BridgeClaims{}

	var invalidExpectedTxs []*core.BridgeExpectedCardanoTx
	var processedExpectedTxs []*core.BridgeExpectedCardanoTx
	expectedTxs, err := bp.db.GetExpectedTxs(0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get expected txs. error: %v\n", err)
		bp.logger.Error("Failed to get expected txs", "err", err)
		return
	}

	expectedTxsMap := make(map[string]*core.BridgeExpectedCardanoTx, len(expectedTxs))
	for _, expectedTx := range expectedTxs {
		expectedTxsMap[expectedTx.StrKey()] = expectedTx
	}

	unprocessedTxs, err := bp.db.GetUnprocessedTxs(0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get unprocessed txs. error: %v\n", err)
		bp.logger.Error("Failed to get unprocessed txs", "err", err)
		return
	}

	var processedTxs []*core.CardanoTx
	var invalidTxHashes []string

	// check unprocessed txs from indexers
	if len(unprocessedTxs) > 0 {
		processedExpectedTxs, processedTxs, invalidTxHashes = bp.checkUnprocessedTxs(bridgeClaims, unprocessedTxs, expectedTxsMap)
	}

	// check expected txs from bridge
	if bridgeClaims.Count() < bp.appConfig.Settings.MaxBridgingClaimsToGroup && len(expectedTxsMap) > 0 {
		processed, invalid := bp.checkExpectedTxs(bridgeClaims, expectedTxsMap)
		processedExpectedTxs = append(processedExpectedTxs, processed...)
		invalidExpectedTxs = invalid
	}

	// if expected tx is invalid, we should mark them regardless of if submit failed or not
	if len(invalidExpectedTxs) > 0 {
		bp.logger.Debug("Marking expected txs as invalid", "txs", invalidExpectedTxs)
		err := bp.db.MarkExpectedTxsAsInvalid(invalidExpectedTxs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to mark expected txs as invalid. error: %v\n", err)
			bp.logger.Error("Failed to mark expected txs as invalid", "err", err)
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

	// we should only change this in db if submit succeeded
	if len(processedExpectedTxs) > 0 {
		bp.logger.Debug("Marking expected txs as processed", "txs", processedExpectedTxs)
		err := bp.db.MarkExpectedTxsAsProcessed(processedExpectedTxs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to mark expected txs as processed. error: %v\n", err)
			bp.logger.Error("Failed to mark expected txs as processed", "err", err)
		}
	}

	// we should only change this in db if submit succeeded
	if len(processedTxs) > 0 {
		bp.logger.Debug("Marking txs as processed", "txs", processedTxs)
		err := bp.db.MarkUnprocessedTxsAsProcessed(processedTxs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to mark txs as processed. error: %v\n", err)
			bp.logger.Error("Failed to mark txs as processed", "err", err)
			return
		}
	}

	// we should only change this in db if submit succeeded
	if len(invalidTxHashes) > 0 {
		bp.logger.Debug("Saving invalid txs", "txs", invalidTxHashes)
		err := bp.db.AddInvalidTxHashes(invalidTxHashes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save invalid txs. error: %v\n", err)
			bp.logger.Error("Failed to save invalid txs", "err", err)
			return
		}
	}
}

func (bp *CardanoTxsProcessorImpl) checkUnprocessedTxs(
	bridgeClaims *core.BridgeClaims,
	unprocessedTxs []*core.CardanoTx,
	expectedTxsMap map[string]*core.BridgeExpectedCardanoTx,
) (
	processedExpectedTxs []*core.BridgeExpectedCardanoTx,
	processedTxs []*core.CardanoTx,
	invalidTxHashes []string,
) {
unprocessedTxsLoop:
	for _, unprocessedTx := range unprocessedTxs {
		var txProcessed = false
	txProcessorsLoop:
		for _, txProcessor := range bp.txProcessors {
			relevant, err := txProcessor.IsTxRelevant(unprocessedTx, bp.appConfig)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to check if tx is relevant. error: %v\n", err)
				bp.logger.Error("Failed to check if tx is relevant", "tx", unprocessedTx, "err", err)
				continue txProcessorsLoop
			}

			if relevant {
				err := txProcessor.ValidateAndAddClaim(bridgeClaims, unprocessedTx, bp.appConfig)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to ValidateAndAddClaim. error: %v\n", err)
					bp.logger.Error("Failed to ValidateAndAddClaim", "tx", unprocessedTx, "err", err)
					continue txProcessorsLoop
				}

				expectedTx := expectedTxsMap[unprocessedTx.StrKey()]
				if expectedTx != nil {
					processedExpectedTxs = append(processedExpectedTxs, expectedTx)
					delete(expectedTxsMap, expectedTx.StrKey())
				}

				processedTxs = append(processedTxs, unprocessedTx)
				txProcessed = true

				if bridgeClaims.Count() >= bp.appConfig.Settings.MaxBridgingClaimsToGroup {
					break unprocessedTxsLoop
				} else {
					break txProcessorsLoop
				}
			}
		}

		if !txProcessed {
			// transfer an unprocessed tx to invalid txs bucket, to keep as history
			invalidTxHashes = append(invalidTxHashes, unprocessedTx.Hash)
			// and mark it as processed to prevent it from being fetched again as unprocessed
			processedTxs = append(processedTxs, unprocessedTx)
		}
	}

	return processedExpectedTxs, processedTxs, invalidTxHashes
}

func (bp *CardanoTxsProcessorImpl) checkExpectedTxs(
	bridgeClaims *core.BridgeClaims,
	expectedTxsMap map[string]*core.BridgeExpectedCardanoTx,
) (
	processedExpectedTxs []*core.BridgeExpectedCardanoTx,
	invalidExpectedTxs []*core.BridgeExpectedCardanoTx,
) {
expectedTxsLoop:
	for _, expectedTx := range expectedTxsMap {
		ccoDb := bp.getCardanoChainObserverDb(expectedTx.ChainId)
		if ccoDb == nil {
			fmt.Fprintf(os.Stderr, "Failed to get cardano chain observer db for: %v\n", expectedTx.ChainId)
			bp.logger.Error("Failed to get cardano chain observer db", "chainId", expectedTx.ChainId)
			invalidExpectedTxs = append(invalidExpectedTxs, expectedTx)
			continue
		}

		latestBlockPoint, err := ccoDb.GetLatestBlockPoint()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get latest block point for: %v. error: %v\n", expectedTx.ChainId, err)
			bp.logger.Error("Failed to get latest block point", "chainId", expectedTx.ChainId, "err", err)
			continue
		}

		if expectedTx.Ttl+LastExpectedBlockSlotOffset >= latestBlockPoint.BlockSlot {
			// not expired yet
			continue
		}

		var expectedTxProcessed = false
	failedTxProcessorsLoop:
		for _, txProcessor := range bp.failedTxProcessors {
			relevant, err := txProcessor.IsTxRelevant(expectedTx, bp.appConfig)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to check if expected tx is relevant. error: %v\n", err)
				bp.logger.Error("Failed to check if expected tx is relevant", "expectedTx", expectedTx, "err", err)
				continue failedTxProcessorsLoop
			}

			if relevant {
				err := txProcessor.ValidateAndAddClaim(bridgeClaims, expectedTx, bp.appConfig)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to ValidateAndAddClaim. error: %v\n", err)
					bp.logger.Error("Failed to ValidateAndAddClaim", "expectedTx", expectedTx, "err", err)
					continue failedTxProcessorsLoop
				}

				processedExpectedTxs = append(processedExpectedTxs, expectedTx)
				expectedTxProcessed = true

				if bridgeClaims.Count() >= bp.appConfig.Settings.MaxBridgingClaimsToGroup {
					break expectedTxsLoop
				} else {
					break failedTxProcessorsLoop
				}
			}
		}

		if !expectedTxProcessed {
			// expired, but can not process, so we mark it as invalid
			invalidExpectedTxs = append(invalidExpectedTxs, expectedTx)
		}
	}

	return processedExpectedTxs, invalidExpectedTxs
}
