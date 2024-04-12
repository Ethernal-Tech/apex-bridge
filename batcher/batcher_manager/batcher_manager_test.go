package batcher_manager

import (
	"encoding/json"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/batcher/batcher"
	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	cardano "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/cardano-infrastructure/common"
	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatcherManagerOperations(t *testing.T) {
	testDir, err := os.MkdirTemp("", "cardano-prime")
	require.NoError(t, err)

	defer func() {
		os.RemoveAll(testDir)
		os.Remove(testDir)
	}()

	jsonData := []byte(`{
		"testnetMagic": 2,
		"blockfrostUrl": "https://cardano-preview.blockfrost.io/api/v0",
		"blockfrostApiKey": "preview7mGSjpyEKb24OxQ4cCxomxZ5axMs5PvE",
		"atLeastValidators": 0.6666666666666666,
		"potentialFee": 300000
		}`)

	rawMessage := json.RawMessage(jsonData)

	config := &core.BatcherManagerConfiguration{
		Chains: map[string]core.ChainConfig{
			"prime": {
				Base: core.BaseConfig{
					ChainId:     "prime",
					KeysDirPath: testDir,
				},
				ChainSpecific: core.ChainSpecific{
					ChainType: "Cardano",
					Config:    rawMessage,
				},
			},
		},
		Bridge: core.BridgeConfig{
			NodeUrl:              "https://polygon-mumbai-pokt.nodies.app", // will be our node,
			SmartContractAddress: "0x816402271eE6D9078Fc8Cb537aDBDD58219485BA",
		},
		PullTimeMilis: 2500,
	}

	for _, chain := range config.Chains {
		wallet, err := cardano.GenerateWallet(testDir, false, true)
		require.NoError(t, err)

		chainOp, err := batcher.GetChainSpecificOperations(chain.ChainSpecific, testDir)
		assert.NoError(t, err)

		operationsType := reflect.TypeOf(chainOp)
		assert.NotNil(t, operationsType)

		// check keys
		concreteChainOp, ok := chainOp.(*batcher.CardanoChainOperations)
		if ok {
			// check config
			cardanoChainConfig, err := core.ToCardanoChainConfig(chain.ChainSpecific)
			assert.NoError(t, err)
			assert.Equal(t, cardanoChainConfig, &concreteChainOp.Config)

			// remove cbor prefix
			assert.Equal(t, wallet.MultiSig.GetSigningKey(), concreteChainOp.CardanoWallet.MultiSig.GetSigningKey())
			assert.Equal(t, wallet.MultiSigFee.GetSigningKey(), concreteChainOp.CardanoWallet.MultiSigFee.GetSigningKey())

			// test signatures
			sigWithString, err := cardano.CreateTxWitness("b335adf170a3df72dfba3864a1d09eb87d3848c98aac54d58bce1d544d1a63ea", wallet.MultiSig)
			assert.NoError(t, err)
			sigWithWallet, err := cardano.CreateTxWitness("b335adf170a3df72dfba3864a1d09eb87d3848c98aac54d58bce1d544d1a63ea", concreteChainOp.CardanoWallet.MultiSig)
			assert.NoError(t, err)
			assert.Equal(t, sigWithString, sigWithWallet)
		}
	}
}

func TestBatcherManagerCreation(t *testing.T) {
	testDir, err := os.MkdirTemp("", "cardano-prime")
	require.NoError(t, err)

	defer func() {
		os.RemoveAll(testDir)
		os.Remove(testDir)
	}()

	_, err = cardano.GenerateWallet(testDir, false, true)
	require.NoError(t, err)

	ecdsaValidatoSecretDirPath := path.Join(testDir, secrets.ConsensusFolderLocal)
	require.NoError(t, common.CreateDirSafe(ecdsaValidatoSecretDirPath, 0770))

	ecdsaValidatoSecretFilePath := path.Join(ecdsaValidatoSecretDirPath, secrets.ValidatorKeyLocal)
	require.NoError(t, os.WriteFile(ecdsaValidatoSecretFilePath, []byte(
		"6a9d5cf2d80878afcd6c268fc4972f23eab59ac258435d8c9ac5790b5e15da6d",
	), 0770))

	jsonData := []byte(`{
		"testnetMagic": 2,
		"blockfrostUrl": "https://cardano-preview.blockfrost.io/api/v0",
		"blockfrostApiKey": "preview7mGSjpyEKb24OxQ4cCxomxZ5axMs5PvE",
		"atLeastValidators": 0.6666666666666666,
		"potentialFee": 300000
		}`)

	rawMessage := json.RawMessage(jsonData)

	config := &core.BatcherManagerConfiguration{
		Chains: map[string]core.ChainConfig{
			"prime": {
				Base: core.BaseConfig{
					ChainId:     "prime",
					KeysDirPath: testDir,
				},
				ChainSpecific: core.ChainSpecific{
					ChainType: "Cardano",
					Config:    rawMessage,
				},
			},
		},
		Bridge: core.BridgeConfig{
			NodeUrl:              "https://polygon-mumbai-pokt.nodies.app", // will be our node,
			SmartContractAddress: "0x816402271eE6D9078Fc8Cb537aDBDD58219485BA",
			SecretsManager: &secrets.SecretsManagerConfig{
				Type: "local",
				Path: testDir,
			},
		},
		PullTimeMilis: 2500,
	}

	t.Run("creation fails - invalid operations", func(t *testing.T) {
		invalidConfig := &core.BatcherManagerConfiguration{
			Chains: map[string]core.ChainConfig{
				"prime": {
					Base: core.BaseConfig{
						ChainId: "prime",
					},
					ChainSpecific: core.ChainSpecific{
						ChainType: "Cardano",
						Config:    json.RawMessage(""),
					},
				},
			},
		}

		manager := NewBatcherManager(invalidConfig, make(map[string]core.ChainOperations))
		require.Nil(t, manager)
	})

	t.Run("create manager without mocks", func(t *testing.T) {
		manager := NewBatcherManager(config, make(map[string]core.ChainOperations))
		require.NotNil(t, manager)
	})

	t.Run("create manager with chain operations mock", func(t *testing.T) {
		manager := NewBatcherManager(config, map[string]core.ChainOperations{"prime": nil})
		require.NotNil(t, manager)
	})

	t.Run("create manager with chain operations and bridge mock", func(t *testing.T) {
		manager := NewBatcherManager(config, map[string]core.ChainOperations{"prime": nil}, nil)
		require.NotNil(t, manager)
	})
}
