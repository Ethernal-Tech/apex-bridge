package processor

import (
	"context"
	"fmt"
	"math"
	"os"
	"sort"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"

	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/hashicorp/go-hclog"
)

const (
	TickTimeMs         = 2000
	TTLInsuranceOffset = 2
)

type CardanoTxsProcessorImpl struct {
	ctx                         context.Context
	appConfig                   *core.AppConfig
	db                          core.CardanoTxsProcessorDB
	txProcessors                []core.CardanoTxProcessor
	failedTxProcessors          []core.CardanoTxFailedProcessor
	bridgeSubmitter             core.BridgeSubmitter
	indexerDbs                  map[string]indexer.Database
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater
	logger                      hclog.Logger
	tickTime                    time.Duration
}

var _ core.CardanoTxsProcessor = (*CardanoTxsProcessorImpl)(nil)

func NewCardanoTxsProcessor(
	ctx context.Context,
	appConfig *core.AppConfig,
	db core.CardanoTxsProcessorDB,
	txProcessors []core.CardanoTxProcessor,
	failedTxProcessors []core.CardanoTxFailedProcessor,
	bridgeSubmitter core.BridgeSubmitter,
	indexerDbs map[string]indexer.Database,
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater,
	logger hclog.Logger,
) *CardanoTxsProcessorImpl {
	return &CardanoTxsProcessorImpl{
		ctx:                         ctx,
		appConfig:                   appConfig,
		db:                          db,
		txProcessors:                txProcessors,
		failedTxProcessors:          failedTxProcessors,
		bridgeSubmitter:             bridgeSubmitter,
		indexerDbs:                  indexerDbs,
		bridgingRequestStateUpdater: bridgingRequestStateUpdater,
		logger:                      logger,
		tickTime:                    TickTimeMs,
	}
}

func (bp *CardanoTxsProcessorImpl) NewUnprocessedTxs(originChainID string, txs []*indexer.Tx) error {
	bp.logger.Debug("NewUnprocessedTxs", "txs", txs)

	var (
		bridgingRequests []*indexer.Tx
		relevantTxs      []*core.CardanoTx
	)

	for _, tx := range txs {
		cardanoTx := &core.CardanoTx{
			OriginChainID: originChainID,
			Tx:            *tx,
		}

		for _, txProcessor := range bp.txProcessors {
			relevant, err := txProcessor.IsTxRelevant(cardanoTx)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to check if tx is relevant. error: %v\n", err)
				bp.logger.Error("Failed to check if tx is relevant", "err", err)

				continue
			}

			if relevant {
				relevantTxs = append(relevantTxs, cardanoTx)

				if txProcessor.GetType() == core.TxProcessorTypeBridgingRequest {
					bridgingRequests = append(bridgingRequests, tx)
				}

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

	err := bp.bridgingRequestStateUpdater.NewMultiple(originChainID, bridgingRequests)
	if err != nil {
		bp.logger.Error("error while adding new bridging request states", "err", err)
	}

	return nil
}

func (bp *CardanoTxsProcessorImpl) Start() {
	bp.logger.Debug("Starting CardanoTxsProcessor")

	for {
		if !bp.checkShouldGenerateClaims() {
			return
		}
	}
}

func (bp *CardanoTxsProcessorImpl) checkShouldGenerateClaims() bool {
	bp.logger.Debug("Checking if should generate claims")

	// ensure always same order of iterating through bp.appConfig.CardanoChains
	keys := make([]string, 0, len(bp.appConfig.CardanoChains))
	for k := range bp.appConfig.CardanoChains {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	ticker := time.NewTicker(bp.tickTime * time.Millisecond)
	defer ticker.Stop()

	for _, key := range keys {
		select {
		case <-bp.ctx.Done():
			return false
		case <-ticker.C:
		}

		bp.processAllForChain(bp.appConfig.CardanoChains[key].ChainID)
	}

	return true
}

func (bp *CardanoTxsProcessorImpl) constructBridgeClaimsBlockInfo(
	chainID string,
	unprocessedTxs []*core.CardanoTx,
	expectedTxs []*core.BridgeExpectedCardanoTx,
) (
	*core.BridgeClaimsBlockInfo,
	indexer.Database,
) {
	ccoDB := bp.indexerDbs[chainID]
	if ccoDB == nil {
		fmt.Fprintf(os.Stderr, "Failed to get cardano chain observer db for: %v\n", chainID)
		bp.logger.Error("Failed to get cardano chain observer db", "chainId", chainID)
	}

	found := false
	minSlot := uint64(math.MaxUint64)

	var blockHash string

	if len(unprocessedTxs) > 0 {
		// unprocessed are ordered by slot, so first in collection is min
		minSlot = unprocessedTxs[0].BlockSlot
		blockHash = unprocessedTxs[0].BlockHash
		found = true
	}

	if len(expectedTxs) > 0 {
		// expected are ordered by ttl, so first in collection is min
		expectedTx := expectedTxs[0]
		fromSlot := expectedTx.TTL + TTLInsuranceOffset

		if ccoDB != nil {
			blocks, err := ccoDB.GetConfirmedBlocksFrom(fromSlot, 1)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to get confirmed blocks from slot: %v, for %v. error: %v\n", fromSlot, chainID, err)
				bp.logger.Error("Failed to get confirmed blocks", "fromSlot", fromSlot, "chainId", chainID, "err", err)
			} else if len(blocks) > 0 && blocks[0].Slot < minSlot {
				minSlot = blocks[0].Slot
				blockHash = blocks[0].Hash
				found = true
			}
		}
	}

	if found {
		return &core.BridgeClaimsBlockInfo{
			ChainID:            chainID,
			Slot:               minSlot,
			Hash:               blockHash,
			BlockFullyObserved: false,
		}, ccoDB
	}

	return nil, ccoDB
}

func (bp *CardanoTxsProcessorImpl) checkUnprocessedTxs(
	blockInfo *core.BridgeClaimsBlockInfo,
	bridgeClaims *core.BridgeClaims,
	unprocessedTxs []*core.CardanoTx,
	expectedTxsMap map[string]*core.BridgeExpectedCardanoTx,
) (
	[]*core.CardanoTx,
	[]*core.ProcessedCardanoTx,
	[]*core.BridgeExpectedCardanoTx,
) {
	var relevantUnprocessedTxs []*core.CardanoTx

	for _, unprocessedTx := range unprocessedTxs {
		if blockInfo.EqualWithUnprocessed(unprocessedTx) {
			relevantUnprocessedTxs = append(relevantUnprocessedTxs, unprocessedTx)
		}
	}

	var (
		processedTxs         []*core.ProcessedCardanoTx
		processedExpectedTxs []*core.BridgeExpectedCardanoTx
	)

	// check unprocessed txs from indexers
	if len(relevantUnprocessedTxs) > 0 {
	unprocessedTxsLoop:
		for _, unprocessedTx := range relevantUnprocessedTxs {
			var txProcessed = false
		txProcessorsLoop:
			for _, txProcessor := range bp.txProcessors {
				relevant, err := txProcessor.IsTxRelevant(unprocessedTx)
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

					expectedTx := expectedTxsMap[unprocessedTx.ToCardanoTxKey()]
					if expectedTx != nil {
						processedExpectedTxs = append(processedExpectedTxs, expectedTx)
						delete(expectedTxsMap, expectedTx.ToCardanoTxKey())
					}

					processedTxs = append(processedTxs, unprocessedTx.ToProcessedCardanoTx(false))
					txProcessed = true

					if bridgeClaims.Count() >= bp.appConfig.BridgingSettings.MaxBridgingClaimsToGroup {
						break unprocessedTxsLoop
					} else {
						break txProcessorsLoop
					}
				}
			}

			if !txProcessed {
				processedTxs = append(processedTxs, unprocessedTx.ToProcessedCardanoTx(true))
			}
		}
	}

	return relevantUnprocessedTxs, processedTxs, processedExpectedTxs
}

func (bp *CardanoTxsProcessorImpl) checkExpectedTxs(
	blockInfo *core.BridgeClaimsBlockInfo,
	bridgeClaims *core.BridgeClaims,
	ccoDB indexer.Database,
	expectedTxsMap map[string]*core.BridgeExpectedCardanoTx,
) (
	[]*core.BridgeExpectedCardanoTx,
	[]*core.BridgeExpectedCardanoTx,
	[]*core.BridgeExpectedCardanoTx,
) {
	var relevantExpiredTxs []*core.BridgeExpectedCardanoTx

	// ensure always same order of iterating through expectedTxsMap
	keys := make([]string, 0, len(bp.appConfig.CardanoChains))
	for k := range expectedTxsMap {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, key := range keys {
		expectedTx := expectedTxsMap[key]

		if ccoDB == nil {
			break
		}

		fromSlot := expectedTx.TTL + TTLInsuranceOffset

		blocks, err := ccoDB.GetConfirmedBlocksFrom(fromSlot, 1)
		if err != nil {
			bp.logger.Error("Failed to get confirmed blocks", "fromSlot", fromSlot, "chainId", expectedTx.ChainID, "err", err)

			break
		}

		if len(blocks) == 1 && blockInfo.EqualWithExpected(expectedTx, blocks[0]) {
			relevantExpiredTxs = append(relevantExpiredTxs, expectedTx)
		}
	}

	var (
		invalidRelevantExpiredTxs   []*core.BridgeExpectedCardanoTx
		processedRelevantExpiredTxs []*core.BridgeExpectedCardanoTx
	)

	if bridgeClaims.Count() < bp.appConfig.BridgingSettings.MaxBridgingClaimsToGroup && len(relevantExpiredTxs) > 0 {
	expiredTxsLoop:
		for _, expiredTx := range relevantExpiredTxs {
			processedTx, _ := bp.db.GetProcessedTx(expiredTx.ChainID, expiredTx.Hash)
			if processedTx != nil && !processedTx.IsInvalid {
				// already sent the success claim
				processedRelevantExpiredTxs = append(processedRelevantExpiredTxs, expiredTx)

				continue
			}

			var expiredTxProcessed = false
		failedTxProcessorsLoop:
			for _, txProcessor := range bp.failedTxProcessors {
				relevant, err := txProcessor.IsTxRelevant(expiredTx)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to check if expired tx is relevant. error: %v\n", err)
					bp.logger.Error("Failed to check if expired tx is relevant", "expiredTx", expiredTx, "err", err)

					continue failedTxProcessorsLoop
				}

				if relevant {
					err := txProcessor.ValidateAndAddClaim(bridgeClaims, expiredTx, bp.appConfig)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Failed to ValidateAndAddClaim. error: %v\n", err)
						bp.logger.Error("Failed to ValidateAndAddClaim", "expiredTx", expiredTx, "err", err)

						continue failedTxProcessorsLoop
					}

					processedRelevantExpiredTxs = append(processedRelevantExpiredTxs, expiredTx)
					expiredTxProcessed = true

					if bridgeClaims.Count() >= bp.appConfig.BridgingSettings.MaxBridgingClaimsToGroup {
						break expiredTxsLoop
					} else {
						break failedTxProcessorsLoop
					}
				}
			}

			if !expiredTxProcessed {
				// expired, but can not process, so we mark it as invalid
				invalidRelevantExpiredTxs = append(invalidRelevantExpiredTxs, expiredTx)
			}
		}
	}

	return relevantExpiredTxs, processedRelevantExpiredTxs, invalidRelevantExpiredTxs
}

func (bp *CardanoTxsProcessorImpl) notifyBridgingRequestStateUpdater(
	bridgeClaims *core.BridgeClaims,
	unprocessedTxs []*core.CardanoTx,
	processedTxs []*core.ProcessedCardanoTx,
) error {
	if len(bridgeClaims.BridgingRequestClaims) > 0 {
		for _, brClaim := range bridgeClaims.BridgingRequestClaims {
			err := bp.bridgingRequestStateUpdater.SubmittedToBridge(common.BridgingRequestStateKey{
				SourceChainID: brClaim.SourceChainID,
				SourceTxHash:  brClaim.ObservedTransactionHash,
			}, brClaim.DestinationChainID)

			if err != nil {
				bp.logger.Error(
					"error while updating a bridging request state to SubmittedToBridge",
					"sourceChainId", brClaim.SourceChainID, "sourceTxHash", brClaim.ObservedTransactionHash)
			}
		}
	}

	if len(bridgeClaims.BatchExecutedClaims) > 0 {
		for _, beClaim := range bridgeClaims.BatchExecutedClaims {
			err := bp.bridgingRequestStateUpdater.ExecutedOnDestination(
				beClaim.ChainID, beClaim.BatchNonceID.Uint64(), beClaim.ObservedTransactionHash)

			if err != nil {
				bp.logger.Error(
					"error while updating bridging request states to ExecutedOnDestination",
					"destinationChainId", beClaim.ChainID, "batchId", beClaim.BatchNonceID.Uint64(),
					"destinationTxHash", beClaim.ObservedTransactionHash)
			}
		}
	}

	if len(bridgeClaims.BatchExecutionFailedClaims) > 0 {
		for _, befClaim := range bridgeClaims.BatchExecutionFailedClaims {
			err := bp.bridgingRequestStateUpdater.FailedToExecuteOnDestination(befClaim.ChainID, befClaim.BatchNonceID.Uint64())

			if err != nil {
				bp.logger.Error(
					"error while updating bridging request states to FailedToExecuteOnDestination",
					"destinationChainId", befClaim.ChainID, "batchId", befClaim.BatchNonceID.Uint64())
			}
		}
	}

	var bridgingRequestTxProcessor core.CardanoTxProcessor

	for _, txProcessor := range bp.txProcessors {
		if txProcessor.GetType() == core.TxProcessorTypeBridgingRequest {
			bridgingRequestTxProcessor = txProcessor

			break
		}
	}

	if bridgingRequestTxProcessor == nil {
		return fmt.Errorf("failed to find bridging request tx processor")
	}

	for _, tx := range processedTxs {
		if tx.IsInvalid {
			for _, unprocessedTx := range unprocessedTxs {
				if unprocessedTx.ToCardanoTxKey() == tx.ToCardanoTxKey() {
					relevant, err := bridgingRequestTxProcessor.IsTxRelevant(unprocessedTx)
					if err != nil {
						bp.logger.Error("Failed to check if unprocessedTx is relevant", "unprocessedTx", unprocessedTx, "err", err)
					} else if relevant {
						err := bp.bridgingRequestStateUpdater.Invalid(common.BridgingRequestStateKey{
							SourceChainID: tx.OriginChainID,
							SourceTxHash:  tx.Hash,
						})

						if err != nil {
							bp.logger.Error(
								"error while updating a bridging request state to Invalid",
								"sourceChainId", tx.OriginChainID, "sourceTxHash", tx.Hash)
						}
					}

					break
				}
			}
		}
	}

	return nil
}

func (bp *CardanoTxsProcessorImpl) processAllForChain(
	chainID string,
) {
	expectedTxs, err := bp.db.GetExpectedTxs(chainID, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get expected txs. error: %v\n", err)
		bp.logger.Error("Failed to get expected txs", "err", err)

		return
	}

	unprocessedTxs, err := bp.db.GetUnprocessedTxs(chainID, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get unprocessed txs. error: %v\n", err)
		bp.logger.Error("Failed to get unprocessed txs", "err", err)

		return
	}

	blockInfo, ccoDB := bp.constructBridgeClaimsBlockInfo(chainID, unprocessedTxs, expectedTxs)
	if blockInfo == nil {
		return
	}

	bridgeClaims := &core.BridgeClaims{}

	expectedTxsMap := make(map[string]*core.BridgeExpectedCardanoTx, len(expectedTxs))
	for _, expectedTx := range expectedTxs {
		expectedTxsMap[expectedTx.ToCardanoTxKey()] = expectedTx
	}

	relevantUnprocessedTxs, processedTxs, processedExpectedTxs := bp.checkUnprocessedTxs(
		blockInfo, bridgeClaims, unprocessedTxs, expectedTxsMap)
	relevantExpiredTxs, processedRelevantExpiredTxs, invalidRelevantExpiredTxs := bp.checkExpectedTxs(
		blockInfo, bridgeClaims, ccoDB, expectedTxsMap)
	processedExpectedTxs = append(processedExpectedTxs, processedRelevantExpiredTxs...)

	blockInfo.BlockFullyObserved = len(processedTxs) == len(relevantUnprocessedTxs) &&
		len(processedRelevantExpiredTxs)+len(invalidRelevantExpiredTxs) == len(relevantExpiredTxs)

	// if expected/expired tx is invalid, we should mark them regardless of if submit failed or not
	if len(invalidRelevantExpiredTxs) > 0 {
		bp.logger.Debug("Marking expected txs as invalid", "txs", invalidRelevantExpiredTxs)

		err := bp.db.MarkExpectedTxsAsInvalid(invalidRelevantExpiredTxs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to mark expected txs as invalid. error: %v\n", err)
			bp.logger.Error("Failed to mark expected txs as invalid", "err", err)
		}
	}

	bp.logger.Debug("Submitting bridge claims", "claims", bridgeClaims)

	err = bp.bridgeSubmitter.SubmitClaims(bridgeClaims)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to submit claims. error: %v\n", err)
		bp.logger.Error("Failed to submit claims", "err", err)

		return
	}

	err = bp.notifyBridgingRequestStateUpdater(bridgeClaims, unprocessedTxs, processedTxs)
	if err != nil {
		bp.logger.Error("Error while updating bridging request states", "err", err)
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
		}
	}
}
