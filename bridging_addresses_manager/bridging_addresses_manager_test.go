package bridgingaddressmanager

import (
	"context"
	"math/big"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	oracleCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBridgingAddressesManager(t *testing.T) {
	cardanoChains := map[string]*oracleCore.CardanoChainConfig{
		common.ChainIDStrPrime: {
			ChainID: common.ChainIDStrPrime,
		},
	}

	bridgeSmartContractMock := &eth.BridgeSmartContractMock{}
	bridgeSmartContractMock.On("GetAllRegisteredChains", mock.Anything).Return([]eth.Chain{
		{
			Id:              1,
			ChainType:       0,
			AddressMultisig: "",
			AddressFeePayer: "",
		},
	}, nil)

	bridgeSmartContractMock.On("GetValidatorsChainData", mock.Anything, mock.Anything).Return([]contractbinding.IBridgeStructsValidatorChainData{
		{
			Key: [4]*big.Int{
				big.NewInt(1),
				big.NewInt(2),
				big.NewInt(3),
				big.NewInt(4),
			},
		},
	}, nil)

	bridgeSmartContractMock.On("GetBridgingAddressesCount", mock.Anything, "prime").Return(uint8(1), nil)

	chainIDConverter := common.NewChainIDConverterForTest()

	bridgingAddressesManager, err := NewBridgingAdressesManager(
		context.Background(),
		cardanoChains,
		chainIDConverter,
		bridgeSmartContractMock,
		hclog.NewNullLogger(),
	)

	require.NoError(t, err)
	require.NotNil(t, bridgingAddressesManager)

	expectedPaymentAddress := "addr1x8xqwh43y8auuzw9h0mdgwv4d7f4pn40klvpxjyejvmz807djgaeaq93dzfh6e5dvyhg0eqzmlce6mnml30hku2ds4nq9d70gz"
	expectedStakeAddress := "stake178xeywu7szck3ymav6xkzt58uspdluvadealchmmw9xc2esjp3n8j"
	expectedCustodialAddress := "addr1x8xqwh43y8auuzw9h0mdgwv4d7f4pn40klvpxjyejvmz807djgaeaq93dzfh6e5dvyhg0eqzmlce6mnml30hku2ds4nq9d70gz"

	t.Run("GetAllPaymentAddresses", func(t *testing.T) {
		paymentAddresses := bridgingAddressesManager.GetAllPaymentAddresses(1)
		require.Equal(t, paymentAddresses, []string{expectedPaymentAddress})
	})

	t.Run("GetAllStakeAddresses", func(t *testing.T) {
		stakeAddresses := bridgingAddressesManager.GetAllStakeAddresses(1)
		require.Equal(t, stakeAddresses, []string{expectedStakeAddress})
	})

	t.Run("GetPaymentAddressFromIndex", func(t *testing.T) {
		paymentAddress, ok := bridgingAddressesManager.GetPaymentAddressFromIndex(1, 0)
		require.True(t, ok)
		require.Equal(t, paymentAddress, expectedPaymentAddress)
	})

	t.Run("GetStakeAddressFromIndex", func(t *testing.T) {
		stakeAddress, ok := bridgingAddressesManager.GetStakeAddressFromIndex(1, 0)
		require.True(t, ok)
		require.Equal(t, stakeAddress, expectedStakeAddress)
	})

	t.Run("GetPaymentAddressIndex ok", func(t *testing.T) {
		paymentAddressIndex, ok := bridgingAddressesManager.GetPaymentAddressIndex(1, expectedPaymentAddress)
		require.True(t, ok)
		require.Equal(t, paymentAddressIndex, uint8(0))
	})

	t.Run("GetPaymentAddressIndex not ok", func(t *testing.T) {
		paymentAddressIndex, ok := bridgingAddressesManager.GetPaymentAddressIndex(1, "not_existing_address")
		require.False(t, ok)
		require.Equal(t, paymentAddressIndex, uint8(0))
	})

	t.Run("GetStakeAddressIndex ok", func(t *testing.T) {
		stakeAddressIndex, ok := bridgingAddressesManager.GetStakeAddressIndex(1, expectedStakeAddress)
		require.True(t, ok)
		require.Equal(t, stakeAddressIndex, uint8(0))
	})

	t.Run("GetStakeAddressIndex not ok", func(t *testing.T) {
		stakeAddressIndex, ok := bridgingAddressesManager.GetStakeAddressIndex(1, "not_existing_address")
		require.False(t, ok)
		require.Equal(t, stakeAddressIndex, uint8(0))
	})

	t.Run("GetPaymentAddressFromIndex not ok", func(t *testing.T) {
		paymentAddress, ok := bridgingAddressesManager.GetPaymentAddressFromIndex(1, 1)
		require.False(t, ok)
		require.Equal(t, paymentAddress, "")
	})

	t.Run("GetStakeAddressFromIndex not ok", func(t *testing.T) {
		stakeAddress, ok := bridgingAddressesManager.GetStakeAddressFromIndex(1, 1)
		require.False(t, ok)
		require.Equal(t, stakeAddress, "")
	})

	t.Run("GetAllPaymentAddresses not ok", func(t *testing.T) {
		paymentAddresses := bridgingAddressesManager.GetAllPaymentAddresses(2)
		require.Equal(t, paymentAddresses, []string(nil))
	})

	t.Run("GetAllStakeAddresses not ok", func(t *testing.T) {
		stakeAddresses := bridgingAddressesManager.GetAllStakeAddresses(2)
		require.Equal(t, stakeAddresses, []string(nil))
	})

	t.Run("GetCustodialAddress ok", func(t *testing.T) {
		custodialAddress, ok := bridgingAddressesManager.GetCustodialAddress(1)
		require.True(t, ok)
		require.Equal(t, custodialAddress, expectedCustodialAddress)
	})

	t.Run("GetCustodialAddress not ok", func(t *testing.T) {
		custodialAddress, ok := bridgingAddressesManager.GetCustodialAddress(2)
		require.False(t, ok)
		require.Equal(t, custodialAddress, "")
	})
}
