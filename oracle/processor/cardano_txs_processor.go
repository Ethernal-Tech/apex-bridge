package processor

import (
	"context"
	"fmt"
	"math"
	"os"
	"sort"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/apex-bridge/telemetry"

	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/hashicorp/go-hclog"
)

const (
	TickTimeMs                  = 2000
	TTLInsuranceOffset          = 2
	MinBridgingClaimsToGroup    = 1
	GasLimitMultiplierDefault   = float32(1)
	GasLimitMultiplierIncrement = float32(0.5)
	GasLimitMultiplierMax       = float32(3)
)

type CardanoTxsProcessorImpl struct {
	ctx                         context.Context
	appConfig                   *core.AppConfig
	db                          core.CardanoTxsProcessorDB
	txProcessors                map[string]core.CardanoTxProcessor
	failedTxProcessors          map[string]core.CardanoTxFailedProcessor
	bridgeSubmitter             core.BridgeSubmitter
	indexerDbs                  map[string]indexer.Database
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater
	logger                      hclog.Logger
	tickTime                    time.Duration

	maxBridgingClaimsToGroup map[string]int
	gasLimitMultiplier       map[string]float32
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
	txProcessorsMap := make(map[string]core.CardanoTxProcessor, len(txProcessors))
	for _, txProcessor := range txProcessors {
		txProcessorsMap[string(txProcessor.GetType())] = txProcessor
	}

	failedTxProcessorsMap := make(map[string]core.CardanoTxFailedProcessor, len(failedTxProcessors))
	for _, txProcessor := range failedTxProcessors {
		failedTxProcessorsMap[string(txProcessor.GetType())] = txProcessor
	}

	maxBridgingClaimsToGroup := make(map[string]int, len(appConfig.CardanoChains))
	for _, chain := range appConfig.CardanoChains {
		maxBridgingClaimsToGroup[chain.ChainID] = appConfig.BridgingSettings.MaxBridgingClaimsToGroup
	}

	gasLimitMultiplier := make(map[string]float32, len(appConfig.CardanoChains))
	for _, chain := range appConfig.CardanoChains {
		gasLimitMultiplier[chain.ChainID] = 1
	}

	return &CardanoTxsProcessorImpl{
		ctx:                         ctx,
		appConfig:                   appConfig,
		db:                          db,
		txProcessors:                txProcessorsMap,
		failedTxProcessors:          failedTxProcessorsMap,
		bridgeSubmitter:             bridgeSubmitter,
		indexerDbs:                  indexerDbs,
		bridgingRequestStateUpdater: bridgingRequestStateUpdater,
		logger:                      logger,
		tickTime:                    TickTimeMs,

		maxBridgingClaimsToGroup: maxBridgingClaimsToGroup,
		gasLimitMultiplier:       gasLimitMultiplier,
	}
}

func (bp *CardanoTxsProcessorImpl) NewUnprocessedTxs(originChainID string, txs []*indexer.Tx) error {
	bp.logger.Info("NewUnprocessedTxs", "txs", txs)

	var (
		bridgingRequests  []*indexer.Tx
		relevantTxs       []*core.CardanoTx
		processedTxs      []*core.ProcessedCardanoTx
		invalidTxsCounter int
	)

	onIrrelevantTx := func(cardanoTx *core.CardanoTx) {
		processedTxs = append(processedTxs, cardanoTx.ToProcessedCardanoTx(false))
		invalidTxsCounter++
	}

	for _, tx := range txs {
		cardanoTx := &core.CardanoTx{
			OriginChainID: originChainID,
			Tx:            *tx,
			Priority:      1,
		}

		bp.logger.Debug("Checking if tx is relevant", "tx", tx)

		txProcessor, err := bp.getTxProcessor(tx.Metadata)
		if err != nil {
			bp.logger.Error("Failed to get tx processor for new tx", "tx", tx, "err", err)

			onIrrelevantTx(cardanoTx)

			continue
		}

		if txProcessor.GetType() == common.BridgingTxTypeBatchExecution {
			cardanoTx.Priority = 0
		}

		relevantTxs = append(relevantTxs, cardanoTx)

		if txProcessor.GetType() == common.BridgingTxTypeBridgingRequest {
			bridgingRequests = append(bridgingRequests, tx)
		}
	}

	if len(processedTxs) > 0 {
		bp.logger.Debug("Adding already processed txs to db", "txs", processedTxs)

		err := bp.db.AddProcessedTxs(processedTxs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to add already processed txs. error: %v\n", err)
			bp.logger.Error("Failed to add already processed txs", "err", err)

			return err
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

	if invalidTxsCounter > 0 {
		telemetry.UpdateOracleClaimsInvalidMetaDataCounter(originChainID, invalidTxsCounter) // update telemetry
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

func (bp *CardanoTxsProcessorImpl) getTxProcessor(metadataBytes []byte) (
	core.CardanoTxProcessor, error,
) {
	metadata, err := common.UnmarshalMetadata[common.BaseMetadata](common.MetadataEncodingTypeCbor, metadataBytes)
	if err != nil {
		return nil, err
	}

	txProcessor, relevant := bp.txProcessors[string(metadata.BridgingTxType)]
	if !relevant {
		return nil, fmt.Errorf("irrelevant tx. Tx type: %s", metadata.BridgingTxType)
	}

	return txProcessor, nil
}

func (bp *CardanoTxsProcessorImpl) getFailedTxProcessor(metadataBytes []byte) (
	core.CardanoTxFailedProcessor, error,
) {
	metadata, err := common.UnmarshalMetadata[common.BaseMetadata](common.MetadataEncodingTypeCbor, metadataBytes)
	if err != nil {
		return nil, err
	}

	txProcessor, relevant := bp.failedTxProcessors[string(metadata.BridgingTxType)]
	if !relevant {
		return nil, fmt.Errorf("irrelevant tx. Tx type: %s", metadata.BridgingTxType)
	}

	return txProcessor, nil
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
		allInvalidRelevantExpiredTxs []*core.BridgeExpectedCardanoTx
		allProcessedExpectedTxs      []*core.BridgeExpectedCardanoTx
		allProcessedTxs              []*core.ProcessedCardanoTx
		allUnprocessedTxs            []*core.CardanoTx
	)

	bridgeClaims := &core.BridgeClaims{}

	maxClaimsToGroup := bp.maxBridgingClaimsToGroup[startChainID]

	for priority := uint(0); priority <= core.LastProcessingPriority; priority++ {
		invalidRelevantExpiredTxs, processedExpectedTxs,
			processedTxs, unprocessedTxs := bp.processAllForChain(bridgeClaims, startChainID, maxClaimsToGroup, priority)

		allInvalidRelevantExpiredTxs = append(allInvalidRelevantExpiredTxs, invalidRelevantExpiredTxs...)
		allProcessedExpectedTxs = append(allProcessedExpectedTxs, processedExpectedTxs...)
		allProcessedTxs = append(allProcessedTxs, processedTxs...)
		allUnprocessedTxs = append(allUnprocessedTxs, unprocessedTxs...)

		if !bridgeClaims.CanAddMore(maxClaimsToGroup) {
			break
		}
	}

	// ensure always same order of iterating through bp.appConfig.CardanoChains
	keys := make([]string, 0, len(bp.appConfig.CardanoChains))
	for k := range bp.appConfig.CardanoChains {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, key := range keys {

		chainID := bp.appConfig.CardanoChains[key].ChainID
		if chainID != startChainID {
			for priority := uint(0); priority <= core.LastProcessingPriority; priority++ {
				if !bridgeClaims.CanAddMore(maxClaimsToGroup) {
					break
				}

				invalidRelevantExpiredTxs, processedExpectedTxs,
					processedTxs, unprocessedTxs := bp.processAllForChain(bridgeClaims, chainID, maxClaimsToGroup, priority)

				allInvalidRelevantExpiredTxs = append(allInvalidRelevantExpiredTxs, invalidRelevantExpiredTxs...)
				allProcessedExpectedTxs = append(allProcessedExpectedTxs, processedExpectedTxs...)
				allProcessedTxs = append(allProcessedTxs, processedTxs...)
				allUnprocessedTxs = append(allUnprocessedTxs, unprocessedTxs...)
			}
		}
	}

	// if expected/expired tx is invalid, we should mark them regardless of if submit failed or not
	if len(allInvalidRelevantExpiredTxs) > 0 {
		bp.logger.Info("Marking expected txs as invalid", "txs", allInvalidRelevantExpiredTxs)

		err := bp.db.MarkExpectedTxsAsInvalid(allInvalidRelevantExpiredTxs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to mark expected txs as invalid. error: %v\n", err)
			bp.logger.Error("Failed to mark expected txs as invalid", "err", err)
		}
	}

	if bridgeClaims.Count() > 0 {
		bp.logger.Info("Submitting bridge claims", "claims", bridgeClaims)

		err := bp.bridgeSubmitter.SubmitClaims(
			bridgeClaims, &eth.SubmitOpts{GasLimitMultiplier: bp.gasLimitMultiplier[startChainID]})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to submit claims. error: %v\n", err)
			bp.logger.Error("Failed to submit claims", "err", err)

			bp.maxBridgingClaimsToGroup[startChainID] = bridgeClaims.Count() - 1
			if bp.maxBridgingClaimsToGroup[startChainID] < MinBridgingClaimsToGroup {
				bp.maxBridgingClaimsToGroup[startChainID] = MinBridgingClaimsToGroup
			}

			bp.logger.Warn("set maxBridgingClaimsToGroup",
				"startChainID", startChainID, "newValue", bp.maxBridgingClaimsToGroup[startChainID])

			if bridgeClaims.Count() <= MinBridgingClaimsToGroup &&
				bp.gasLimitMultiplier[startChainID]+GasLimitMultiplierIncrement <= GasLimitMultiplierMax {
				bp.gasLimitMultiplier[startChainID] += GasLimitMultiplierIncrement

				bp.logger.Warn("Increased gasLimitMultiplier",
					"startChainID", startChainID, "newValue", bp.gasLimitMultiplier[startChainID])
			}

			return
		}

		if bp.maxBridgingClaimsToGroup[startChainID] != bp.appConfig.BridgingSettings.MaxBridgingClaimsToGroup {
			bp.maxBridgingClaimsToGroup[startChainID] = bp.appConfig.BridgingSettings.MaxBridgingClaimsToGroup

			bp.logger.Info("Reset maxBridgingClaimsToGroup",
				"startChainID", startChainID, "newValue", bp.maxBridgingClaimsToGroup[startChainID])
		}

		if bp.gasLimitMultiplier[startChainID] != GasLimitMultiplierDefault {
			bp.gasLimitMultiplier[startChainID] = GasLimitMultiplierDefault

			bp.logger.Info("Reset gasLimitMultiplier",
				"startChainID", startChainID, "newValue", bp.gasLimitMultiplier[startChainID])
		}

		telemetry.UpdateOracleClaimsSubmitCounter(bridgeClaims.Count()) // update telemetry
	}

	err := bp.notifyBridgingRequestStateUpdater(bridgeClaims, allUnprocessedTxs, allProcessedTxs)
	if err != nil {
		bp.logger.Error("Error while updating bridging request states", "err", err)
	}

	// we should only change this in db if submit succeeded
	if len(allProcessedExpectedTxs) > 0 {
		bp.logger.Info("Marking expected txs as processed", "txs", allProcessedExpectedTxs)

		err := bp.db.MarkExpectedTxsAsProcessed(allProcessedExpectedTxs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to mark expected txs as processed. error: %v\n", err)
			bp.logger.Error("Failed to mark expected txs as processed", "err", err)
		}
	}

	// we should only change this in db if submit succeeded
	if len(allProcessedTxs) > 0 {
		bp.logger.Info("Marking txs as processed", "txs", allProcessedTxs)

		err := bp.db.MarkUnprocessedTxsAsProcessed(allProcessedTxs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to mark txs as processed. error: %v\n", err)
			bp.logger.Error("Failed to mark txs as processed", "err", err)
		}
	}
}

func (bp *CardanoTxsProcessorImpl) processAllForChain(
	bridgeClaims *core.BridgeClaims,
	chainID string,
	maxClaimsToGroup int,
	priority uint,
) (
	allInvalidRelevantExpiredTxs []*core.BridgeExpectedCardanoTx,
	allProcessedExpectedTxs []*core.BridgeExpectedCardanoTx,
	allProcessedTxs []*core.ProcessedCardanoTx,
	unprocessedTxs []*core.CardanoTx,
) {
	expectedTxs, err := bp.db.GetExpectedTxs(chainID, priority, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get expected txs. error: %v\n", err)
		bp.logger.Error("Failed to get expected txs", "err", err)

		return
	}

	unprocessedTxs, err = bp.db.GetUnprocessedTxs(chainID, priority, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get unprocessed txs. error: %v\n", err)
		bp.logger.Error("Failed to get unprocessed txs", "err", err)

		return
	}

	ccoDB := bp.indexerDbs[chainID]
	if ccoDB == nil {
		fmt.Fprintf(os.Stderr, "Failed to get cardano chain observer db for: %v\n", chainID)
		bp.logger.Error("Failed to get cardano chain observer db", "chainId", chainID)
	}

	// needed for the guarantee that both unprocessedTxs and expectedTxs are processed in order of slot
	// and prevent the situation when there are always enough unprocessedTxs to fill out claims,
	// that all claims are filled only from unprocessedTxs and never from expectedTxs
	blockInfo := bp.constructBridgeClaimsBlockInfo(
		chainID, ccoDB, unprocessedTxs, expectedTxs, nil)
	if blockInfo == nil {
		return
	}

	expectedTxsMap := make(map[string]*core.BridgeExpectedCardanoTx, len(expectedTxs))
	for _, expectedTx := range expectedTxs {
		expectedTxsMap[expectedTx.ToCardanoTxKey()] = expectedTx
	}

	for {
		bp.logger.Debug("Processing", "for chainID", chainID, "blockInfo", blockInfo)

		_, processedTxs, processedExpectedTxs := bp.checkUnprocessedTxs(
			blockInfo, bridgeClaims, unprocessedTxs, expectedTxsMap, maxClaimsToGroup)

		_, processedRelevantExpiredTxs, invalidRelevantExpiredTxs := bp.checkExpectedTxs(
			blockInfo, bridgeClaims, ccoDB, expectedTxsMap, maxClaimsToGroup)

		processedExpectedTxs = append(processedExpectedTxs, processedRelevantExpiredTxs...)

		bp.logger.Debug("Checked all", "for chainID", chainID,
			"processedTxs", processedTxs, "processedExpectedTxs", processedExpectedTxs,
			"invalidRelevantExpiredTxs", invalidRelevantExpiredTxs)

		allProcessedTxs = append(allProcessedTxs, processedTxs...)
		allProcessedExpectedTxs = append(allProcessedExpectedTxs, processedExpectedTxs...)
		allInvalidRelevantExpiredTxs = append(allInvalidRelevantExpiredTxs, invalidRelevantExpiredTxs...)

		if !bridgeClaims.CanAddMore(maxClaimsToGroup) {
			break
		}

		blockInfo = bp.constructBridgeClaimsBlockInfo(
			chainID, ccoDB, unprocessedTxs, expectedTxs, blockInfo)
		if blockInfo == nil {
			break
		}
	}

	return allInvalidRelevantExpiredTxs, allProcessedExpectedTxs, allProcessedTxs, unprocessedTxs
}

func (bp *CardanoTxsProcessorImpl) constructBridgeClaimsBlockInfo(
	chainID string,
	ccoDB indexer.Database,
	unprocessedTxs []*core.CardanoTx,
	expectedTxs []*core.BridgeExpectedCardanoTx,
	prevBlockInfo *core.BridgeClaimsBlockInfo,
) *core.BridgeClaimsBlockInfo {
	found := false
	minSlot := uint64(math.MaxUint64)

	var blockHash string

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
		// expected are ordered by ttl, so first in collection is min
		for _, tx := range expectedTxs {
			fromSlot := tx.TTL + TTLInsuranceOffset

			if ccoDB != nil {
				blocks, err := ccoDB.GetConfirmedBlocksFrom(fromSlot, 1)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to get confirmed blocks from slot: %v, for %v. error: %v\n", fromSlot, chainID, err)
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
		processedTxs         []*core.ProcessedCardanoTx
		processedExpectedTxs []*core.BridgeExpectedCardanoTx
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

		txProcessor, err := bp.getTxProcessor(unprocessedTx.Metadata)
		if err != nil {
			bp.logger.Error("Failed to get tx processor for unprocessed tx", "tx", unprocessedTx, "err", err)

			onInvalidTx(unprocessedTx)

			continue
		}

		if txProcessor.GetType() == common.BridgingTxTypeBatchExecution &&
			!bridgeClaims.CanAddBatchExecutedClaim() {
			continue
		}

		err = txProcessor.ValidateAndAddClaim(bridgeClaims, unprocessedTx, bp.appConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to ValidateAndAddClaim. error: %v\n", err)
			bp.logger.Error("Failed to ValidateAndAddClaim", "tx", unprocessedTx, "err", err)

			onInvalidTx(unprocessedTx)

			continue
		}

		expectedTx := expectedTxsMap[unprocessedTx.ToCardanoTxKey()]
		if expectedTx != nil {
			processedExpectedTxs = append(processedExpectedTxs, expectedTx)
			delete(expectedTxsMap, expectedTx.ToCardanoTxKey())
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
	ccoDB indexer.Database,
	expectedTxsMap map[string]*core.BridgeExpectedCardanoTx,
	maxClaimsToGroup int,
) (
	[]*core.BridgeExpectedCardanoTx,
	[]*core.BridgeExpectedCardanoTx,
	[]*core.BridgeExpectedCardanoTx,
) {
	var relevantExpiredTxs []*core.BridgeExpectedCardanoTx

	// ensure always same order of iterating through expectedTxsMap
	keys := make([]string, 0, len(expectedTxsMap))
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

		txProcessor, err := bp.getFailedTxProcessor(expiredTx.Metadata)
		if err != nil {
			bp.logger.Error("Failed to get tx processor for expired tx", "tx", expiredTx, "err", err)

			onInvalidTx(expiredTx)

			continue
		}

		err = txProcessor.ValidateAndAddClaim(bridgeClaims, expiredTx, bp.appConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to ValidateAndAddClaim. error: %v\n", err)
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

	for _, tx := range processedTxs {
		if tx.IsInvalid {
			for _, unprocessedTx := range unprocessedTxs {
				if unprocessedTx.ToCardanoTxKey() == tx.ToCardanoTxKey() {
					txProcessor, err := bp.getTxProcessor(unprocessedTx.Metadata)
					if err != nil {
						bp.logger.Error("Failed to get tx processor for processed tx", "tx", tx, "err", err)
					} else if txProcessor.GetType() == common.BridgingTxTypeBridgingRequest {
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
