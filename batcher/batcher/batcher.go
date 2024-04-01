package batcher

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	wallet "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/hashicorp/go-hclog"
)

type BatcherImpl struct {
	config              *core.BatcherConfiguration
	logger              hclog.Logger
	operations          core.ChainOperations
	bridgeSmartContract eth.IBridgeSmartContract
}

var _ core.Batcher = (*BatcherImpl)(nil)

func NewBatcher(
	config *core.BatcherConfiguration,
	logger hclog.Logger,
	operations core.ChainOperations, bridgeSmartContract eth.IBridgeSmartContract) *BatcherImpl {
	return &BatcherImpl{
		config:              config,
		logger:              logger,
		operations:          operations,
		bridgeSmartContract: bridgeSmartContract,
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

		b.execute(ctx)
	}
}

func (b *BatcherImpl) execute(ctx context.Context) {
	var (
		err error
	)

	// Check if I should create batch
	shouldCreateBatch, err := b.bridgeSmartContract.ShouldCreateBatch(ctx, b.config.Base.ChainId)
	if err != nil {
		b.logger.Error("Failed to query bridge.ShouldCreateBatch", "err", err)
		return
	}

	if !shouldCreateBatch {
		b.logger.Info("Called ShouldCreateBatch before it supposed to or already created this batch")
		return
	}
	b.logger.Info("Starting batch creation process")

	b.logger.Info("Query smart contract for confirmed transactions")
	// Get confirmed transactions from smart contract
	confirmedTransactions, err := b.bridgeSmartContract.GetConfirmedTransactions(ctx, b.config.Base.ChainId)
	if err != nil {
		b.logger.Error("Failed to query bridge.GetConfirmedTransactions", "err", err)
		return
	}
	b.logger.Info("Successfully queried smart contract for confirmed transactions")

	// Generate batch transaction
	rawTx, txHash, utxos, err := b.operations.GenerateBatchTransaction(ctx, b.bridgeSmartContract, b.config.Base.ChainId, confirmedTransactions)
	if err != nil {
		b.logger.Error("Failed to generate batch transaction", "err", err)
		return
	}

	b.logger.Info("Created tx", "txHash", txHash)

	// Sign batch transaction
	multisigSignature, multisigFeeSignature, err := b.operations.SignBatchTransaction(txHash)
	if err != nil {
		b.logger.Error("Failed to sign batch transaction", "err", err)
		return
	}

	b.logger.Info("Batch successfully signed")

	// TODO: Update ID
	// Submit batch to smart contract
	signedBatch := eth.SignedBatch{
		Id:                        big.NewInt(0),
		DestinationChainId:        b.config.Base.ChainId,
		RawTransaction:            hex.EncodeToString(rawTx),
		MultisigSignature:         hex.EncodeToString(multisigSignature),
		FeePayerMultisigSignature: hex.EncodeToString(multisigFeeSignature),
		IncludedTransactions:      confirmedTransactions,
		UsedUTXOs:                 *utxos,
	}

	b.logger.Info("Submiting signed batch to smart contract")
	err = b.bridgeSmartContract.SubmitSignedBatch(ctx, signedBatch)
	if err != nil {
		b.logger.Error("Failed to submit signed batch", "err", err)
		return
	}
	b.logger.Info("Batch successfully submited")
}

// GetChainSpecificOperations returns the chain-specific operations based on the chain type
func GetChainSpecificOperations(config core.ChainSpecific, pkPath string) (core.ChainOperations, error) {
	var operations core.ChainOperations

	// Create the appropriate chain-specific configuration based on the chain type
	switch strings.ToLower(config.ChainType) {
	case "cardano":
		var cardanoChainConfig core.CardanoChainConfig
		if err := json.Unmarshal(config.Config, &cardanoChainConfig); err != nil {
			return nil, fmt.Errorf("failed to unmarshal Cardano configuration: %v", err)
		}

		cardanoWallet, err := wallet.LoadWallet(pkPath, false)
		if err != nil {
			return nil, fmt.Errorf("error while loading wallet info: %v", err)
		}

		operations = NewCardanoChainOperations(cardanoChainConfig, *cardanoWallet)
	default:
		return nil, fmt.Errorf("unknown chain type: %s", config.ChainType)
	}

	return operations, nil
}
