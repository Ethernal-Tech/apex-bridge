package core

import (
	"context"

	cardano "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/validatorobserver"
)

type GeneratedBatchTxData struct {
	BatchType uint8
	TxRaw     []byte
	TxHash    string
}

type BatcherManager interface {
	Start()
}

type Batcher interface {
	Start(ctx context.Context)
	UpdateValidatorSet(validators *validatorobserver.ValidatorsPerChain)
}

type ChainOperations interface {
	GenerateBatchTransaction(
		ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract,
		destinationChain string, confirmedTransactions []eth.ConfirmedTransaction, batchNonceID uint64,
	) (*GeneratedBatchTxData, error)
	SignBatchTransaction(generatedBatchData *GeneratedBatchTxData) ([]byte, []byte, error)
	IsSynchronized(
		ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract, chainID string,
	) (bool, error)
	Submit(ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract, batch eth.SignedBatch) error

	// Update & transfer to new multisig
	CreateValidatorSetChangeTx(ctx context.Context,
		chainID string, nextBatchID uint64,
		bridgeSmartContract eth.IBridgeSmartContract,
		validatorsKeys validatorobserver.ValidatorsPerChain,
		lastBatchID uint64, lastBatchType uint8,
	) (bool, *GeneratedBatchTxData, error)
	GeneratePolicyAndMultisig(
		validators *validatorobserver.ValidatorsPerChain,
		chainID string) (*cardano.ApexPolicyScripts, *cardano.ApexAddresses, error)
}

// ChainSpecificConfig defines the interface for chain-specific configurations
type ChainSpecificConfig interface {
	GetChainType() string
}
