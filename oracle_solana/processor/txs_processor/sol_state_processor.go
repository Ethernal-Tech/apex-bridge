package processor

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	oracleCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_solana/core"
	"github.com/Ethernal-Tech/apex-bridge/telemetry"
	solanaTxsStore "github.com/Ethernal-Tech/solana-infrastructure/tracker/store"
	solana "github.com/gagliardetto/solana-go"
	"github.com/hashicorp/go-hclog"
)

type evtErr struct {
	evt *oracleCore.DBBatchInfoEvent
	err error
}

const (
	TTLInsuranceOffset             = 2
	logLastNBatchInfoSkippedEvents = 10
)

var _ oracleCore.SpecificChainTxsProcessorState = (*SolStateProcessor)(nil)

type SolStateProcessor struct {
	ctx          context.Context
	appConfig    *oracleCore.AppConfig
	db           core.SolanaTxsProcessorDB
	txProcessors *txProcessorsCollection
	indexerDbs   map[string]solanaTxsStore.StorageHandler
	logger       hclog.Logger

	state *perTickState
}

func NewSolStateProcessor(
	ctx context.Context,
	appConfig *oracleCore.AppConfig,
	db core.SolanaTxsProcessorDB,
	txProcessors *txProcessorsCollection,
	indexerDbs map[string]solanaTxsStore.StorageHandler,
	logger hclog.Logger,
) *SolStateProcessor {
	return &SolStateProcessor{
		ctx:          ctx,
		appConfig:    appConfig,
		db:           db,
		txProcessors: txProcessors,
		indexerDbs:   indexerDbs,
		logger:       logger,
	}
}

func (s *SolStateProcessor) GetChainType() string {
	return common.ChainTypeSolanaStr
}

func (s *SolStateProcessor) Reset() {
	s.state = &perTickState{updateData: &core.SolanaUpdateTxsData{}}
}

func (s *SolStateProcessor) ProcessSavedEvents() {
	var batchEvents []*oracleCore.DBBatchInfoEvent

	for _, chain := range s.appConfig.SolanaChains {
		chainBatchEvents, err := s.db.GetUnprocessedBatchEvents(chain.ChainID)
		if err != nil {
			s.logger.Error("Failed to get unprocessed batch events", "err", err)

			continue
		}

		batchEvents = append(batchEvents, chainBatchEvents...)
	}

	if len(batchEvents) > 0 {
		s.logger.Debug("Processing stored BatchExecutionInfoEvent events", "cnt", len(batchEvents))

		processedBatchEvents, _ := s.processBatchExecutionInfoEvents(batchEvents)

		if len(processedBatchEvents) > 0 {
			s.logger.Debug("Removing BatchExecutionInfoEvent events from db", "events", processedBatchEvents)
			s.state.updateData.RemoveBatchInfoEvents = processedBatchEvents
		}
	}
}

func (s *SolStateProcessor) RunChecks(
	bridgeClaims *oracleCore.BridgeClaims,
	chainID string,
	maxClaimsToGroup int,
	priority uint8,
) {
	expectedTxs, err := s.db.GetExpectedTxs(chainID, priority, 0)
	if err != nil {
		s.logger.Error("Failed to get expected txs", "err", err)

		return
	}

	s.state.unprocessedTxs, err = s.db.GetUnprocessedTxs(chainID, priority, 0)
	if err != nil {
		s.logger.Error("Failed to get unprocessed txs", "err", err)

		return
	}

	// needed for the guarantee that both unprocessedTxs and expectedTxs are processed in order of slot
	// and prevent the situation when there are always enough unprocessedTxs to fill out claims,
	// that all claims are filled only from unprocessedTxs and never from expectedTxs
	s.state.blockInfo = s.constructBridgeClaimsBlockInfo(
		chainID, s.state.unprocessedTxs, expectedTxs, nil)
	if s.state.blockInfo == nil {
		return
	}

	s.state.expectedTxsMap = make(map[string]*core.BridgeExpectedSolanaTx, len(expectedTxs))
	for _, expectedTx := range expectedTxs {
		s.state.expectedTxsMap[string(expectedTx.ToSolanaTxKey())] = expectedTx
	}

	for {
		s.logger.Debug("Processing",
			"for chainID", s.state.blockInfo.ChainID,
			"blockInfo", s.state.blockInfo)

		s.checkUnprocessedTxs(bridgeClaims, maxClaimsToGroup)
		s.checkExpectedTxs(bridgeClaims, maxClaimsToGroup)

		if !bridgeClaims.CanAddMore(maxClaimsToGroup) {
			break
		}

		s.state.blockInfo = s.constructBridgeClaimsBlockInfo(
			chainID, s.state.unprocessedTxs, expectedTxs, s.state.blockInfo)
		if s.state.blockInfo == nil {
			break
		}
	}
}

func (s *SolStateProcessor) ProcessSubmitClaimsEvents(
	events *oracleCore.SubmitClaimsEvents, claims *oracleCore.BridgeClaims) {
	if len(events.NotEnoughFunds) > 0 {
		s.processNotEnoughFundsEvents(events.NotEnoughFunds, claims)
	}

	if len(events.BatchExecutionInfo) > 0 {
		_, skippedEvents := s.processBatchExecutionInfoEvents(events.BatchExecutionInfo)
		if len(skippedEvents) > 0 {
			s.logger.Debug("Storing BatchExecutionInfoEvent events", "cnt", len(skippedEvents))
			s.state.updateData.AddBatchInfoEvents = skippedEvents
		}
	}
}

func (s *SolStateProcessor) processBatchExecutionInfoEvents(
	events []*oracleCore.DBBatchInfoEvent,
) ([]*oracleCore.DBBatchInfoEvent, []*oracleCore.DBBatchInfoEvent) {
	var (
		processedEvents      = make([]*oracleCore.DBBatchInfoEvent, 0, len(events))
		newProcessedTxs      []oracleCore.BaseProcessedTx
		newUnprocessedTxs    []oracleCore.BaseTx
		skippedEventsWithErr []evtErr
	)

	for _, event := range events {
		txs, err := s.getTxsFromBatchEvent(event)
		if err != nil {
			skippedEventsWithErr = append(skippedEventsWithErr, evtErr{evt: event, err: err})

			continue
		}

		processedEvents = append(processedEvents, event)

		if event.IsFailedClaim {
			for _, tx := range txs {
				tx.IncrementBatchTryCount()
				tx.ResetSubmitTryCount()
				tx.SetLastTimeTried(time.Time{})

				for _, batchTx := range event.TxHashes {
					if s.appConfig.ChainIDConverter.ToChainIDStr(batchTx.SourceChainID) == tx.GetChainID() &&
						batchTx.ObservedTransactionHash == common.Hash(tx.GetTxHash()) &&
						batchTx.TransactionType == uint8(common.RefundConfirmedTxType) {
						tx.IncrementRefundTryCount()

						break
					}
				}

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

		s.logger.Info(
			fmt.Sprintf("couldn't find txs for some BatchExecutionInfoEvent events. listing last %d",
				logLastNBatchInfoSkippedEvents))

		for _, item := range lastNSkippedEventsWithErr {
			s.logger.Info(
				"couldn't find txs for BatchExecutionInfoEvent event",
				"event", item.evt, "err", item.err)
		}
	}

	skippedEvents := make([]*oracleCore.DBBatchInfoEvent, len(skippedEventsWithErr))
	for idx, item := range skippedEventsWithErr {
		skippedEvents[idx] = item.evt
	}

	s.state.updateData.MovePendingToProcessed = newProcessedTxs
	s.state.updateData.MovePendingToUnprocessed = newUnprocessedTxs

	return processedEvents, skippedEvents
}

func (s *SolStateProcessor) getTxsFromBatchEvent(
	event *oracleCore.DBBatchInfoEvent,
) ([]oracleCore.BaseTx, error) {
	result := make([]oracleCore.BaseTx, len(event.TxHashes))

	for idx, hash := range event.TxHashes {
		tx, err := s.db.GetPendingTx(
			oracleCore.DBTxID{
				ChainID: s.appConfig.ChainIDConverter.ToChainIDStr(hash.SourceChainID),
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

func (s *SolStateProcessor) processNotEnoughFundsEvents(
	events []*oracleCore.NotEnoughFundsEvent, claims *oracleCore.BridgeClaims,
) {
	allPendingMap := make(map[string]*core.SolanaTx, len(s.state.updateData.MoveUnprocessedToPending))
	for _, tx := range s.state.updateData.MoveUnprocessedToPending {
		allPendingMap[string(tx.ToSolanaTxKey())] = tx
	}

	now := time.Now().UTC()
	unprocessedToUpdate := make([]*core.SolanaTx, 0, len(events))

	for _, event := range events {
		txToUpdate, err := s.findRejectedTxInPending(event, claims, allPendingMap)
		if err != nil {
			s.logger.Error("couldn't find tx for NotEnoughFunds event", "event", event, "err", err)

			continue
		}

		delete(allPendingMap, string(txToUpdate.ToSolanaTxKey()))

		txToUpdate.SubmitTryCount++
		txToUpdate.LastTimeTried = now
		unprocessedToUpdate = append(unprocessedToUpdate, txToUpdate)

		s.logger.Debug("updated unprocessedTx SubmitTryCount and LastTimeTried", "tx", txToUpdate)
	}

	filteredAllPending := make([]*core.SolanaTx, 0, len(allPendingMap))
	for _, tx := range allPendingMap {
		filteredAllPending = append(filteredAllPending, tx)
	}

	s.state.updateData.MoveUnprocessedToPending = filteredAllPending
	s.state.updateData.UpdateUnprocessed = append(s.state.updateData.UpdateUnprocessed, unprocessedToUpdate...)
}

func (s *SolStateProcessor) findRejectedTxInPending(
	event *oracleCore.NotEnoughFundsEvent, claims *oracleCore.BridgeClaims,
	allPendingMap map[string]*core.SolanaTx,
) (*core.SolanaTx, error) {
	chainIDConverter := s.appConfig.ChainIDConverter

	if strings.HasPrefix(event.ClaimeType, oracleCore.BRCClaimType) {
		brcIndex := event.Index.Uint64()
		if brcIndex >= uint64(len(claims.BridgingRequestClaims)) {
			return nil, fmt.Errorf(
				"invalid NotEnoughFundsEvent.Index: %d. BRCs len: %d", brcIndex, len(claims.BridgingRequestClaims))
		}

		brc := claims.BridgingRequestClaims[brcIndex]

		var sig solana.Signature

		copy(sig[:], brc.ObservedTransactionHash)

		tx, exists := allPendingMap[string(
			core.ToSolanaTxKey(chainIDConverter.ToChainIDStr(brc.SourceChainId), sig))]
		if !exists {
			return nil, fmt.Errorf(
				"BRC not found in MoveUnprocessedToPending for index: %d", brcIndex)
		}

		return tx, nil
	} else if strings.HasPrefix(event.ClaimeType, oracleCore.RRCClaimType) {
		rrcIndex := event.Index.Uint64()
		if rrcIndex >= uint64(len(claims.RefundRequestClaims)) {
			return nil, fmt.Errorf(
				"invalid NotEnoughFundsEvent.Index: %d. RRCs len: %d", rrcIndex, len(claims.RefundRequestClaims))
		}

		rrc := claims.RefundRequestClaims[rrcIndex]

		var sig solana.Signature

		copy(sig[:], rrc.OriginTransactionHash[:])

		tx, exists := allPendingMap[string(
			core.ToSolanaTxKey(chainIDConverter.ToChainIDStr(rrc.OriginChainId), sig))]
		if !exists {
			return nil, fmt.Errorf(
				"RRC not found in MoveUnprocessedToPending for index: %d", rrcIndex)
		}

		return tx, nil
	}

	return nil, fmt.Errorf(
		"unsupported NotEnoughFundsEvent.claimType: %s", event.ClaimeType)
}

func (s *SolStateProcessor) PersistNew() {
	if s.state.updateData.Count() > 0 {
		s.logger.Info("Updating txs", "data", s.state.updateData)

		if err := s.db.UpdateTxs(s.state.updateData, s.appConfig.ChainIDConverter); err != nil {
			s.logger.Error("Failed to update txs", "err", err)
		}
	}
}

func (s *SolStateProcessor) constructBridgeClaimsBlockInfo(
	chainID string,
	unprocessedTxs []*core.SolanaTx,
	expectedTxs []*core.BridgeExpectedSolanaTx,
	prevBlockInfo *core.BridgeClaimsSlotInfo,
) *core.BridgeClaimsSlotInfo {
	found := false
	minSlot := uint64(math.MaxUint64)

	if len(unprocessedTxs) > 0 {
		// unprocessed are ordered by slot number, so first in collection is min
		for _, tx := range unprocessedTxs {
			if prevBlockInfo == nil || prevBlockInfo.Number < tx.SlotNumber {
				minSlot = tx.SlotNumber
				found = true

				break
			}
		}
	}

	if len(expectedTxs) > 0 {
		soDB := s.indexerDbs[chainID]
		if soDB == nil {
			s.logger.Error("Failed to get solana chain observer db", "chainId", chainID)
		} else {
			// expected are ordered by ttl, so first in collection is min
			for _, tx := range expectedTxs {
				fromSlot := tx.TTL + TTLInsuranceOffset

				lastProcessedSlot, err := soDB.ReadSlot()
				if err != nil {
					s.logger.Error("Failed to get last processed slot",
						"chainId", chainID, "err", err)
				} else if lastProcessedSlot >= fromSlot && fromSlot < minSlot &&
					(prevBlockInfo == nil || prevBlockInfo.Number < fromSlot) {
					minSlot = fromSlot
					found = true

					break
				}
			}
		}
	}

	if found {
		return &core.BridgeClaimsSlotInfo{
			ChainID: chainID,
			Number:  minSlot,
		}
	}

	return nil
}

func (s *SolStateProcessor) checkUnprocessedTxs(
	bridgeClaims *oracleCore.BridgeClaims,
	maxClaimsToGroup int,
) {
	var relevantUnprocessedTxs []*core.SolanaTx

	for _, unprocessedTx := range s.state.unprocessedTxs {
		if s.state.blockInfo.EqualWithUnprocessed(unprocessedTx) && oracleCore.IsTxReady(
			unprocessedTx.SubmitTryCount, unprocessedTx.LastTimeTried, s.appConfig.RetryUnprocessedSettings) {
			relevantUnprocessedTxs = append(relevantUnprocessedTxs, unprocessedTx)
		}
	}

	if len(relevantUnprocessedTxs) == 0 {
		return
	}

	var (
		processedInvalidTxs  []*core.SolanaTx
		processedValidTxs    []*core.SolanaTx
		pendingTxs           []*core.SolanaTx
		processedExpectedTxs []*core.BridgeExpectedSolanaTx
		invalidTxsCounter    int
	)

	onInvalidTx := func(tx *core.SolanaTx) {
		processedInvalidTxs = append(processedInvalidTxs, tx)
		invalidTxsCounter++
	}

	// check unprocessed txs from indexers
	for _, unprocessedTx := range relevantUnprocessedTxs {
		s.logger.Debug("Checking if tx is relevant", "tx", unprocessedTx)

		txProcessor, err := s.txProcessors.getSuccess(unprocessedTx, s.appConfig)
		if err != nil {
			s.logger.Error("Failed to get tx processor for unprocessed tx", "tx", unprocessedTx, "err", err)

			onInvalidTx(unprocessedTx)

			continue
		}

		err = txProcessor.ValidateAndAddClaim(bridgeClaims, unprocessedTx, s.appConfig)
		if err != nil {
			s.logger.Error("Failed to ValidateAndAddClaim", "tx", unprocessedTx, "err", err)

			onInvalidTx(unprocessedTx)

			continue
		}

		if txProcessor.GetType() == common.BridgingTxTypeBridgingRequest ||
			txProcessor.GetType() == common.TxTypeRefundRequest {
			pendingTxs = append(pendingTxs, unprocessedTx)
		} else {
			key := string(unprocessedTx.ToExpectedSolanaTxKey())

			if expectedTx, exists := s.state.expectedTxsMap[key]; exists {
				processedExpectedTxs = append(processedExpectedTxs, expectedTx)

				delete(s.state.expectedTxsMap, key)
			}

			processedValidTxs = append(processedValidTxs, unprocessedTx)
		}

		if !bridgeClaims.CanAddMore(maxClaimsToGroup) {
			break
		}
	}

	if invalidTxsCounter > 0 {
		telemetry.UpdateOracleClaimsInvalidCounter(s.state.blockInfo.ChainID, invalidTxsCounter) // update telemetry
	}

	for _, tx := range processedValidTxs {
		s.state.updateData.MoveUnprocessedToProcessed = append(
			s.state.updateData.MoveUnprocessedToProcessed, tx.ToProcessedSolanaTx(false))
	}

	for _, tx := range processedInvalidTxs {
		s.state.updateData.MoveUnprocessedToProcessed = append(
			s.state.updateData.MoveUnprocessedToProcessed, tx.ToProcessedSolanaTx(true))

		s.state.allProcessedInvalid = append(s.state.allProcessedInvalid, tx)
	}

	s.state.updateData.MoveUnprocessedToPending = append(s.state.updateData.MoveUnprocessedToPending, pendingTxs...)
	s.state.updateData.ExpectedProcessed = append(s.state.updateData.ExpectedProcessed, processedExpectedTxs...)

	s.logger.Debug("Checked all unprocessed",
		"for chainID", s.state.blockInfo.ChainID,
		"processedValidTxs", processedValidTxs,
		"processedInvalidTxs", processedInvalidTxs,
		"pendingTxs", pendingTxs,
		"processedExpectedTxs", processedExpectedTxs)
}

func (s *SolStateProcessor) checkExpectedTxs(
	bridgeClaims *oracleCore.BridgeClaims,
	maxClaimsToGroup int,
) {
	var relevantExpiredTxs []*core.BridgeExpectedSolanaTx

	soDB := s.indexerDbs[s.state.blockInfo.ChainID]
	if soDB == nil {
		s.logger.Error("Failed to get solana chain observer db", "chainId", s.state.blockInfo.ChainID)
	} else {
		// ensure always same order of iterating through expectedTxsMap
		keys := make([]string, 0, len(s.state.expectedTxsMap))
		for k := range s.state.expectedTxsMap {
			keys = append(keys, k)
		}

		sort.Strings(keys)

		for _, key := range keys {
			expectedTx := s.state.expectedTxsMap[key]

			fromSlot := expectedTx.TTL + TTLInsuranceOffset

			lastProcessedSlot, err := soDB.ReadSlot()
			if err != nil {
				s.logger.Error("Failed to get last processed slot",
					"chainId", expectedTx.ChainID, "err", err)

				break
			}

			if lastProcessedSlot >= fromSlot && s.state.blockInfo.EqualWithExpected(expectedTx, fromSlot) {
				relevantExpiredTxs = append(relevantExpiredTxs, expectedTx)
			}
		}
	}

	if !bridgeClaims.CanAddMore(maxClaimsToGroup) || len(relevantExpiredTxs) == 0 {
		return
	}

	//nolint:prealloc
	var (
		invalidRelevantExpiredTxs   []*core.BridgeExpectedSolanaTx
		processedRelevantExpiredTxs []*core.BridgeExpectedSolanaTx
	)

	onInvalidTx := func(tx *core.BridgeExpectedSolanaTx) {
		// expired, but can not process, so we mark it as invalid
		invalidRelevantExpiredTxs = append(invalidRelevantExpiredTxs, tx)
	}

	for _, expiredTx := range relevantExpiredTxs {
		processedTx, _ := s.db.GetProcessedTxByInnerActionTxHash(expiredTx.ChainID, expiredTx.TxSignature[:])
		if processedTx != nil && !processedTx.IsInvalid {
			// already sent the success claim
			processedRelevantExpiredTxs = append(processedRelevantExpiredTxs, expiredTx)

			continue
		}

		s.logger.Debug("Checking if expired tx is relevant", "expiredTx", expiredTx)

		txProcessor, err := s.txProcessors.getFailed(expiredTx, s.appConfig)
		if err != nil {
			s.logger.Error("Failed to get tx processor for expired tx", "tx", expiredTx, "err", err)

			onInvalidTx(expiredTx)

			continue
		}

		err = txProcessor.ValidateAndAddClaim(bridgeClaims, expiredTx, s.appConfig)
		if err != nil {
			s.logger.Error("Failed to ValidateAndAddClaim", "expiredTx", expiredTx, "err", err)

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
			s.state.blockInfo.ChainID, len(invalidRelevantExpiredTxs)) // update telemetry
	}

	s.state.updateData.ExpectedProcessed = append(s.state.updateData.ExpectedProcessed, processedRelevantExpiredTxs...)
	s.state.updateData.ExpectedInvalid = append(s.state.updateData.ExpectedInvalid, invalidRelevantExpiredTxs...)

	s.logger.Debug("Checked all expected",
		"for chainID", s.state.blockInfo.ChainID,
		"processedExpectedTxs", processedRelevantExpiredTxs,
		"invalidRelevantExpiredTxs", invalidRelevantExpiredTxs)
}

func (s *SolStateProcessor) UpdateBridgingRequestStates(
	bridgeClaims *oracleCore.BridgeClaims,
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater,
) {
	chainIDConverter := s.appConfig.ChainIDConverter

	if len(bridgeClaims.BridgingRequestClaims) > 0 || len(bridgeClaims.RefundRequestClaims) > 0 {
		notRejectedMap := make(map[string]bool, len(s.state.updateData.MoveUnprocessedToPending))
		for _, tx := range s.state.updateData.MoveUnprocessedToPending {
			notRejectedMap[string(tx.ToSolanaTxKey())] = true
		}

		updateToSubmittedToBridge := func(
			sourceChainId uint8, observedTransactionHash []byte, destinationChainId uint8, isRefund bool,
		) {
			srcChainID := chainIDConverter.ToChainIDStr(sourceChainId)

			var sig solana.Signature

			copy(sig[:], observedTransactionHash)

			key := core.ToSolanaTxKey(srcChainID, sig)
			if !notRejectedMap[string(key)] {
				return
			}

			var txHash common.Hash

			copy(txHash[:], observedTransactionHash)

			err := bridgingRequestStateUpdater.SubmittedToBridge(
				common.NewBridgingRequestStateKey(srcChainID, txHash, isRefund),
				chainIDConverter.ToChainIDStr(destinationChainId))

			if err != nil {
				s.logger.Error(
					"error while updating a bridging request state to",
					"state", common.BridgingRequestStateStatusStr(common.BridgingRequestStatusSubmittedToBridge, isRefund),
					"srcChainId", srcChainID, "srcTxHash", observedTransactionHash, "err", err)
			}
		}

		for _, brClaim := range bridgeClaims.BridgingRequestClaims {
			updateToSubmittedToBridge(
				brClaim.SourceChainId, brClaim.ObservedTransactionHash, brClaim.DestinationChainId, false)
		}

		for _, rrClaim := range bridgeClaims.RefundRequestClaims {
			updateToSubmittedToBridge(
				rrClaim.OriginChainId, rrClaim.OriginTransactionHash[:], rrClaim.OriginChainId, true)
		}
	}

	for _, tx := range s.state.allProcessedInvalid {
		txProcessor, err := s.txProcessors.getSuccess(tx, s.appConfig)
		if err != nil {
			s.logger.Error("Failed to get tx processor for processed tx", "tx", tx, "err", err)
		} else if txProcessor.GetType() == common.BridgingTxTypeBridgingRequest ||
			txProcessor.GetType() == common.TxTypeRefundRequest {
			var txHash common.Hash

			copy(txHash[:], tx.TxSignature[:])

			err := bridgingRequestStateUpdater.Invalid(common.NewBridgingRequestStateKey(
				tx.OriginChainID, txHash, false))
			if err != nil {
				s.logger.Error(
					"error while updating a bridging request state to Invalid",
					"sourceChainId", tx.OriginChainID,
					"sourceTxHash", tx.TxSignature, "err", err)
			}
		}
	}
}
