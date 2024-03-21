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
	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
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

	loadedConfig, err := LoadConfig("../config.json")
	assert.NoError(t, err)

	assert.NotEmpty(t, loadedConfig.Chains)
	assert.Equal(t, expectedConfig.Bridge, loadedConfig.Bridge)
	assert.Equal(t, expectedConfig.PullTimeMilis, loadedConfig.PullTimeMilis)
	assert.Equal(t, expectedConfig.Logger, loadedConfig.Logger)
}

func TestBatcherManagerOperations(t *testing.T) {
	loadedConfig, err := LoadConfig("../config.json")
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
		}
	}
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
