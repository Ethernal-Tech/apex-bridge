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
					ChainId:               "prime",
					SigningKeyMultiSig:    "58201217236ac24d8ac12684b308cf9468f68ef5283096896dc1c5c3caf8351e2847",
					SigningKeyMultiSigFee: "5820f2c3b9527ec2f0d70e6ee2db5752e27066fe63f5c84d1aa5bf20a5fc4d2411e6",
				},
				ChainSpecific: core.ChainSpecific{
					ChainType: "Cardano",
					Config:    rawMessage,
				},
			},
			"vector": {
				Base: core.BaseConfig{
					ChainId:               "vector",
					SigningKeyMultiSig:    "58201217236ac24d8ac12684b308cf9468f68ef5283096896dc1c5c3caf8351e2847",
					SigningKeyMultiSigFee: "5820f2c3b9527ec2f0d70e6ee2db5752e27066fe63f5c84d1aa5bf20a5fc4d2411e6",
				},
				ChainSpecific: core.ChainSpecific{
					ChainType: "Cardano",
					Config:    rawMessage,
				},
			},
		},
		Bridge: core.BridgeConfig{
			NodeUrl:              "https://polygon-mumbai-pokt.nodies.app", // will be our node,
			SmartContractAddress: "0xaE9d7040978152349c488b1A29b653e04dcca1f3",
			SigningKey:           "93c91e490bfd3736d17d04f53a10093e9cf2435309f4be1f5751381c8e201d23",
		},
		PullTimeMilis: 1000,
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

	cardanoOperationsPrime, err := GetChainSpecificOperations(loadedConfig.Chains["prime"].ChainSpecific)
	assert.NoError(t, err)

	operationsType := reflect.TypeOf(cardanoOperationsPrime)
	assert.NotNil(t, operationsType)
}
