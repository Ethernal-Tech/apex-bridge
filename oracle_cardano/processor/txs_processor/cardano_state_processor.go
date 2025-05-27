package processor

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/telemetry"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/hashicorp/go-hclog"
)

const (
	TTLInsuranceOffset             = 2
	logLastNBatchInfoSkippedEvents = 10
)

var _ cCore.SpecificChainTxsProcessorState = (*CardanoStateProcessor)(nil)

type CardanoStateProcessor struct {
	ctx          context.Context
	appConfig    *cCore.AppConfig
	db           core.CardanoTxsProcessorDB
	txProcessors *txProcessorsCollection
	indexerDbs   map[string]indexer.Database
	logger       hclog.Logger

	state *perTickState
}

func NewCardanoStateProcessor(
	ctx context.Context,
	appConfig *cCore.AppConfig,
	db core.CardanoTxsProcessorDB,
	txProcessors *txProcessorsCollection,
	indexerDbs map[string]indexer.Database,
	logger hclog.Logger,
) *CardanoStateProcessor {
	return &CardanoStateProcessor{
		ctx:          ctx,
		appConfig:    appConfig,
		db:           db,
		txProcessors: txProcessors,
		indexerDbs:   indexerDbs,
		logger:       logger,
	}
}

func (sp *CardanoStateProcessor) GetChainType() string {
	return common.ChainTypeCardanoStr
}

func (sp *CardanoStateProcessor) Reset() {
	sp.state = &perTickState{updateData: &core.CardanoUpdateTxsData{}}
}

func (sp *CardanoStateProcessor) ProcessSavedEvents() {
	var batchEvents []*cCore.DBBatchInfoEvent

	for _, chain := range sp.appConfig.CardanoChains {
		chainBatchEvents, err := sp.db.GetUnprocessedBatchEvents(chain.ChainID)
		if err != nil {
			sp.logger.Error("Failed to get unprocessed batch events", "err", err)

			continue
		}

		batchEvents = append(batchEvents, chainBatchEvents...)
	}

	if len(batchEvents) > 0 {
		sp.logger.Debug("Processing stored BatchExecutionInfoEvent events", "cnt", len(batchEvents))

		processedBatchEvents, _ := sp.processBatchExecutionInfoEvents(batchEvents)

		if len(processedBatchEvents) > 0 {
			sp.logger.Debug("Removing BatchExecutionInfoEvent events from db", "events", processedBatchEvents)
			sp.state.updateData.RemoveBatchInfoEvents = processedBatchEvents
		}
	}
}

func (sp *CardanoStateProcessor) RunChecks(
	bridgeClaims *cCore.BridgeClaims,
	chainID string,
	maxClaimsToGroup int,
	priority uint8,
) {
	expectedTxs, err := sp.db.GetExpectedTxs(chainID, priority, 0)
	if err != nil {
		sp.logger.Error("Failed to get expected txs", "err", err)

		return
	}

	sp.state.unprocessedTxs, err = sp.db.GetUnprocessedTxs(chainID, priority, 0)
	if err != nil {
		sp.logger.Error("Failed to get unprocessed txs", "err", err)

		return
	}

	// needed for the guarantee that both unprocessedTxs and expectedTxs are processed in order of slot
	// and prevent the situation when there are always enough unprocessedTxs to fill out claims,
	// that all claims are filled only from unprocessedTxs and never from expectedTxs
	sp.state.blockInfo = sp.constructBridgeClaimsBlockInfo(
		chainID, sp.state.unprocessedTxs, expectedTxs, nil)
	if sp.state.blockInfo == nil {
		return
	}

	sp.state.expectedTxsMap = make(map[string]*core.BridgeExpectedCardanoTx, len(expectedTxs))
	for _, expectedTx := range expectedTxs {
		sp.state.expectedTxsMap[string(expectedTx.ToCardanoTxKey())] = expectedTx
	}

	for {
		sp.logger.Debug("Processing",
			"for chainID", sp.state.blockInfo.ChainID,
			"blockInfo", sp.state.blockInfo)

		sp.checkUnprocessedTxs(bridgeClaims, maxClaimsToGroup)
		sp.checkExpectedTxs(bridgeClaims, maxClaimsToGroup)

		if !bridgeClaims.CanAddMore(maxClaimsToGroup) {
			break
		}

		sp.state.blockInfo = sp.constructBridgeClaimsBlockInfo(
			chainID, sp.state.unprocessedTxs, expectedTxs, sp.state.blockInfo)
		if sp.state.blockInfo == nil {
			break
		}
	}
}

func (sp *CardanoStateProcessor) ProcessSubmitClaimsEvents(
	events *cCore.SubmitClaimsEvents, claims *cCore.BridgeClaims) {
	if len(events.NotEnoughFunds) > 0 {
		sp.processNotEnoughFundsEvents(events.NotEnoughFunds, claims)
	}

	if len(events.BatchExecutionInfo) > 0 {
		_, skippedEvents := sp.processBatchExecutionInfoEvents(events.BatchExecutionInfo)
		if len(skippedEvents) > 0 {
			sp.logger.Debug("Storing BatchExecutionInfoEvent events", "cnt", len(skippedEvents))
			sp.state.updateData.AddBatchInfoEvents = skippedEvents
		}
	}
}

func (sp *CardanoStateProcessor) processBatchExecutionInfoEvents(
	events []*cCore.DBBatchInfoEvent,
) ([]*cCore.DBBatchInfoEvent, []*cCore.DBBatchInfoEvent) {
	var (
		processedEvents      = make([]*cCore.DBBatchInfoEvent, 0, len(events))
		newProcessedTxs      []cCore.BaseProcessedTx
		newUnprocessedTxs    []cCore.BaseTx
		skippedEventsWithErr []struct {
			evt *cCore.DBBatchInfoEvent
			err error
		}
	)

	for _, event := range events {
		txs, err := sp.getTxsFromBatchEvent(event)
		if err != nil {
			skippedEventsWithErr = append(
				skippedEventsWithErr,
				struct {
					evt *cCore.DBBatchInfoEvent
					err error
				}{evt: event, err: err})

			continue
		}

		processedEvents = append(processedEvents, event)

		if event.IsFailedClaim {
			for _, tx := range txs {
				tx.IncrementBatchTryCount()
				tx.IncrementSubmitTryCount()
				tx.SetLastTimeTried(time.Time{})
				newUnprocessedTxs = append(newUnprocessedTxs, tx)
			}
		} else {
			for _, tx := range txs {
				processedTx := tx.ToProcessed(false)
				newProcessedTxs = append(newProcessedTxs, processedTx)
			}
		}
	}

	if len(skippedEventsWithErr) > 0 {
		lastNSkippedEventsWithErr := common.LastN(skippedEventsWithErr, logLastNBatchInfoSkippedEvents)

		sp.logger.Info(
			fmt.Sprintf("couldn't find txs for some BatchExecutionInfoEvent events. listing last %d",
				logLastNBatchInfoSkippedEvents))

		for _, item := range lastNSkippedEventsWithErr {
			sp.logger.Info(
				"couldn't find txs for BatchExecutionInfoEvent event",
				"event", item.evt, "err", item.err)
		}
	}

	skippedEvents := make([]*cCore.DBBatchInfoEvent, len(skippedEventsWithErr))
	for idx, item := range skippedEventsWithErr {
		skippedEvents[idx] = item.evt
	}

	sp.state.updateData.MovePendingToProcessed = newProcessedTxs
	sp.state.updateData.MovePendingToUnprocessed = newUnprocessedTxs

	return processedEvents, skippedEvents
}

func (sp *CardanoStateProcessor) getTxsFromBatchEvent(
	event *cCore.DBBatchInfoEvent,
) ([]cCore.BaseTx, error) {
	result := make([]cCore.BaseTx, len(event.TxHashes))

	for idx, hash := range event.TxHashes {
		tx, err := sp.db.GetPendingTx(
			cCore.DBTxID{
				ChainID: common.ToStrChainID(hash.SourceChainID),
				DBKey:   hash.ObservedTransactionHash[:],
			},
		)
		if err != nil {
			return nil, err
		}

		result[idx] = tx
	}

	return result, nil
}

func (sp *CardanoStateProcessor) processNotEnoughFundsEvents(
	events []*cCore.NotEnoughFundsEvent, claims *cCore.BridgeClaims,
) {
	allPendingMap := make(map[string]*core.CardanoTx, len(sp.state.updateData.MoveUnprocessedToPending))
	for _, tx := range sp.state.updateData.MoveUnprocessedToPending {
		allPendingMap[string(tx.ToCardanoTxKey())] = tx
	}

	now := time.Now().UTC()
	unprocessedToUpdate := make([]*core.CardanoTx, 0, len(events))

	for _, event := range events {
		txToUpdate, err := sp.findRejectedTxInPending(event, claims, allPendingMap)
		if err != nil {
			sp.logger.Error("couldn't find tx for NotEnoughFunds event", "event", event, "err", err)

			continue
		}

		delete(allPendingMap, string(txToUpdate.ToCardanoTxKey()))

		txToUpdate.SubmitTryCount++
		txToUpdate.LastTimeTried = now
		unprocessedToUpdate = append(unprocessedToUpdate, txToUpdate)

		sp.logger.Debug("updated unprocessedTx TryCount and LastTimeTried", "tx", txToUpdate)
	}

	filteredAllPending := make([]*core.CardanoTx, 0, len(allPendingMap))
	for _, tx := range allPendingMap {
		filteredAllPending = append(filteredAllPending, tx)
	}

	sp.state.updateData.MoveUnprocessedToPending = filteredAllPending
	sp.state.updateData.UpdateUnprocessed = append(sp.state.updateData.UpdateUnprocessed, unprocessedToUpdate...)
}

func (sp *CardanoStateProcessor) findRejectedTxInPending(
	event *cCore.NotEnoughFundsEvent, claims *cCore.BridgeClaims,
	allPendingMap map[string]*core.CardanoTx,
) (*core.CardanoTx, error) {
	switch event.ClaimeType {
	case cCore.BRCClaimType:
		brcIndex := event.Index.Uint64()
		if brcIndex >= uint64(len(claims.BridgingRequestClaims)) {
			return nil, fmt.Errorf(
				"invalid NotEnoughFundsEvent.Index: %d. BRCs len: %d", brcIndex, len(claims.BridgingRequestClaims))
		}

		brc := claims.BridgingRequestClaims[brcIndex]

		tx, exists := allPendingMap[string(
			core.ToCardanoTxKey(common.ToStrChainID(brc.SourceChainId), brc.ObservedTransactionHash))]
		if !exists {
			return nil, fmt.Errorf(
				"BRC not found in MoveUnprocessedToPending for index: %d", brcIndex)
		}

		return tx, nil
	default:
		return nil, fmt.Errorf(
			"unsupported NotEnoughFundsEvent.claimType: %s", event.ClaimeType)
	}
}

func (sp *CardanoStateProcessor) PersistNew() {
	if sp.state.updateData.Count() > 0 {
		sp.logger.Info("Updating txs", "data", sp.state.updateData)

		if err := sp.db.UpdateTxs(sp.state.updateData); err != nil {
			sp.logger.Error("Failed to update txs", "err", err)
		}
	}
}

func (sp *CardanoStateProcessor) constructBridgeClaimsBlockInfo(
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
		ccoDB := sp.indexerDbs[chainID]
		if ccoDB == nil {
			sp.logger.Error("Failed to get cardano chain observer db", "chainId", chainID)
		} else {
			// expected are ordered by ttl, so first in collection is min
			for _, tx := range expectedTxs {
				fromSlot := tx.TTL + TTLInsuranceOffset

				blocks, err := ccoDB.GetConfirmedBlocksFrom(fromSlot, 1)
				if err != nil {
					sp.logger.Error("Failed to get confirmed blocks", "fromSlot", fromSlot, "chainId", chainID, "err", err)
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

func (sp *CardanoStateProcessor) checkUnprocessedTxs(
	bridgeClaims *cCore.BridgeClaims,
	maxClaimsToGroup int,
) {
	var relevantUnprocessedTxs []*core.CardanoTx

	for _, unprocessedTx := range sp.state.unprocessedTxs {
		if sp.state.blockInfo.EqualWithUnprocessed(unprocessedTx) && cCore.IsTxReady(
			unprocessedTx.SubmitTryCount, unprocessedTx.LastTimeTried, sp.appConfig.RetryUnprocessedSettings) {
			relevantUnprocessedTxs = append(relevantUnprocessedTxs, unprocessedTx)
		}
	}

	if len(relevantUnprocessedTxs) == 0 {
		return
	}

	//nolint:prealloc
	var (
		processedInvalidTxs  []*core.CardanoTx
		processedValidTxs    []*core.CardanoTx
		pendingTxs           []*core.CardanoTx
		processedExpectedTxs []*core.BridgeExpectedCardanoTx
		invalidTxsCounter    int
	)

	onInvalidTx := func(tx *core.CardanoTx) {
		processedInvalidTxs = append(processedInvalidTxs, tx)
		invalidTxsCounter++
	}

	// check unprocessed txs from indexers
	for _, unprocessedTx := range relevantUnprocessedTxs {
		sp.logger.Debug("Checking if tx is relevant", "tx", unprocessedTx)

		txProcessor, err := sp.txProcessors.getSuccess(unprocessedTx, sp.appConfig)
		if err != nil {
			sp.logger.Error("Failed to get tx processor for unprocessed tx", "tx", unprocessedTx, "err", err)

			onInvalidTx(unprocessedTx)

			continue
		}

		err = txProcessor.ValidateAndAddClaim(bridgeClaims, unprocessedTx, sp.appConfig)
		if err != nil {
			sp.logger.Error("Failed to ValidateAndAddClaim", "tx", unprocessedTx, "err", err)

			onInvalidTx(unprocessedTx)

			continue
		}

		if txProcessor.GetType() == common.BridgingTxTypeBridgingRequest {
			pendingTxs = append(pendingTxs, unprocessedTx)
		} else {
			key := string(unprocessedTx.ToCardanoTxKey())

			if expectedTx, exists := sp.state.expectedTxsMap[key]; exists {
				processedExpectedTxs = append(processedExpectedTxs, expectedTx)

				delete(sp.state.expectedTxsMap, key)
			}

			processedValidTxs = append(processedValidTxs, unprocessedTx)
		}

		if !bridgeClaims.CanAddMore(maxClaimsToGroup) {
			break
		}
	}

	if invalidTxsCounter > 0 {
		telemetry.UpdateOracleClaimsInvalidCounter(sp.state.blockInfo.ChainID, invalidTxsCounter) // update telemetry
	}

	for _, tx := range processedValidTxs {
		sp.state.updateData.MoveUnprocessedToProcessed = append(
			sp.state.updateData.MoveUnprocessedToProcessed, tx.ToProcessedCardanoTx(false))
	}

	for _, tx := range processedInvalidTxs {
		sp.state.updateData.MoveUnprocessedToProcessed = append(
			sp.state.updateData.MoveUnprocessedToProcessed, tx.ToProcessedCardanoTx(true))

		sp.state.allProcessedInvalid = append(sp.state.allProcessedInvalid, tx)
	}

	sp.state.updateData.MoveUnprocessedToPending = append(sp.state.updateData.MoveUnprocessedToPending, pendingTxs...)
	sp.state.updateData.ExpectedProcessed = append(sp.state.updateData.ExpectedProcessed, processedExpectedTxs...)

	sp.logger.Debug("Checked all unprocessed",
		"for chainID", sp.state.blockInfo.ChainID,
		"processedValidTxs", processedValidTxs,
		"processedInvalidTxs", processedInvalidTxs,
		"pendingTxs", pendingTxs,
		"processedExpectedTxs", processedExpectedTxs)
}

func (sp *CardanoStateProcessor) checkExpectedTxs(
	bridgeClaims *cCore.BridgeClaims,
	maxClaimsToGroup int,
) {
	var relevantExpiredTxs []*core.BridgeExpectedCardanoTx

	ccoDB := sp.indexerDbs[sp.state.blockInfo.ChainID]
	if ccoDB == nil {
		sp.logger.Error("Failed to get cardano chain observer db", "chainId", sp.state.blockInfo.ChainID)
	} else {
		// ensure always same order of iterating through expectedTxsMap
		keys := make([]string, 0, len(sp.state.expectedTxsMap))
		for k := range sp.state.expectedTxsMap {
			keys = append(keys, k)
		}

		sort.Strings(keys)

		for _, key := range keys {
			expectedTx := sp.state.expectedTxsMap[key]

			fromSlot := expectedTx.TTL + TTLInsuranceOffset

			blocks, err := ccoDB.GetConfirmedBlocksFrom(fromSlot, 1)
			if err != nil {
				sp.logger.Error("Failed to get confirmed blocks", "fromSlot", fromSlot, "chainId", expectedTx.ChainID, "err", err)

				break
			}

			if len(blocks) == 1 && sp.state.blockInfo.EqualWithExpected(expectedTx, blocks[0]) {
				relevantExpiredTxs = append(relevantExpiredTxs, expectedTx)
			}
		}
	}

	if !bridgeClaims.CanAddMore(maxClaimsToGroup) || len(relevantExpiredTxs) == 0 {
		return
	}

	var (
		invalidRelevantExpiredTxs   []*core.BridgeExpectedCardanoTx
		processedRelevantExpiredTxs = make([]*core.BridgeExpectedCardanoTx, 0)
	)

	onInvalidTx := func(tx *core.BridgeExpectedCardanoTx) {
		// expired, but can not process, so we mark it as invalid
		invalidRelevantExpiredTxs = append(invalidRelevantExpiredTxs, tx)
	}

	for _, expiredTx := range relevantExpiredTxs {
		processedTx, _ := sp.db.GetProcessedTx(
			cCore.DBTxID{
				ChainID: expiredTx.ChainID,
				DBKey:   expiredTx.Hash[:],
			},
		)
		if processedTx != nil && !processedTx.IsInvalid {
			// already sent the success claim
			processedRelevantExpiredTxs = append(processedRelevantExpiredTxs, expiredTx)

			continue
		}

		sp.logger.Debug("Checking if expired tx is relevant", "expiredTx", expiredTx)

		txProcessor, err := sp.txProcessors.getFailed(expiredTx, sp.appConfig)
		if err != nil {
			sp.logger.Error("Failed to get tx processor for expired tx", "tx", expiredTx, "err", err)

			onInvalidTx(expiredTx)

			continue
		}

		err = txProcessor.ValidateAndAddClaim(bridgeClaims, expiredTx, sp.appConfig)
		if err != nil {
			sp.logger.Error("Failed to ValidateAndAddClaim", "expiredTx", expiredTx, "err", err)

			onInvalidTx(expiredTx)

			continue
		}

		processedRelevantExpiredTxs = append(processedRelevantExpiredTxs, expiredTx)

		if !bridgeClaims.CanAddMore(maxClaimsToGroup) {
			break
		}
	}

	if len(invalidRelevantExpiredTxs) > 0 {
		telemetry.UpdateOracleClaimsInvalidCounter(
			sp.state.blockInfo.ChainID, len(invalidRelevantExpiredTxs)) // update telemetry
	}

	sp.state.updateData.ExpectedProcessed = append(sp.state.updateData.ExpectedProcessed, processedRelevantExpiredTxs...)
	sp.state.updateData.ExpectedInvalid = append(sp.state.updateData.ExpectedInvalid, invalidRelevantExpiredTxs...)

	sp.logger.Debug("Checked all expected",
		"for chainID", sp.state.blockInfo.ChainID,
		"processedExpectedTxs", processedRelevantExpiredTxs,
		"invalidRelevantExpiredTxs", invalidRelevantExpiredTxs)
}

func (sp *CardanoStateProcessor) UpdateBridgingRequestStates(
	bridgeClaims *cCore.BridgeClaims,
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater,
) {
	if len(bridgeClaims.BridgingRequestClaims) > 0 {
		notRejectedMap := make(map[string]bool, len(sp.state.updateData.MoveUnprocessedToPending))
		for _, tx := range sp.state.updateData.MoveUnprocessedToPending {
			notRejectedMap[string(tx.ToCardanoTxKey())] = true
		}

		for _, brClaim := range bridgeClaims.BridgingRequestClaims {
			srcChainID := common.ToStrChainID(brClaim.SourceChainId)
			key := core.ToCardanoTxKey(srcChainID, brClaim.ObservedTransactionHash)

			if !notRejectedMap[string(key)] {
				continue
			}

			err := bridgingRequestStateUpdater.SubmittedToBridge(
				common.NewBridgingRequestStateKey(srcChainID, brClaim.ObservedTransactionHash),
				common.ToStrChainID(brClaim.DestinationChainId))

			if err != nil {
				sp.logger.Error(
					"error while updating a bridging request state to SubmittedToBridge",
					"srcChainId", srcChainID, "srcTxHash", brClaim.ObservedTransactionHash, "err", err)
			}
		}
	}

	for _, tx := range sp.state.allProcessedInvalid {
		txProcessor, err := sp.txProcessors.getSuccess(tx, sp.appConfig)
		if err != nil {
			sp.logger.Error("Failed to get tx processor for processed tx", "tx", tx, "err", err)
		} else if txProcessor.GetType() == common.BridgingTxTypeBridgingRequest {
			err := bridgingRequestStateUpdater.Invalid(common.NewBridgingRequestStateKey(
				tx.OriginChainID, common.Hash(tx.Hash)))

			if err != nil {
				sp.logger.Error(
					"error while updating a bridging request state to Invalid",
					"srcChainId", tx.OriginChainID,
					"srcTxHash", tx.Hash, "err", err)
			}
		}
	}
}
