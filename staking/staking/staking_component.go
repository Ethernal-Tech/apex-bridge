package stakingcomponent

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
	"time"

	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	"github.com/Ethernal-Tech/apex-bridge/staking/core"
	"github.com/hashicorp/go-hclog"
)

type StakingComponentImpl struct {
	config               *core.StakingConfiguration
	cardanoChainObserver oCore.CardanoChainObserver
	stakingDB            core.Database
	logger               hclog.Logger
}

var _ core.StakingComponent = (*StakingComponentImpl)(nil)

func NewStakingComponent(
	config *core.StakingConfiguration,
	cardanoChainObserver oCore.CardanoChainObserver,
	stakingDB core.Database,
	logger hclog.Logger,
) (*StakingComponentImpl, error) {
	chainID := config.Chain.ChainID

	for _, addr := range config.Chain.StakingAddresses {
		if _, err := stakingDB.GetStakingAddress(chainID, addr); err != nil {
			sa, err := NewStakingAddress(addr, config.UsersRewardsPercentage)
			if err != nil {
				return nil, fmt.Errorf("failed to create staking address %s: %w", addr, err)
			}

			err = stakingDB.UpdateStakingAddress(chainID, sa)
			if err != nil {
				return nil, fmt.Errorf("failed to add new staking address %s to the database: %w", addr, err)
			}
		}
	}

	exchangeRate, err := stakingDB.GetLastExchangeRate(chainID)
	if err != nil {
		logger.Info("Could not get last exchange rate, initializing to 1", "chainID", chainID, "error", err)

		if err := stakingDB.UpdateExchangeRate(chainID, 1); err != nil {
			return nil, fmt.Errorf("failed to update initial exchange rate for chainID %s: %w", chainID, err)
		}

		exchangeRate = 1
	}

	logger.Info("Creating new staking component", "exchangeRate", exchangeRate)

	return &StakingComponentImpl{
		config:               config,
		cardanoChainObserver: cardanoChainObserver,
		stakingDB:            stakingDB,
		logger:               logger,
	}, nil
}

// Start launches the staking component's background process.
// It starts the Cardano chain observer and enters a loop that periodically executes staking logic.
// The loop exits gracefully when the context is cancelled.
func (sc *StakingComponentImpl) Start(ctx context.Context) error {
	sc.logger.Debug("Starting Staking Component...")

	// Start the Cardano chain observer
	if err := sc.cardanoChainObserver.Start(); err != nil {
		return fmt.Errorf("failed to start observer for chain %s: %w",
			sc.cardanoChainObserver.GetConfig().GetChainID(), err)
	}

	waitInterval := time.Millisecond * time.Duration(sc.config.PullTimeMilis)

	// Background loop that runs until context is cancelled
	for {
		select {
		case <-ctx.Done():
			sc.logger.Info("Staking Component shutting down...")

			return nil

		case <-time.After(waitInterval):
			sc.logger.Debug("Staking Component executing...")
		}
	}
}

// GetLastExchangeRate returns the current exchange rate between the native tokens and stTokens.
func (sc *StakingComponentImpl) GetLastExchangeRate() (float64, error) {
	return sc.stakingDB.GetLastExchangeRate(sc.config.Chain.ChainID)
}

// ChooseStakeAddrForStaking selects the staking address with the lowest current load.
func (sc *StakingComponentImpl) ChooseStakeAddrForStaking(amount *big.Int) (string, error) {
	stakingAddresses, err := sc.stakingDB.GetAllStakingAddresses(sc.config.Chain.ChainID)
	if err != nil {
		return "", fmt.Errorf("failed to get staking addresses for chainID %s", sc.config.Chain.ChainID)
	}

	if len(stakingAddresses) == 0 {
		return "", fmt.Errorf("no staking addresses configured for chainID %s", sc.config.Chain.ChainID)
	}

	minTotalTokens := new(big.Int)
	addrForStaking := ""

	for _, stakeAddr := range stakingAddresses {
		totalTokens := stakeAddr.GetTotalTokensWithRewards()
		if addrForStaking == "" || totalTokens.Cmp(minTotalTokens) < 0 {
			minTotalTokens = totalTokens
			addrForStaking = stakeAddr.GetAddress()
		}
	}

	return addrForStaking, nil
}

// ChooseStakeAddrForUnstaking selects the staking address with the most available tokens for unstaking.
// Currently returns the address with the highest load.
func (sc *StakingComponentImpl) ChooseStakeAddrForUnstaking(amount *big.Int) (map[string]*big.Int, error) {
	stakingAddresses, err := sc.stakingDB.GetAllStakingAddresses(sc.config.Chain.ChainID)
	if err != nil {
		return nil, fmt.Errorf("failed to get staking addresses for chainID %s", sc.config.Chain.ChainID)
	}

	if len(stakingAddresses) == 0 {
		return nil, fmt.Errorf("no staking addresses configured for chainID %s", sc.config.Chain.ChainID)
	}

	sorted := make([]core.StakingAddress, 0, len(stakingAddresses))
	sorted = append(sorted, stakingAddresses...)

	// Sort in descending order by total tokens with rewards
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].GetTotalTokensWithRewards().Cmp(sorted[j].GetTotalTokensWithRewards()) > 0
	})

	remaining := new(big.Int).Set(amount)
	addrsForUnstake := make(map[string]*big.Int)

	for _, addr := range sorted {
		if remaining.Sign() == 0 || addr.GetTotalTokensWithRewards().Sign() == 0 {
			break
		}

		// Determine how much can be unstaked from this address
		toUnstake := new(big.Int).Set(minBigInt(remaining, addr.GetTotalTokensWithRewards()))
		addrsForUnstake[addr.GetAddress()] = toUnstake
		remaining.Sub(remaining, toUnstake)
	}

	if remaining.Sign() > 0 {
		available := new(big.Int).Sub(amount, remaining)

		return nil, fmt.Errorf(
			"insufficient funds to unstake %s tokens for chainID %s: only %s available",
			amount.String(), sc.config.Chain.ChainID, available.String(),
		)
	}

	return addrsForUnstake, nil
}

// Stake updates the staking address state after receiving a user's staked tokens.
//
// This function assumes the staking address has already received the actual tokens.
func (sc *StakingComponentImpl) Stake(amount *big.Int, stakingAddress string) error {
	chainID := sc.config.Chain.ChainID

	if amount == nil || amount.Sign() <= 0 {
		return fmt.Errorf("stake amount must be greater than zero")
	}

	sa, err := sc.stakingDB.GetStakingAddress(chainID, stakingAddress)
	if err != nil {
		return fmt.Errorf("failed to stake - failed to get staking address %s from the database: %w", stakingAddress, err)
	}

	exchangeRate, err := sc.GetLastExchangeRate()
	if err != nil {
		return fmt.Errorf("failed to get exchange rate: %w", err)
	}

	if err = sa.Stake(amount, exchangeRate); err != nil {
		return fmt.Errorf("failed to stake into staking address %s: %w", stakingAddress, err)
	}

	return sc.stakingDB.UpdateStakingAddress(chainID, sa)
}

// Unstake processes a user's request to withdraw staked tokens from a specific staking address.
func (sc *StakingComponentImpl) Unstake(amount *big.Int, stakingAddress string) error {
	chainID := sc.config.Chain.ChainID

	if amount == nil || amount.Sign() <= 0 {
		return fmt.Errorf("unstake amount must be greater than zero")
	}

	sa, err := sc.stakingDB.GetStakingAddress(chainID, stakingAddress)
	if err != nil {
		return fmt.Errorf("failed to unstake - failed to get staking address %s from the database: %w", stakingAddress, err)
	}

	exchangeRate, err := sc.GetLastExchangeRate()
	if err != nil {
		return fmt.Errorf("failed to get exchange rate: %w", err)
	}

	if err = sa.Unstake(amount, exchangeRate); err != nil {
		return fmt.Errorf("failed to unstake from staking address %s: %w", stakingAddress, err)
	}

	return sc.stakingDB.UpdateStakingAddress(chainID, sa)
}

// ReceiveReward adds a reward amount to the specified staking address
// and updates the global exchange rate accordingly.
func (sc *StakingComponentImpl) ReceiveReward(reward *big.Int, stakingAddress string) error {
	chainID := sc.config.Chain.ChainID

	if reward == nil || reward.Sign() < 0 {
		return fmt.Errorf("reward must not be negative")
	}

	stakingAddresses, err := sc.stakingDB.GetAllStakingAddresses(sc.config.Chain.ChainID)
	if err != nil {
		return fmt.Errorf("failed to receive reward - failed to get staking addresses from the database: %w", err)
	}

	sa := findStakingAddress(stakingAddresses, stakingAddress)
	if sa == nil {
		return fmt.Errorf("staking address %s not found", stakingAddress)
	}

	if err := sa.ReceiveReward(reward); err != nil {
		return fmt.Errorf("failed to distribute reward to staking address %s: %w", stakingAddress, err)
	}

	newExchangeRate, err := sc.calculateExchangeRate(stakingAddresses)
	if err != nil {
		return fmt.Errorf("failed to calculate exchange rate: %w", err)
	}

	err = sc.stakingDB.UpdateStakingAddressAndExRate(chainID, sa, &newExchangeRate)
	if err != nil {
		return fmt.Errorf("failed to update staking address %s and exchange rate in db: %w", stakingAddress, err)
	}

	return nil
}

func DecodeStakingAddress(data []byte) (core.StakingAddress, error) {
	var sa StakingAddressImpl
	err := json.Unmarshal(data, &sa)

	return &sa, err
}

// calculateExchangeRate computes the exchange rate between total tokens (including rewards)
// and the total staked tokens across all configured staking addresses.
//
// If no staking addresses are configured, it returns an error.
// If the total staked tokens is zero, it returns the current exchange rate
func (sc *StakingComponentImpl) calculateExchangeRate(stakingAddresses []core.StakingAddress) (float64, error) {
	if len(stakingAddresses) == 0 {
		return 0, fmt.Errorf("no staking addresses configured for chainID %s", sc.config.Chain.ChainID)
	}

	totalStTokens := totalStTokens(stakingAddresses)
	if totalStTokens.Sign() == 0 {
		exchangeRate, err := sc.GetLastExchangeRate()
		if err != nil {
			return 0, fmt.Errorf("no staked tokens, failed to get last exchange rate: %w", err)
		}

		return exchangeRate, nil
	}

	// Convert integers to floats for division
	totalTokensWithRewardsFloat := new(big.Float).SetInt(totalTokensWithRewards(stakingAddresses))
	totalStTokensFloat := new(big.Float).SetInt(totalStTokens)

	// Compute exchange rate: totalTokensWithRewards / totalStTokens
	exchangeRate, _ := new(big.Float).Quo(totalTokensWithRewardsFloat, totalStTokensFloat).Float64()

	return exchangeRate, nil
}

// totalTokensWithRewards returns the sum of total tokens including rewards
// for all configured staking addresses.
func totalTokensWithRewards(stakingAddresses []core.StakingAddress) *big.Int {
	sum := new(big.Int)
	for _, addr := range stakingAddresses {
		sum.Add(sum, addr.GetTotalTokensWithRewards())
	}

	return sum
}

// totalStTokens returns the total number of staked tokens across all staking addresses.
func totalStTokens(stakingAddresses []core.StakingAddress) *big.Int {
	sum := new(big.Int)
	for _, addr := range stakingAddresses {
		sum.Add(sum, addr.GetTotalStTokens())
	}

	return sum
}

func findStakingAddress(addresses []core.StakingAddress, targetAddr string) core.StakingAddress {
	for _, addr := range addresses {
		if addr.GetAddress() == targetAddr {
			return addr
		}
	}

	return nil
}

func minBigInt(a, b *big.Int) *big.Int {
	if a.Cmp(b) <= 0 {
		return a
	}

	return b
}
