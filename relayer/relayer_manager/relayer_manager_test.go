package relayermanager

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/relayer/core"
	"github.com/Ethernal-Tech/apex-bridge/relayer/relayer"
	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	secretsHelper "github.com/Ethernal-Tech/cardano-infrastructure/secrets/helper"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRelayerManagerConfig(t *testing.T) {
	testDir, _, cleanup := setupTestSecretsManager(t)
	defer cleanup()

	jsonData := []byte(`{
		"blockFrostUrl": "http://hello.com",
		"testnetMagic": 2,
		"potentialFee": 300000
		}`)

	rawMessage := json.RawMessage(jsonData)

	expectedConfig := &core.RelayerManagerConfiguration{
		Chains: map[string]core.ChainConfig{
			common.ChainIDStrPrime: {
				ChainType:      "Cardano",
				ChainSpecific:  rawMessage,
				RelayerDataDir: testDir,
			},
			common.ChainIDStrVector: {
				ChainType:      "CardaNo",
				ChainSpecific:  rawMessage,
				RelayerDataDir: testDir,
			},
		},
		Bridge: core.BridgeConfig{
			NodeURL:              "dummyNode", // will be our node,
			SmartContractAddress: "0x3786783",
		},
		PullTimeMilis: 1000,
		Logger: logger.LoggerConfig{
			LogFilePath:   filepath.Join(testDir, "relayer_logs"),
			LogLevel:      hclog.Debug,
			JSONLogFormat: false,
			AppendFile:    true,
		},
	}

	configFilePath := filepath.Join(testDir, "config.json")

	bytes, err := json.Marshal(expectedConfig)
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(configFilePath, bytes, 0770))

	loadedConfig, err := LoadConfig(configFilePath)
	require.NoError(t, err)

	assert.NotEmpty(t, loadedConfig.Chains)

	for chainId, chainConfig := range loadedConfig.Chains {
		chainConfig.ChainID = chainId
		expectedOp, err := relayer.GetChainSpecificOperations(chainConfig, eth.Chain{}, hclog.NewNullLogger())
		require.NoError(t, err)

		loadedOp, err := relayer.GetChainSpecificOperations(chainConfig, eth.Chain{}, hclog.NewNullLogger())
		require.NoError(t, err)

		assert.Equal(t, expectedOp, loadedOp)
	}

	assert.Equal(t, expectedConfig.Bridge, loadedConfig.Bridge)
	assert.Equal(t, expectedConfig.PullTimeMilis, loadedConfig.PullTimeMilis)
	assert.Equal(t, expectedConfig.Logger, loadedConfig.Logger)
}

func Test_getRelayersAndConfigurations(t *testing.T) {
	testDir, secretsMngr, cleanup := setupTestSecretsManager(t)
	defer cleanup()

	_, err := eth.CreateAndSaveRelayerEVMPrivateKey(secretsMngr, common.ChainIDStrNexus, true)
	require.NoError(t, err)

	allRegisteredChains := []eth.Chain{
		{
			Id:        common.ChainIDIntPrime,
			ChainType: 0,
		},
		{
			Id:        common.ChainIDIntNexus,
			ChainType: 1,
		},
		{
			Id:        0x73,
			ChainType: 1,
		},
	}
	config := &core.RelayerManagerConfiguration{
		Chains: map[string]core.ChainConfig{
			common.ChainIDStrPrime: {
				ChainType: common.ChainTypeCardanoStr,
				DbsPath:   testDir,
				ChainSpecific: json.RawMessage([]byte(`{
					"blockFrostUrl": "http://hello.com"
				}`)),
				RelayerDataDir: testDir,
			},
			common.ChainIDStrVector: {
				ChainType:     common.ChainTypeCardanoStr,
				DbsPath:       testDir,
				ChainSpecific: json.RawMessage("{}"),
			},
			common.ChainIDStrNexus: {
				ChainType: common.ChainTypeEVMStr,
				DbsPath:   testDir,
				ChainSpecific: json.RawMessage([]byte(fmt.Sprintf(`{
					"dataDir": "%s"
				}`, testDir))),
			},
		},
	}

	relayers, chainsConfigs, err := getRelayersAndConfigurations(
		&eth.BridgeSmartContractMock{}, allRegisteredChains, config, hclog.NewNullLogger())
	require.NoError(t, err, err)
	require.Len(t, relayers, 2)
	require.Len(t, chainsConfigs, 2)
	require.True(t, chainsConfigs[common.ChainIDStrPrime].ChainID != "")
	require.True(t, chainsConfigs[common.ChainIDStrNexus].ChainID != "")
}

func setupTestSecretsManager(t *testing.T) (string, secrets.SecretsManager, func()) {
	t.Helper()

	testDir, err := os.MkdirTemp("", "rl-mngr-config")
	require.NoError(t, err)

	secretsMngr, err := secretsHelper.CreateSecretsManager(&secrets.SecretsManagerConfig{
		Path: testDir,
		Type: secrets.Local,
	})
	require.NoError(t, err)

	for _, chainID := range []string{common.ChainIDStrPrime, common.ChainIDStrVector} {
		_, err = cardanotx.GenerateWallet(secretsMngr, chainID, true, true)
		require.NoError(t, err)
	}

	cleanup := func() { os.RemoveAll(testDir) }

	return testDir, secretsMngr, cleanup
}
