package core

import (
	"context"

	"github.com/Ethernal-Tech/apex-bridge/eth"
)

type GeneratedBatchTxData struct {
	BatchType           eth.BatchTypes
	IsStakeSignNeeded   bool
	IsPaymentSignNeeded bool
	TxRaw               []byte
	TxHash              string
}

func (gb GeneratedBatchTxData) IsConsolidation() bool {
	return gb.BatchType == eth.BatchTypeConsolidation
}

type BatchSignatures struct {
	Multisig, MultsigStake, Fee []byte
}

type BatcherManager interface {
	Start()
}

type Batcher interface {
	Start(ctx context.Context)
}

type ChainOperations interface {
	GenerateBatchTransaction(
		ctx context.Context, destinationChain string, confirmedTransactions []eth.ConfirmedTransaction, batchNonceID uint64,
	) (*GeneratedBatchTxData, error)
	SignBatchTransaction(generatedBatchData *GeneratedBatchTxData) (*BatchSignatures, error)
	IsSynchronized(
		ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract, chainID string,
	) (bool, error)
	Submit(ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract, batch eth.SignedBatch) error
}

// ChainSpecificConfig defines the interface for chain-specific configurations
type ChainSpecificConfig interface {
	GetChainType() string
}
