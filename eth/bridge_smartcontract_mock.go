package eth

import (
	"context"
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
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
	panic("unimplemented")
}

func (m *BridgeSmartContractMock) GetAvailableUTXOs(ctx context.Context, destinationChain string, txCost *big.Int) (*contractbinding.IBridgeContractStructsUTXOs, error) {
	panic("unimplemented")
}

func (m *BridgeSmartContractMock) GetConfirmedTransactions(ctx context.Context, destinationChain string) ([]contractbinding.IBridgeContractStructsConfirmedTransaction, error) {
	panic("unimplemented")
}
