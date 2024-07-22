package eth

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type EVMGatewaySmartContractMock struct {
	mock.Mock
}

var _ IEVMGatewaySmartContract = (*EVMGatewaySmartContractMock)(nil)

func (m *EVMGatewaySmartContractMock) Deposit(
	ctx context.Context, signature []byte, bitmap []byte, data []byte,
) error {
	args := m.Called(ctx, signature, bitmap, data)

	return args.Error(0)
}
