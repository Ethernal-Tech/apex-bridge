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

		if err := b.execute(ctx); err != nil {
			b.logger.Error("execute failed", "err", err)
		}
	}
}

func (b *BatcherImpl) execute(ctx context.Context) error {
	var (
		err error
	)

	// Check if I should create batch
	batchId, err := b.bridgeSmartContract.GetNextBatchId(ctx, b.config.Base.ChainId)
	if err != nil {
		return fmt.Errorf("failed to query bridge.GetNextBatchId: %v", err)
	}

	if batchId == big.NewInt(0) {
		b.logger.Info("Waiting on a new batch")
		return nil
	}
	b.logger.Info("Starting batch creation process")

	b.logger.Info("Query smart contract for confirmed transactions")
	// Get confirmed transactions from smart contract
	confirmedTransactions, err := b.bridgeSmartContract.GetConfirmedTransactions(ctx, b.config.Base.ChainId)
	if err != nil {
		return fmt.Errorf("failed to query bridge.GetConfirmedTransactions: %v", err)
	}
	b.logger.Info("Successfully queried smart contract for confirmed transactions")

	// Generate batch transaction
	rawTx, txHash, utxos, err := b.operations.GenerateBatchTransaction(ctx, b.bridgeSmartContract, b.config.Base.ChainId, confirmedTransactions, batchId)
	if err != nil {
		return fmt.Errorf("failed to generate batch transaction: %v", err)
	}

	b.logger.Info("Created tx", "txHash", txHash)

	// Sign batch transaction
	multisigSignature, multisigFeeSignature, err := b.operations.SignBatchTransaction(txHash)
	if err != nil {
		return fmt.Errorf("failed to sign batch transaction: %v", err)
	}

	b.logger.Info("Batch successfully signed")

	// Submit batch to smart contract
	signedBatch := eth.SignedBatch{
		Id:                        batchId,
		DestinationChainId:        b.config.Base.ChainId,
		RawTransaction:            hex.EncodeToString(rawTx),
		MultisigSignature:         hex.EncodeToString(multisigSignature),
		FeePayerMultisigSignature: hex.EncodeToString(multisigFeeSignature),
		IncludedTransactions:      []*big.Int{},
		UsedUTXOs:                 *utxos,
	}

	b.logger.Info("Submiting signed batch to smart contract")
	err = b.bridgeSmartContract.SubmitSignedBatch(ctx, signedBatch)
	if err != nil {
		return fmt.Errorf("failed to submit signed batch: %v", err)
	}
	b.logger.Info("Batch successfully submited")

	return nil
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
