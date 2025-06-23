package stakingmanager

import (
	"context"
	"encoding/json"
	"math/big"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/staking/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ocCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	databaseaccess "github.com/Ethernal-Tech/apex-bridge/staking/database_access"
)

var stakingAddresses = []string{
	"addr_test1wpphktxx847q52fzn5g7efu2ftwxzuwwjkgsuvgwn2m7sdgn0z1z1",
	"addr_test1wpphktxx847q52fzn5g7efu2ftwxzuwwjkgsuvgwn2m7sdgn0z8z2",
}

func TestStakingManagerConfig(t *testing.T) {
	testDir := createTempDir(t)
	defer os.RemoveAll(testDir)

	expectedConfig := createConfigWithStAddrs(testDir)
	configFilePath := filepath.Join(testDir, "config.json")

	writeJSONFile(t, configFilePath, expectedConfig)

	loadedConfig, err := common.LoadConfig[core.StakingManagerConfiguration](configFilePath, "")
	require.NoError(t, err)
	loadedConfig.FillOut()

	assert.Equal(t, expectedConfig.DbsPath, loadedConfig.DbsPath)
	assert.Equal(t, expectedConfig.PullTimeMilis, loadedConfig.PullTimeMilis)
	assert.Equal(t, expectedConfig.Logger, loadedConfig.Logger)

	for chainID, chainConfig := range loadedConfig.Chains {
		expectedChainConfig, ok := expectedConfig.Chains[chainID]
		assert.True(t, ok)
		assert.Equal(t, expectedChainConfig, chainConfig)
	}
}

func TestGetExchangeRate(t *testing.T) {
	testDir := createTempDir(t)
	defer os.RemoveAll(testDir)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	dbPath := filepath.Join(testDir, "temp_test.db")
	indexerDbs := map[string]indexer.Database{
		common.ChainIDStrPrime: &indexer.DatabaseMock{},
	}

	tests := []struct {
		name           string
		configBuilder  func(string) core.StakingManagerConfiguration
		chainID        string
		expectError    bool
		expectedErrMsg string
		expectedRate   float64
	}{
		{
			name:          "cardano exchange rate returns initial value",
			configBuilder: createConfigWithStAddrs,
			chainID:       common.ChainIDStrCardano,
			expectedRate:  1,
		},
		{
			name:          "prime exchange rate returns initial value",
			configBuilder: createConfigWithStAddrs,
			chainID:       common.ChainIDStrPrime,
			expectedRate:  1,
		},
		{
			name:           "exchange rate fails for non-existing chain",
			configBuilder:  createConfigWithStAddrs,
			chainID:        common.ChainIDStrNexus,
			expectError:    true,
			expectedErrMsg: "failed to get staking component for chainID",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Cleanup(func() {
				os.Remove(dbPath)
			})

			smConfig := test.configBuilder(testDir)

			stakingDB, err := databaseaccess.NewDatabase(dbPath, &smConfig)
			require.NoError(t, err)

			stakingManager, err := NewStakingManager(ctx, &smConfig, stakingDB, indexerDbs, hclog.Default())
			require.NoError(t, err)

			stakingComponent, err := stakingManager.GetStakingComponent(test.chainID)

			if test.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.expectedErrMsg)
			} else {
				require.NoError(t, err)

				rate, err := stakingComponent.GetLastExchangeRate()
				require.NoError(t, err)

				assert.Equal(t, test.expectedRate, rate)
			}
		})
	}
}

func TestChooseAddrForStaking(t *testing.T) {
	testDir := createTempDir(t)
	defer os.RemoveAll(testDir)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	dbPath := filepath.Join(testDir, "temp_test.db")
	indexerDbs := map[string]indexer.Database{
		common.ChainIDStrPrime: &indexer.DatabaseMock{},
	}

	tests := []struct {
		name           string
		configBuilder  func(string) core.StakingManagerConfiguration
		chainID        string
		expectError    bool
		expectedErrMsg string
	}{
		{
			name:          "cardano: choose address for staking when all staking addresses have zero tokens and rewards",
			configBuilder: createConfigWithStAddrs,
			chainID:       common.ChainIDStrCardano,
		},
		{
			name:          "prime: choose address for staking when all staking addresses have zero tokens and rewards",
			configBuilder: createConfigWithStAddrs,
			chainID:       common.ChainIDStrPrime,
		},
		{
			name:           "choose address for staking fails when no staking addresses configured",
			configBuilder:  createConfigWithoutStAddrs,
			chainID:        common.ChainIDStrPrime,
			expectError:    true,
			expectedErrMsg: "no staking addresses configured",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Cleanup(func() {
				os.Remove(dbPath)
			})

			smConfig := test.configBuilder(testDir)

			stakingDB, err := databaseaccess.NewDatabase(dbPath, &smConfig)
			require.NoError(t, err)

			stakingManager, err := NewStakingManager(ctx, &smConfig, stakingDB, indexerDbs, hclog.Default())
			require.NoError(t, err)

			stakingComponent, err := stakingManager.GetStakingComponent(test.chainID)
			require.NoError(t, err)

			stakeAddr, err := stakingComponent.ChooseStakeAddrForStaking(big.NewInt(5))

			if test.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.expectedErrMsg)
			} else {
				require.NoError(t, err)
				assert.True(t, slices.Contains(stakingAddresses, stakeAddr))
			}
		})
	}
}

func TestChooseAddrForUnstaking(t *testing.T) {
	testDir := createTempDir(t)
	defer os.RemoveAll(testDir)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	dbPath := filepath.Join(testDir, "temp_test.db")
	indexerDbs := map[string]indexer.Database{
		common.ChainIDStrPrime: &indexer.DatabaseMock{},
	}

	tests := []struct {
		name           string
		configBuilder  func(string) core.StakingManagerConfiguration
		chainID        string
		expectError    bool
		expectedErrMsg string
	}{
		{
			name:           "cardano: choose address for unstaking when all staking addresses have zero tokens",
			configBuilder:  createConfigWithStAddrs,
			chainID:        common.ChainIDStrCardano,
			expectedErrMsg: "insufficient funds to unstake",
		},
		{
			name:           "prime: choose address for unstaking when all staking addresses have zero tokens",
			configBuilder:  createConfigWithStAddrs,
			chainID:        common.ChainIDStrPrime,
			expectedErrMsg: "insufficient funds to unstake",
		},
		{
			name:           "choose address for unstaking fails when no staking addresses configured",
			configBuilder:  createConfigWithoutStAddrs,
			chainID:        common.ChainIDStrPrime,
			expectedErrMsg: "no staking addresses configured",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Cleanup(func() {
				os.Remove(dbPath)
			})

			smConfig := test.configBuilder(testDir)

			stakingDB, err := databaseaccess.NewDatabase(dbPath, &smConfig)
			require.NoError(t, err)

			stakingManager, err := NewStakingManager(ctx, &smConfig, stakingDB, indexerDbs, hclog.Default())
			require.NoError(t, err)

			stakingComponent, err := stakingManager.GetStakingComponent(test.chainID)
			require.NoError(t, err)

			_, err = stakingComponent.ChooseStakeAddrForUnstaking(big.NewInt(5))

			require.Error(t, err)
			require.Contains(t, err.Error(), test.expectedErrMsg)
		})
	}
}

func createTempDir(t *testing.T) string {
	t.Helper()

	dir, err := os.MkdirTemp("", "staking-test-*")
	require.NoError(t, err)

	return dir
}

func writeJSONFile(t *testing.T, path string, data any) {
	t.Helper()

	bytes, err := json.Marshal(data)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, bytes, 0770))
}

func createConfigWithStAddrs(dir string) core.StakingManagerConfiguration {
	return createStakingManagerConfig(dir, stakingAddresses)
}

func createConfigWithoutStAddrs(testDir string) core.StakingManagerConfiguration {
	return createStakingManagerConfig(testDir, []string{})
}

func createStakingManagerConfig(testDir string, stakingAddresses []string) core.StakingManagerConfiguration {
	return core.StakingManagerConfiguration{
		Chains: map[string]*core.CardanoChainConfig{
			common.ChainIDStrCardano: {
				BaseCardanoChainConfig: ocCore.BaseCardanoChainConfig{
					ChainID:                common.ChainIDStrCardano,
					NetworkAddress:         "localhost:5500",
					StartBlockHash:         "",
					StartSlot:              0,
					ConfirmationBlockCount: 10,
				},
				ChainType:        "Cardano",
				NetworkMagic:     2,
				StakingAddresses: stakingAddresses,
				StakingBridgingAddr: core.StakingBridgingAddresses{
					StakingBridgingAddr: "addr_test1wpphktxx847q52fzn5g7efu2ftwxzuwwjkgsuvgwn2m7sdgn0z8zg",
					FeeAddress:          "addr_test1wz4f8kcdy3yue80gq5qmd4902pu8n3wf0k35m54k7g2qmsggfjqck",
				},
			},
			common.ChainIDStrPrime: {
				BaseCardanoChainConfig: ocCore.BaseCardanoChainConfig{
					ChainID:                common.ChainIDStrPrime,
					NetworkAddress:         "localhost:5100",
					StartBlockHash:         "",
					StartSlot:              0,
					ConfirmationBlockCount: 10,
				},
				ChainType:        "Cardano",
				NetworkMagic:     3311,
				StakingAddresses: stakingAddresses,
				StakingBridgingAddr: core.StakingBridgingAddresses{
					StakingBridgingAddr: "addr_test1wrslpa7pfd5qqpk20jvy0sku653yc8sg8lway5k7sssr8ygdefnxh",
					FeeAddress:          "addr_test1wpg7lpsdaslegalggl4wnjs0qjnc0eh00aw59pgvtzs597cumml62",
				},
			},
		},
		PullTimeMilis:          1000,
		DbsPath:                testDir,
		UsersRewardsPercentage: 0.2,
		Logger: logger.LoggerConfig{
			LogFilePath:   filepath.Join(testDir, "staking_logs"),
			LogLevel:      hclog.Debug,
			JSONLogFormat: false,
			AppendFile:    true,
		},
	}
}
