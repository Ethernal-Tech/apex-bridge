package core

import (
	"context"

	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
)

type BatcherManager interface {
	Start() error
	Stop() error
}

type Batcher interface {
	Start(ctx context.Context)
}

type ChainOperations interface {
	GenerateBatchTransaction(ctx context.Context, ethTxHelper ethtxhelper.IEthTxHelper, smartContractAddress string, destinationChain string, confirmedTransactions []contractbinding.ConfirmedTransaction) ([]byte, string, *contractbinding.UTXOs, error)
	SignBatchTransaction(txHash string, signingKey string, signingKeyFee string) ([]byte, []byte, error)
}

// ChainSpecificConfig defines the interface for chain-specific configurations
type ChainSpecificConfig interface {
	GetChainType() string
}
