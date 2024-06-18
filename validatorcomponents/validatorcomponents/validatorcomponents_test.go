package validatorcomponents

import (
	"context"
	"errors"
	"testing"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	oracleCore "github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
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

		txProviderMock := &cardanotx.TxProviderTestMock{}
		txProviderMock.On("GetUtxos", mock.Anything, mock.Anything).Return([]cardanowallet.Utxo{}, nil)

		txProviders := map[string]cardanowallet.ITxProvider{
			common.ChainIDStrVector: txProviderMock,
		}

		err := populateUtxosAndAddresses(context.Background(), getConfig(), scMock, txProviders, hclog.NewNullLogger())
		require.ErrorContains(t, err, "no config for registered chain")
	})

	t.Run("failed to retrieve available utxos once", func(t *testing.T) {
		scMock := &eth.BridgeSmartContractMock{}
		scMock.On("GetAllRegisteredChains", mock.Anything).Return([]eth.Chain{
			{
				Id: common.ToNumChainID(common.ChainIDStrVector),
			},
		}, error(nil))

		txProviderMock := &cardanotx.TxProviderTestMock{}
		txProviderMock.On("GetUtxos", mock.Anything, mock.Anything).Return(nil, errors.New("error")).Once()
		txProviderMock.On("GetUtxos", mock.Anything, mock.Anything).Return([]cardanowallet.Utxo{}, nil)

		txProviders := map[string]cardanowallet.ITxProvider{
			common.ChainIDStrVector: txProviderMock,
		}

		err := populateUtxosAndAddresses(context.Background(), getConfig(), scMock, txProviders, hclog.NewNullLogger())
		require.NoError(t, err)
	})

	t.Run("failed to retrieve registered chains once", func(t *testing.T) {
		scMock := &eth.BridgeSmartContractMock{}
		scMock.On("GetAllRegisteredChains", mock.Anything).Once().Return(nil, errors.New("er")).Once()
		scMock.On("GetAllRegisteredChains", mock.Anything).Once().Return([]eth.Chain{
			{
				Id: common.ToNumChainID(common.ChainIDStrVector),
			},
			{
				Id: common.ToNumChainID(common.ChainIDStrPrime),
			},
		}, nil)

		txProviderMock := &cardanotx.TxProviderTestMock{}
		txProviderMock.On("GetUtxos", mock.Anything, mock.Anything).Return([]cardanowallet.Utxo{}, nil)

		txProviders := map[string]cardanowallet.ITxProvider{
			common.ChainIDStrVector: txProviderMock,
			common.ChainIDStrPrime:  txProviderMock,
		}

		err := populateUtxosAndAddresses(context.Background(), getConfig(), scMock, txProviders, hclog.NewNullLogger())
		require.NoError(t, err)
	})

	t.Run("happy path", func(t *testing.T) {
		const (
			multisigPrime  = "addr_1"
			multisigVector = "addr_2"
			feePayerPrime  = "addr_3"
			feePayerVector = "addr_4"
		)

		utxos := []cardanowallet.Utxo{
			{
				Hash:   "0x01",
				Index:  2,
				Amount: 200,
			},
			{
				Hash:   "0x02",
				Index:  0,
				Amount: 100,
			},
			{
				Hash:   "0x03",
				Index:  129,
				Amount: 10,
			},
			{
				Hash:   "0x04",
				Index:  0,
				Amount: 1000,
			},
			{
				Hash:   "0x05",
				Index:  1,
				Amount: 1,
			},
			{
				Hash:   "0x06",
				Index:  2,
				Amount: 2,
			},
			{
				Hash:   "0x07",
				Index:  0,
				Amount: 100,
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

		txProviderMock := &cardanotx.TxProviderTestMock{}
		txProviderMock.On("GetUtxos", mock.Anything, multisigVector).Return(utxos[0:1], error(nil))
		txProviderMock.On("GetUtxos", mock.Anything, feePayerVector).Return(utxos[1:3], error(nil))
		txProviderMock.On("GetUtxos", mock.Anything, multisigPrime).Return(utxos[3:6], error(nil))
		txProviderMock.On("GetUtxos", mock.Anything, feePayerPrime).Return(utxos[6:7], error(nil))

		txProviders := map[string]cardanowallet.ITxProvider{
			common.ChainIDStrVector: txProviderMock,
			common.ChainIDStrPrime:  txProviderMock,
		}

		config := getConfig()
		err := populateUtxosAndAddresses(context.Background(), config, scMock, txProviders, hclog.NewNullLogger())
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
				assert.Equal(t, indexer.NewHashFromHexString(x.Hash),
					config.CardanoChains[common.ChainIDStrVector].InitialUtxos[i].Input.Hash)
				assert.Equal(t, x.Index, config.CardanoChains[common.ChainIDStrVector].InitialUtxos[i].Input.Index)
			} else {
				if i < 6 {
					assert.Equal(t, multisigPrime, config.CardanoChains[common.ChainIDStrPrime].InitialUtxos[i-3].Output.Address)
				} else {
					assert.Equal(t, feePayerPrime, config.CardanoChains[common.ChainIDStrPrime].InitialUtxos[i-3].Output.Address)
				}

				assert.Equal(t, x.Amount, config.CardanoChains[common.ChainIDStrPrime].InitialUtxos[i-3].Output.Amount)
				assert.Equal(t, indexer.NewHashFromHexString(x.Hash),
					config.CardanoChains[common.ChainIDStrPrime].InitialUtxos[i-3].Input.Hash)
				assert.Equal(t, x.Index, config.CardanoChains[common.ChainIDStrPrime].InitialUtxos[i-3].Input.Index)
			}
		}
	})
}
