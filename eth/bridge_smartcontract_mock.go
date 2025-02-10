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

func (m *BridgeSmartContractMock) SubmitSignedBatch(
	ctx context.Context, signedBatch SignedBatch, gasLimit uint64,
) error {
	return m.Called(ctx, signedBatch, gasLimit).Error(0)
}

func (m *BridgeSmartContractMock) SubmitSignedBatchEVM(
	ctx context.Context, signedBatch SignedBatch, gasLimit uint64,
) error {
	return m.Called(ctx, signedBatch, gasLimit).Error(0)
}

func (m *BridgeSmartContractMock) ShouldCreateBatch(ctx context.Context, destinationChain string) (bool, error) {
	args := m.Called(ctx, destinationChain)
	arg0, _ := args.Get(0).(bool)

	return arg0, args.Error(1)
}

func (m *BridgeSmartContractMock) GetConfirmedTransactions(
	ctx context.Context, destinationChain string,
) ([]ConfirmedTransaction, error) {
	args := m.Called(ctx, destinationChain)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	arg0, _ := args.Get(0).([]ConfirmedTransaction)

	return arg0, args.Error(1)
}

func (m *BridgeSmartContractMock) GetLastObservedBlock(
	ctx context.Context, destinationChain string,
) (CardanoBlock, error) {
	args := m.Called(ctx, destinationChain)

	if args.Get(0) == nil {
		return CardanoBlock{}, args.Error(1)
	}

	arg0, _ := args.Get(0).(CardanoBlock)

	return arg0, args.Error(1)
}

func (m *BridgeSmartContractMock) GetValidatorsChainData(
	ctx context.Context, destinationChain string,
) ([]ValidatorChainData, error) {
	args := m.Called(ctx, destinationChain)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	arg0, _ := args.Get(0).([]ValidatorChainData)

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

func (m *BridgeSmartContractMock) GetBlockNumber(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	if args.Get(0) != nil {
		arg0, _ := args.Get(0).(uint64)

		return arg0, args.Error(1)
	}

	return 0, args.Error(1)
}

func (m *BridgeSmartContractMock) SetChainAdditionalData(
	ctx context.Context, chainID, multisigAddr, feeAddr string,
) error {
	return m.Called(ctx, chainID, multisigAddr, feeAddr).Error(0)
}

// GetBatchTransactions implements IOracleBridgeSmartContract.
func (m *BridgeSmartContractMock) GetBatchTransactions(
	ctx context.Context, chainID string, batchID uint64,
) ([]TxDataInfo, error) {
	args := m.Called(ctx, chainID, batchID)
	result, _ := args.Get(0).([]TxDataInfo)
	return result, args.Error(1)
}
