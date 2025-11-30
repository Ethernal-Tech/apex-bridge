package eth

import (
	"context"
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/mock"
)

type EVMGatewaySmartContractMock struct {
	mock.Mock
}

var _ IEVMGatewaySmartContract = (*EVMGatewaySmartContractMock)(nil)

func (m *EVMGatewaySmartContractMock) Deposit(
	ctx context.Context, signature []byte, bitmap *big.Int, data []byte,
) error {
	args := m.Called(ctx, signature, bitmap, data)

	return args.Error(0)
}

// RegisterToken implements IEVMGatewaySmartContract.
func (m *EVMGatewaySmartContractMock) RegisterToken(
	ctx context.Context, lockUnlockSCAddress common.Address,
	tokenID uint16, name string, symbol string,
) (*contractbinding.GatewayTokenRegistered, error) {
	args := m.Called(ctx, lockUnlockSCAddress, tokenID, name, symbol)
	if args.Get(0) != nil {
		arg0, _ := args.Get(0).(*contractbinding.GatewayTokenRegistered)

		return arg0, args.Error(1)
	}

	return nil, args.Error(1)
}
