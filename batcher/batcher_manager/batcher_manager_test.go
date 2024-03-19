package batcher_manager

import (
	"encoding/json"
	"reflect"
	"testing"

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
			"prime1": {
				Base: core.BaseConfig{
					ChainId:               "prime1",
					SigningKeyMultiSig:    "58201825bce09711e1563fc1702587da6892d1d869894386323bd4378ea5e3d6cba0",
					SigningKeyMultiSigFee: "58204cd84bf321e70ab223fbdbfe5eba249a5249bd9becbeb82109d45e56c9c610a9",
				},
				ChainSpecific: core.ChainSpecific{
					ChainType: "Cardano",
					Config:    rawMessage,
				},
			},
			"prime2": {
				Base: core.BaseConfig{
					ChainId:               "prime2",
					SigningKeyMultiSig:    "5820ccdae0d1cd3fa9be16a497941acff33b9aa20bdbf2f9aa5715942d152988e083",
					SigningKeyMultiSigFee: "58208fcc8cac6b7fedf4c30aed170633df487642cb22f7e8615684e2b98e367fcaa3",
				},
				ChainSpecific: core.ChainSpecific{
					ChainType: "Cardano",
					Config:    rawMessage,
				},
			},
			"prime3": {
				Base: core.BaseConfig{
					ChainId:               "prime3",
					SigningKeyMultiSig:    "582094bfc7d65a5d936e7b527c93ea6bf75de51029290b1ef8c8877bffe070398b40",
					SigningKeyMultiSigFee: "582058fb35da120c65855ad691dadf5681a2e4fc62e9dcda0d0774ff6fdc463a679a",
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

	cardanoOperationsPrime, err := GetChainSpecificOperations(loadedConfig.Chains["prime1"].ChainSpecific)
	assert.NoError(t, err)

	operationsType := reflect.TypeOf(cardanoOperationsPrime)
	assert.NotNil(t, operationsType)
}
