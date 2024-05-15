package core

import (
	"context"
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/eth"
)

type GeneratedBatchTxData struct {
	TxRaw  []byte
	TxHash string
	Utxos  eth.UTXOs
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
		destinationChain string, confirmedTransactions []eth.ConfirmedTransaction, batchNonceID *big.Int,
	) (*GeneratedBatchTxData, error)
	SignBatchTransaction(txHash string) ([]byte, []byte, error)
}

// ChainSpecificConfig defines the interface for chain-specific configurations
type ChainSpecificConfig interface {
	GetChainType() string
}
