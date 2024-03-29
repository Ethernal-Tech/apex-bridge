package core

import (
	"context"

	"github.com/Ethernal-Tech/apex-bridge/eth"
)

type RelayerManager interface {
	Start() error
	Stop() error
}

type Relayer interface {
	Start(ctx context.Context)
}

type ChainOperations interface {
	SendTx(smartContractData *eth.ConfirmedBatch) error
}

// ChainSpecificConfig defines the interface for chain-specific configurations
type ChainSpecificConfig interface {
	GetChainType() string
}
