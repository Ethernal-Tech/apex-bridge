package database_access

import (
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/relayer/core"
	"github.com/stretchr/testify/mock"
)

type DbMock struct {
	mock.Mock
}

var _ core.Database = (*DbMock)(nil)

func (d *DbMock) AddLastSubmittedBatchId(chainId string, batchId *big.Int) error {
	return d.Called(chainId, batchId).Error(0)
}

func (d *DbMock) Close() error {
	return nil
}

func (d *DbMock) GetLastSubmittedBatchId(chainId string) (*big.Int, error) {
	args := d.Called(chainId)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*big.Int), args.Error(1)
}

func (d *DbMock) Init(filePath string) error {
	return nil
}

type CardanoChainOperationsMock struct {
	mock.Mock
}

var _ core.ChainOperations = (*CardanoChainOperationsMock)(nil)

func (m *CardanoChainOperationsMock) SendTx(smartContractData *eth.ConfirmedBatch) error {
	args := m.Called(smartContractData)

	return args.Error(0)
}
