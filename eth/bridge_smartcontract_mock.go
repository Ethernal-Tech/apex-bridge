package eth

import (
	"context"
	"math/big"

	"github.com/stretchr/testify/mock"
)

type BridgeSmartContractMock struct {
	mock.Mock
}

var _ IBridgeSmartContract = (*BridgeSmartContractMock)(nil)

func (m *BridgeSmartContractMock) GetConfirmedBatch(
	ctx context.Context, destinationChain string) (*ConfirmedBatch, error) {
	args := m.Called(ctx, destinationChain)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*ConfirmedBatch), args.Error(1)
}

func (m *BridgeSmartContractMock) SubmitSignedBatch(ctx context.Context, signedBatch SignedBatch) error {
	return m.Called(ctx, signedBatch).Error(0)
}

func (m *BridgeSmartContractMock) ShouldCreateBatch(ctx context.Context, destinationChain string) (bool, error) {
	args := m.Called(ctx, destinationChain)
	return args.Get(0).(bool), args.Error(1)
}

func (m *BridgeSmartContractMock) GetAvailableUTXOs(ctx context.Context, destinationChain string) (*UTXOs, error) {
	args := m.Called(ctx, destinationChain)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*UTXOs), args.Error(1)
}

func (m *BridgeSmartContractMock) GetConfirmedTransactions(
	ctx context.Context, destinationChain string,
) (
	[]ConfirmedTransaction, error,
) {
	args := m.Called(ctx, destinationChain)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]ConfirmedTransaction), args.Error(1)
}

func (m *BridgeSmartContractMock) GetLastObservedBlock(ctx context.Context, destinationChain string) (*CardanoBlock, error) {
	args := m.Called(ctx, destinationChain)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*CardanoBlock), args.Error(1)
}

func (m *BridgeSmartContractMock) GetValidatorsCardanoData(ctx context.Context, destinationChain string) ([]ValidatorCardanoData, error) {
	args := m.Called(ctx, destinationChain)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]ValidatorCardanoData), args.Error(1)
}

func (m *BridgeSmartContractMock) GetNextBatchId(ctx context.Context, destinationChain string) (*big.Int, error) {
	args := m.Called(ctx, destinationChain)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *BridgeSmartContractMock) GetAllRegisteredChains(ctx context.Context) ([]Chain, error) {
	args := m.Called(ctx)
	if args.Get(0) != nil {
		return args.Get(0).([]Chain), args.Error(1)
	}

	return nil, args.Error(1)
}
