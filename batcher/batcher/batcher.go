package batcher

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/hashicorp/go-hclog"
)

type BatcherImpl struct {
	config                      *core.BatcherConfiguration
	logger                      hclog.Logger
	operations                  core.ChainOperations
	bridgeSmartContract         eth.IBridgeSmartContract
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater
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
	}
}

func (b *BatcherImpl) Start(ctx context.Context) {
	b.logger.Debug("Batcher started")

	ticker := time.NewTicker(time.Millisecond * time.Duration(b.config.PullTimeMilis))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		if err := b.execute(ctx); err != nil {
			b.logger.Error("execute failed", "err", err)
		}
	}
}

func (b *BatcherImpl) execute(ctx context.Context) error {
	// Check if I should create batch
	batchId, err := b.bridgeSmartContract.GetNextBatchId(ctx, b.config.Chain.ChainId)
	if err != nil {
		return fmt.Errorf("failed to query bridge.GetNextBatchId: %w", err)
	}

	if batchId.Cmp(big.NewInt(0)) == 0 {
		b.logger.Info("Waiting on a new batch")
		return nil
	}

	b.logger.Info("Starting batch creation process")

	b.logger.Info("Query smart contract for confirmed transactions")
	// Get confirmed transactions from smart contract
	confirmedTransactions, err := b.bridgeSmartContract.GetConfirmedTransactions(ctx, b.config.Chain.ChainId)
	if err != nil {
		return fmt.Errorf("failed to query bridge.GetConfirmedTransactions: %w", err)
	}

	b.logger.Info("Successfully queried smart contract for confirmed transactions")

	// Generate batch transaction
	rawTx, txHash, utxos, includedConfirmedTransactions, err := b.operations.GenerateBatchTransaction(
		ctx, b.bridgeSmartContract, b.config.Chain.ChainId, confirmedTransactions, batchId)
	if err != nil {
		return fmt.Errorf("failed to generate batch transaction: %w", err)
	}

	includedConfirmedTransactionsNonces := make([]*big.Int, 0, len(includedConfirmedTransactions))
	for _, tx := range includedConfirmedTransactions {
		includedConfirmedTransactionsNonces = append(includedConfirmedTransactionsNonces, tx.Nonce)
	}

	b.logger.Info("Created tx", "txHash", txHash)

	// Sign batch transaction
	multisigSignature, multisigFeeSignature, err := b.operations.SignBatchTransaction(txHash)
	if err != nil {
		return fmt.Errorf("failed to sign batch transaction: %w", err)
	}

	b.logger.Info("Batch successfully signed")

	// Submit batch to smart contract
	signedBatch := eth.SignedBatch{
		Id:                        batchId,
		DestinationChainId:        b.config.Chain.ChainId,
		RawTransaction:            hex.EncodeToString(rawTx),
		MultisigSignature:         hex.EncodeToString(multisigSignature),
		FeePayerMultisigSignature: hex.EncodeToString(multisigFeeSignature),
		IncludedTransactions:      includedConfirmedTransactionsNonces,
		UsedUTXOs:                 *utxos,
	}

	b.logger.Info("Submitting signed batch to smart contract")
	err = b.bridgeSmartContract.SubmitSignedBatch(ctx, signedBatch)
	if err != nil {
		return fmt.Errorf("failed to submit signed batch: %w", err)
	}

	b.logger.Info("Batch successfully submitted")

	txsInBatch := make([]common.BridgingRequestStateKey, 0, len(includedConfirmedTransactionsNonces))
	for _, confirmedTx := range confirmedTransactions {
		if _, exists := includedConfirmedTransactions[confirmedTx.Nonce.Uint64()]; exists {
			txsInBatch = append(txsInBatch, common.BridgingRequestStateKey{
				SourceChainId: confirmedTx.SourceChainID,
				SourceTxHash:  confirmedTx.ObservedTransactionHash,
			})
		}
	}

	err = b.bridgingRequestStateUpdater.IncludedInBatch(
		signedBatch.DestinationChainId, signedBatch.Id.Uint64(), txsInBatch)
	if err != nil {
		b.logger.Error(
			"error while updating bridging request states to IncludedInBatch",
			"destinationChainId", signedBatch.DestinationChainId, "batchId", signedBatch.Id.Uint64())
	}

	return nil
}

// GetChainSpecificOperations returns the chain-specific operations based on the chain type
func GetChainSpecificOperations(config core.ChainConfig) (core.ChainOperations, error) {
	// Create the appropriate chain-specific configuration based on the chain type
	switch strings.ToLower(config.ChainType) {
	case "cardano":
		return NewCardanoChainOperations(config.ChainSpecific)
	default:
		return nil, fmt.Errorf("unknown chain type: %s", config.ChainType)
	}
}
