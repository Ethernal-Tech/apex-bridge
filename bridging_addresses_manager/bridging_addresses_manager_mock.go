package bridgingaddressmanager

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/stretchr/testify/mock"
)

type BridgingAddressesManagerMock struct {
	mock.Mock
}

// Ensure interface compliance
var _ common.BridgingAddressesManager = (*BridgingAddressesManagerMock)(nil)

func (m *BridgingAddressesManagerMock) GetAllPaymentAddresses(chainID uint8) []string {
	args := m.Called(chainID)

	arg0, _ := args.Get(0).([]string)

	return arg0
}

func (m *BridgingAddressesManagerMock) GetAllStakeAddresses(chainID uint8) []string {
	args := m.Called(chainID)

	arg0, _ := args.Get(0).([]string)

	return arg0
}

func (m *BridgingAddressesManagerMock) GetPaymentAddressIndex(chainID uint8, address string) (uint8, bool) {
	args := m.Called(chainID, address)

	arg0, _ := args.Get(0).(uint8)

	return arg0, args.Bool(1)
}

func (m *BridgingAddressesManagerMock) GetStakeAddressIndex(chainID uint8, address string) (uint8, bool) {
	args := m.Called(chainID, address)

	arg0, _ := args.Get(0).(uint8)

	return arg0, args.Bool(1)
}

func (m *BridgingAddressesManagerMock) GetPaymentAddressFromIndex(chainID uint8, index uint8) (string, bool) {
	args := m.Called(chainID, index)

	return args.String(0), args.Bool(1)
}

func (m *BridgingAddressesManagerMock) GetStakeAddressFromIndex(chainID uint8, index uint8) (string, bool) {
	args := m.Called(chainID, index)

	return args.String(0), args.Bool(1)
}

func (m *BridgingAddressesManagerMock) GetFirstIndexAddress(chainID uint8) (string, bool) {
	args := m.Called(chainID)

	return args.String(0), args.Bool(1)
}

func (m *BridgingAddressesManagerMock) GetFirstIndex() uint8 {
	args := m.Called()

	return args.Get(0).(uint8)
}

func (m *BridgingAddressesManagerMock) GetPaymentPolicyScript(chainID uint8, index uint8) (*wallet.PolicyScript, bool) {
	args := m.Called(chainID, index)
	ps, _ := args.Get(0).(*wallet.PolicyScript)

	return ps, args.Bool(1)
}

func (m *BridgingAddressesManagerMock) GetStakePolicyScript(chainID uint8, index uint8) (*wallet.PolicyScript, bool) {
	args := m.Called(chainID, index)
	ps, _ := args.Get(0).(*wallet.PolicyScript)

	return ps, args.Bool(1)
}

func (m *BridgingAddressesManagerMock) GetFeeMultisigAddress(chainID uint8) string {
	args := m.Called(chainID)

	return args.String(0)
}

func (m *BridgingAddressesManagerMock) GetFeeMultisigPolicyScript(chainID uint8) (*wallet.PolicyScript, bool) {
	args := m.Called(chainID)
	ps, _ := args.Get(0).(*wallet.PolicyScript)

	return ps, args.Bool(1)
}
