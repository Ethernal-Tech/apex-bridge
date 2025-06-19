package stakingcomponent

import (
	"context"
	"fmt"
	"math/big"
	"sort"
	"sync"
	"time"

	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	"github.com/Ethernal-Tech/apex-bridge/staking/core"
	"github.com/hashicorp/go-hclog"
)

type StakingComponentImpl struct {
	config               *core.StakingConfiguration
	cardanoChainObserver oCore.CardanoChainObserver
	stakingAddresses     map[string]core.StakingAddress
	exchangeRate         float64
	mutex                sync.RWMutex
	logger               hclog.Logger
}

var _ core.StakingComponent = (*StakingComponentImpl)(nil)

func NewStakingComponent(
	config *core.StakingConfiguration,
	cardanoChainObserver oCore.CardanoChainObserver,
	logger hclog.Logger,
) (*StakingComponentImpl, error) {
	stakingAddresses := make(map[string]core.StakingAddress, len(config.Chain.StakingAddresses))

	for _, addr := range config.Chain.StakingAddresses {
		sa, err := NewStakingAddress(addr, config.UsersRewardsPercentage)
		if err != nil {
			return nil, fmt.Errorf("failed to create staking address %s: %w", addr, err)
		}

		stakingAddresses[addr] = sa
	}

	return &StakingComponentImpl{
		config:               config,
		cardanoChainObserver: cardanoChainObserver,
		stakingAddresses:     stakingAddresses,
		exchangeRate:         1,
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

// GetExchangeRate returns the current exchange rate between the native tokens and stTokens.
func (sc *StakingComponentImpl) GetExchangeRate() float64 {
	sc.mutex.RLock()
	defer sc.mutex.RUnlock()

	return sc.exchangeRate
}

// ChooseStakeAddrForStaking selects the staking address with the lowest current load.
func (sc *StakingComponentImpl) ChooseStakeAddrForStaking(amount *big.Int) (string, error) {
	if len(sc.stakingAddresses) == 0 {
		return "", fmt.Errorf("no staking addresses configured for chainID %s", sc.config.Chain.ChainID)
	}

	minTotalTokens := new(big.Int)
	addrForStaking := ""

	for addr, stakeAddr := range sc.stakingAddresses {
		totalTokens := stakeAddr.GetTotalTokensWithRewards()
		if addrForStaking == "" || totalTokens.Cmp(minTotalTokens) < 0 {
			minTotalTokens = totalTokens
			addrForStaking = addr
		}
	}

	return addrForStaking, nil
}

// ChooseStakeAddrForUnstaking selects the staking address with the most available tokens for unstaking.
// Currently returns the address with the highest load.
func (sc *StakingComponentImpl) ChooseStakeAddrForUnstaking(amount *big.Int) (map[string]*big.Int, error) {
	if len(sc.stakingAddresses) == 0 {
		return nil, fmt.Errorf("no staking addresses configured for chainID %s", sc.config.Chain.ChainID)
	}

	sorted := make([]core.StakingAddress, 0, len(sc.stakingAddresses))
	for _, addr := range sc.stakingAddresses {
		sorted = append(sorted, addr)
	}

	// Sort in descending order by total tokens with rewards
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].GetTotalTokensWithRewards().Cmp(sorted[j].GetTotalTokensWithRewards()) > 0
	})

	remaining := new(big.Int).Set(amount)
	addrsForUnstake := make(map[string]*big.Int)

	for _, addr := range sorted {
		if remaining.Cmp(big.NewInt(0)) == 0 || addr.GetTotalTokensWithRewards().Cmp(big.NewInt(0)) == 0 {
			break
		}

		// Determine how much can be unstaked from this address
		toUnstake := new(big.Int).Set(minBigInt(remaining, addr.GetTotalTokensWithRewards()))
		addrsForUnstake[addr.GetAddress()] = toUnstake
		remaining.Sub(remaining, toUnstake)
	}

	if remaining.Cmp(big.NewInt(0)) > 0 {
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
	sa, ok := sc.stakingAddresses[stakingAddress]
	if !ok {
		return fmt.Errorf("staking address %s not found for chainID %s", stakingAddress, sc.config.Chain.ChainID)
	}

	return sa.Stake(amount, sc.GetExchangeRate())
}

// Unstake processes a user's request to withdraw staked tokens from a specific staking address.
func (sc *StakingComponentImpl) Unstake(amount *big.Int, stakingAddress string) error {
	sa, ok := sc.stakingAddresses[stakingAddress]
	if !ok {
		return fmt.Errorf("staking address %s not found for chainID %s", stakingAddress, sc.config.Chain.ChainID)
	}

	return sa.Unstake(amount, sc.GetExchangeRate())
}

// ReceiveReward adds a reward amount to the specified staking address
// and updates the global exchange rate accordingly.
func (sc *StakingComponentImpl) ReceiveReward(reward *big.Int, stakingAddress string) error {
	sa, ok := sc.stakingAddresses[stakingAddress]
	if !ok {
		return fmt.Errorf("staking address %s not found for chainID %s", stakingAddress, sc.config.Chain.ChainID)
	}

	if err := sa.ReceiveReward(reward); err != nil {
		return fmt.Errorf("failed to distribute reward to staking address %s: %w", stakingAddress, err)
	}

	newExchangeRate, err := sc.calculateExchangeRate()
	if err != nil {
		return fmt.Errorf("failed to calculate exchange rate: %w", err)
	}

	sc.mutex.Lock()
	sc.exchangeRate = newExchangeRate
	sc.mutex.Unlock()

	return nil
}

// calculateExchangeRate computes the exchange rate between total tokens (including rewards)
// and the total staked tokens across all configured staking addresses.
//
// If no staking addresses are configured, it returns an error.
// If the total staked tokens is zero, it returns the current exchange rate
func (sc *StakingComponentImpl) calculateExchangeRate() (float64, error) {
	if len(sc.stakingAddresses) == 0 {
		return 0, fmt.Errorf("no staking addresses configured for chainID %s", sc.config.Chain.ChainID)
	}

	totalStTokens := sc.totalStTokens()
	if totalStTokens.Cmp(big.NewInt(0)) == 0 {
		return sc.GetExchangeRate(), nil
	}

	// Convert integers to floats for division
	totalTokensWithRewardsFloat := new(big.Float).SetInt(sc.totalTokensWithRewards())
	totalStTokensFloat := new(big.Float).SetInt(totalStTokens)

	// Compute exchange rate: totalTokensWithRewards / totalStTokens
	exchangeRate, _ := new(big.Float).Quo(totalTokensWithRewardsFloat, totalStTokensFloat).Float64()

	return exchangeRate, nil
}

// totalTokensWithRewards returns the sum of total tokens including rewards
// for all configured staking addresses.
func (sc *StakingComponentImpl) totalTokensWithRewards() *big.Int {
	sum := new(big.Int)
	for _, addr := range sc.stakingAddresses {
		sum.Add(sum, addr.GetTotalTokensWithRewards())
	}

	return sum
}

// totalStTokens returns the total number of staked tokens across all staking addresses.
func (sc *StakingComponentImpl) totalStTokens() *big.Int {
	sum := new(big.Int)
	for _, addr := range sc.stakingAddresses {
		sum.Add(sum, addr.GetTotalStTokens())
	}

	return sum
}

func minBigInt(a, b *big.Int) *big.Int {
	if a.Cmp(b) <= 0 {
		return a
	}

	return b
}
