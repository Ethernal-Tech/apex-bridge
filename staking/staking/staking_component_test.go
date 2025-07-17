package stakingcomponent

import (
	"math/big"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/staking/core"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ocCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	databaseaccess "github.com/Ethernal-Tech/apex-bridge/staking/database_access"
)

const usersRewardsPercentage = 0.2

var stakingAddresses = []string{
	"addr_test1wpphktxx847q52fzn5g7efu2ftwxzuwwjkgsuvgwn2m7sdgn0z8z5",
	"addr_test1wpphktxx847q52fzn5g7efu2ftwxzuwwjkgsuvgwn2m7sdgn0z8z3",
	"addr_test1wpphktxx847q52fzn5g7efu2ftwxzuwwjkgsuvgwn2m7sdgn0z8z4",
}

func TestCalculateExchangeRate(t *testing.T) {
	testDir := createTempDir(t)
	defer os.RemoveAll(testDir)

	dbPath := filepath.Join(testDir, "temp_test.db")

	dbCleanup := func() {
		if _, err := os.Stat(dbPath); err == nil {
			os.Remove(dbPath)
		}
	}

	stakeAmounts := []*big.Int{
		big.NewInt(50_000_000_000_000_000),
		big.NewInt(20_000_000_000_000_000),
		big.NewInt(30_000_000_000_000_000),
	}
	rewards := []*big.Int{
		big.NewInt(10_000_000_000),
		big.NewInt(5_000_000_000),
		big.NewInt(20_000_000_000),
	}

	tests := []struct {
		name           string
		configBuilder  func() core.StakingConfiguration
		expectError    bool
		expectedErrMsg string
		expectedRate   float64
		stakeAmount    map[string]*big.Int
		unstakeAmount  map[string]*big.Int
	}{
		{
			name:          "initial exchange rate without staking",
			configBuilder: createConfigWithStAddrs,
			expectedRate:  1,
		},
		{
			name:          "exchange rate remains 1 after staking",
			configBuilder: createConfigWithStAddrs,
			expectedRate:  1,
			stakeAmount: map[string]*big.Int{
				stakingAddresses[0]: big.NewInt(5),
			},
		},
		{
			name:          "exchange rate remains 1 after both staking and unstaking",
			configBuilder: createConfigWithStAddrs,
			expectedRate:  1,
			stakeAmount: map[string]*big.Int{
				stakingAddresses[0]: big.NewInt(5),
				stakingAddresses[1]: big.NewInt(4),
			},
			unstakeAmount: map[string]*big.Int{
				stakingAddresses[0]: big.NewInt(3),
				stakingAddresses[1]: big.NewInt(4),
			},
		},
		{
			name:           "exchange rate fails when no staking addresses configured",
			configBuilder:  createConfigWithoutStAddrs,
			expectError:    true,
			expectedErrMsg: "no staking addresses configured",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Cleanup(dbCleanup)
			stakingDB := createStakingDB(t, dbPath)

			stakingComponent := newStakingComponent(t, stakingDB, test.configBuilder)
			assert.NotNil(t, stakingComponent)

			for addr, amount := range test.stakeAmount {
				err := stakingComponent.Stake(amount, addr)
				require.NoError(t, err)
			}

			checkExchangeRate(t, stakingComponent, test.expectedRate, test.expectError, test.expectedErrMsg)

			for addr, amount := range test.unstakeAmount {
				err := stakingComponent.Unstake(amount, addr)
				require.NoError(t, err)
			}

			checkExchangeRate(t, stakingComponent, test.expectedRate, test.expectError, test.expectedErrMsg)
		})
	}

	t.Run("calculate exchange rate after rewards are received", func(t *testing.T) {
		t.Cleanup(dbCleanup)
		stakingDB := createStakingDB(t, dbPath)

		sc := newStakingComponent(t, stakingDB, createConfigWithStAddrs)
		assert.NotNil(t, sc)

		expectedRates := calculateExpectedExchangeRates(stakeAmounts, rewards)

		checkExchangeRate(t, sc, 1.0, false, "")

		for i, amt := range stakeAmounts {
			require.NoError(t, sc.Stake(amt, stakingAddresses[i]))
			checkExchangeRate(t, sc, 1.0, false, "")
		}

		for i, reward := range rewards {
			require.NoError(t, sc.ReceiveReward(reward, stakingAddresses[i]))
			checkExchangeRate(t, sc, expectedRates[i], false, "")
		}
	})

	t.Run("exchange rate after stake, rewards, and unstake", func(t *testing.T) {
		t.Cleanup(dbCleanup)
		stakingDB := createStakingDB(t, dbPath)

		sc := newStakingComponent(t, stakingDB, createConfigWithStAddrs)
		assert.NotNil(t, sc)

		expectedRates := calculateExpectedExchangeRates(stakeAmounts, rewards)

		checkExchangeRate(t, sc, 1.0, false, "")

		// stake all to one address
		for _, amt := range stakeAmounts {
			require.NoError(t, sc.Stake(amt, stakingAddresses[0]))
			checkExchangeRate(t, sc, 1.0, false, "")
		}

		for idx, reward := range rewards {
			require.NoError(t, sc.ReceiveReward(reward, stakingAddresses[0]))
			checkExchangeRate(t, sc, expectedRates[idx], false, "")
		}

		for _, amount := range stakeAmounts {
			require.NoError(t, sc.Unstake(amount, stakingAddresses[0]))
			checkExchangeRate(t, sc, expectedRates[len(expectedRates)-1], false, "")
		}

		stakingAddresses, err := sc.stakingDB.GetAllStakingAddresses(sc.config.Chain.ChainID)
		require.NoError(t, err)

		remainingTotalTokens := totalTokensWithRewards(stakingAddresses)
		assert.Zero(t, remainingTotalTokens.Sign())
	})
}

func TestChooseAddrForStaking(t *testing.T) {
	testDir := createTempDir(t)
	defer os.RemoveAll(testDir)

	dbPath := filepath.Join(testDir, "temp_test.db")

	dbCleanup := func() {
		if _, err := os.Stat(dbPath); err == nil {
			os.Remove(dbPath)
		}
	}

	tests := []struct {
		name            string
		configBuilder   func() core.StakingConfiguration
		stakeTokens     map[string]*big.Int
		amount          *big.Int
		expectError     bool
		expectedErrMsg  string
		expectedAddress string
	}{
		{
			name:            "choose addresses for staking when all staking addresses have zero tokens and rewards",
			configBuilder:   createConfigWithStAddrs,
			amount:          big.NewInt(5),
			expectedAddress: "any",
		},
		{
			name:          "choose addresses for staking after the first and third have been staked to",
			configBuilder: createConfigWithStAddrs,
			amount:        big.NewInt(3),
			stakeTokens: map[string]*big.Int{
				stakingAddresses[0]: big.NewInt(5),
				stakingAddresses[2]: big.NewInt(5),
			},
			expectedAddress: stakingAddresses[1],
		},
		{
			name:          "choose addresses for staking after staking on all addresses",
			configBuilder: createConfigWithStAddrs,
			amount:        big.NewInt(3),
			stakeTokens: map[string]*big.Int{
				stakingAddresses[0]: big.NewInt(5),
				stakingAddresses[1]: big.NewInt(15),
				stakingAddresses[2]: big.NewInt(3),
			},
			expectedAddress: stakingAddresses[2],
		},
		{
			name:           "choose addresses for staking fails when no staking addresses configured",
			configBuilder:  createConfigWithoutStAddrs,
			amount:         big.NewInt(3),
			expectError:    true,
			expectedErrMsg: "no staking addresses configured",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Cleanup(dbCleanup)
			stakingDB := createStakingDB(t, dbPath)

			stakingComponent := newStakingComponent(t, stakingDB, test.configBuilder)
			assert.NotNil(t, stakingComponent)

			stakeTokens(t, stakingComponent, test.stakeTokens)

			stakeAddr, err := stakingComponent.ChooseStakeAddrForStaking(test.amount)

			if test.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.expectedErrMsg)
			} else {
				require.NoError(t, err)

				if test.expectedAddress == "any" {
					assert.True(t, slices.Contains(stakingAddresses, stakeAddr))
				} else {
					assert.Equal(t, test.expectedAddress, stakeAddr)
				}
			}
		})
	}
}

func TestChooseAddrForUnstaking(t *testing.T) {
	testDir := createTempDir(t)
	defer os.RemoveAll(testDir)

	dbPath := filepath.Join(testDir, "temp_test.db")

	dbCleanup := func() {
		if _, err := os.Stat(dbPath); err == nil {
			os.Remove(dbPath)
		}
	}

	tests := []struct {
		name           string
		configBuilder  func() core.StakingConfiguration
		stakeTokens    map[string]*big.Int
		amount         *big.Int
		expectError    bool
		expectedErrMsg string
		expectedResult map[string]*big.Int
	}{
		{
			name:          "choose one address for unstaking",
			configBuilder: createConfigWithStAddrs,
			amount:        big.NewInt(5),
			stakeTokens: map[string]*big.Int{
				stakingAddresses[0]: big.NewInt(6),
				stakingAddresses[1]: big.NewInt(7),
				stakingAddresses[2]: big.NewInt(2),
			},
			expectedResult: map[string]*big.Int{
				stakingAddresses[1]: big.NewInt(5),
			},
		},
		{
			name:          "choose two addresses for unstaking",
			configBuilder: createConfigWithStAddrs,
			amount:        big.NewInt(10),
			stakeTokens: map[string]*big.Int{
				stakingAddresses[0]: big.NewInt(6),
				stakingAddresses[1]: big.NewInt(7),
				stakingAddresses[2]: big.NewInt(2),
			},
			expectedResult: map[string]*big.Int{
				stakingAddresses[1]: big.NewInt(7),
				stakingAddresses[0]: big.NewInt(3),
			},
		},
		{
			name:          "choose two addresses for unstaking",
			configBuilder: createConfigWithStAddrs,
			amount:        big.NewInt(14),
			stakeTokens: map[string]*big.Int{
				stakingAddresses[0]: big.NewInt(6),
				stakingAddresses[1]: big.NewInt(7),
				stakingAddresses[2]: big.NewInt(2),
			},
			expectedResult: map[string]*big.Int{
				stakingAddresses[1]: big.NewInt(7),
				stakingAddresses[0]: big.NewInt(6),
				stakingAddresses[2]: big.NewInt(1),
			},
		},
		{
			name:           "unstake zero tokens",
			configBuilder:  createConfigWithStAddrs,
			amount:         big.NewInt(0),
			expectedResult: map[string]*big.Int{},
		},
		{
			name:          "choose address for unstaking fails when there is not enough funds",
			configBuilder: createConfigWithStAddrs,
			amount:        big.NewInt(16),
			stakeTokens: map[string]*big.Int{
				stakingAddresses[0]: big.NewInt(6),
				stakingAddresses[1]: big.NewInt(7),
				stakingAddresses[2]: big.NewInt(2),
			},
			expectError:    true,
			expectedErrMsg: "insufficient funds to unstake",
		},
		{
			name:           "choose address for unstaking fails when no staking addresses configured",
			configBuilder:  createConfigWithoutStAddrs,
			amount:         big.NewInt(3),
			expectError:    true,
			expectedErrMsg: "no staking addresses configured",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Cleanup(dbCleanup)
			stakingDB := createStakingDB(t, dbPath)

			stakingComponent := newStakingComponent(t, stakingDB, test.configBuilder)
			assert.NotNil(t, stakingComponent)

			stakeTokens(t, stakingComponent, test.stakeTokens)

			stakeAddr, err := stakingComponent.ChooseStakeAddrForUnstaking(test.amount)

			if test.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.expectedErrMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectedResult, stakeAddr)
			}
		})
	}
}

func TestStake(t *testing.T) {
	testDir := createTempDir(t)
	defer os.RemoveAll(testDir)

	dbPath := filepath.Join(testDir, "temp_test.db")

	dbCleanup := func() {
		if _, err := os.Stat(dbPath); err == nil {
			os.Remove(dbPath)
		}
	}

	t.Run("stake to valid address updates state correctly", func(t *testing.T) {
		t.Cleanup(dbCleanup)
		stakingDB := createStakingDB(t, dbPath)

		sc := newStakingComponent(t, stakingDB, createConfigWithStAddrs)
		assert.NotNil(t, sc)

		amount := big.NewInt(1_000_000_000_000_000) // 1 token
		addr := stakingAddresses[0]

		initialExchangeRate, err := sc.GetLastExchangeRate()
		require.NoError(t, err)
		assert.Equal(t, float64(1), initialExchangeRate)

		err = sc.Stake(amount, addr)
		require.NoError(t, err)

		sa, err := sc.stakingDB.GetStakingAddress(sc.config.Chain.ChainID, addr)
		require.NoError(t, err)

		assert.Equal(t, amount.String(), sa.GetTotalTokensWithRewards().String())
		assert.Equal(t, amount.String(), sa.GetTotalStTokens().String()) // 1:1 exchange rate
	})

	t.Run("stake fails with unknown staking address", func(t *testing.T) {
		t.Cleanup(dbCleanup)
		stakingDB := createStakingDB(t, dbPath)

		sc := newStakingComponent(t, stakingDB, createConfigWithStAddrs)
		assert.NotNil(t, sc)

		err := sc.Stake(big.NewInt(1_000), "unknown_address")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("stake fails when exchange rate is 0", func(t *testing.T) {
		t.Cleanup(dbCleanup)
		stakingDB := createStakingDB(t, dbPath)

		sc := newStakingComponent(t, stakingDB, createConfigWithStAddrs)
		assert.NotNil(t, sc)

		addr := stakingAddresses[0]

		// manually override exchange rate to simulate 0
		require.NoError(t, sc.stakingDB.UpdateExchangeRate(sc.config.Chain.ChainID, 0))

		err := sc.Stake(big.NewInt(1_000), addr)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exchange rate cannot be less than 1")
	})
}

func TestUnstake(t *testing.T) {
	testDir := createTempDir(t)
	defer os.RemoveAll(testDir)

	dbPath := filepath.Join(testDir, "temp_test.db")

	dbCleanup := func() {
		if _, err := os.Stat(dbPath); err == nil {
			os.Remove(dbPath)
		}
	}

	t.Run("unstake successfully with valid address and sufficient balance", func(t *testing.T) {
		t.Cleanup(dbCleanup)
		stakingDB := createStakingDB(t, dbPath)

		sc := newStakingComponent(t, stakingDB, createConfigWithStAddrs)
		assert.NotNil(t, sc)

		addr := stakingAddresses[0]
		stakeAmount := big.NewInt(1_000_000_000_000_000) // 1 token

		// Stake first
		require.NoError(t, sc.Stake(stakeAmount, addr))

		// Unstake the same amount of stTokens (1:1 exchange rate)
		require.NoError(t, sc.Unstake(stakeAmount, addr))

		sa, err := sc.stakingDB.GetStakingAddress(sc.config.Chain.ChainID, addr)
		require.NoError(t, err)

		assert.Zero(t, sa.GetTotalTokensWithRewards().Sign())
		assert.Zero(t, sa.GetTotalStTokens().Sign())
	})

	t.Run("fail to unstake from unknown address", func(t *testing.T) {
		t.Cleanup(dbCleanup)
		stakingDB := createStakingDB(t, dbPath)

		sc := newStakingComponent(t, stakingDB, createConfigWithStAddrs)
		err := sc.Unstake(big.NewInt(1000), "unknown_address")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("fail to unstake with zero exchange rate", func(t *testing.T) {
		t.Cleanup(dbCleanup)
		stakingDB := createStakingDB(t, dbPath)

		sc := newStakingComponent(t, stakingDB, createConfigWithStAddrs)
		addr := stakingAddresses[0]

		// Simulate staking
		require.NoError(t, sc.Stake(big.NewInt(1_000), addr))

		// Force exchange rate to 0
		require.NoError(t, sc.stakingDB.UpdateExchangeRate(sc.config.Chain.ChainID, 0))

		err := sc.Unstake(big.NewInt(100), addr)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "exchange rate cannot be less than 1")
	})

	t.Run("fail to unstake more stTokens than available", func(t *testing.T) {
		t.Cleanup(dbCleanup)
		stakingDB := createStakingDB(t, dbPath)

		sc := newStakingComponent(t, stakingDB, createConfigWithStAddrs)
		addr := stakingAddresses[0]

		require.NoError(t, sc.Stake(big.NewInt(1_000), addr))

		// Try to unstake more than was staked
		err := sc.Unstake(big.NewInt(2_000), addr)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "exceeds available stTokens")
	})

	t.Run("fail to unstake if underlying tokens with rewards are insufficient", func(t *testing.T) {
		t.Cleanup(dbCleanup)
		stakingDB := createStakingDB(t, dbPath)

		sc := newStakingComponent(t, stakingDB, createConfigWithStAddrs)
		chainID := sc.config.Chain.ChainID
		addr := stakingAddresses[0]

		// Stake 1000 tokens
		require.NoError(t, sc.Stake(big.NewInt(1_000), addr))

		// Simulate reduction of totalTokensWithRewards to a small value
		sa, err := sc.stakingDB.GetStakingAddress(chainID, addr)
		require.NoError(t, err)

		sa.(*StakingAddressImpl).TotalTokensWithRewards = big.NewInt(500)
		require.NoError(t, sc.stakingDB.UpdateStakingAddress(chainID, sa))

		err = sc.Unstake(big.NewInt(1_000), addr)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "exceed available tokens with rewards")
	})
}

func TestReceiveReward(t *testing.T) {
	testDir := createTempDir(t)
	defer os.RemoveAll(testDir)

	dbPath := filepath.Join(testDir, "temp_test.db")

	dbCleanup := func() {
		if _, err := os.Stat(dbPath); err == nil {
			os.Remove(dbPath)
		}
	}

	t.Run("successfully distribute reward and update exchange rate", func(t *testing.T) {
		t.Cleanup(dbCleanup)
		stakingDB := createStakingDB(t, dbPath)

		sc := newStakingComponent(t, stakingDB, createConfigWithStAddrs)
		addr := stakingAddresses[0]
		stakeAmount := big.NewInt(1_000_000_000_000_000)
		reward := big.NewInt(1_000_000_000)

		// Stake first
		require.NoError(t, sc.Stake(stakeAmount, addr))
		initialRate, err := sc.GetLastExchangeRate()
		require.NoError(t, err)
		assert.Equal(t, 1.0, initialRate)

		// Receive reward
		require.NoError(t, sc.ReceiveReward(reward, addr))

		// Check exchange rate increased
		newRate, err := sc.GetLastExchangeRate()
		require.NoError(t, err)
		assert.Greater(t, newRate, initialRate)

		// Check that tokens with rewards increased
		sa, err := sc.stakingDB.GetStakingAddress(sc.config.Chain.ChainID, addr)
		require.NoError(t, err)
		assert.True(t, sa.GetTotalTokensWithRewards().Cmp(stakeAmount) > 0)
	})

	t.Run("successfully distribute zero reward without updating exchange rate", func(t *testing.T) {
		t.Cleanup(dbCleanup)
		stakingDB := createStakingDB(t, dbPath)

		sc := newStakingComponent(t, stakingDB, createConfigWithStAddrs)
		addr := stakingAddresses[0]
		stakeAmount := big.NewInt(1_000_000_000_000_000)
		reward := big.NewInt(0)

		// Stake first
		require.NoError(t, sc.Stake(stakeAmount, addr))
		initialRate, err := sc.GetLastExchangeRate()
		require.NoError(t, err)
		assert.Equal(t, 1.0, initialRate)

		// Receive reward
		require.NoError(t, sc.ReceiveReward(reward, addr))

		// Check that exchange rate  does not change
		newRate, err := sc.GetLastExchangeRate()
		require.NoError(t, err)
		assert.True(t, newRate == initialRate)

		// Check that tokens with rewards  does not change
		sa, err := sc.stakingDB.GetStakingAddress(sc.config.Chain.ChainID, addr)
		require.NoError(t, err)
		assert.True(t, sa.GetTotalTokensWithRewards().Cmp(stakeAmount) == 0)
	})

	t.Run("successfully distribute reward with zero rewards percentage without updating exchange rate", func(t *testing.T) {
		t.Cleanup(dbCleanup)
		stakingDB := createStakingDB(t, dbPath)

		observer := core.CardanoChainObserverMock{}
		cfg := createConfigWithStAddrs()
		cfg.UsersRewardsPercentage = 0

		sc, err := NewStakingComponent(&cfg, &observer, stakingDB, hclog.Default())
		require.NoError(t, err)

		addr := stakingAddresses[0]
		stakeAmount := big.NewInt(1_000_000_000_000_000)
		reward := big.NewInt(1_000_000_000)

		// Stake first
		require.NoError(t, sc.Stake(stakeAmount, addr))
		initialRate, err := sc.GetLastExchangeRate()
		require.NoError(t, err)
		assert.Equal(t, 1.0, initialRate)

		// Receive reward
		require.NoError(t, sc.ReceiveReward(reward, addr))

		// Check that exchange rate  does not change
		newRate, err := sc.GetLastExchangeRate()
		require.NoError(t, err)
		assert.True(t, newRate == initialRate)

		// Check that tokens with rewards  does not change
		sa, err := sc.stakingDB.GetStakingAddress(sc.config.Chain.ChainID, addr)
		require.NoError(t, err)
		assert.True(t, sa.GetTotalTokensWithRewards().Cmp(stakeAmount) == 0)
	})

	t.Run("fail to distribute reward if no stTokens are present", func(t *testing.T) {
		t.Cleanup(dbCleanup)
		stakingDB := createStakingDB(t, dbPath)

		sc := newStakingComponent(t, stakingDB, createConfigWithStAddrs)
		addr := stakingAddresses[0]
		reward := big.NewInt(1_000_000_000)

		err := sc.ReceiveReward(reward, addr)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "reward cannot be distributed")
	})

	t.Run("fail to distribute reward to unknown address", func(t *testing.T) {
		t.Cleanup(dbCleanup)
		stakingDB := createStakingDB(t, dbPath)

		sc := newStakingComponent(t, stakingDB, createConfigWithStAddrs)
		reward := big.NewInt(1_000_000_000)

		err := sc.ReceiveReward(reward, "nonexistent_address")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func createTempDir(t *testing.T) string {
	t.Helper()

	dir, err := os.MkdirTemp("", "staking-test-*")
	require.NoError(t, err)

	return dir
}

func createStakingDB(t *testing.T, dbPath string) *databaseaccess.BBoltDatabase {
	t.Helper()

	cfg := createConfigWithStAddrs()

	smConfig := createSMConfig(&cfg)
	boltDB, err := databaseaccess.NewDatabase(dbPath, &smConfig)
	require.NoError(t, err)

	stakingDB := databaseaccess.NewBBoltDatabase(DecodeStakingAddress)
	stakingDB.Init(boltDB, &smConfig)

	return stakingDB
}

func newStakingComponent(t *testing.T, stakingDB *databaseaccess.BBoltDatabase, configBuilder func() core.StakingConfiguration) *StakingComponentImpl {
	t.Helper()

	observer := core.CardanoChainObserverMock{}
	cfg := configBuilder()

	sc, err := NewStakingComponent(&cfg, &observer, stakingDB, hclog.Default())
	require.NoError(t, err)

	return sc
}

func stakeTokens(t *testing.T, sc *StakingComponentImpl, tokens map[string]*big.Int) {
	t.Helper()

	chainID := sc.config.Chain.ChainID
	stakingAddresses, err := sc.stakingDB.GetAllStakingAddresses(chainID)
	require.NoError(t, err)

	for _, addr := range stakingAddresses {
		if amount, ok := tokens[addr.GetAddress()]; ok {
			require.NoError(t, addr.Stake(amount, 1))
			require.NoError(t, sc.stakingDB.UpdateStakingAddress(chainID, addr))
		}
	}
}

func createSMConfig(stakingConfig *core.StakingConfiguration) core.StakingManagerConfiguration {
	chains := map[string]*core.CardanoChainConfig{
		stakingConfig.Chain.ChainID: {
			BaseCardanoChainConfig: ocCore.BaseCardanoChainConfig{
				ChainID: stakingConfig.Chain.ChainID,
			},
		},
	}

	return core.StakingManagerConfiguration{
		Chains: chains,
	}
}

func createConfigWithStAddrs() core.StakingConfiguration {
	return createStakingConfig(stakingAddresses)
}

func createConfigWithoutStAddrs() core.StakingConfiguration {
	return createStakingConfig([]string{})
}

func createStakingConfig(stakingAddresses []string) core.StakingConfiguration {
	return core.StakingConfiguration{
		Chain: core.CardanoChainConfig{
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
		PullTimeMilis:          1000,
		UsersRewardsPercentage: usersRewardsPercentage,
	}
}

func checkExchangeRate(t *testing.T, stakingComponent *StakingComponentImpl, expectedRate float64, expectError bool, expectedErrMsg string) {
	t.Helper()

	stakingAddresses, err := stakingComponent.stakingDB.GetAllStakingAddresses(stakingComponent.config.Chain.ChainID)
	require.NoError(t, err)
	rate, err := stakingComponent.calculateExchangeRate(stakingAddresses)

	if expectError {
		require.Error(t, err)
		require.Contains(t, err.Error(), expectedErrMsg)
	} else {
		require.NoError(t, err)
		assert.Equal(t, expectedRate, rate)
	}
}

func sumBigInts(values []*big.Int) *big.Int {
	total := big.NewInt(0)
	for _, v := range values {
		total.Add(total, v)
	}

	return total
}

func calculateExpectedExchangeRates(stakeAmts, rewards []*big.Int) []float64 {
	totalTokens := sumBigInts(stakeAmts)
	totalStaked := new(big.Int).Set(totalTokens)
	result := make([]float64, 0, len(rewards))

	for _, reward := range rewards {
		// Apply user rewards percentage
		rewardFloat := new(big.Float).SetInt(reward)
		usersRewardFloat := new(big.Float).Mul(big.NewFloat(usersRewardsPercentage), rewardFloat)

		usersReward := new(big.Int)
		usersReward, _ = usersRewardFloat.Int(usersReward)

		totalTokens.Add(totalTokens, usersReward)

		totalTokensFloat := new(big.Float).SetInt(totalTokens)
		totalStakedFloat := new(big.Float).SetInt(totalStaked)
		exRate, _ := new(big.Float).Quo(totalTokensFloat, totalStakedFloat).Float64()

		result = append(result, exRate)
	}

	return result
}
