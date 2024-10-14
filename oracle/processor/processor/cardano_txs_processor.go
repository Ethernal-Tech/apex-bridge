package processor

import (
	"bytes"
	"context"
	"math"
	"sort"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/apex-bridge/telemetry"

	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/hashicorp/go-hclog"
)

type CardanoTxsProcessorImpl struct {
	ctx                         context.Context
	appConfig                   *core.AppConfig
	db                          core.CardanoTxsProcessorDB
	txProcessors                *txProcessorsCollection
	settings                    *txsProcessorSettings
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
	successTxProcessors []core.CardanoTxProcessor,
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
		txProcessors:                NewTxProcessorsCollection(successTxProcessors, failedTxProcessors),
		settings:                    NewTxsProcessorSettings(appConfig),
		bridgeSubmitter:             bridgeSubmitter,
		indexerDbs:                  indexerDbs,
		bridgingRequestStateUpdater: bridgingRequestStateUpdater,
		logger:                      logger,
		tickTime:                    TickTimeMs,
	}
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
	// ensure always same order of iterating through bp.appConfig.CardanoChains
	keys := make([]string, 0, len(bp.appConfig.CardanoChains))
	for k := range bp.appConfig.CardanoChains {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, key := range keys {
		select {
		case <-bp.ctx.Done():
			return false
		case <-time.After(bp.tickTime * time.Millisecond):
		}

		bp.processAllStartingWithChain(bp.appConfig.CardanoChains[key].ChainID)
	}

	return true
}

// first process for a specific chainID, to give every chainID the chance
// and then, if max claims not reached, rest of the chains can be processed too
func (bp *CardanoTxsProcessorImpl) processAllStartingWithChain(
	startChainID string,
) {
	var (
		newTxsState      = &cardanoTxsState{}
		bridgeClaims     = &core.BridgeClaims{}
		maxClaimsToGroup = bp.settings.maxBridgingClaimsToGroup[startChainID]
	)

	bp.processAllForChain(bridgeClaims, newTxsState, startChainID, maxClaimsToGroup)

	// ensure always same order of iterating through bp.appConfig.CardanoChains
	keys := make([]string, 0, len(bp.appConfig.CardanoChains))
	for k := range bp.appConfig.CardanoChains {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, key := range keys {
		chainID := bp.appConfig.CardanoChains[key].ChainID
		if chainID != startChainID {
			bp.processAllForChain(bridgeClaims, newTxsState, chainID, maxClaimsToGroup)
		}
	}

	if bridgeClaims.Count() > 0 {
		bp.logger.Info("Submitting bridge claims", "claims", bridgeClaims)

		err := bp.bridgeSubmitter.SubmitClaims(
			bridgeClaims, &eth.SubmitOpts{GasLimitMultiplier: bp.settings.gasLimitMultiplier[startChainID]})
		if err != nil {
			bp.logger.Error("Failed to submit claims", "err", err)

			bp.settings.OnSubmitClaimsFailed(startChainID, bridgeClaims.Count())

			bp.logger.Warn("Adjusted submit claims settings",
				"startChainID", startChainID,
				"maxBridgingClaimsToGroup", bp.settings.maxBridgingClaimsToGroup[startChainID],
				"gasLimitMultiplier", bp.settings.gasLimitMultiplier[startChainID],
			)

			return
		}

		bp.settings.ResetSubmitClaimsSettings(startChainID)

		telemetry.UpdateOracleClaimsSubmitCounter(bridgeClaims.Count()) // update telemetry
	}

	bp.persistNewState(bridgeClaims, newTxsState)
}

func (bp *CardanoTxsProcessorImpl) processAllForChain(
	bridgeClaims *core.BridgeClaims,
	newTxsState *cardanoTxsState,
	chainID string,
	maxClaimsToGroup int,
) {
	for priority := uint8(0); priority <= core.LastProcessingPriority; priority++ {
		if !bridgeClaims.CanAddMore(maxClaimsToGroup) {
			break
		}

		bp.processAllForChainAndPriority(bridgeClaims, newTxsState, chainID, maxClaimsToGroup, priority)
	}
}

func (bp *CardanoTxsProcessorImpl) processAllForChainAndPriority(
	bridgeClaims *core.BridgeClaims,
	newTxsState *cardanoTxsState,
	chainID string,
	maxClaimsToGroup int,
	priority uint8,
) {
	expectedTxs, err := bp.db.GetExpectedTxs(chainID, priority, 0)
	if err != nil {
		bp.logger.Error("Failed to get expected txs", "err", err)

		return
	}

	unprocessedTxs, err := bp.db.GetUnprocessedTxs(chainID, priority, 0)
	if err != nil {
		bp.logger.Error("Failed to get unprocessed txs", "err", err)

		return
	}

	newTxsState.addToUnprocessed(unprocessedTxs)

	// needed for the guarantee that both unprocessedTxs and expectedTxs are processed in order of slot
	// and prevent the situation when there are always enough unprocessedTxs to fill out claims,
	// that all claims are filled only from unprocessedTxs and never from expectedTxs
	blockInfo := bp.constructBridgeClaimsBlockInfo(
		chainID, unprocessedTxs, expectedTxs, nil)
	if blockInfo == nil {
		return
	}

	expectedTxsMap := make(map[string]*core.BridgeExpectedCardanoTx, len(expectedTxs))
	for _, expectedTx := range expectedTxs {
		expectedTxsMap[string(expectedTx.ToCardanoTxKey())] = expectedTx
	}

	for {
		bp.logger.Debug("Processing", "for chainID", chainID, "blockInfo", blockInfo)

		_, processedTxs, processedExpectedTxs := bp.checkUnprocessedTxs(
			blockInfo, bridgeClaims, unprocessedTxs, expectedTxsMap, maxClaimsToGroup)

		_, processedRelevantExpiredTxs, invalidRelevantExpiredTxs := bp.checkExpectedTxs(
			blockInfo, bridgeClaims, expectedTxsMap, maxClaimsToGroup)

		processedExpectedTxs = append(processedExpectedTxs, processedRelevantExpiredTxs...)

		bp.logger.Debug("Checked all", "for chainID", chainID,
			"processedTxs", processedTxs, "processedExpectedTxs", processedExpectedTxs,
			"invalidRelevantExpiredTxs", invalidRelevantExpiredTxs)

		newTxsState.addToProcessed(processedTxs)
		newTxsState.addToProcessedExpected(processedExpectedTxs)
		newTxsState.addToInvalidRelevantExpired(invalidRelevantExpiredTxs)

		if !bridgeClaims.CanAddMore(maxClaimsToGroup) {
			break
		}

		blockInfo = bp.constructBridgeClaimsBlockInfo(
			chainID, unprocessedTxs, expectedTxs, blockInfo)
		if blockInfo == nil {
			break
		}
	}
}

func (bp *CardanoTxsProcessorImpl) constructBridgeClaimsBlockInfo(
	chainID string,
	unprocessedTxs []*core.CardanoTx,
	expectedTxs []*core.BridgeExpectedCardanoTx,
	prevBlockInfo *core.BridgeClaimsBlockInfo,
) *core.BridgeClaimsBlockInfo {
	found := false
	minSlot := uint64(math.MaxUint64)

	var blockHash indexer.Hash

	if len(unprocessedTxs) > 0 {
		// unprocessed are ordered by slot, so first in collection is min
		for _, tx := range unprocessedTxs {
			if prevBlockInfo == nil || prevBlockInfo.Slot < tx.BlockSlot {
				minSlot = tx.BlockSlot
				blockHash = tx.BlockHash
				found = true

				break
			}
		}
	}

	if len(expectedTxs) > 0 {
		ccoDB := bp.indexerDbs[chainID]
		if ccoDB == nil {
			bp.logger.Error("Failed to get cardano chain observer db", "chainId", chainID)
		} else {
			// expected are ordered by ttl, so first in collection is min
			for _, tx := range expectedTxs {
				fromSlot := tx.TTL + TTLInsuranceOffset

				blocks, err := ccoDB.GetConfirmedBlocksFrom(fromSlot, 1)
				if err != nil {
					bp.logger.Error("Failed to get confirmed blocks", "fromSlot", fromSlot, "chainId", chainID, "err", err)
				} else if len(blocks) > 0 && blocks[0].Slot < minSlot &&
					(prevBlockInfo == nil || prevBlockInfo.Slot < blocks[0].Slot) {
					minSlot = blocks[0].Slot
					blockHash = blocks[0].Hash
					found = true

					break
				}
			}
		}
	}

	if found {
		return &core.BridgeClaimsBlockInfo{
			ChainID: chainID,
			Slot:    minSlot,
			Hash:    blockHash,
		}
	}

	return nil
}

func (bp *CardanoTxsProcessorImpl) checkUnprocessedTxs(
	blockInfo *core.BridgeClaimsBlockInfo,
	bridgeClaims *core.BridgeClaims,
	unprocessedTxs []*core.CardanoTx,
	expectedTxsMap map[string]*core.BridgeExpectedCardanoTx,
	maxClaimsToGroup int,
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
		processedTxs         = make([]*core.ProcessedCardanoTx, 0)
		processedExpectedTxs = make([]*core.BridgeExpectedCardanoTx, 0)
		invalidTxsCounter    int
	)

	if len(relevantUnprocessedTxs) == 0 {
		return relevantUnprocessedTxs, processedTxs, processedExpectedTxs
	}

	onInvalidTx := func(tx *core.CardanoTx) {
		processedTxs = append(processedTxs, tx.ToProcessedCardanoTx(true))
		invalidTxsCounter++
	}

	// check unprocessed txs from indexers
	for _, unprocessedTx := range relevantUnprocessedTxs {
		bp.logger.Debug("Checking if tx is relevant", "tx", unprocessedTx)

		txProcessor, err := bp.txProcessors.getSuccess(unprocessedTx.Metadata)
		if err != nil {
			bp.logger.Error("Failed to get tx processor for unprocessed tx", "tx", unprocessedTx, "err", err)

			onInvalidTx(unprocessedTx)

			continue
		}

		err = txProcessor.ValidateAndAddClaim(bridgeClaims, unprocessedTx, bp.appConfig)
		if err != nil {
			bp.logger.Error("Failed to ValidateAndAddClaim", "tx", unprocessedTx, "err", err)

			onInvalidTx(unprocessedTx)

			continue
		}

		key := string(unprocessedTx.ToCardanoTxKey())

		if expectedTx, exists := expectedTxsMap[key]; exists {
			processedExpectedTxs = append(processedExpectedTxs, expectedTx)

			delete(expectedTxsMap, key)
		}

		processedTxs = append(processedTxs, unprocessedTx.ToProcessedCardanoTx(false))

		if !bridgeClaims.CanAddMore(maxClaimsToGroup) {
			break
		}
	}

	if invalidTxsCounter > 0 {
		telemetry.UpdateOracleClaimsInvalidCounter(blockInfo.ChainID, invalidTxsCounter) // update telemetry
	}

	return relevantUnprocessedTxs, processedTxs, processedExpectedTxs
}

func (bp *CardanoTxsProcessorImpl) checkExpectedTxs(
	blockInfo *core.BridgeClaimsBlockInfo,
	bridgeClaims *core.BridgeClaims,
	expectedTxsMap map[string]*core.BridgeExpectedCardanoTx,
	maxClaimsToGroup int,
) (
	[]*core.BridgeExpectedCardanoTx,
	[]*core.BridgeExpectedCardanoTx,
	[]*core.BridgeExpectedCardanoTx,
) {
	var relevantExpiredTxs []*core.BridgeExpectedCardanoTx

	ccoDB := bp.indexerDbs[blockInfo.ChainID]
	if ccoDB == nil {
		bp.logger.Error("Failed to get cardano chain observer db", "chainId", blockInfo.ChainID)
	} else {
		// ensure always same order of iterating through expectedTxsMap
		keys := make([]string, 0, len(expectedTxsMap))
		for k := range expectedTxsMap {
			keys = append(keys, k)
		}

		sort.Strings(keys)

		for _, key := range keys {
			expectedTx := expectedTxsMap[key]

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
	}

	var (
		invalidRelevantExpiredTxs   []*core.BridgeExpectedCardanoTx
		processedRelevantExpiredTxs = make([]*core.BridgeExpectedCardanoTx, 0)
	)

	if !bridgeClaims.CanAddMore(maxClaimsToGroup) ||
		len(relevantExpiredTxs) == 0 {
		return relevantExpiredTxs, processedRelevantExpiredTxs, invalidRelevantExpiredTxs
	}

	onInvalidTx := func(tx *core.BridgeExpectedCardanoTx) {
		// expired, but can not process, so we mark it as invalid
		invalidRelevantExpiredTxs = append(invalidRelevantExpiredTxs, tx)
	}

	for _, expiredTx := range relevantExpiredTxs {
		processedTx, _ := bp.db.GetProcessedTx(expiredTx.ChainID, expiredTx.Hash)
		if processedTx != nil && !processedTx.IsInvalid {
			// already sent the success claim
			processedRelevantExpiredTxs = append(processedRelevantExpiredTxs, expiredTx)

			continue
		}

		bp.logger.Debug("Checking if expired tx is relevant", "expiredTx", expiredTx)

		txProcessor, err := bp.txProcessors.getFailed(expiredTx.Metadata)
		if err != nil {
			bp.logger.Error("Failed to get tx processor for expired tx", "tx", expiredTx, "err", err)

			onInvalidTx(expiredTx)

			continue
		}

		err = txProcessor.ValidateAndAddClaim(bridgeClaims, expiredTx, bp.appConfig)
		if err != nil {
			bp.logger.Error("Failed to ValidateAndAddClaim", "expiredTx", expiredTx, "err", err)

			onInvalidTx(expiredTx)

			continue
		}

		processedRelevantExpiredTxs = append(processedRelevantExpiredTxs, expiredTx)

		if !bridgeClaims.CanAddMore(maxClaimsToGroup) {
			break
		}
	}

	if len(invalidRelevantExpiredTxs) > 0 {
		telemetry.UpdateOracleClaimsInvalidCounter(blockInfo.ChainID, len(invalidRelevantExpiredTxs)) // update telemetry
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
				SourceChainID: common.ToStrChainID(brClaim.SourceChainId),
				SourceTxHash:  brClaim.ObservedTransactionHash,
			}, common.ToStrChainID(brClaim.DestinationChainId))

			if err != nil {
				bp.logger.Error(
					"error while updating a bridging request state to SubmittedToBridge",
					"sourceChainId", common.ToStrChainID(brClaim.SourceChainId),
					"sourceTxHash", brClaim.ObservedTransactionHash, "err", err)
			}
		}
	}

	if len(bridgeClaims.BatchExecutedClaims) > 0 {
		for _, beClaim := range bridgeClaims.BatchExecutedClaims {
			err := bp.bridgingRequestStateUpdater.ExecutedOnDestination(
				common.ToStrChainID(beClaim.ChainId),
				beClaim.BatchNonceId,
				beClaim.ObservedTransactionHash)

			if err != nil {
				bp.logger.Error(
					"error while updating bridging request states to ExecutedOnDestination",
					"destinationChainId", common.ToStrChainID(beClaim.ChainId), "batchId", beClaim.BatchNonceId,
					"destinationTxHash", beClaim.ObservedTransactionHash, "err", err)
			}
		}
	}

	if len(bridgeClaims.BatchExecutionFailedClaims) > 0 {
		for _, befClaim := range bridgeClaims.BatchExecutionFailedClaims {
			err := bp.bridgingRequestStateUpdater.FailedToExecuteOnDestination(
				common.ToStrChainID(befClaim.ChainId),
				befClaim.BatchNonceId)

			if err != nil {
				bp.logger.Error(
					"error while updating bridging request states to FailedToExecuteOnDestination",
					"destinationChainId", common.ToStrChainID(befClaim.ChainId),
					"batchId", befClaim.BatchNonceId, "err", err)
			}
		}
	}

	for _, tx := range processedTxs {
		if tx.IsInvalid {
			for _, unprocessedTx := range unprocessedTxs {
				if bytes.Equal(unprocessedTx.ToCardanoTxKey(), tx.ToCardanoTxKey()) {
					txProcessor, err := bp.txProcessors.getSuccess(unprocessedTx.Metadata)
					if err != nil {
						bp.logger.Error("Failed to get tx processor for processed tx", "tx", tx, "err", err)
					} else if txProcessor.GetType() == common.BridgingTxTypeBridgingRequest {
						err := bp.bridgingRequestStateUpdater.Invalid(common.BridgingRequestStateKey{
							SourceChainID: tx.OriginChainID,
							SourceTxHash:  common.Hash(tx.Hash),
						})

						if err != nil {
							bp.logger.Error(
								"error while updating a bridging request state to Invalid",
								"sourceChainId", tx.OriginChainID,
								"sourceTxHash", tx.Hash, "err", err)
						}
					}

					break
				}
			}
		}
	}

	return nil
}

func (bp *CardanoTxsProcessorImpl) persistNewState(
	bridgeClaims *core.BridgeClaims, newTxsState *cardanoTxsState,
) {
	err := bp.notifyBridgingRequestStateUpdater(bridgeClaims, newTxsState.unprocessed, newTxsState.processed)
	if err != nil {
		bp.logger.Error("Error while updating bridging request states", "err", err)
	}

	// we should only change this in db if submit succeeded (not really, but for convenience)
	if len(newTxsState.invalidRelevantExpired) > 0 {
		bp.logger.Info("Marking expected txs as invalid", "txs", newTxsState.invalidRelevantExpired)

		err := bp.db.MarkExpectedTxsAsInvalid(newTxsState.invalidRelevantExpired)
		if err != nil {
			bp.logger.Error("Failed to mark expected txs as invalid", "err", err)
		}
	}

	// we should only change this in db if submit succeeded
	if len(newTxsState.processedExpected) > 0 {
		bp.logger.Info("Marking expected txs as processed", "txs", newTxsState.processedExpected)

		err := bp.db.MarkExpectedTxsAsProcessed(newTxsState.processedExpected)
		if err != nil {
			bp.logger.Error("Failed to mark expected txs as processed", "err", err)
		}
	}

	// we should only change this in db if submit succeeded
	if len(newTxsState.processed) > 0 {
		bp.logger.Info("Marking txs as processed", "txs", newTxsState.processed)

		err := bp.db.MarkUnprocessedTxsAsProcessed(newTxsState.processed)
		if err != nil {
			bp.logger.Error("Failed to mark txs as processed", "err", err)
		}
	}
}
