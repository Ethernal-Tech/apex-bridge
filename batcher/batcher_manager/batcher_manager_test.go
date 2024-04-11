package batcher_manager

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/batcher/batcher"
	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	cardano "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatcherManagerConfig(t *testing.T) {
	jsonData := []byte(`{
		"testnetMagic": 2,
		"blockfrostUrl": "https://cardano-preview.blockfrost.io/api/v0",
		"blockfrostApiKey": "preview7mGSjpyEKb24OxQ4cCxomxZ5axMs5PvE",
		"atLeastValidators": 0.6666666666666666,
		"potentialFee": 300000
		}`)

	rawMessage := json.RawMessage(jsonData)

	expectedConfig := &core.BatcherManagerConfiguration{
		Chains: map[string]core.ChainConfig{
			"prime": {
				Base: core.BaseConfig{
					ChainId:     "prime",
					KeysDirPath: "./keys/prime",
				},
				ChainSpecific: core.ChainSpecific{
					ChainType: "Cardano",
					Config:    rawMessage,
				},
			},
			"vector": {
				Base: core.BaseConfig{
					ChainId:     "vector",
					KeysDirPath: "./keys/vector",
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
			SigningKey:           "93c91e490bfd3736d17d04f53a10093e9cf2435309f4be1f5751381c8e201d23",
		},
		PullTimeMilis: 2500,
		Logger: logger.LoggerConfig{
			LogFilePath:   "./batcher_logs",
			LogLevel:      hclog.Debug,
			JSONLogFormat: false,
			AppendFile:    true,
		},
	}

	loadedConfig, err := common.LoadJson[core.BatcherManagerConfiguration]("../config.json")
	assert.NoError(t, err)

	assert.NotEmpty(t, loadedConfig.Chains)
	assert.Equal(t, expectedConfig.Bridge, loadedConfig.Bridge)
	assert.Equal(t, expectedConfig.PullTimeMilis, loadedConfig.PullTimeMilis)
	assert.Equal(t, expectedConfig.Logger, loadedConfig.Logger)
}

func TestBatcherManagerOperations(t *testing.T) {
	loadedConfig, err := common.LoadJson[core.BatcherManagerConfiguration]("../config.json")
	assert.NoError(t, err)

	for _, chain := range loadedConfig.Chains {
		testKeysPath := "../" + chain.Base.KeysDirPath[1:]

		chainOp, err := batcher.GetChainSpecificOperations(chain.ChainSpecific, testKeysPath)
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

			multisigAddress, multisigFeeAddress, err := walletLoadingHelper(testKeysPath)
			assert.NoError(t, err)

			// remove cbor prefix
			assert.Equal(t, multisigAddress[4:], hex.EncodeToString(concreteChainOp.CardanoWallet.MultiSig.GetSigningKey()))
			assert.Equal(t, multisigFeeAddress[4:], hex.EncodeToString(concreteChainOp.CardanoWallet.MultiSigFee.GetSigningKey()))

			// test signatures
			sigWithString, err := cardano.CreateTxWitness("b335adf170a3df72dfba3864a1d09eb87d3848c98aac54d58bce1d544d1a63ea", cardano.NewSigningKey(multisigAddress))
			assert.NoError(t, err)
			sigWithWallet, err := cardano.CreateTxWitness("b335adf170a3df72dfba3864a1d09eb87d3848c98aac54d58bce1d544d1a63ea", concreteChainOp.CardanoWallet.MultiSig)
			assert.NoError(t, err)
			assert.Equal(t, sigWithString, sigWithWallet)
		}
	}
}

func TestBatcherManagerCreation(t *testing.T) {
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
					KeysDirPath: "../keys/prime",
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
			SigningKey:           "93c91e490bfd3736d17d04f53a10093e9cf2435309f4be1f5751381c8e201d23",
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

func walletLoadingHelper(directory string) (string, string, error) {
	type FileContent struct {
		Type        string `json:"type"`
		Description string `json:"description"`
		CborHex     string `json:"cborHex"`
	}

	var multisigAddress, multisigFeeAddress string

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(info.Name(), "payment.skey") {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			var fileContent FileContent
			if err := json.Unmarshal(content, &fileContent); err != nil {
				return err
			}

			if strings.Contains(path, "multisig/") {
				multisigAddress = fileContent.CborHex
			} else if strings.Contains(path, "multisigfee/") {
				multisigFeeAddress = fileContent.CborHex
			}
		}

		return nil
	})
	if err != nil {
		return "", "", err
	}

	if multisigAddress == "" || multisigFeeAddress == "" {
		return "", "", fmt.Errorf("payment.skey files not found in both multisig and multisigfee directories")
	}

	return multisigAddress, multisigFeeAddress, nil
}
