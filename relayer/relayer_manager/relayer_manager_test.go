package relayer_manager

import (
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/relayer/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
)

func TestRelayerManagerConfig(t *testing.T) {
	expectedConfig := &core.RelayerManagerConfiguration{
		CardanoChains: map[string]core.CardanoChainConfig{
			"prime": {
				TestNetMagic:      uint(2),
				ChainId:           "prime",
				BlockfrostUrl:     "https://cardano-preview.blockfrost.io/api/v0",
				BlockfrostAPIKey:  "preview7mGSjpyEKb24OxQ4cCxomxZ5axMs5PvE",
				AtLeastValidators: 2.0 / 3.0,
				PotentialFee:      300_000,
			},
			"vector": {
				TestNetMagic:      uint(2),
				ChainId:           "vector",
				BlockfrostUrl:     "https://cardano-preview.blockfrost.io/api/v0",
				BlockfrostAPIKey:  "preview7mGSjpyEKb24OxQ4cCxomxZ5axMs5PvE",
				AtLeastValidators: 2.0 / 3.0,
				PotentialFee:      300_000,
			},
		},
		Bridge: core.BridgeConfig{
			NodeUrl:              "https://polygon-mumbai-pokt.nodies.app", // will be our node,
			SmartContractAddress: "0xaE9d7040978152349c488b1A29b653e04dcca1f3",
		},
		PullTimeMilis: 1000,
		Logger: logger.LoggerConfig{
			LogFilePath:   "./relayer_logs",
			LogLevel:      hclog.Debug,
			JSONLogFormat: false,
			AppendFile:    true,
		},
	}

	loadedConfig, err := LoadConfig("../config.json")
	assert.NoError(t, err)

	assert.Equal(t, expectedConfig.CardanoChains, loadedConfig.CardanoChains)
	assert.Equal(t, expectedConfig.Bridge, loadedConfig.Bridge)
	assert.Equal(t, expectedConfig.PullTimeMilis, loadedConfig.PullTimeMilis)
	assert.Equal(t, expectedConfig.Logger, loadedConfig.Logger)
}
