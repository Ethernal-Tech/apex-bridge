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
	SendTx(smartContractData *eth.ConfirmedBatch) error
}

type BatchIdDb interface {
	AddLastSubmittedBatchId(chainId string, batchId *big.Int) error
	GetLastSubmittedBatchId(chainId string) (*big.Int, error)
}
type Database interface {
	BatchIdDb
	Init(filePath string) error
	Close() error
}
