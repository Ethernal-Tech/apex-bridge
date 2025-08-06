package bridgingaddressscoordinator

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/stretchr/testify/mock"
)

type BridgingAddressesCoordinatorMock struct {
	mock.Mock
}

// Ensure interface compliance
var _ common.BridgingAddressesCoordinator = (*BridgingAddressesCoordinatorMock)(nil)

func (m *BridgingAddressesCoordinatorMock) GetAddressesAndAmountsToPayFrom(
	chainID uint8,
	cardanoCliBinary string,
	protocolParams []byte,
	txOutputs []cardanowallet.TxOutput,
) ([]common.AddressAndAmount, error) {
	args := m.Called(chainID, cardanoCliBinary, protocolParams, txOutputs)

	arg0, _ := args.Get(0).([]common.AddressAndAmount)

	return arg0, args.Error(1)
}

func (m *BridgingAddressesCoordinatorMock) GetAddressesAndAmountsToStakeTo(
	chainID uint8, amount uint64,
) (common.AddressAndAmount, error) {
	args := m.Called(chainID, amount)

	arg0, _ := args.Get(0).(common.AddressAndAmount)

	return arg0, args.Error(1)
}
