package core

import (
	"context"

	"github.com/Ethernal-Tech/apex-bridge/eth"
)

type BatcherManager interface {
	Start() error
	Stop() error
}

type Batcher interface {
	Start(ctx context.Context)
}

type ChainOperations interface {
	GenerateBatchTransaction(ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract, destinationChain string, confirmedTransactions []eth.ConfirmedTransaction) ([]byte, string, *eth.UTXOs, error)
	SignBatchTransaction(txHash string) ([]byte, []byte, error)
}

// ChainSpecificConfig defines the interface for chain-specific configurations
type ChainSpecificConfig interface {
	GetChainType() string
}
