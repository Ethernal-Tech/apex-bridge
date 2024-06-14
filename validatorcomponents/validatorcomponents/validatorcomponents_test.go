package validatorcomponents

import (
	"context"
	"errors"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	oracleCore "github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_populateUtxosAndAddresses(t *testing.T) {
	getConfig := func() *oracleCore.AppConfig {
		return &oracleCore.AppConfig{
			CardanoChains: map[string]*oracleCore.CardanoChainConfig{
				common.ChainIDStrVector: {
					NetworkAddress: "http://vector.com",
				},
				common.ChainIDStrPrime: {
					NetworkAddress: "http://prime.com",
				},
				"dummy": {
					NetworkAddress: "http://dummy.com",
				},
			},
		}
	}

	t.Run("chain not in config", func(t *testing.T) {
		scMock := &eth.BridgeSmartContractMock{}
		scMock.On("GetAllRegisteredChains", mock.Anything).Return([]eth.Chain{
			{
				Id: common.ToNumChainID(common.ChainIDStrVector),
			},
			{
				Id: 0,
			},
		}, error(nil))
		scMock.On("GetAvailableUTXOs", mock.Anything, common.ChainIDStrVector).Return(&eth.UTXOs{
			MultisigOwnedUTXOs: []contractbinding.IBridgeStructsUTXO{},
			FeePayerOwnedUTXOs: []contractbinding.IBridgeStructsUTXO{},
		}, error(nil))

		err := populateUtxosAndAddresses(context.Background(), getConfig(), scMock, hclog.NewNullLogger())
		require.ErrorContains(t, err, "no config for registered chain")
	})

	t.Run("failed to retrieve available utxos once", func(t *testing.T) {
		scMock := &eth.BridgeSmartContractMock{}
		scMock.On("GetAllRegisteredChains", mock.Anything).Return([]eth.Chain{
			{
				Id: common.ToNumChainID(common.ChainIDStrVector),
			},
		}, error(nil))
		scMock.On("GetAvailableUTXOs", mock.Anything, common.ChainIDStrVector).Once().Return(nil, errors.New("er"))
		scMock.On("GetAvailableUTXOs", mock.Anything, common.ChainIDStrVector).Once().Return(&eth.UTXOs{
			MultisigOwnedUTXOs: []contractbinding.IBridgeStructsUTXO{},
			FeePayerOwnedUTXOs: []contractbinding.IBridgeStructsUTXO{},
		}, nil)

		err := populateUtxosAndAddresses(context.Background(), getConfig(), scMock, hclog.NewNullLogger())
		require.NoError(t, err)
	})

	t.Run("failed to retrieve registered chains once", func(t *testing.T) {
		scMock := &eth.BridgeSmartContractMock{}
		scMock.On("GetAllRegisteredChains", mock.Anything).Once().Return(nil, errors.New("er"))
		scMock.On("GetAllRegisteredChains", mock.Anything).Once().Return([]eth.Chain{
			{
				Id: common.ToNumChainID(common.ChainIDStrVector),
			},
			{
				Id: common.ToNumChainID(common.ChainIDStrPrime),
			},
		}, nil)
		scMock.On("GetAvailableUTXOs", mock.Anything, mock.Anything).Return(&eth.UTXOs{
			MultisigOwnedUTXOs: []contractbinding.IBridgeStructsUTXO{},
			FeePayerOwnedUTXOs: []contractbinding.IBridgeStructsUTXO{},
		}, nil)

		err := populateUtxosAndAddresses(context.Background(), getConfig(), scMock, hclog.NewNullLogger())
		require.NoError(t, err)
	})

	t.Run("happy path", func(t *testing.T) {
		const (
			multisigPrime  = "addr_1"
			multisigVector = "addr_2"
			feePayerPrime  = "addr_3"
			feePayerVector = "addr_3"
		)

		utxos := []contractbinding.IBridgeStructsUTXO{
			{
				TxHash:  common.MustHashToBytes32("0x01"),
				TxIndex: 2,
				Amount:  200,
			},
			{
				TxHash:  common.MustHashToBytes32("0x02"),
				TxIndex: 0,
				Amount:  100,
			},
			{
				TxHash:  common.MustHashToBytes32("0x03"),
				TxIndex: 129,
				Amount:  10,
			},
			{
				TxHash:  common.MustHashToBytes32("0x04"),
				TxIndex: 0,
				Amount:  1000,
			},
			{
				TxHash:  common.MustHashToBytes32("0x05"),
				TxIndex: 1,
				Amount:  1,
			},
			{
				TxHash:  common.MustHashToBytes32("0x06"),
				TxIndex: 2,
				Amount:  2,
			},
			{
				TxHash:  common.MustHashToBytes32("0x07"),
				TxIndex: 0,
				Amount:  100,
			},
		}

		scMock := &eth.BridgeSmartContractMock{}
		scMock.On("GetAllRegisteredChains", mock.Anything).Return([]eth.Chain{
			{
				Id:              common.ToNumChainID(common.ChainIDStrVector),
				AddressMultisig: multisigVector,
				AddressFeePayer: feePayerVector,
			},
			{
				Id:              common.ToNumChainID(common.ChainIDStrPrime),
				AddressMultisig: multisigPrime,
				AddressFeePayer: feePayerPrime,
			},
		}, error(nil))
		scMock.On("GetAvailableUTXOs", mock.Anything, common.ChainIDStrVector).Return(eth.UTXOs{
			MultisigOwnedUTXOs: utxos[0:1],
			FeePayerOwnedUTXOs: utxos[1:3],
		}, error(nil))
		scMock.On("GetAvailableUTXOs", mock.Anything, common.ChainIDStrPrime).Return(eth.UTXOs{
			MultisigOwnedUTXOs: utxos[3:6],
			FeePayerOwnedUTXOs: utxos[6:7],
		}, error(nil))

		config := getConfig()
		err := populateUtxosAndAddresses(context.Background(), config, scMock, hclog.NewNullLogger())
		require.NoError(t, err)

		require.Len(t, config.CardanoChains, 2)

		assert.Equal(t, multisigVector, config.CardanoChains[common.ChainIDStrVector].BridgingAddresses.BridgingAddress)
		assert.Equal(t, feePayerVector, config.CardanoChains[common.ChainIDStrVector].BridgingAddresses.FeeAddress)
		assert.Equal(t, multisigPrime, config.CardanoChains[common.ChainIDStrPrime].BridgingAddresses.BridgingAddress)
		assert.Equal(t, feePayerPrime, config.CardanoChains[common.ChainIDStrPrime].BridgingAddresses.FeeAddress)

		require.Len(t, config.CardanoChains[common.ChainIDStrVector].InitialUtxos, 3)
		require.Len(t, config.CardanoChains[common.ChainIDStrPrime].InitialUtxos, 4)

		for i, x := range utxos {
			if i < 3 {
				if i == 0 {
					assert.Equal(t, multisigVector, config.CardanoChains[common.ChainIDStrVector].InitialUtxos[i].Output.Address)
				} else {
					assert.Equal(t, feePayerVector, config.CardanoChains[common.ChainIDStrVector].InitialUtxos[i].Output.Address)
				}

				assert.Equal(t, x.Amount, config.CardanoChains[common.ChainIDStrVector].InitialUtxos[i].Output.Amount)
				assert.Equal(t, x.TxHash[:], config.CardanoChains[common.ChainIDStrVector].InitialUtxos[i].Input.Hash[:])
				assert.Equal(t, uint32(x.TxIndex), config.CardanoChains[common.ChainIDStrVector].InitialUtxos[i].Input.Index)
			} else {
				if i < 6 {
					assert.Equal(t, multisigPrime, config.CardanoChains[common.ChainIDStrPrime].InitialUtxos[i-3].Output.Address)
				} else {
					assert.Equal(t, feePayerPrime, config.CardanoChains[common.ChainIDStrPrime].InitialUtxos[i-3].Output.Address)
				}

				assert.Equal(t, x.Amount, config.CardanoChains[common.ChainIDStrPrime].InitialUtxos[i-3].Output.Amount)
				assert.Equal(t, x.TxHash[:], config.CardanoChains[common.ChainIDStrPrime].InitialUtxos[i-3].Input.Hash[:])
				assert.Equal(t, uint32(x.TxIndex), config.CardanoChains[common.ChainIDStrPrime].InitialUtxos[i-3].Input.Index)
			}
		}
	})
}
