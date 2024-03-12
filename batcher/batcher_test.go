package batcher

import (
	"testing"

	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
)

func TestBatcherManagerConfig(t *testing.T) {
	expectedConfig := &BatcherManagerConfiguration{
		CardanoChains: map[string]CardanoChainConfig{
			"prime": {
				TestNetMagic:          uint(2),
				BlockfrostUrl:         "https://cardano-preview.blockfrost.io/api/v0",
				BlockfrostAPIKey:      "preview7mGSjpyEKb24OxQ4cCxomxZ5axMs5PvE",
				AtLeastValidators:     2.0 / 3.0,
				PotentialFee:          300_000,
				SigningKeyMultiSig:    "58201217236ac24d8ac12684b308cf9468f68ef5283096896dc1c5c3caf8351e2847",
				SigningKeyMultiSigFee: "5820f2c3b9527ec2f0d70e6ee2db5752e27066fe63f5c84d1aa5bf20a5fc4d2411e6",
			},
			"vector": {
				TestNetMagic:          uint(2),
				BlockfrostUrl:         "https://cardano-preview.blockfrost.io/api/v0",
				BlockfrostAPIKey:      "preview7mGSjpyEKb24OxQ4cCxomxZ5axMs5PvE",
				AtLeastValidators:     2.0 / 3.0,
				PotentialFee:          300_000,
				SigningKeyMultiSig:    "58201217236ac24d8ac12684b308cf9468f68ef5283096896dc1c5c3caf8351e2847",
				SigningKeyMultiSigFee: "5820f2c3b9527ec2f0d70e6ee2db5752e27066fe63f5c84d1aa5bf20a5fc4d2411e6",
			},
		},
		Bridge: BridgeConfig{
			NodeUrl:              "https://polygon-mumbai-pokt.nodies.app", // will be our node,
			SmartContractAddress: "0xb2B87f7e652Aa847F98Cc05e130d030b91c7B37d",
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

	loadedConfig, err := LoadConfig()
	assert.NoError(t, err)

	assert.Equal(t, expectedConfig.CardanoChains, loadedConfig.CardanoChains)
	assert.Equal(t, expectedConfig.Bridge, loadedConfig.Bridge)
	assert.Equal(t, expectedConfig.PullTimeMilis, loadedConfig.PullTimeMilis)
	assert.Equal(t, expectedConfig.Logger, loadedConfig.Logger)
}
