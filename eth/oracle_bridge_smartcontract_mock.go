package eth

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type OracleBridgeSmartContractMock struct {
	mock.Mock
}

// GetLastObservedBlock implements IOracleBridgeSmartContract.
func (m *OracleBridgeSmartContractMock) GetLastObservedBlock(ctx context.Context, sourceChain string) (*CardanoBlock, error) {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).(*CardanoBlock), args.Error(1)
	}

	return nil, args.Error(1)
}

// GetRawTransactionFromLastBatch implements IOracleBridgeSmartContract.
func (m *OracleBridgeSmartContractMock) GetRawTransactionFromLastBatch(ctx context.Context, chainID string) (*LastBatchRawTx, error) {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).(*LastBatchRawTx), args.Error(1)
	}

	return nil, args.Error(1)
}

// SubmitClaims implements IOracleBridgeSmartContract.
func (m *OracleBridgeSmartContractMock) SubmitClaims(ctx context.Context, claims Claims) error {
	args := m.Called()
	return args.Error(0)
}

// SubmitLastObservableBlocks implements IOracleBridgeSmartContract.
func (m *OracleBridgeSmartContractMock) SubmitLastObservableBlocks(ctx context.Context, chainID string, blocks []CardanoBlock) error {
	args := m.Called()
	return args.Error(0)
}

var _ IOracleBridgeSmartContract = (*OracleBridgeSmartContractMock)(nil)
