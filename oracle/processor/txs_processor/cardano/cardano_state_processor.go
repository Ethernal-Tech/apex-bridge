package cardanotxsprocessor

import (
	"bytes"
	"context"
	"math"
	"sort"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/apex-bridge/telemetry"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/hashicorp/go-hclog"
)

const (
	TTLInsuranceOffset = 2
)

var _ core.SpecificChainTxsProcessorState = (*CardanoStateProcessor)(nil)

type CardanoStateProcessor struct {
	ctx          context.Context
	appConfig    *core.AppConfig
	db           core.CardanoTxsProcessorDB
	txProcessors *txProcessorsCollection
	indexerDbs   map[string]indexer.Database
	logger       hclog.Logger

	state *perTickState
}

func NewCardanoStateProcessor(
	ctx context.Context,
	appConfig *core.AppConfig,
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

func (c *CardanoStateProcessor) Reset() {
	c.state = &perTickState{}
}

// PrepareRun implements new_processor.SpecificChainStateProcessor.
func (c *CardanoStateProcessor) PrepareFirstCheck(chainID string, priority uint8) bool {
	var err error

	c.state.expectedTxs, err = c.db.GetExpectedTxs(chainID, priority, 0)
	if err != nil {
		c.logger.Error("Failed to get expected txs", "err", err)

		return false
	}

	c.state.unprocessedTxs, err = c.db.GetUnprocessedTxs(chainID, priority, 0)
	if err != nil {
		c.logger.Error("Failed to get unprocessed txs", "err", err)

		return false
	}

	c.state.unprocessed = append(c.state.unprocessed, c.state.unprocessedTxs...)

	// needed for the guarantee that both unprocessedTxs and expectedTxs are processed in order of slot
	// and prevent the situation when there are always enough unprocessedTxs to fill out claims,
	// that all claims are filled only from unprocessedTxs and never from expectedTxs
	c.state.blockInfo = c.constructBridgeClaimsBlockInfo(
		chainID, c.state.unprocessedTxs, c.state.expectedTxs, nil)
	if c.state.blockInfo == nil {
		return false
	}

	c.state.expectedTxsMap = make(map[string]*core.BridgeExpectedCardanoTx, len(c.state.expectedTxs))
	for _, expectedTx := range c.state.expectedTxs {
		c.state.expectedTxsMap[string(expectedTx.ToCardanoTxKey())] = expectedTx
	}

	return true
}

func (c *CardanoStateProcessor) RunCheck(
	bridgeClaims *core.BridgeClaims,
	maxClaimsToGroup int,
) {
	c.logger.Debug("Processing",
		"for chainID", c.state.blockInfo.ChainID,
		"blockInfo", c.state.blockInfo)

	_, processedTxs, processedExpectedTxs := c.checkUnprocessedTxs(bridgeClaims, maxClaimsToGroup)

	_, processedRelevantExpiredTxs, invalidRelevantExpiredTxs := c.checkExpectedTxs(bridgeClaims, maxClaimsToGroup)

	processedExpectedTxs = append(processedExpectedTxs, processedRelevantExpiredTxs...)

	c.logger.Debug("Checked all", "for chainID", c.state.blockInfo.ChainID,
		"processedTxs", processedTxs, "processedExpectedTxs", processedExpectedTxs,
		"invalidRelevantExpiredTxs", invalidRelevantExpiredTxs)
}

func (c *CardanoStateProcessor) NextBlockInfo() bool {
	c.state.blockInfo = c.constructBridgeClaimsBlockInfo(
		c.state.blockInfo.ChainID, c.state.unprocessedTxs,
		c.state.expectedTxs, c.state.blockInfo)

	return c.state.blockInfo != nil
}

func (c *CardanoStateProcessor) PersistNew(
	bridgeClaims *core.BridgeClaims,
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater,
) {
	err := c.notifyBridgingRequestStateUpdater(bridgeClaims, bridgingRequestStateUpdater)
	if err != nil {
		c.logger.Error("Error while updating bridging request states", "err", err)
	}

	// we should only change this in db if submit succeeded (not really, but for convenience)
	if len(c.state.invalidRelevantExpired) > 0 {
		c.logger.Info("Marking expected txs as invalid", "txs", c.state.invalidRelevantExpired)

		err := c.db.MarkExpectedTxsAsInvalid(c.state.invalidRelevantExpired)
		if err != nil {
			c.logger.Error("Failed to mark expected txs as invalid", "err", err)
		}
	}

	// we should only change this in db if submit succeeded
	if len(c.state.processedExpected) > 0 {
		c.logger.Info("Marking expected txs as processed", "txs", c.state.processedExpected)

		err := c.db.MarkExpectedTxsAsProcessed(c.state.processedExpected)
		if err != nil {
			c.logger.Error("Failed to mark expected txs as processed", "err", err)
		}
	}

	// we should only change this in db if submit succeeded
	if len(c.state.processed) > 0 {
		c.logger.Info("Marking txs as processed", "txs", c.state.processed)

		err := c.db.MarkUnprocessedTxsAsProcessed(c.state.processed)
		if err != nil {
			c.logger.Error("Failed to mark txs as processed", "err", err)
		}
	}
}

func (c *CardanoStateProcessor) constructBridgeClaimsBlockInfo(
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
		ccoDB := c.indexerDbs[chainID]
		if ccoDB == nil {
			c.logger.Error("Failed to get cardano chain observer db", "chainId", chainID)
		} else {
			// expected are ordered by ttl, so first in collection is min
			for _, tx := range expectedTxs {
				fromSlot := tx.TTL + TTLInsuranceOffset

				blocks, err := ccoDB.GetConfirmedBlocksFrom(fromSlot, 1)
				if err != nil {
					c.logger.Error("Failed to get confirmed blocks", "fromSlot", fromSlot, "chainId", chainID, "err", err)
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

func (c *CardanoStateProcessor) checkUnprocessedTxs(
	bridgeClaims *core.BridgeClaims,
	maxClaimsToGroup int,
) (
	[]*core.CardanoTx,
	[]*core.ProcessedCardanoTx,
	[]*core.BridgeExpectedCardanoTx,
) {
	var relevantUnprocessedTxs []*core.CardanoTx

	for _, unprocessedTx := range c.state.unprocessedTxs {
		if c.state.blockInfo.EqualWithUnprocessed(unprocessedTx) {
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
		c.logger.Debug("Checking if tx is relevant", "tx", unprocessedTx)

		txProcessor, err := c.txProcessors.getSuccess(unprocessedTx.Metadata)
		if err != nil {
			c.logger.Error("Failed to get tx processor for unprocessed tx", "tx", unprocessedTx, "err", err)

			onInvalidTx(unprocessedTx)

			continue
		}

		err = txProcessor.ValidateAndAddClaim(bridgeClaims, unprocessedTx, c.appConfig)
		if err != nil {
			c.logger.Error("Failed to ValidateAndAddClaim", "tx", unprocessedTx, "err", err)

			onInvalidTx(unprocessedTx)

			continue
		}

		key := string(unprocessedTx.ToCardanoTxKey())

		if expectedTx, exists := c.state.expectedTxsMap[key]; exists {
			processedExpectedTxs = append(processedExpectedTxs, expectedTx)

			delete(c.state.expectedTxsMap, key)
		}

		processedTxs = append(processedTxs, unprocessedTx.ToProcessedCardanoTx(false))

		if !bridgeClaims.CanAddMore(maxClaimsToGroup) {
			break
		}
	}

	if invalidTxsCounter > 0 {
		telemetry.UpdateOracleClaimsInvalidCounter(c.state.blockInfo.ChainID, invalidTxsCounter) // update telemetry
	}

	c.state.processed = append(c.state.processed, processedTxs...)
	c.state.processedExpected = append(c.state.processedExpected, processedExpectedTxs...)

	return relevantUnprocessedTxs, processedTxs, processedExpectedTxs
}

func (c *CardanoStateProcessor) checkExpectedTxs(
	bridgeClaims *core.BridgeClaims,
	maxClaimsToGroup int,
) (
	[]*core.BridgeExpectedCardanoTx,
	[]*core.BridgeExpectedCardanoTx,
	[]*core.BridgeExpectedCardanoTx,
) {
	var relevantExpiredTxs []*core.BridgeExpectedCardanoTx

	ccoDB := c.indexerDbs[c.state.blockInfo.ChainID]
	if ccoDB == nil {
		c.logger.Error("Failed to get cardano chain observer db", "chainId", c.state.blockInfo.ChainID)
	} else {
		// ensure always same order of iterating through expectedTxsMap
		keys := make([]string, 0, len(c.state.expectedTxsMap))
		for k := range c.state.expectedTxsMap {
			keys = append(keys, k)
		}

		sort.Strings(keys)

		for _, key := range keys {
			expectedTx := c.state.expectedTxsMap[key]

			fromSlot := expectedTx.TTL + TTLInsuranceOffset

			blocks, err := ccoDB.GetConfirmedBlocksFrom(fromSlot, 1)
			if err != nil {
				c.logger.Error("Failed to get confirmed blocks", "fromSlot", fromSlot, "chainId", expectedTx.ChainID, "err", err)

				break
			}

			if len(blocks) == 1 && c.state.blockInfo.EqualWithExpected(expectedTx, blocks[0]) {
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
		processedTx, _ := c.db.GetProcessedTx(expiredTx.ChainID, expiredTx.Hash)
		if processedTx != nil && !processedTx.IsInvalid {
			// already sent the success claim
			processedRelevantExpiredTxs = append(processedRelevantExpiredTxs, expiredTx)

			continue
		}

		c.logger.Debug("Checking if expired tx is relevant", "expiredTx", expiredTx)

		txProcessor, err := c.txProcessors.getFailed(expiredTx.Metadata)
		if err != nil {
			c.logger.Error("Failed to get tx processor for expired tx", "tx", expiredTx, "err", err)

			onInvalidTx(expiredTx)

			continue
		}

		err = txProcessor.ValidateAndAddClaim(bridgeClaims, expiredTx, c.appConfig)
		if err != nil {
			c.logger.Error("Failed to ValidateAndAddClaim", "expiredTx", expiredTx, "err", err)

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
			c.state.blockInfo.ChainID, len(invalidRelevantExpiredTxs)) // update telemetry
	}

	c.state.processedExpected = append(c.state.processedExpected, processedRelevantExpiredTxs...)
	c.state.invalidRelevantExpired = append(c.state.invalidRelevantExpired, invalidRelevantExpiredTxs...)

	return relevantExpiredTxs, processedRelevantExpiredTxs, invalidRelevantExpiredTxs
}

func (c *CardanoStateProcessor) notifyBridgingRequestStateUpdater(
	bridgeClaims *core.BridgeClaims,
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater,
) error {
	if len(bridgeClaims.BridgingRequestClaims) > 0 {
		for _, brClaim := range bridgeClaims.BridgingRequestClaims {
			err := bridgingRequestStateUpdater.SubmittedToBridge(common.BridgingRequestStateKey{
				SourceChainID: common.ToStrChainID(brClaim.SourceChainId),
				SourceTxHash:  brClaim.ObservedTransactionHash,
			}, common.ToStrChainID(brClaim.DestinationChainId))

			if err != nil {
				c.logger.Error(
					"error while updating a bridging request state to SubmittedToBridge",
					"sourceChainId", common.ToStrChainID(brClaim.SourceChainId),
					"sourceTxHash", brClaim.ObservedTransactionHash, "err", err)
			}
		}
	}

	if len(bridgeClaims.BatchExecutedClaims) > 0 {
		for _, beClaim := range bridgeClaims.BatchExecutedClaims {
			err := bridgingRequestStateUpdater.ExecutedOnDestination(
				common.ToStrChainID(beClaim.ChainId),
				beClaim.BatchNonceId,
				beClaim.ObservedTransactionHash)

			if err != nil {
				c.logger.Error(
					"error while updating bridging request states to ExecutedOnDestination",
					"destinationChainId", common.ToStrChainID(beClaim.ChainId), "batchId", beClaim.BatchNonceId,
					"destinationTxHash", beClaim.ObservedTransactionHash, "err", err)
			}
		}
	}

	if len(bridgeClaims.BatchExecutionFailedClaims) > 0 {
		for _, befClaim := range bridgeClaims.BatchExecutionFailedClaims {
			err := bridgingRequestStateUpdater.FailedToExecuteOnDestination(
				common.ToStrChainID(befClaim.ChainId),
				befClaim.BatchNonceId)

			if err != nil {
				c.logger.Error(
					"error while updating bridging request states to FailedToExecuteOnDestination",
					"destinationChainId", common.ToStrChainID(befClaim.ChainId),
					"batchId", befClaim.BatchNonceId, "err", err)
			}
		}
	}

	for _, tx := range c.state.processed {
		if tx.IsInvalid {
			for _, unprocessedTx := range c.state.unprocessed {
				if bytes.Equal(unprocessedTx.ToCardanoTxKey(), tx.ToCardanoTxKey()) {
					txProcessor, err := c.txProcessors.getSuccess(unprocessedTx.Metadata)
					if err != nil {
						c.logger.Error("Failed to get tx processor for processed tx", "tx", tx, "err", err)
					} else if txProcessor.GetType() == common.BridgingTxTypeBridgingRequest {
						err := bridgingRequestStateUpdater.Invalid(common.BridgingRequestStateKey{
							SourceChainID: tx.OriginChainID,
							SourceTxHash:  common.Hash(tx.Hash),
						})

						if err != nil {
							c.logger.Error(
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
