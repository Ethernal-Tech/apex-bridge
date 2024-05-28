package batcher

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/telemetry"
	"github.com/hashicorp/go-hclog"
)

type BatcherImpl struct {
	config                      *core.BatcherConfiguration
	logger                      hclog.Logger
	operations                  core.ChainOperations
	bridgeSmartContract         eth.IBridgeSmartContract
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater
	lastBatchID                 *big.Int
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
		lastBatchID:                 big.NewInt(0),
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
				if b.lastBatchID.BitLen() == 0 {
					batchID, err := b.bridgeSmartContract.GetNextBatchID(ctx, b.config.Chain.ChainID)
					if err == nil {
						telemetry.UpdateBatcherBatchSubmitFailed(b.config.Chain.ChainID, batchID.Uint64())
					}
				} else {
					telemetry.UpdateBatcherBatchSubmitFailed(b.config.Chain.ChainID, b.lastBatchID.Uint64()+1)
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

	if batchID.Cmp(big.NewInt(0)) == 0 {
		b.logger.Info("Waiting on a new batch", "chainID", b.config.Chain.ChainID)

		return nil
	}

	if batchID.Cmp(b.lastBatchID) <= 0 {
		b.logger.Info("retrieved batch id not good", "chainID", b.config.Chain.ChainID,
			"old", b.lastBatchID, "new", batchID)

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
		DestinationChainId:        b.config.Chain.ChainID,
		RawTransaction:            hex.EncodeToString(generatedBatchData.TxRaw),
		MultisigSignature:         hex.EncodeToString(multisigSignature),
		FeePayerMultisigSignature: hex.EncodeToString(multisigFeeSignature),
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

	brStateKeys := getBridgingRequestStateKeys(confirmedTransactions, firstTxNonceID, lastTxNonceID)

	b.logger.Info("Batch successfully submitted", "chainID", b.config.Chain.ChainID,
		"batchID", batchID, "txs cnt", len(confirmedTransactions), "txs", brStateKeys)

	err = b.bridgingRequestStateUpdater.IncludedInBatch(b.config.Chain.ChainID, batchID.Uint64(), brStateKeys)
	if err != nil {
		b.logger.Error(
			"error while updating bridging request states to IncludedInBatch",
			"chain", b.config.Chain.ChainID, "batchId", batchID)
	}

	b.lastBatchID = batchID // update last batch id

	telemetry.UpdateBatcherBatchSubmitSucceeded(b.config.Chain.ChainID, batchID.Uint64())

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

func getFirstAndLastTxNonceID(confirmedTxs []eth.ConfirmedTransaction) (*big.Int, *big.Int) {
	first, last := confirmedTxs[0].Nonce, confirmedTxs[0].Nonce

	for _, x := range confirmedTxs[1:] {
		if first.Cmp(x.Nonce) > 0 {
			first = x.Nonce
		}

		if last.Cmp(x.Nonce) < 0 {
			last = x.Nonce
		}
	}

	return first, last
}

func getBridgingRequestStateKeys(
	txs []eth.ConfirmedTransaction, firstTxNonceID, lastTxNonceID *big.Int,
) []common.BridgingRequestStateKey {
	txsInBatch := make([]common.BridgingRequestStateKey, 0, lastTxNonceID.Uint64()-firstTxNonceID.Uint64()+1)

	for _, confirmedTx := range txs {
		if confirmedTx.Nonce.Cmp(firstTxNonceID) >= 0 && confirmedTx.Nonce.Cmp(lastTxNonceID) <= 0 {
			txsInBatch = append(txsInBatch, common.BridgingRequestStateKey{
				SourceChainID: confirmedTx.SourceChainID,
				SourceTxHash:  confirmedTx.ObservedTransactionHash,
			})
		}
	}

	return txsInBatch
}
