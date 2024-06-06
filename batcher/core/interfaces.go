package core

import (
	"context"

	"github.com/Ethernal-Tech/apex-bridge/eth"
)

type GeneratedBatchTxData struct {
	TxRaw  []byte
	TxHash string
	Utxos  eth.UTXOs
	Slot   uint64
}

type BatcherManager interface {
	Start()
}

type Batcher interface {
	Start(ctx context.Context)
}

type ChainOperations interface {
	GenerateBatchTransaction(
		ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract,
		destinationChain string, confirmedTransactions []eth.ConfirmedTransaction, batchNonceID uint64,
	) (*GeneratedBatchTxData, error)
	SignBatchTransaction(txHash string) ([]byte, []byte, error)
}

// ChainSpecificConfig defines the interface for chain-specific configurations
type ChainSpecificConfig interface {
	GetChainType() string
}
