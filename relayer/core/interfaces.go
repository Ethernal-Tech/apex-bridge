package core

import (
	"context"
	"math/big"

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
	SendTx(ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract, data *eth.ConfirmedBatch) error
}

// BalanceReporter is optionally implemented by ChainOperations that can report the
// balance of the account the relayer submits batches with.
type BalanceReporter interface {
	GetRelayerAddress() string
	GetRelayerBalance(ctx context.Context) (*big.Int, error)
}

type BatchIDDB interface {
	AddLastSubmittedBatchID(chainID string, batchID *big.Int) error
	GetLastSubmittedBatchID(chainID string) (*big.Int, error)
}
type Database interface {
	BatchIDDB
	Init(filePath string) error
	Close() error
}
