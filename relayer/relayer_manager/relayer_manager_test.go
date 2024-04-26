package relayermanager

import (
	"encoding/json"
	"os"
	"path"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/relayer/core"
	"github.com/Ethernal-Tech/apex-bridge/relayer/relayer"
	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRelayerManagerConfig(t *testing.T) {
	testDir, err := os.MkdirTemp("", "rl-mngr-config")
	require.NoError(t, err)

	defer func() {
		os.RemoveAll(testDir)
		os.Remove(testDir)
	}()

	jsonData := []byte(`{
		"blockFrostUrl": "http://hello.com",
		"testnetMagic": 2,
		"potentialFee": 300000
		}`)

	rawMessage := json.RawMessage(jsonData)

	expectedConfig := &core.RelayerManagerConfiguration{
		Chains: map[string]core.ChainConfig{
			"prime": {
				ChainType:     "Cardano",
				ChainSpecific: rawMessage,
			},
			"vector": {
				ChainType:     "CardaNo",
				ChainSpecific: rawMessage,
			},
		},
		Bridge: core.BridgeConfig{
			NodeURL:              "dummyNode", // will be our node,
			SmartContractAddress: "0x3786783",
		},
		PullTimeMilis: 1000,
		Logger: logger.LoggerConfig{
			LogFilePath:   path.Join(testDir, "relayer_logs"),
			LogLevel:      hclog.Debug,
			JSONLogFormat: false,
			AppendFile:    true,
		},
	}

	configFilePath := path.Join(testDir, "config.json")

	bytes, err := json.Marshal(expectedConfig)
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(configFilePath, bytes, 0770))

	loadedConfig, err := LoadConfig(configFilePath)
	require.NoError(t, err)

	assert.NotEmpty(t, loadedConfig.Chains)

	for _, chainConfig := range loadedConfig.Chains {
		expectedOp, err := relayer.GetChainSpecificOperations(chainConfig)
		require.NoError(t, err)

		loadedOp, err := relayer.GetChainSpecificOperations(chainConfig)
		require.NoError(t, err)

		assert.Equal(t, expectedOp, loadedOp)
	}

	assert.Equal(t, expectedConfig.Bridge, loadedConfig.Bridge)
	assert.Equal(t, expectedConfig.PullTimeMilis, loadedConfig.PullTimeMilis)
	assert.Equal(t, expectedConfig.Logger, loadedConfig.Logger)
}

func TestRelayerManagerCreation(t *testing.T) {
	t.Run("create manager fail - invalid operations", func(t *testing.T) {
		config := &core.RelayerManagerConfiguration{
			Chains: map[string]core.ChainConfig{
				"prime": {
					ChainType:     "Cardano",
					ChainSpecific: json.RawMessage(""),
				},
			},
		}
		manager, err := NewRelayerManager(config, hclog.NewNullLogger())
		require.Nil(t, manager)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to unmarshal Cardano configuration")
	})
}
