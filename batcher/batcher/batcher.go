package batcher

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/batcher/bridge"
	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	wallet "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hashicorp/go-hclog"
)

type BatcherImpl struct {
	config     *core.BatcherConfiguration
	logger     hclog.Logger
	ethClient  *ethclient.Client
	operations core.ChainOperations
}

var _ core.Batcher = (*BatcherImpl)(nil)

func NewBatcher(
	config *core.BatcherConfiguration,
	logger hclog.Logger,
	operations core.ChainOperations) *BatcherImpl {
	return &BatcherImpl{
		config:     config,
		logger:     logger,
		ethClient:  nil,
		operations: operations,
	}
}

func (b *BatcherImpl) Start(ctx context.Context) {
	var (
		timerTime = time.Millisecond * time.Duration(b.config.PullTimeMilis)
	)

	b.logger.Debug("Batcher started")

	timer := time.NewTimer(timerTime)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
		case <-ctx.Done():
			return
		}

		b.execute(ctx)

		timer.Reset(timerTime)
	}
}

func (b *BatcherImpl) execute(ctx context.Context) {
	var (
		err error
	)

	if b.ethClient == nil {
		b.ethClient, err = ethclient.Dial(b.config.Bridge.NodeUrl)
		if err != nil {
			b.logger.Error("Failed to dial bridge", "err", err)
			return
		}
	}

	ethTxHelper, err := ethtxhelper.NewEThTxHelper(ethtxhelper.WithClient(b.ethClient))
	if err != nil {
		// In case of error, reset ethClient to nil to try again in the next iteration.
		b.ethClient = nil
		return
	}

	// Check if I should create batch
	shouldCreateBatch, err := bridge.ShouldCreateBatch(ctx, ethTxHelper, b.config.Bridge.SmartContractAddress, b.config.Base.ChainId)
	if err != nil {
		b.logger.Error("Failed to query bridge.ShouldCreateBatch", "err", err)

		b.ethClient = nil
		return
	}

	if !shouldCreateBatch {
		b.logger.Info("Called ShouldCreateBatch before it supposed to or already created this batch")
		return
	}
	b.logger.Info("Starting batch creation process")

	// Get confirmed transactions from smart contract
	confirmedTransactions, err := bridge.GetConfirmedTransactions(ctx, ethTxHelper, b.config.Bridge.SmartContractAddress, b.config.Base.ChainId)
	if err != nil {
		b.logger.Error("Failed to query bridge.GetConfirmedTransactions", "err", err)

		b.ethClient = nil
		return
	}

	// Generate batch transaction
	rawTx, txHash, utxos, err := b.operations.GenerateBatchTransaction(ctx, ethTxHelper, b.config.Bridge.SmartContractAddress, b.config.Base.ChainId, confirmedTransactions)
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
	signedBatch := contractbinding.TestContractSignedBatch{
		Id:                        "",
		DestinationChainId:        b.config.Base.ChainId,
		RawTransaction:            hex.EncodeToString(rawTx),
		MultisigSignature:         hex.EncodeToString(multisigSignature),
		FeePayerMultisigSignature: hex.EncodeToString(multisigFeeSignature),
		IncludedTransactions:      confirmedTransactions,
		UsedUTXOs:                 *utxos,
	}

	err = bridge.SubmitSignedBatch(ctx, ethTxHelper, b.config.Bridge.SmartContractAddress, signedBatch, b.config.Bridge.SigningKey)
	if err != nil {
		b.ethClient = nil
		b.logger.Error("Failed to submit signed batch", "err", err)
		return
	}
	b.logger.Info("Batch successfully submited")
}

// GetChainSpecificOperations returns the chain-specific operations based on the chain type
func GetChainSpecificOperations(config core.ChainSpecific, pkPath string) (core.ChainOperations, error) {
	var operations core.ChainOperations

	// Create the appropriate chain-specific configuration based on the chain type
	switch config.ChainType {
	case "Cardano":
		var cardanoChainConfig core.CardanoChainConfig
		if err := json.Unmarshal(config.Config, &cardanoChainConfig); err != nil {
			return nil, fmt.Errorf("failed to unmarshal Cardano configuration: %v", err)
		}

		cardanoWallet, err := wallet.LoadWallet(pkPath, false)
		if err != nil {
			return nil, fmt.Errorf("error while loading wallet info: %v\n", err)
		}

		operations = NewCardanoChainOperations(cardanoChainConfig, *cardanoWallet)
	default:
		return nil, fmt.Errorf("unknown chain type: %s", config.ChainType)
	}

	return operations, nil
}
