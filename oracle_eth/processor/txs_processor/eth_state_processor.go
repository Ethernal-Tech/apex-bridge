package processor

import (
	"context"
	"math"
	"sort"

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
	sp.state = &perTickState{}
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

	sp.state.allUnprocessed = append(sp.state.allUnprocessed, sp.state.unprocessedTxs...)

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

		_, processedValidTxs, processedInvalidTxs, processedExpectedTxs := sp.checkUnprocessedTxs(
			bridgeClaims, maxClaimsToGroup)

		_, processedRelevantExpiredTxs, invalidRelevantExpiredTxs := sp.checkExpectedTxs(
			bridgeClaims, maxClaimsToGroup)

		processedExpectedTxs = append(processedExpectedTxs, processedRelevantExpiredTxs...)

		sp.logger.Debug("Checked all",
			"for chainID", sp.state.blockInfo.ChainID,
			"processedValidTxs", processedValidTxs,
			"processedInvalidTxs", processedInvalidTxs,
			"processedExpectedTxs", processedExpectedTxs,
			"invalidRelevantExpiredTxs", invalidRelevantExpiredTxs)

		sp.state.allProcessedValid = append(sp.state.allProcessedValid, processedValidTxs...)
		sp.state.allProcessedInvalid = append(sp.state.allProcessedInvalid, processedInvalidTxs...)
		sp.state.allProcessedExpected = append(sp.state.allProcessedExpected, processedExpectedTxs...)
		sp.state.allInvalidRelevantExpired = append(sp.state.allInvalidRelevantExpired, invalidRelevantExpiredTxs...)

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
	// a TODO: implement this
}

func (sp *EthStateProcessor) PersistNew(
	bridgeClaims *oracleCore.BridgeClaims,
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater,
) {
	err := sp.notifyBridgingRequestStateUpdater(bridgeClaims, bridgingRequestStateUpdater)
	if err != nil {
		sp.logger.Error("Error while updating bridging request states", "err", err)
	}

	expectedInvalid := sp.state.allInvalidRelevantExpired
	expectedProcessed := sp.state.allProcessedExpected
	allProcessed := make(
		[]*core.ProcessedEthTx, 0, len(sp.state.allProcessedValid)+len(sp.state.allProcessedInvalid))

	for _, tx := range sp.state.allProcessedInvalid {
		allProcessed = append(allProcessed, tx.ToProcessedEthTx(true))
	}

	for _, tx := range sp.state.allProcessedValid {
		allProcessed = append(allProcessed, tx.ToProcessedEthTx(false))
	}

	// we should update db only if there are some changes needed
	if len(expectedInvalid)+len(expectedProcessed)+len(allProcessed) > 0 {
		sp.logger.Info("Marking expected txs", "invalid", expectedInvalid,
			"expected", expectedProcessed, "processed", allProcessed)

		if err := sp.db.MarkTxs(expectedInvalid, expectedProcessed, allProcessed); err != nil {
			sp.logger.Error("Failed to mark expected txs as invalid", "err", err)
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
) (
	[]*core.EthTx,
	[]*core.EthTx,
	[]*core.EthTx,
	[]*core.BridgeExpectedEthTx,
) {
	var relevantUnprocessedTxs []*core.EthTx

	for _, unprocessedTx := range sp.state.unprocessedTxs {
		if sp.state.blockInfo.EqualWithUnprocessed(unprocessedTx) {
			relevantUnprocessedTxs = append(relevantUnprocessedTxs, unprocessedTx)
		}
	}

	//nolint:prealloc
	var (
		processedInvalidTxs  []*core.EthTx
		processedValidTxs    []*core.EthTx
		processedExpectedTxs []*core.BridgeExpectedEthTx
		invalidTxsCounter    int
	)

	if len(relevantUnprocessedTxs) == 0 {
		return relevantUnprocessedTxs, processedValidTxs, processedInvalidTxs, processedExpectedTxs
	}

	onInvalidTx := func(tx *core.EthTx) {
		processedInvalidTxs = append(processedInvalidTxs, tx)
		invalidTxsCounter++
	}

	// check unprocessed txs from indexers
	for _, unprocessedTx := range relevantUnprocessedTxs {
		sp.logger.Debug("Checking if tx is relevant", "tx", unprocessedTx)

		txProcessor, err := sp.txProcessors.getSuccess(unprocessedTx.Metadata)
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

		if txProcessor.GetType() == common.BridgingTxTypeBatchExecution {
			key := string(unprocessedTx.ToExpectedEthTxKey())

			if expectedTx, exists := sp.state.expectedTxsMap[key]; exists {
				processedExpectedTxs = append(processedExpectedTxs, expectedTx)

				delete(sp.state.expectedTxsMap, key)
			}
		}

		processedValidTxs = append(processedValidTxs, unprocessedTx)

		if !bridgeClaims.CanAddMore(maxClaimsToGroup) {
			break
		}
	}

	if invalidTxsCounter > 0 {
		telemetry.UpdateOracleClaimsInvalidCounter(sp.state.blockInfo.ChainID, invalidTxsCounter) // update telemetry
	}

	return relevantUnprocessedTxs, processedValidTxs, processedInvalidTxs, processedExpectedTxs
}

func (sp *EthStateProcessor) checkExpectedTxs(
	bridgeClaims *oracleCore.BridgeClaims,
	maxClaimsToGroup int,
) (
	[]*core.BridgeExpectedEthTx,
	[]*core.BridgeExpectedEthTx,
	[]*core.BridgeExpectedEthTx,
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

	//nolint:prealloc
	var (
		invalidRelevantExpiredTxs   []*core.BridgeExpectedEthTx
		processedRelevantExpiredTxs []*core.BridgeExpectedEthTx
	)

	if !bridgeClaims.CanAddMore(maxClaimsToGroup) ||
		len(relevantExpiredTxs) == 0 {
		return relevantExpiredTxs, processedRelevantExpiredTxs, invalidRelevantExpiredTxs
	}

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

		txProcessor, err := sp.txProcessors.getFailed(expiredTx.Metadata)
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

	return relevantExpiredTxs, processedRelevantExpiredTxs, invalidRelevantExpiredTxs
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
			var (
				txHash common.Hash
				found  bool
			)

			for _, processedTx := range sp.state.allProcessedValid {
				if processedTx.OriginChainID == common.ToStrChainID(beClaim.ChainId) &&
					processedTx.InnerActionHash == beClaim.ObservedTransactionHash {
					txHash = common.Hash(processedTx.Hash)
					found = true

					break
				}
			}

			if !found {
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
		txProcessor, err := sp.txProcessors.getSuccess(tx.Metadata)
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
