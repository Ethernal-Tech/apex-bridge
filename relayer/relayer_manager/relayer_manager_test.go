package relayer_manager

import (
	"encoding/json"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/relayer/core"
	"github.com/Ethernal-Tech/apex-bridge/relayer/relayer"
	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRelayerManagerConfig(t *testing.T) {
	jsonData := []byte(`{
		"testnetMagic": 2,
		"blockfrostUrl": "https://cardano-preview.blockfrost.io/api/v0",
		"blockfrostApiKey": "preview7mGSjpyEKb24OxQ4cCxomxZ5axMs5PvE",
		"atLeastValidators": 0.6666666666666666,
		"potentialFee": 300000
		}`)

	rawMessage := json.RawMessage(jsonData)

	expectedConfig := &core.RelayerManagerConfiguration{
		Chains: map[string]core.ChainConfig{
			"prime": {
				Base: core.BaseConfig{
					ChainId: "prime",
				},
				ChainSpecific: core.ChainSpecific{
					ChainType: "Cardano",
					Config:    rawMessage,
				},
			},
			"vector": {
				Base: core.BaseConfig{
					ChainId: "vector",
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

	assert.NotEmpty(t, loadedConfig.Chains)
	for _, chainConfig := range loadedConfig.Chains {
		assert.Equal(t, chainConfig.Base, chainConfig.Base)

		expectedOp, err := relayer.GetChainSpecificOperations(chainConfig.ChainSpecific)
		assert.NoError(t, err)

		loadedOp, err := relayer.GetChainSpecificOperations(chainConfig.ChainSpecific)
		assert.NoError(t, err)

		assert.Equal(t, expectedOp, loadedOp)
	}
	assert.Equal(t, expectedConfig.Bridge, loadedConfig.Bridge)
	assert.Equal(t, expectedConfig.PullTimeMilis, loadedConfig.PullTimeMilis)
	assert.Equal(t, expectedConfig.Logger, loadedConfig.Logger)
}

func TestRelayerManagerCreation(t *testing.T) {
	jsonData := []byte(`{
		"testnetMagic": 2,
		"blockfrostUrl": "https://cardano-preview.blockfrost.io/api/v0",
		"blockfrostApiKey": "preview7mGSjpyEKb24OxQ4cCxomxZ5axMs5PvE",
		"atLeastValidators": 0.6666666666666666,
		"potentialFee": 300000
		}`)

	rawMessage := json.RawMessage(jsonData)

	config := &core.RelayerManagerConfiguration{
		Chains: map[string]core.ChainConfig{
			"prime": {
				Base: core.BaseConfig{
					ChainId: "prime",
				},
				ChainSpecific: core.ChainSpecific{
					ChainType: "Cardano",
					Config:    rawMessage,
				},
			},
			"vector": {
				Base: core.BaseConfig{
					ChainId: "vector",
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
		PullTimeMilis: 1000,
		Logger: logger.LoggerConfig{
			LogFilePath:   "./relayer_logs",
			LogLevel:      hclog.Debug,
			JSONLogFormat: false,
			AppendFile:    true,
		},
	}

	t.Run("create manager fail - invalid operations", func(t *testing.T) {
		config := &core.RelayerManagerConfiguration{
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
		manager := NewRelayerManager(config, make(map[string]core.ChainOperations))
		require.Nil(t, manager)
	})

	t.Run("create manager without mocks", func(t *testing.T) {
		manager := NewRelayerManager(config, make(map[string]core.ChainOperations))
		require.NotNil(t, manager)
	})

	t.Run("create manager with chain operations mock", func(t *testing.T) {
		manager := NewRelayerManager(config, map[string]core.ChainOperations{"prime": nil})
		require.NotNil(t, manager)
	})

	t.Run("create manager with chain operations and bridge mock", func(t *testing.T) {
		manager := NewRelayerManager(config, map[string]core.ChainOperations{"prime": nil}, nil)
		require.NotNil(t, manager)
	})

}
