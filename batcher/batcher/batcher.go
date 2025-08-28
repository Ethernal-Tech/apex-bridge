package batcher

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/telemetry"
	"github.com/hashicorp/go-hclog"
)

type lastBatchData struct {
	id     uint64
	txHash string
}

type BatcherImpl struct {
	config                      *core.BatcherConfiguration
	operations                  core.ChainOperations
	bridgeSmartContract         eth.IBridgeSmartContract
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater
	lastBatch                   lastBatchData
	logger                      hclog.Logger
}

var _ core.Batcher = (*BatcherImpl)(nil)

func NewBatcher(
	config *core.BatcherConfiguration,
	operations core.ChainOperations,
	bridgeSmartContract eth.IBridgeSmartContract,
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater,
	bridgingAddressesManager common.BridgingAddressesManager,
	bridgingAddressesCoordinator common.BridgingAddressesCoordinator,
	logger hclog.Logger,
) *BatcherImpl {
	return &BatcherImpl{
		config:                      config,
		operations:                  operations,
		bridgeSmartContract:         bridgeSmartContract,
		bridgingRequestStateUpdater: bridgingRequestStateUpdater,
		lastBatch:                   lastBatchData{},
		logger:                      logger,
	}
}

func (b *BatcherImpl) Start(ctx context.Context) {
	b.logger.Debug("Batcher started")

	waitTime := time.Millisecond * time.Duration(b.config.PullTimeMilis)

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(waitTime):
		}

		isSync, err := b.operations.IsSynchronized(ctx, b.bridgeSmartContract, b.config.Chain.ChainID)
		if err != nil {
			b.logger.Error("is synchronized check failed", "err", err)

			continue
		} else if !isSync {
			b.logger.Info("batcher is not synchronized - creating batch skipped")

			continue
		}

		batchID, err := b.execute(ctx)
		if err != nil {
			if errors.Is(err, errNonActiveBatchPeriod) {
				b.logger.Info("execution skipped", "reason", err)
			} else {
				// update telemetry only if batchID is specified
				if batchID != 0 {
					telemetry.UpdateBatcherBatchSubmitFailed(b.config.Chain.ChainID, batchID)
				}

				b.logger.Error("execution failed", "err", err)
			}
		}
	}
}

func (b *BatcherImpl) execute(ctx context.Context) (uint64, error) {
	// Check if I should create batch
	batchID, err := b.bridgeSmartContract.GetNextBatchID(ctx, b.config.Chain.ChainID)
	if err != nil {
		return 0, fmt.Errorf("failed to query bridge.GetNextBatchID for chainID: %s. err: %w",
			b.config.Chain.ChainID, err)
	}

	if batchID == 0 {
		b.logger.Info("Waiting on a new batch")

		return 0, nil
	}

	if batchID < b.lastBatch.id {
		return 0, fmt.Errorf("retrieved batch id is not good for chainID: %s. old: %d vs new: %d",
			b.config.Chain.ChainID, b.lastBatch.id, batchID)
	}

	b.logger.Info("Starting batch creation process", "batchID", batchID)

	// Get confirmed transactions from smart contract
	confirmedTransactions, err := b.bridgeSmartContract.GetConfirmedTransactions(ctx, b.config.Chain.ChainID)
	if err != nil {
		return batchID, fmt.Errorf("failed to query bridge.GetConfirmedTransactions for chainID: %s. err: %w",
			b.config.Chain.ChainID, err)
	}

	if len(confirmedTransactions) == 0 {
		return batchID, fmt.Errorf("batch should not be created for zero number of confirmed transactions. chainID: %s",
			b.config.Chain.ChainID)
	}

	b.logger.Debug("Successfully queried smart contract for confirmed transactions",
		"batchID", batchID, "txs", eth.ConfirmedTransactionsWrapper{Txs: confirmedTransactions})

	// Generate batch transaction
	generatedBatchData, err := b.operations.GenerateBatchTransaction(
		ctx, b.config.Chain.ChainID, confirmedTransactions, batchID)
	if err != nil {
		return batchID, fmt.Errorf("failed to generate batch transaction for chainID: %s. err: %w",
			b.config.Chain.ChainID, err)
	}

	if generatedBatchData.TxHash == b.lastBatch.txHash {
		// there is nothing different to submit
		b.logger.Debug("generated batch is the same as the previous one",
			"batchID", batchID, "txHash", b.lastBatch.txHash,
			"batchType", generatedBatchData.BatchType)

		return batchID, nil
	}

	b.logger.Info("Created batch tx", "batchID", batchID, "txHash", generatedBatchData.TxHash,
		"batchType", generatedBatchData.BatchType, "txs", len(confirmedTransactions))

	// Sign batch transaction
	signatures, err := b.operations.SignBatchTransaction(generatedBatchData)
	if err != nil {
		return batchID, fmt.Errorf("failed to sign batch transaction for chainID: %s. err: %w",
			b.config.Chain.ChainID, err)
	}

	b.logger.Info("Batch successfully signed", "batchID", batchID, "txs", len(confirmedTransactions))

	// Submit batch to smart contract
	signedBatch := b.createSignedBatch(
		batchID, generatedBatchData, signatures, confirmedTransactions)

	b.logger.Debug("Submitting signed batch to smart contract", "batchID", batchID,
		"signedBatch", eth.SignedBatchWrapper{SignedBatch: signedBatch})

	err = b.operations.Submit(ctx, b.bridgeSmartContract, *signedBatch)
	if err != nil {
		return batchID, fmt.Errorf("failed to submit signed batch: %w", err)
	}

	if b.lastBatch.id != batchID {
		stateKeys := b.updateBatchTxsStates(batchID, confirmedTransactions, signedBatch)

		telemetry.UpdateBatcherBatchSubmitSucceeded(b.config.Chain.ChainID, batchID)

		b.logger.Info("Batch successfully submitted", "batchID", batchID, "stateKeys", stateKeys)
	} else {
		b.logger.Info("Batch successfully re-submitted", "batchID", batchID)
	}

	// update last batch data
	b.lastBatch = lastBatchData{
		id:     batchID,
		txHash: generatedBatchData.TxHash,
	}

	return batchID, nil
}

func (b *BatcherImpl) createSignedBatch(
	batchID uint64, generatedBatchData *core.GeneratedBatchTxData,
	signatures *core.BatchSignatures,
	confirmedTxs []eth.ConfirmedTransaction,
) *eth.SignedBatch {
	firstTxNonceID, lastTxNonceID := uint64(0), uint64(0)
	if !generatedBatchData.IsConsolidation() {
		firstTxNonceID, lastTxNonceID = getFirstAndLastTxNonceID(confirmedTxs)
	}

	return &eth.SignedBatch{
		Id:                 batchID,
		DestinationChainId: common.ToNumChainID(b.config.Chain.ChainID),
		RawTransaction:     generatedBatchData.TxRaw,
		Signature:          signatures.Multisig,
		StakeSignature:     signatures.MultsigStake,
		FeeSignature:       signatures.Fee,
		FirstTxNonceId:     firstTxNonceID,
		LastTxNonceId:      lastTxNonceID,
		BatchType:          uint8(generatedBatchData.BatchType),
	}
}

func (b *BatcherImpl) updateBatchTxsStates(
	batchID uint64, txs []eth.ConfirmedTransaction, signedBatch *eth.SignedBatch,
) []common.BridgingRequestStateKey {
	if signedBatch.BatchType == uint8(eth.BatchTypeConsolidation) {
		return nil
	}

	txsInBatch := getBridgingRequestStateKeys(txs, signedBatch.FirstTxNonceId, signedBatch.LastTxNonceId)

	if err := b.bridgingRequestStateUpdater.IncludedInBatch(txsInBatch, b.config.Chain.ChainID); err != nil {
		b.logger.Error("failed to update batch txs states", "batchID", batchID, "err", err)
	}

	return txsInBatch
}

func getFirstAndLastTxNonceID(confirmedTxs []eth.ConfirmedTransaction) (uint64, uint64) {
	first, last := confirmedTxs[0].Nonce, confirmedTxs[0].Nonce

	for _, x := range confirmedTxs[1:] {
		if x.Nonce < first {
			first = x.Nonce
		}

		if x.Nonce > last {
			last = x.Nonce
		}
	}

	return first, last
}

func getBridgingRequestStateKeys(
	txs []eth.ConfirmedTransaction, firstTxNonceID uint64, lastTxNonceID uint64,
) []common.BridgingRequestStateKey {
	txsInBatch := make([]common.BridgingRequestStateKey, 0, lastTxNonceID-firstTxNonceID+1)

	for _, confirmedTx := range txs {
		if firstTxNonceID <= confirmedTx.Nonce && confirmedTx.Nonce <= lastTxNonceID &&
			!common.IsDirectlyConfirmedTransaction(confirmedTx.TransactionType) {
			txsInBatch = append(txsInBatch, common.NewBridgingRequestStateKey(
				common.ToStrChainID(confirmedTx.SourceChainId), confirmedTx.ObservedTransactionHash))
		}
	}

	return txsInBatch
}
