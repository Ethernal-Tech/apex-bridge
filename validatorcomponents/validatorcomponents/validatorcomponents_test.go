package validatorcomponents

import (
	"context"
	"errors"
	"math/big"
	"testing"

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
				"vector": {
					NetworkAddress: "http://vector.com",
				},
				"prime": {
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
				Id: "vector",
			},
			{
				Id: "non_existing",
			},
		}, error(nil))
		scMock.On("GetAvailableUTXOs", mock.Anything, "vector").Return(&eth.UTXOs{
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
				Id: "vector",
			},
		}, error(nil))
		scMock.On("GetAvailableUTXOs", mock.Anything, "vector").Once().Return(nil, errors.New("er"))
		scMock.On("GetAvailableUTXOs", mock.Anything, "vector").Once().Return(&eth.UTXOs{
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
				Id: "vector",
			},
			{
				Id: "prime",
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
				TxHash:  "0x001",
				TxIndex: new(big.Int).SetUint64(2),
				Amount:  new(big.Int).SetUint64(200),
			},
			{
				TxHash:  "0x002",
				TxIndex: new(big.Int).SetUint64(0),
				Amount:  new(big.Int).SetUint64(100),
			},
			{
				TxHash:  "0x003",
				TxIndex: new(big.Int).SetUint64(129),
				Amount:  new(big.Int).SetUint64(10),
			},
			{
				TxHash:  "0x004",
				TxIndex: new(big.Int).SetUint64(0),
				Amount:  new(big.Int).SetUint64(1000),
			},
			{
				TxHash:  "0x005",
				TxIndex: new(big.Int).SetUint64(1),
				Amount:  new(big.Int).SetUint64(1),
			},
			{
				TxHash:  "0x006",
				TxIndex: new(big.Int).SetUint64(2),
				Amount:  new(big.Int).SetUint64(2),
			},
			{
				TxHash:  "0x007",
				TxIndex: new(big.Int).SetUint64(0),
				Amount:  new(big.Int).SetUint64(100),
			},
		}

		scMock := &eth.BridgeSmartContractMock{}
		scMock.On("GetAllRegisteredChains", mock.Anything).Return([]eth.Chain{
			{
				Id:              "vector",
				AddressMultisig: multisigVector,
				AddressFeePayer: feePayerVector,
			},
			{
				Id:              "prime",
				AddressMultisig: multisigPrime,
				AddressFeePayer: feePayerPrime,
			},
		}, error(nil))
		scMock.On("GetAvailableUTXOs", mock.Anything, "vector").Return(eth.UTXOs{
			MultisigOwnedUTXOs: utxos[0:1],
			FeePayerOwnedUTXOs: utxos[1:3],
		}, error(nil))
		scMock.On("GetAvailableUTXOs", mock.Anything, "prime").Return(eth.UTXOs{
			MultisigOwnedUTXOs: utxos[3:6],
			FeePayerOwnedUTXOs: utxos[6:7],
		}, error(nil))

		config := getConfig()
		err := populateUtxosAndAddresses(context.Background(), config, scMock, hclog.NewNullLogger())
		require.NoError(t, err)

		require.Len(t, config.CardanoChains, 2)

		assert.Equal(t, multisigVector, config.CardanoChains["vector"].BridgingAddresses.BridgingAddress)
		assert.Equal(t, feePayerVector, config.CardanoChains["vector"].BridgingAddresses.FeeAddress)
		assert.Equal(t, multisigPrime, config.CardanoChains["prime"].BridgingAddresses.BridgingAddress)
		assert.Equal(t, feePayerPrime, config.CardanoChains["prime"].BridgingAddresses.FeeAddress)

		require.Len(t, config.CardanoChains["vector"].InitialUtxos, 3)
		require.Len(t, config.CardanoChains["prime"].InitialUtxos, 4)

		for i, x := range utxos {
			if i < 3 {
				if i == 0 {
					assert.Equal(t, multisigVector, config.CardanoChains["vector"].InitialUtxos[i].Output.Address)
				} else {
					assert.Equal(t, feePayerVector, config.CardanoChains["vector"].InitialUtxos[i].Output.Address)
				}

				assert.Equal(t, x.Amount.Uint64(), config.CardanoChains["vector"].InitialUtxos[i].Output.Amount)
				assert.Equal(t, x.TxHash, config.CardanoChains["vector"].InitialUtxos[i].Input.Hash)
				assert.Equal(t, uint32(x.TxIndex.Uint64()), config.CardanoChains["vector"].InitialUtxos[i].Input.Index)
			} else {
				if i < 6 {
					assert.Equal(t, multisigPrime, config.CardanoChains["prime"].InitialUtxos[i-3].Output.Address)
				} else {
					assert.Equal(t, feePayerPrime, config.CardanoChains["prime"].InitialUtxos[i-3].Output.Address)
				}

				assert.Equal(t, x.Amount.Uint64(), config.CardanoChains["prime"].InitialUtxos[i-3].Output.Amount)
				assert.Equal(t, x.TxHash, config.CardanoChains["prime"].InitialUtxos[i-3].Input.Hash)
				assert.Equal(t, uint32(x.TxIndex.Uint64()), config.CardanoChains["prime"].InitialUtxos[i-3].Input.Index)
			}
		}
	})
}
