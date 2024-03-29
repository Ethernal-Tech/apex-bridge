package processor

import (
	"fmt"
	"math"
	"os"
	"sort"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/oracle/core"

	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/hashicorp/go-hclog"
)

const (
	TickTimeMs         = 2000
	TtlInsuranceOffset = 2
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
	tickTime                  time.Duration
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
		tickTime:                  TickTimeMs,
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

	for {
		bp.checkShouldGenerateClaims()
	}
}

func (bp *CardanoTxsProcessorImpl) Stop() error {
	bp.logger.Debug("Stopping CardanoTxsProcessor")
	close(bp.closeCh)
	return nil
}

func (bp *CardanoTxsProcessorImpl) checkShouldGenerateClaims() {
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
		case <-bp.closeCh:
			return
		case <-ticker.C:
		}

		bp.processAllForChain(bp.appConfig.CardanoChains[key].ChainId)
	}
}

func (bp *CardanoTxsProcessorImpl) constructBridgeClaims(
	chainId string,
	unprocessedTxs []*core.CardanoTx,
	expectedTxs []*core.BridgeExpectedCardanoTx,
) (
	*core.BridgeClaims,
	indexer.Database,
) {
	ccoDb := bp.getCardanoChainObserverDb(chainId)
	if ccoDb == nil {
		fmt.Fprintf(os.Stderr, "Failed to get cardano chain observer db for: %v\n", chainId)
		bp.logger.Error("Failed to get cardano chain observer db", "chainId", chainId)
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
		fromSlot := expectedTx.Ttl + TtlInsuranceOffset
		if ccoDb != nil {
			blocks, err := ccoDb.GetConfirmedBlocksFrom(fromSlot, 1)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to get confirmed blocks from slot: %v, for %v. error: %v\n", fromSlot, chainId, err)
				bp.logger.Error("Failed to get confirmed blocks", "fromSlot", fromSlot, "chainId", chainId, "err", err)
			} else {
				if len(blocks) > 0 && blocks[0].Slot < minSlot {
					minSlot = blocks[0].Slot
					blockHash = blocks[0].Hash
					found = true
				}
			}
		}
	}

	if found {
		return &core.BridgeClaims{
			BlockFullyObserved: false,
			BlockInfo: &core.BridgeClaimsBlockInfo{
				ChainId: chainId,
				Slot:    minSlot,
				Hash:    blockHash,
			},
		}, ccoDb
	}

	return nil, ccoDb
}

func (bp *CardanoTxsProcessorImpl) checkUnprocessedTxs(
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
		if bridgeClaims.BlockInfoEqualWithUnprocessed(unprocessedTx) {
			relevantUnprocessedTxs = append(relevantUnprocessedTxs, unprocessedTx)
		}
	}

	var processedTxs []*core.ProcessedCardanoTx
	var processedExpectedTxs []*core.BridgeExpectedCardanoTx

	// check unprocessed txs from indexers
	if len(relevantUnprocessedTxs) > 0 {
	unprocessedTxsLoop:
		for _, unprocessedTx := range relevantUnprocessedTxs {
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

					expectedTx := expectedTxsMap[unprocessedTx.ToCardanoTxKey()]
					if expectedTx != nil {
						processedExpectedTxs = append(processedExpectedTxs, expectedTx)
						delete(expectedTxsMap, expectedTx.ToCardanoTxKey())
					}

					processedTxs = append(processedTxs, unprocessedTx.ToProcessedCardanoTx(false))
					txProcessed = true

					if bridgeClaims.Count() >= bp.appConfig.Settings.MaxBridgingClaimsToGroup {
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
	bridgeClaims *core.BridgeClaims,
	ccoDb indexer.Database,
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
		if ccoDb == nil {
			break
		}

		fromSlot := expectedTx.Ttl + TtlInsuranceOffset
		blocks, err := ccoDb.GetConfirmedBlocksFrom(fromSlot, 1)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get confirmed blocks from slot: %v, for %v. error: %v\n", fromSlot, expectedTx.ChainId, err)
			bp.logger.Error("Failed to get confirmed blocks", "fromSlot", fromSlot, "chainId", expectedTx.ChainId, "err", err)
			break
		}

		if len(blocks) == 1 && bridgeClaims.BlockInfoEqualWithExpected(expectedTx, blocks[0]) {
			relevantExpiredTxs = append(relevantExpiredTxs, expectedTx)
		}
	}

	var invalidRelevantExpiredTxs []*core.BridgeExpectedCardanoTx
	var processedRelevantExpiredTxs []*core.BridgeExpectedCardanoTx

	if bridgeClaims.Count() < bp.appConfig.Settings.MaxBridgingClaimsToGroup && len(relevantExpiredTxs) > 0 {
	expiredTxsLoop:
		for _, expiredTx := range relevantExpiredTxs {
			processedTx, _ := bp.db.GetProcessedTx(expiredTx.ChainId, expiredTx.Hash)
			if processedTx != nil && !processedTx.IsInvalid {
				// already sent the success claim
				processedRelevantExpiredTxs = append(processedRelevantExpiredTxs, expiredTx)
				continue
			}

			var expiredTxProcessed = false
		failedTxProcessorsLoop:
			for _, txProcessor := range bp.failedTxProcessors {
				relevant, err := txProcessor.IsTxRelevant(expiredTx, bp.appConfig)
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

					if bridgeClaims.Count() >= bp.appConfig.Settings.MaxBridgingClaimsToGroup {
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

func (bp *CardanoTxsProcessorImpl) processAllForChain(
	chainId string,
) {
	expectedTxs, err := bp.db.GetExpectedTxs(chainId, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get expected txs. error: %v\n", err)
		bp.logger.Error("Failed to get expected txs", "err", err)
		return
	}

	unprocessedTxs, err := bp.db.GetUnprocessedTxs(chainId, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get unprocessed txs. error: %v\n", err)
		bp.logger.Error("Failed to get unprocessed txs", "err", err)
		return
	}

	bridgeClaims, ccoDb := bp.constructBridgeClaims(chainId, unprocessedTxs, expectedTxs)
	if bridgeClaims == nil {
		return
	}

	expectedTxsMap := make(map[string]*core.BridgeExpectedCardanoTx, len(expectedTxs))
	for _, expectedTx := range expectedTxs {
		expectedTxsMap[expectedTx.ToCardanoTxKey()] = expectedTx
	}

	relevantUnprocessedTxs, processedTxs, processedExpectedTxs := bp.checkUnprocessedTxs(bridgeClaims, unprocessedTxs, expectedTxsMap)
	relevantExpiredTxs, processedRelevantExpiredTxs, invalidRelevantExpiredTxs := bp.checkExpectedTxs(bridgeClaims, ccoDb, expectedTxsMap)
	processedExpectedTxs = append(processedExpectedTxs, processedRelevantExpiredTxs...)

	bridgeClaims.BlockFullyObserved = len(processedTxs) == len(relevantUnprocessedTxs) &&
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
	err = bp.claimsSubmitter.SubmitClaims(bridgeClaims)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to submit claims. error: %v\n", err)
		bp.logger.Error("Failed to submit claims", "err", err)
		return
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
