package processor

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	oracleCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_eth/core"
	"github.com/Ethernal-Tech/apex-bridge/telemetry"
	eventTrackerStore "github.com/Ethernal-Tech/blockchain-event-tracker/store"
	"github.com/hashicorp/go-hclog"
)

const (
	TTLInsuranceOffset = 2
)

var _ oracleCore.SpecificChainTxsProcessorState = (*EthStateProcessor)(nil)

type EthStateProcessor struct {
	ctx          context.Context
	appConfig    *oracleCore.AppConfig
	db           core.EthTxsProcessorDB
	txProcessors *txProcessorsCollection
	indexerDbs   map[string]eventTrackerStore.EventTrackerStore
	logger       hclog.Logger

	state *perTickState
}

func NewEthStateProcessor(
	ctx context.Context,
	appConfig *oracleCore.AppConfig,
	db core.EthTxsProcessorDB,
	txProcessors *txProcessorsCollection,
	indexerDbs map[string]eventTrackerStore.EventTrackerStore,
	logger hclog.Logger,
) *EthStateProcessor {
	return &EthStateProcessor{
		ctx:          ctx,
		appConfig:    appConfig,
		db:           db,
		txProcessors: txProcessors,
		indexerDbs:   indexerDbs,
		logger:       logger,
	}
}

func (sp *EthStateProcessor) GetChainType() string {
	return common.ChainTypeEVMStr
}

func (sp *EthStateProcessor) Reset() {
	sp.state = &perTickState{
		updateData:                    &core.EthUpdateTxsData{},
		innerActionHashToActualTxHash: make(map[string]common.Hash),
	}
}

func (sp *EthStateProcessor) RunChecks(
	bridgeClaims *oracleCore.BridgeClaims,
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

	sp.state.expectedTxsMap = make(map[string]*core.BridgeExpectedEthTx, len(expectedTxs))
	for _, expectedTx := range expectedTxs {
		sp.state.expectedTxsMap[string(expectedTx.ToEthTxKey())] = expectedTx
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

func (sp *EthStateProcessor) ProcessSubmitClaimsEvents(
	events *oracleCore.SubmitClaimsEvents, claims *oracleCore.BridgeClaims) {
	if len(events.NotEnoughFunds) > 0 {
		sp.processNotEnoughFundsEvents(events.NotEnoughFunds, claims)
	}

	if len(events.BatchExecutionInfo) > 0 {
		sp.processBatchExecutionInfoEvent(events.BatchExecutionInfo, claims)
	}
}

func (sp *EthStateProcessor) processBatchExecutionInfoEvent(
	events []*oracleCore.BatchExecutionInfoEvent, claims *oracleCore.BridgeClaims,
) {
	newProcessedTxsMap := make(map[string]*core.ProcessedEthTx, 0)
	newUnprocessedTxs := make([]*core.EthTx, 0)

	keys := make([][]byte, 0)
	for _, brc := range claims.BridgingRequestClaims {
		key := core.ToEthTxKey(common.ToStrChainID(brc.SourceChainId), brc.ObservedTransactionHash)
		keys = append(keys, key)
	}

	txs, err := sp.db.GetPendingTxs(keys)
	if err != nil {
		sp.logger.Error("couldn't find txs for BatchExecutionInfoEvent event", "err", err)
	}

	for _, tx := range txs {
		processedTx := tx.ToProcessedEthTx(false)
		newProcessedTxsMap[string(processedTx.Key())] = processedTx
	}

	for _, event := range events {
		txs, err := sp.getTxsFromBatchEvent(event)
		if err != nil {
			sp.logger.Error("couldn't find txs for BatchExecutionInfoEvent event", "event", event, "err", err)

			continue
		}

		if event.IsFailedClaim {
			for _, tx := range txs {
				tx.TryCount++
				tx.LastTimeTried = time.Time{}
				failedProcessedTx := tx.ToProcessedEthTx(false)
				delete(newProcessedTxsMap, string(failedProcessedTx.Key()))
			}

			newUnprocessedTxs = append(newUnprocessedTxs, txs...)
		}
	}

	newProcessedTxs := make([]*core.ProcessedEthTx, 0)
	for _, tx := range newProcessedTxsMap {
		newProcessedTxs = append(newProcessedTxs, tx)
	}

	sp.state.updateData.MovePendingToProcessed = newProcessedTxs
	sp.state.updateData.MovePendingToUnprocessed = newUnprocessedTxs
}

func (sp *EthStateProcessor) getTxsFromBatchEvent(
	event *oracleCore.BatchExecutionInfoEvent,
) ([]*core.EthTx, error) {
	keys := make([][]byte, len(event.TxHashes))

	for idx, hash := range event.TxHashes {
		key := core.ToEthTxKey(common.ToStrChainID(hash.SourceChainId), hash.ObservedTransactionHash)
		keys[idx] = key
	}

	txs, err := sp.db.GetPendingTxs(keys)
	if err != nil {
		return nil, err
	}

	return txs, nil
}

func (sp *EthStateProcessor) processNotEnoughFundsEvents(
	events []*oracleCore.NotEnoughFundsEvent, claims *oracleCore.BridgeClaims,
) {
	allPendingMap := make(map[string]*core.EthTx, len(sp.state.updateData.MoveUnprocessedToPending))
	for _, tx := range sp.state.updateData.MoveUnprocessedToPending {
		allPendingMap[string(tx.ToEthTxKey())] = tx
	}

	now := time.Now().UTC()
	unprocessedToUpdate := make([]*core.EthTx, 0, len(events))

	for _, event := range events {
		txToUpdate, err := sp.findRejectedTxInPending(event, claims, allPendingMap)
		if err != nil {
			sp.logger.Error("couldn't find tx for NotEnoughFunds event", "event", event, "err", err)

			continue
		}

		delete(allPendingMap, string(txToUpdate.ToEthTxKey()))

		txToUpdate.TryCount++
		txToUpdate.LastTimeTried = now
		unprocessedToUpdate = append(unprocessedToUpdate, txToUpdate)

		sp.logger.Debug("updated unprocessedTx TryCount and LastTimeTried", "tx", txToUpdate)
	}

	filteredAllPending := make([]*core.EthTx, 0, len(allPendingMap))
	for _, tx := range allPendingMap {
		filteredAllPending = append(filteredAllPending, tx)
	}

	sp.state.updateData.MoveUnprocessedToPending = filteredAllPending
	sp.state.updateData.UpdateUnprocessed = append(sp.state.updateData.UpdateUnprocessed, unprocessedToUpdate...)
}

func (sp *EthStateProcessor) findRejectedTxInPending(
	event *oracleCore.NotEnoughFundsEvent, claims *oracleCore.BridgeClaims,
	allPendingMap map[string]*core.EthTx,
) (*core.EthTx, error) {
	switch event.ClaimeType {
	case oracleCore.BRCClaimType:
		brcIndex := event.Index.Uint64()
		if brcIndex >= uint64(len(claims.BridgingRequestClaims)) {
			return nil, fmt.Errorf(
				"invalid NotEnoughFundsEvent.Index: %d. BRCs len: %d", brcIndex, len(claims.BridgingRequestClaims))
		}

		brc := claims.BridgingRequestClaims[brcIndex]

		tx, exists := allPendingMap[string(
			core.ToEthTxKey(common.ToStrChainID(brc.SourceChainId), brc.ObservedTransactionHash))]
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

func (sp *EthStateProcessor) PersistNew(
	bridgeClaims *oracleCore.BridgeClaims,
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater,
) {
	err := sp.notifyBridgingRequestStateUpdater(bridgeClaims, bridgingRequestStateUpdater)
	if err != nil {
		sp.logger.Error("Error while updating bridging request states", "err", err)
	}

	// we should update db only if there are some changes needed
	if sp.state.updateData.Count() > 0 {
		sp.logger.Info("Updating txs", "data", sp.state.updateData)

		// see EthUpdateTxsData struct for comments
		if err := sp.db.UpdateTxs(sp.state.updateData); err != nil {
			sp.logger.Error("Failed to update txs", "err", err)
		}
	}
}

func (sp *EthStateProcessor) constructBridgeClaimsBlockInfo(
	chainID string,
	unprocessedTxs []*core.EthTx,
	expectedTxs []*core.BridgeExpectedEthTx,
	prevBlockInfo *core.BridgeClaimsBlockInfo,
) *core.BridgeClaimsBlockInfo {
	found := false
	minBlockNumber := uint64(math.MaxUint64)

	if len(unprocessedTxs) > 0 {
		// unprocessed are ordered by block number, so first in collection is min
		for _, tx := range unprocessedTxs {
			if prevBlockInfo == nil || prevBlockInfo.Number < tx.BlockNumber {
				minBlockNumber = tx.BlockNumber
				found = true

				break
			}
		}
	}

	if len(expectedTxs) > 0 {
		ecoDB := sp.indexerDbs[chainID]
		if ecoDB == nil {
			sp.logger.Error("Failed to get eth chain observer db", "chainId", chainID)
		} else {
			// expected are ordered by ttl, so first in collection is min
			for _, tx := range expectedTxs {
				fromBlockNumber := tx.TTL + TTLInsuranceOffset

				lastProcessedBlock, err := ecoDB.GetLastProcessedBlock()
				if err != nil {
					sp.logger.Error("Failed to get last processed block",
						"chainId", chainID, "err", err)
				} else if lastProcessedBlock >= fromBlockNumber && fromBlockNumber < minBlockNumber &&
					(prevBlockInfo == nil || prevBlockInfo.Number < fromBlockNumber) {
					minBlockNumber = fromBlockNumber
					found = true

					break
				}
			}
		}
	}

	if found {
		return &core.BridgeClaimsBlockInfo{
			ChainID: chainID,
			Number:  minBlockNumber,
		}
	}

	return nil
}

func (sp *EthStateProcessor) checkUnprocessedTxs(
	bridgeClaims *oracleCore.BridgeClaims,
	maxClaimsToGroup int,
) {
	var relevantUnprocessedTxs []*core.EthTx

	for _, unprocessedTx := range sp.state.unprocessedTxs {
		if sp.state.blockInfo.EqualWithUnprocessed(unprocessedTx) &&
			!unprocessedTx.ShouldSkipForNow() {
			relevantUnprocessedTxs = append(relevantUnprocessedTxs, unprocessedTx)
		}
	}

	if len(relevantUnprocessedTxs) == 0 {
		return
	}

	//nolint:prealloc
	var (
		processedInvalidTxs  []*core.EthTx
		processedValidTxs    []*core.EthTx
		pendingTxs           []*core.EthTx
		processedExpectedTxs []*core.BridgeExpectedEthTx
		invalidTxsCounter    int
	)

	onInvalidTx := func(tx *core.EthTx) {
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
			if txProcessor.GetType() == common.BridgingTxTypeBatchExecution {
				key := string(unprocessedTx.ToExpectedEthTxKey())

				if expectedTx, exists := sp.state.expectedTxsMap[key]; exists {
					processedExpectedTxs = append(processedExpectedTxs, expectedTx)

					delete(sp.state.expectedTxsMap, key)
				}

				sp.state.innerActionHashToActualTxHash[string(core.ToEthTxKey(
					unprocessedTx.OriginChainID, unprocessedTx.InnerActionHash,
				))] = common.Hash(unprocessedTx.Hash)
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
			sp.state.updateData.MoveUnprocessedToProcessed, tx.ToProcessedEthTx(false))
	}

	for _, tx := range processedInvalidTxs {
		sp.state.updateData.MoveUnprocessedToProcessed = append(
			sp.state.updateData.MoveUnprocessedToProcessed, tx.ToProcessedEthTx(true))

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

func (sp *EthStateProcessor) checkExpectedTxs(
	bridgeClaims *oracleCore.BridgeClaims,
	maxClaimsToGroup int,
) {
	var relevantExpiredTxs []*core.BridgeExpectedEthTx

	ecoDB := sp.indexerDbs[sp.state.blockInfo.ChainID]
	if ecoDB == nil {
		sp.logger.Error("Failed to get eth chain observer db", "chainId", sp.state.blockInfo.ChainID)
	} else {
		// ensure always same order of iterating through expectedTxsMap
		keys := make([]string, 0, len(sp.state.expectedTxsMap))
		for k := range sp.state.expectedTxsMap {
			keys = append(keys, k)
		}

		sort.Strings(keys)

		for _, key := range keys {
			expectedTx := sp.state.expectedTxsMap[key]

			fromBlockNumber := expectedTx.TTL + TTLInsuranceOffset

			lastBlockProcessed, err := ecoDB.GetLastProcessedBlock()
			if err != nil {
				sp.logger.Error("Failed to get last processed block",
					"chainId", expectedTx.ChainID, "err", err)

				break
			}

			if lastBlockProcessed >= fromBlockNumber && sp.state.blockInfo.EqualWithExpected(
				expectedTx, fromBlockNumber) {
				relevantExpiredTxs = append(relevantExpiredTxs, expectedTx)
			}
		}
	}

	if !bridgeClaims.CanAddMore(maxClaimsToGroup) || len(relevantExpiredTxs) == 0 {
		return
	}

	//nolint:prealloc
	var (
		invalidRelevantExpiredTxs   []*core.BridgeExpectedEthTx
		processedRelevantExpiredTxs []*core.BridgeExpectedEthTx
	)

	onInvalidTx := func(tx *core.BridgeExpectedEthTx) {
		// expired, but can not process, so we mark it as invalid
		invalidRelevantExpiredTxs = append(invalidRelevantExpiredTxs, tx)
	}

	for _, expiredTx := range relevantExpiredTxs {
		processedTx, _ := sp.db.GetProcessedTxByInnerActionTxHash(expiredTx.ChainID, expiredTx.Hash)
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

func (sp *EthStateProcessor) notifyBridgingRequestStateUpdater(
	bridgeClaims *oracleCore.BridgeClaims,
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater,
) error {
	if len(bridgeClaims.BridgingRequestClaims) > 0 {
		for _, brClaim := range bridgeClaims.BridgingRequestClaims {
			err := bridgingRequestStateUpdater.SubmittedToBridge(common.BridgingRequestStateKey{
				SourceChainID: common.ToStrChainID(brClaim.SourceChainId),
				SourceTxHash:  brClaim.ObservedTransactionHash,
			}, common.ToStrChainID(brClaim.DestinationChainId))

			if err != nil {
				sp.logger.Error(
					"error while updating a bridging request state to SubmittedToBridge",
					"sourceChainId", common.ToStrChainID(brClaim.SourceChainId),
					"sourceTxHash", brClaim.ObservedTransactionHash, "err", err)
			}
		}
	}

	if len(bridgeClaims.BatchExecutedClaims) > 0 {
		for _, beClaim := range bridgeClaims.BatchExecutedClaims {
			txHash, exists := sp.state.innerActionHashToActualTxHash[string(core.ToEthTxKey(
				common.ToStrChainID(beClaim.ChainId), beClaim.ObservedTransactionHash))]
			if !exists {
				sp.logger.Error(
					"Failed to get txHash of a processed tx, based on BatchExecutedClaim.ObservedTransactionHash",
					"bec", beClaim)
			}

			err := bridgingRequestStateUpdater.ExecutedOnDestination(
				common.ToStrChainID(beClaim.ChainId),
				beClaim.BatchNonceId,
				txHash)

			if err != nil {
				sp.logger.Error(
					"error while updating bridging request states to ExecutedOnDestination",
					"destinationChainId", common.ToStrChainID(beClaim.ChainId), "batchId", beClaim.BatchNonceId,
					"destinationTxHash", txHash, "err", err)
			}
		}
	}

	if len(bridgeClaims.BatchExecutionFailedClaims) > 0 {
		for _, befClaim := range bridgeClaims.BatchExecutionFailedClaims {
			err := bridgingRequestStateUpdater.FailedToExecuteOnDestination(
				common.ToStrChainID(befClaim.ChainId),
				befClaim.BatchNonceId)

			if err != nil {
				sp.logger.Error(
					"error while updating bridging request states to FailedToExecuteOnDestination",
					"destinationChainId", common.ToStrChainID(befClaim.ChainId),
					"batchId", befClaim.BatchNonceId, "err", err)
			}
		}
	}

	for _, tx := range sp.state.allProcessedInvalid {
		txProcessor, err := sp.txProcessors.getSuccess(tx, sp.appConfig)
		if err != nil {
			sp.logger.Error("Failed to get tx processor for processed tx", "tx", tx, "err", err)
		} else if txProcessor.GetType() == common.BridgingTxTypeBridgingRequest {
			err := bridgingRequestStateUpdater.Invalid(common.BridgingRequestStateKey{
				SourceChainID: tx.OriginChainID,
				SourceTxHash:  common.Hash(tx.Hash),
			})

			if err != nil {
				sp.logger.Error(
					"error while updating a bridging request state to Invalid",
					"sourceChainId", tx.OriginChainID,
					"sourceTxHash", tx.Hash, "err", err)
			}
		}
	}

	return nil
}
