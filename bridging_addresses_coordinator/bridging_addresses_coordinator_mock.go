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

func (m *BridgingAddressesCoordinatorMock) GetAddressesAndAmountsForBatch(
	chainID uint8,
	cardanoCliBinary string,
	isRedistribution bool,
	protocolParams []byte,
	txOutputs common.TxOutputs,
	tokens []cardanowallet.MintTokenAmount,
) ([]common.AddressAndAmount, bool, error) {
	args := m.Called(chainID, cardanoCliBinary, isRedistribution, protocolParams, txOutputs, tokens)

	arg0, _ := args.Get(0).([]common.AddressAndAmount)
	arg1, _ := args.Get(0).(bool)

	return arg0, arg1, args.Error(2)
}

func (m *BridgingAddressesCoordinatorMock) GetAddressToBridgeTo(
	chainID uint8,
	contansNativeTokens bool,
) (common.AddressAndAmount, error) {
	args := m.Called(chainID)

	arg0, _ := args.Get(0).(common.AddressAndAmount)

	return arg0, args.Error(1)
}
