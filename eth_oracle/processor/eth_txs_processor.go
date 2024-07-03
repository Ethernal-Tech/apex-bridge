package processor

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/eth_oracle/core"
	oracleCore "github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/apex-bridge/telemetry"
	eventTrackerStore "github.com/Ethernal-Tech/blockchain-event-tracker/store"
	"github.com/Ethernal-Tech/ethgo"

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

type EthTxsProcessorImpl struct {
	ctx                         context.Context
	appConfig                   *oracleCore.AppConfig
	db                          core.EthTxsProcessorDB
	txProcessors                map[string]core.EthTxProcessor
	failedTxProcessors          map[string]core.EthTxFailedProcessor
	bridgeSubmitter             oracleCore.BridgeSubmitter
	indexerDbs                  map[string]eventTrackerStore.EventTrackerStore
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater
	logger                      hclog.Logger
	tickTime                    time.Duration

	maxBridgingClaimsToGroup map[string]int
	gasLimitMultiplier       map[string]float32
}

var _ core.EthTxsProcessor = (*EthTxsProcessorImpl)(nil)

func NewEthTxsProcessor(
	ctx context.Context,
	appConfig *oracleCore.AppConfig,
	db core.EthTxsProcessorDB,
	txProcessors []core.EthTxProcessor,
	failedTxProcessors []core.EthTxFailedProcessor,
	bridgeSubmitter oracleCore.BridgeSubmitter,
	indexerDbs map[string]eventTrackerStore.EventTrackerStore,
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater,
	logger hclog.Logger,
) *EthTxsProcessorImpl {
	txProcessorsMap := make(map[string]core.EthTxProcessor, len(txProcessors))
	for _, txProcessor := range txProcessors {
		txProcessorsMap[string(txProcessor.GetType())] = txProcessor
	}

	failedTxProcessorsMap := make(map[string]core.EthTxFailedProcessor, len(failedTxProcessors))
	for _, txProcessor := range failedTxProcessors {
		failedTxProcessorsMap[string(txProcessor.GetType())] = txProcessor
	}

	maxBridgingClaimsToGroup := make(map[string]int, len(appConfig.EthChains))
	for _, chain := range appConfig.EthChains {
		maxBridgingClaimsToGroup[chain.ChainID] = appConfig.BridgingSettings.MaxBridgingClaimsToGroup
	}

	gasLimitMultiplier := make(map[string]float32, len(appConfig.EthChains))
	for _, chain := range appConfig.EthChains {
		gasLimitMultiplier[chain.ChainID] = 1
	}

	return &EthTxsProcessorImpl{
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

func (bp *EthTxsProcessorImpl) NewUnprocessedLog(originChainID string, log *ethgo.Log) error {
	bp.logger.Info("NewUnprocessedLog", "log", log)

	//nolint:prealloc
	var (
		bridgingRequests  []*common.NewBridgingRequestStateModel
		relevantTxs       []*core.EthTx
		processedTxs      []*core.ProcessedEthTx
		invalidTxsCounter int
	)

	onIrrelevantTx := func(ethTx *core.EthTx) {
		processedTxs = append(processedTxs, ethTx.ToProcessedEthTx(false))
		invalidTxsCounter++
	}

	// a TODO: finish this
	// treat every event as a separate "tx"
	txs := make([]*core.EthTx, 0)

	// Unpack log here, and for each relevant event create ethTx and add to txs
	// will probably use binding UnpackLog or binding.<name>ContractFilterer.Parse<event>(log).
	// also fetch the transaction.value() using ethtxhelper
	/*
		eventABIs := make([]string, 0)
		abis := make([]abi.ABI, 0, len(eventABIs))
		for idx, eventABI := range eventABIs {
			abi, err := abi.JSON(strings.NewReader(eventABI))
			if err != nil {
				return fmt.Errorf("failed to create event ABI: %w", err)
			}

			abis[idx] = abi
		}
	*/

	for _, tx := range txs {
		/*
			ethTx := &core.EthTx{
				OriginChainID: originChainID,
				// IndexerEthTx:  *tx,
				Priority: 1,
			}
		*/
		bp.logger.Debug("Checking if tx is relevant", "tx", tx)

		txProcessor, err := bp.getTxProcessor(tx.MetadataJSON)
		if err != nil {
			bp.logger.Error("Failed to get tx processor for new tx", "tx", tx, "err", err)

			onIrrelevantTx(tx)

			continue
		}

		if txProcessor.GetType() == common.BridgingTxTypeBatchExecution {
			tx.Priority = 0
		}

		relevantTxs = append(relevantTxs, tx)

		if txProcessor.GetType() == common.BridgingTxTypeBridgingRequest {
			bridgingRequests = append(
				bridgingRequests,
				&common.NewBridgingRequestStateModel{
					SourceTxHash: common.Hash(tx.Hash),
				},
			)
		}
	}

	if len(processedTxs) > 0 {
		bp.logger.Debug("Adding already processed txs to db", "txs", processedTxs)

		err := bp.db.AddProcessedTxs(processedTxs)
		if err != nil {
			bp.logger.Error("Failed to add already processed txs", "err", err)

			return err
		}
	}

	if len(relevantTxs) > 0 {
		bp.logger.Debug("Adding relevant txs to db", "txs", relevantTxs)

		err := bp.db.AddUnprocessedTxs(relevantTxs)
		if err != nil {
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

func (bp *EthTxsProcessorImpl) Start() {
	bp.logger.Debug("Starting EthTxsProcessor")

	for {
		if !bp.checkShouldGenerateClaims() {
			return
		}
	}
}

func (bp *EthTxsProcessorImpl) getTxProcessor(metadataJSON []byte) (
	core.EthTxProcessor, error,
) {
	metadata, err := common.UnmarshalMetadata[common.BaseMetadata](common.MetadataEncodingTypeJSON, metadataJSON)
	if err != nil {
		return nil, err
	}

	txProcessor, relevant := bp.txProcessors[string(metadata.BridgingTxType)]
	if !relevant {
		return nil, fmt.Errorf("irrelevant tx. Tx type: %s", metadata.BridgingTxType)
	}

	return txProcessor, nil
}

func (bp *EthTxsProcessorImpl) getFailedTxProcessor(metadataJSON []byte) (
	core.EthTxFailedProcessor, error,
) {
	metadata, err := common.UnmarshalMetadata[common.BaseMetadata](common.MetadataEncodingTypeJSON, metadataJSON)
	if err != nil {
		return nil, err
	}

	txProcessor, relevant := bp.failedTxProcessors[string(metadata.BridgingTxType)]
	if !relevant {
		return nil, fmt.Errorf("irrelevant tx. Tx type: %s", metadata.BridgingTxType)
	}

	return txProcessor, nil
}

func (bp *EthTxsProcessorImpl) checkShouldGenerateClaims() bool {
	// ensure always same order of iterating through bp.appConfig.EthChains
	keys := make([]string, 0, len(bp.appConfig.EthChains))
	for k := range bp.appConfig.EthChains {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, key := range keys {
		select {
		case <-bp.ctx.Done():
			return false
		case <-time.After(bp.tickTime * time.Millisecond):
		}

		bp.processAllStartingWithChain(bp.appConfig.EthChains[key].ChainID)
	}

	return true
}

// first process for a specific chainID, to give every chainID the chance
// and then, if max claims not reached, rest of the chains can be processed too
func (bp *EthTxsProcessorImpl) processAllStartingWithChain(
	startChainID string,
) {
	var (
		allInvalidRelevantExpiredTxs []*core.BridgeExpectedEthTx
		allProcessedExpectedTxs      []*core.BridgeExpectedEthTx
		allProcessedTxs              []*core.ProcessedEthTx
		allUnprocessedTxs            []*core.EthTx
	)

	bridgeClaims := &oracleCore.BridgeClaims{}

	maxClaimsToGroup := bp.maxBridgingClaimsToGroup[startChainID]

	for priority := uint8(0); priority <= oracleCore.LastProcessingPriority; priority++ {
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

	// ensure always same order of iterating through bp.appConfig.EthChains
	keys := make([]string, 0, len(bp.appConfig.EthChains))
	for k := range bp.appConfig.EthChains {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, key := range keys {
		chainID := bp.appConfig.EthChains[key].ChainID
		if chainID != startChainID {
			for priority := uint8(0); priority <= oracleCore.LastProcessingPriority; priority++ {
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
			bp.logger.Error("Failed to mark expected txs as invalid", "err", err)
		}
	}

	if bridgeClaims.Count() > 0 {
		bp.logger.Info("Submitting bridge claims", "claims", bridgeClaims)

		err := bp.bridgeSubmitter.SubmitClaims(
			bridgeClaims, &eth.SubmitOpts{GasLimitMultiplier: bp.gasLimitMultiplier[startChainID]})
		if err != nil {
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

		bp.maxBridgingClaimsToGroup[startChainID] = bp.appConfig.BridgingSettings.MaxBridgingClaimsToGroup
		bp.gasLimitMultiplier[startChainID] = GasLimitMultiplierDefault

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
			bp.logger.Error("Failed to mark expected txs as processed", "err", err)
		}
	}

	// we should only change this in db if submit succeeded
	if len(allProcessedTxs) > 0 {
		bp.logger.Info("Marking txs as processed", "txs", allProcessedTxs)

		err := bp.db.MarkUnprocessedTxsAsProcessed(allProcessedTxs)
		if err != nil {
			bp.logger.Error("Failed to mark txs as processed", "err", err)
		}
	}
}

func (bp *EthTxsProcessorImpl) processAllForChain(
	bridgeClaims *oracleCore.BridgeClaims,
	chainID string,
	maxClaimsToGroup int,
	priority uint8,
) (
	allInvalidRelevantExpiredTxs []*core.BridgeExpectedEthTx,
	allProcessedExpectedTxs []*core.BridgeExpectedEthTx,
	allProcessedTxs []*core.ProcessedEthTx,
	unprocessedTxs []*core.EthTx,
) {
	expectedTxs, err := bp.db.GetExpectedTxs(chainID, priority, 0)
	if err != nil {
		bp.logger.Error("Failed to get expected txs", "err", err)

		return
	}

	unprocessedTxs, err = bp.db.GetUnprocessedTxs(chainID, priority, 0)
	if err != nil {
		bp.logger.Error("Failed to get unprocessed txs", "err", err)

		return
	}

	ecoDB := bp.indexerDbs[chainID]
	if ecoDB == nil {
		bp.logger.Error("Failed to get eth chain observer db", "chainId", chainID)
	}

	// needed for the guarantee that both unprocessedTxs and expectedTxs are processed in order of block number
	// and prevent the situation when there are always enough unprocessedTxs to fill out claims,
	// that all claims are filled only from unprocessedTxs and never from expectedTxs
	blockInfo := bp.constructBridgeClaimsBlockInfo(
		chainID, ecoDB, unprocessedTxs, expectedTxs, nil)
	if blockInfo == nil {
		return
	}

	expectedTxsMap := make(map[string]*core.BridgeExpectedEthTx, len(expectedTxs))
	for _, expectedTx := range expectedTxs {
		expectedTxsMap[string(expectedTx.ToEthTxKey())] = expectedTx
	}

	for {
		bp.logger.Debug("Processing", "for chainID", chainID, "blockInfo", blockInfo)

		_, processedTxs, processedExpectedTxs := bp.checkUnprocessedTxs(
			blockInfo, bridgeClaims, unprocessedTxs, expectedTxsMap, maxClaimsToGroup)

		_, processedRelevantExpiredTxs, invalidRelevantExpiredTxs := bp.checkExpectedTxs(
			blockInfo, bridgeClaims, ecoDB, expectedTxsMap, maxClaimsToGroup)

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
			chainID, ecoDB, unprocessedTxs, expectedTxs, blockInfo)
		if blockInfo == nil {
			break
		}
	}

	return allInvalidRelevantExpiredTxs, allProcessedExpectedTxs, allProcessedTxs, unprocessedTxs
}

func (bp *EthTxsProcessorImpl) constructBridgeClaimsBlockInfo(
	chainID string,
	ecoDB eventTrackerStore.EventTrackerStore,
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
		// expected are ordered by ttl, so first in collection is min
		for _, tx := range expectedTxs {
			fromBlockNumber := tx.TTL + TTLInsuranceOffset

			if ecoDB != nil {
				lastProcessedBlock, err := ecoDB.GetLastProcessedBlock()
				if err != nil {
					bp.logger.Error("Failed to get last processed block",
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

func (bp *EthTxsProcessorImpl) checkUnprocessedTxs(
	blockInfo *core.BridgeClaimsBlockInfo,
	bridgeClaims *oracleCore.BridgeClaims,
	unprocessedTxs []*core.EthTx,
	expectedTxsMap map[string]*core.BridgeExpectedEthTx,
	maxClaimsToGroup int,
) (
	[]*core.EthTx,
	[]*core.ProcessedEthTx,
	[]*core.BridgeExpectedEthTx,
) {
	var relevantUnprocessedTxs []*core.EthTx

	for _, unprocessedTx := range unprocessedTxs {
		if blockInfo.EqualWithUnprocessed(unprocessedTx) {
			relevantUnprocessedTxs = append(relevantUnprocessedTxs, unprocessedTx)
		}
	}

	//nolint:prealloc
	var (
		processedTxs         []*core.ProcessedEthTx
		processedExpectedTxs []*core.BridgeExpectedEthTx
		invalidTxsCounter    int
	)

	if len(relevantUnprocessedTxs) == 0 {
		return relevantUnprocessedTxs, processedTxs, processedExpectedTxs
	}

	onInvalidTx := func(tx *core.EthTx) {
		processedTxs = append(processedTxs, tx.ToProcessedEthTx(true))
		invalidTxsCounter++
	}

	// check unprocessed txs from indexers
	for _, unprocessedTx := range relevantUnprocessedTxs {
		bp.logger.Debug("Checking if tx is relevant", "tx", unprocessedTx)

		txProcessor, err := bp.getTxProcessor(unprocessedTx.MetadataJSON)
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
			bp.logger.Error("Failed to ValidateAndAddClaim", "tx", unprocessedTx, "err", err)

			onInvalidTx(unprocessedTx)

			continue
		}

		key := string(unprocessedTx.ToEthTxKey())

		if expectedTx, exists := expectedTxsMap[key]; exists {
			processedExpectedTxs = append(processedExpectedTxs, expectedTx)

			delete(expectedTxsMap, key)
		}

		processedTxs = append(processedTxs, unprocessedTx.ToProcessedEthTx(false))

		if !bridgeClaims.CanAddMore(maxClaimsToGroup) {
			break
		}
	}

	if invalidTxsCounter > 0 {
		telemetry.UpdateOracleClaimsInvalidCounter(blockInfo.ChainID, invalidTxsCounter) // update telemetry
	}

	return relevantUnprocessedTxs, processedTxs, processedExpectedTxs
}

func (bp *EthTxsProcessorImpl) checkExpectedTxs(
	blockInfo *core.BridgeClaimsBlockInfo,
	bridgeClaims *oracleCore.BridgeClaims,
	ecoDB eventTrackerStore.EventTrackerStore,
	expectedTxsMap map[string]*core.BridgeExpectedEthTx,
	maxClaimsToGroup int,
) (
	[]*core.BridgeExpectedEthTx,
	[]*core.BridgeExpectedEthTx,
	[]*core.BridgeExpectedEthTx,
) {
	var relevantExpiredTxs []*core.BridgeExpectedEthTx

	// ensure always same order of iterating through expectedTxsMap
	keys := make([]string, 0, len(expectedTxsMap))
	for k := range expectedTxsMap {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, key := range keys {
		expectedTx := expectedTxsMap[key]

		if ecoDB == nil {
			break
		}

		fromBlockNumber := expectedTx.TTL + TTLInsuranceOffset

		lastBlockProcessed, err := ecoDB.GetLastProcessedBlock()
		if err != nil {
			bp.logger.Error("Failed to get last processed block",
				"chainId", expectedTx.ChainID, "err", err)

			break
		}

		if lastBlockProcessed >= fromBlockNumber && blockInfo.EqualWithExpected(expectedTx, fromBlockNumber) {
			relevantExpiredTxs = append(relevantExpiredTxs, expectedTx)
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
		processedTx, _ := bp.db.GetProcessedTx(expiredTx.ChainID, expiredTx.Hash)
		if processedTx != nil && !processedTx.IsInvalid {
			// already sent the success claim
			processedRelevantExpiredTxs = append(processedRelevantExpiredTxs, expiredTx)

			continue
		}

		bp.logger.Debug("Checking if expired tx is relevant", "expiredTx", expiredTx)

		txProcessor, err := bp.getFailedTxProcessor(expiredTx.MetadataJSON)
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

func (bp *EthTxsProcessorImpl) notifyBridgingRequestStateUpdater(
	bridgeClaims *oracleCore.BridgeClaims,
	unprocessedTxs []*core.EthTx,
	processedTxs []*core.ProcessedEthTx,
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
					"sourceChainId", common.ToStrChainID(brClaim.SourceChainId), "sourceTxHash", brClaim.ObservedTransactionHash)
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
					"destinationTxHash", beClaim.ObservedTransactionHash)
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
					"destinationChainId", common.ToStrChainID(befClaim.ChainId), "batchId", befClaim.BatchNonceId)
			}
		}
	}

	for _, tx := range processedTxs {
		if tx.IsInvalid {
			for _, unprocessedTx := range unprocessedTxs {
				if bytes.Equal(unprocessedTx.ToEthTxKey(), tx.ToEthTxKey()) {
					txProcessor, err := bp.getTxProcessor(unprocessedTx.MetadataJSON)
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
