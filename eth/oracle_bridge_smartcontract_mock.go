package eth

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type OracleBridgeSmartContractMock struct {
	mock.Mock
}

// GetLastObservedBlock implements IOracleBridgeSmartContract.
func (m *OracleBridgeSmartContractMock) GetLastObservedBlock(
	ctx context.Context, sourceChain string,
) (
	*CardanoBlock, error,
) {
	args := m.Called()
	if args.Get(0) != nil {
		arg0, _ := args.Get(0).(*CardanoBlock)

		return arg0, args.Error(1)
	}

	return nil, args.Error(1)
}

// GetRawTransactionFromLastBatch implements IOracleBridgeSmartContract.
func (m *OracleBridgeSmartContractMock) GetRawTransactionFromLastBatch(
	ctx context.Context, chainID string,
) (
	*LastBatchRawTx, error,
) {
	args := m.Called()
	if args.Get(0) != nil {
		arg0, _ := args.Get(0).(*LastBatchRawTx)

		return arg0, args.Error(1)
	}

	return nil, args.Error(1)
}

// SubmitClaims implements IOracleBridgeSmartContract.
func (m *OracleBridgeSmartContractMock) SubmitClaims(ctx context.Context, claims Claims) error {
	args := m.Called()

	return args.Error(0)
}

// SubmitLastObservedBlocks implements IOracleBridgeSmartContract.
func (m *OracleBridgeSmartContractMock) SubmitLastObservedBlocks(
	ctx context.Context, chainID string, blocks []CardanoBlock,
) error {
	args := m.Called()

	return args.Error(0)
}

var _ IOracleBridgeSmartContract = (*OracleBridgeSmartContractMock)(nil)
