package databaseaccess

import (
	"context"
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/relayer/core"
	"github.com/stretchr/testify/mock"
)

type DBMock struct {
	mock.Mock
}

var _ core.Database = (*DBMock)(nil)

func (d *DBMock) AddLastSubmittedBatchID(chainID string, batchID *big.Int) error {
	return d.Called(chainID, batchID).Error(0)
}

func (d *DBMock) Close() error {
	return nil
}

func (d *DBMock) GetLastSubmittedBatchID(chainID string) (*big.Int, error) {
	args := d.Called(chainID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	arg0, _ := args.Get(0).(*big.Int)

	return arg0, args.Error(1)
}

func (d *DBMock) Init(filePath string) error {
	return nil
}

type CardanoChainOperationsMock struct {
	mock.Mock
}

var _ core.ChainOperations = (*CardanoChainOperationsMock)(nil)

func (m *CardanoChainOperationsMock) SendTx(
	ctx context.Context, smartContractData *eth.ConfirmedBatch,
) error {
	args := m.Called(ctx, smartContractData)

	return args.Error(0)
}
