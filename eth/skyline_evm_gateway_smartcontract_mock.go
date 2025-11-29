package eth

import (
	"context"
	"math/big"

	skylinegatewaycontractbinding "github.com/Ethernal-Tech/apex-bridge/contractbinding/gateway/skyline"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/mock"
)

type SkylineEVMGatewaySmartContractMock struct {
	mock.Mock
}

var _ ISkylineEVMGatewaySmartContract = (*SkylineEVMGatewaySmartContractMock)(nil)

func (m *SkylineEVMGatewaySmartContractMock) Deposit(
	ctx context.Context, signature []byte, bitmap *big.Int, data []byte,
) error {
	args := m.Called(ctx, signature, bitmap, data)

	return args.Error(0)
}

// RegisterToken implements IEVMGatewaySmartContract.
func (m *SkylineEVMGatewaySmartContractMock) RegisterToken(
	ctx context.Context, lockUnlockSCAddress common.Address,
	tokenID uint16, name string, symbol string,
) (*skylinegatewaycontractbinding.GatewayTokenRegistered, error) {
	args := m.Called(ctx, lockUnlockSCAddress, tokenID, name, symbol)
	if args.Get(0) != nil {
		arg0, _ := args.Get(0).(*skylinegatewaycontractbinding.GatewayTokenRegistered)

		return arg0, args.Error(1)
	}

	return nil, args.Error(1)
}
