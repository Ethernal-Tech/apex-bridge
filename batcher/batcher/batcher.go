package batcher

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/telemetry"
	"github.com/hashicorp/go-hclog"
)

type lastBatchData struct {
	id   uint64
	slot uint64
}

type BatcherImpl struct {
	config                      *core.BatcherConfiguration
	logger                      hclog.Logger
	operations                  core.ChainOperations
	bridgeSmartContract         eth.IBridgeSmartContract
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater
	lastBatch                   lastBatchData
}

var _ core.Batcher = (*BatcherImpl)(nil)

func NewBatcher(
	config *core.BatcherConfiguration,
	logger hclog.Logger,
	operations core.ChainOperations, bridgeSmartContract eth.IBridgeSmartContract,
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater) *BatcherImpl {
	return &BatcherImpl{
		config:                      config,
		logger:                      logger,
		operations:                  operations,
		bridgeSmartContract:         bridgeSmartContract,
		bridgingRequestStateUpdater: bridgingRequestStateUpdater,
		lastBatch:                   lastBatchData{},
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

		if err := b.execute(ctx); err != nil {
			if errors.Is(err, errNonActiveBatchPeriod) {
				b.logger.Info("execution skipped", "reason", err)
			} else {
				// update telemetry
				if b.lastBatch.id == 0 {
					batchID, err := b.bridgeSmartContract.GetNextBatchID(ctx, b.config.Chain.ChainID)
					if err == nil {
						telemetry.UpdateBatcherBatchSubmitFailed(b.config.Chain.ChainID, batchID)
					}
				} else {
					telemetry.UpdateBatcherBatchSubmitFailed(b.config.Chain.ChainID, b.lastBatch.id+1)
				}

				b.logger.Error("execution failed", "err", err)
			}
		}
	}
}

func (b *BatcherImpl) execute(ctx context.Context) error {
	// Check if I should create batch
	batchID, err := b.bridgeSmartContract.GetNextBatchID(ctx, b.config.Chain.ChainID)
	if err != nil {
		return fmt.Errorf("failed to query bridge.GetNextBatchID for chainID: %s. err: %w", b.config.Chain.ChainID, err)
	}

	if batchID == 0 {
		b.logger.Info("Waiting on a new batch", "chainID", b.config.Chain.ChainID)

		return nil
	}

	if batchID < b.lastBatch.id {
		b.logger.Info("retrieved batch id not good", "chainID", b.config.Chain.ChainID,
			"old", b.lastBatch.id, "new", batchID)

		return nil
	}

	b.logger.Info("Starting batch creation process", "chainID", b.config.Chain.ChainID, "batchID", batchID)

	// Get confirmed transactions from smart contract
	confirmedTransactions, err := b.bridgeSmartContract.GetConfirmedTransactions(ctx, b.config.Chain.ChainID)
	if err != nil {
		return fmt.Errorf("failed to query bridge.GetConfirmedTransactions for chainID: %s. err: %w",
			b.config.Chain.ChainID, err)
	}

	if len(confirmedTransactions) == 0 {
		return fmt.Errorf("batch should not be created for zero number of confirmed transactions. chainID: %s",
			b.config.Chain.ChainID)
	}

	b.logger.Debug("Successfully queried smart contract for confirmed transactions",
		"chainID", b.config.Chain.ChainID, "batchID", batchID, "txs", len(confirmedTransactions))

	// Generate batch transaction
	generatedBatchData, err := b.operations.GenerateBatchTransaction(
		ctx, b.bridgeSmartContract, b.config.Chain.ChainID, confirmedTransactions, batchID)
	if err != nil {
		return fmt.Errorf("failed to generate batch transaction for chainID: %s. err: %w",
			b.config.Chain.ChainID, err)
	}

	if b.lastBatch.id == batchID && b.lastBatch.slot >= generatedBatchData.Slot {
		b.logger.Debug("already submitted", "chainID", b.config.Chain.ChainID,
			"batchID", batchID, "slot", b.lastBatch.slot)

		return nil
	}

	b.logger.Info("Created batch tx", "chainID", b.config.Chain.ChainID, "txHash", generatedBatchData.TxHash,
		"batchID", batchID, "txs", len(confirmedTransactions))

	// Sign batch transaction
	multisigSignature, multisigFeeSignature, err := b.operations.SignBatchTransaction(generatedBatchData.TxHash)
	if err != nil {
		return fmt.Errorf("failed to sign batch transaction for chainID: %s. err: %w",
			b.config.Chain.ChainID, err)
	}

	b.logger.Info("Batch successfully signed", "chainID", b.config.Chain.ChainID,
		"batchID", batchID, "txs", len(confirmedTransactions))

	firstTxNonceID, lastTxNonceID := getFirstAndLastTxNonceID(confirmedTransactions)
	// Submit batch to smart contract
	signedBatch := eth.SignedBatch{
		Id:                        batchID,
		DestinationChainId:        common.ToNumChainID(b.config.Chain.ChainID),
		RawTransaction:            generatedBatchData.TxRaw,
		MultisigSignature:         multisigSignature,
		FeePayerMultisigSignature: multisigFeeSignature,
		FirstTxNonceId:            firstTxNonceID,
		LastTxNonceId:             lastTxNonceID,
		UsedUTXOs:                 generatedBatchData.Utxos,
	}

	b.logger.Debug("Submitting signed batch to smart contract", "chainID", b.config.Chain.ChainID,
		"signedBatch", eth.BatchToString(signedBatch))

	err = b.bridgeSmartContract.SubmitSignedBatch(ctx, signedBatch)
	if err != nil {
		return fmt.Errorf("failed to submit signed batch: %w", err)
	}

	if b.lastBatch.id != batchID {
		brStateKeys := getBridgingRequestStateKeys(confirmedTransactions, firstTxNonceID, lastTxNonceID)

		err = b.bridgingRequestStateUpdater.IncludedInBatch(b.config.Chain.ChainID, batchID, brStateKeys)
		if err != nil {
			b.logger.Error(
				"error while updating bridging request states to IncludedInBatch",
				"chain", b.config.Chain.ChainID, "batchID", batchID)
		}

		telemetry.UpdateBatcherBatchSubmitSucceeded(b.config.Chain.ChainID, batchID)

		b.logger.Info("Batch successfully submitted", "chainID", b.config.Chain.ChainID,
			"batchID", batchID, "first", firstTxNonceID, "last", lastTxNonceID, "txs", brStateKeys)
	} else {
		b.logger.Info("Batch successfully re-submitted", "chainID", b.config.Chain.ChainID,
			"batchID", batchID, "first", firstTxNonceID, "last", lastTxNonceID)
	}

	// update last batch data
	b.lastBatch = lastBatchData{
		id:   batchID,
		slot: generatedBatchData.Slot,
	}

	return nil
}

// GetChainSpecificOperations returns the chain-specific operations based on the chain type
func GetChainSpecificOperations(config core.ChainConfig, logger hclog.Logger) (core.ChainOperations, error) {
	// Create the appropriate chain-specific configuration based on the chain type
	switch strings.ToLower(config.ChainType) {
	case "cardano":
		return NewCardanoChainOperations(config.ChainSpecific, logger)
	default:
		return nil, fmt.Errorf("unknown chain type: %s", config.ChainType)
	}
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
		if firstTxNonceID <= confirmedTx.Nonce && confirmedTx.Nonce <= lastTxNonceID {
			txsInBatch = append(txsInBatch, common.BridgingRequestStateKey{
				SourceChainID: common.ToStrChainID(confirmedTx.SourceChainId),
				SourceTxHash:  hex.EncodeToString(confirmedTx.ObservedTransactionHash[:]),
			})
		}
	}

	return txsInBatch
}
