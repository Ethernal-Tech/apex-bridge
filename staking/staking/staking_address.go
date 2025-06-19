package stakingcomponent

import (
	"fmt"
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/staking/core"
)

// StakingAddressImpl represents an address that delegates to a stake pool,
// along with its associated token balances and user rewards distribution.
//
// Fields:
//   - address: The staking address that delegates to a stake pool.
//   - totalTokensWithRewards: The total tokens staked by users plus accumulated rewards.
//   - totalStTokens: Total staked tokens (stTokens) issued to users by the bridge for this staking address.
//   - usersRewardsPercentage: The percentage of rewards allocated to users
//     staking through this address (value between 0 and 1).
type StakingAddressImpl struct {
	address                string
	totalTokensWithRewards *big.Int
	totalStTokens          *big.Int
	usersRewardsPercentage float64
}

var _ core.StakingAddress = (*StakingAddressImpl)(nil)

// NewStakingAddress creates a new StakingAddressImpl with the given address and user rewards percentage.
// usersRewardsPercentage should be a value between 0 and 1 representing the share of rewards for users.
func NewStakingAddress(address string, usersRewardsPercentage float64) (*StakingAddressImpl, error) {
	if usersRewardsPercentage < 0 || usersRewardsPercentage > 1 {
		return nil, fmt.Errorf("usersRewardsPercentage must be between 0 and 1")
	}

	return &StakingAddressImpl{
		address:                address,
		totalTokensWithRewards: big.NewInt(0),
		totalStTokens:          big.NewInt(0),
		usersRewardsPercentage: usersRewardsPercentage,
	}, nil
}

func (sa *StakingAddressImpl) GetAddress() string {
	return sa.address
}

func (sa *StakingAddressImpl) GetTotalStTokens() *big.Int {
	return sa.totalStTokens
}

func (sa *StakingAddressImpl) GetTotalTokensWithRewards() *big.Int {
	return sa.totalTokensWithRewards
}

// Staking address state is updated after receiving users' staked tokens.
//
// It increases the total tokens (including rewards) by the staked amount,
// and mints the corresponding amount of stTokens based on the provided exchange rate.
func (sa *StakingAddressImpl) Stake(amount *big.Int, exchangeRate float64) error {
	if exchangeRate == 0 {
		return fmt.Errorf("cannot stake tokens: exchange rate cannot be zero")
	}

	sa.totalTokensWithRewards.Add(sa.totalTokensWithRewards, amount)

	amountFloat := new(big.Float).SetInt(amount)
	stTokensToMintFloat := new(big.Float).Quo(amountFloat, big.NewFloat(exchangeRate))
	stTokensToMint := new(big.Int)
	stTokensToMint, _ = stTokensToMintFloat.Int(stTokensToMint)

	sa.totalStTokens.Add(sa.totalStTokens, stTokensToMint)

	return nil
}

// Unstake processes a user's request to withdraw staked tokens.
//
// It decreases the total stTokens by the specified amount and deducts the equivalent
// underlying tokens (including rewards) based on the current exchange rate.
func (sa *StakingAddressImpl) Unstake(amount *big.Int, exchangeRate float64) error {
	if exchangeRate == 0 {
		return fmt.Errorf("cannot unstake tokens: exchange rate cannot be zero")
	}

	if sa.totalStTokens.Cmp(amount) < 0 {
		return fmt.Errorf(
			"cannot unstake: requested stTokens (%s) exceeds available stTokens (%s)",
			amount.String(),
			sa.totalStTokens.String(),
		)
	}

	amountFloat := new(big.Float).SetInt(amount)
	stTokensToUnstakeFloat := new(big.Float).Mul(amountFloat, big.NewFloat(exchangeRate))
	tokensToUnstake := new(big.Int)
	tokensToUnstake, _ = stTokensToUnstakeFloat.Int(tokensToUnstake)

	if sa.totalTokensWithRewards.Cmp(tokensToUnstake) < 0 {
		return fmt.Errorf(
			"cannot unstake: required tokens (%s) exceed available tokens with rewards (%s)",
			tokensToUnstake.String(),
			sa.totalTokensWithRewards.String(),
		)
	}

	sa.totalTokensWithRewards.Sub(sa.totalTokensWithRewards, tokensToUnstake)
	sa.totalStTokens.Sub(sa.totalStTokens, amount)

	return nil
}

// This function should be called when the rewards account receives new tokens.
// It calculates the portion of the reward allocated to users (based on usersRewardsPercentage)
// and adds it to the total tokens including rewards.
func (sa *StakingAddressImpl) ReceiveReward(reward *big.Int) error {
	if sa.totalStTokens.Cmp(big.NewInt(0)) == 0 {
		return fmt.Errorf("no staked tokens: reward cannot be distributed to users")
	}

	// Calculate user's share of the reward
	rewardFloat := new(big.Float).SetInt(reward)
	userRewardFloat := new(big.Float).Mul(rewardFloat, big.NewFloat(sa.usersRewardsPercentage))
	userReward := new(big.Int)
	userReward, _ = userRewardFloat.Int(userReward)

	sa.totalTokensWithRewards.Add(sa.totalTokensWithRewards, userReward)

	return nil
}
