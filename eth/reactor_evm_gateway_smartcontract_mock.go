package eth

import (
	"context"
	"math/big"

	"github.com/stretchr/testify/mock"
)

type ReactorEVMGatewaySmartContractMock struct {
	mock.Mock
}

var _ IEVMGatewaySmartContract = (*ReactorEVMGatewaySmartContractMock)(nil)

func (m *ReactorEVMGatewaySmartContractMock) Deposit(
	ctx context.Context, signature []byte, bitmap *big.Int, data []byte,
) error {
	args := m.Called(ctx, signature, bitmap, data)

	return args.Error(0)
}
