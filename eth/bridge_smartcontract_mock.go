package eth

import (
	"context"

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

	arg0, _ := args.Get(0).(*ConfirmedBatch)

	return arg0, args.Error(1)
}

func (m *BridgeSmartContractMock) SubmitSignedBatch(ctx context.Context, signedBatch SignedBatch) error {
	return m.Called(ctx, signedBatch).Error(0)
}

func (m *BridgeSmartContractMock) ShouldCreateBatch(ctx context.Context, destinationChain string) (bool, error) {
	args := m.Called(ctx, destinationChain)
	arg0, _ := args.Get(0).(bool)

	return arg0, args.Error(1)
}

func (m *BridgeSmartContractMock) GetAvailableUTXOs(ctx context.Context, destinationChain string) (UTXOs, error) {
	args := m.Called(ctx, destinationChain)

	if args.Get(0) == nil {
		return UTXOs{}, args.Error(1)
	}

	arg0, _ := args.Get(0).(UTXOs)

	return arg0, args.Error(1)
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

	arg0, _ := args.Get(0).([]ConfirmedTransaction)

	return arg0, args.Error(1)
}

func (m *BridgeSmartContractMock) GetLastObservedBlock(
	ctx context.Context, destinationChain string,
) (
	*CardanoBlock, error,
) {
	args := m.Called(ctx, destinationChain)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	arg0, _ := args.Get(0).(*CardanoBlock)

	return arg0, args.Error(1)
}

func (m *BridgeSmartContractMock) GetValidatorsCardanoData(
	ctx context.Context, destinationChain string,
) (
	[]ValidatorCardanoData, error,
) {
	args := m.Called(ctx, destinationChain)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	arg0, _ := args.Get(0).([]ValidatorCardanoData)

	return arg0, args.Error(1)
}

func (m *BridgeSmartContractMock) GetNextBatchID(ctx context.Context, destinationChain string) (uint64, error) {
	args := m.Called(ctx, destinationChain)

	arg0, _ := args.Get(0).(uint64)

	return arg0, args.Error(1)
}

func (m *BridgeSmartContractMock) GetAllRegisteredChains(ctx context.Context) ([]Chain, error) {
	args := m.Called(ctx)
	if args.Get(0) != nil {
		arg0, _ := args.Get(0).([]Chain)

		return arg0, args.Error(1)
	}

	return nil, args.Error(1)
}
