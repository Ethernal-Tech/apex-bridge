package core

import (
	"context"
	"math/big"

	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"go.etcd.io/bbolt"
)

type StakingManager interface {
	Start()
	GetStakingComponent(chainID string) (StakingComponent, error)
}

type StakingComponent interface {
	Start(ctx context.Context) error
	GetLastExchangeRate() (float64, error)
	ChooseStakeAddrForStaking(amount *big.Int) (string, error)
	ChooseStakeAddrForUnstaking(amount *big.Int) (map[string]*big.Int, error)
	Stake(amount *big.Int, stakingAddress string) error
	Unstake(amount *big.Int, stakingAddress string) error
	ReceiveReward(reward *big.Int, stakingAddress string) error
}

type StakingAddress interface {
	GetAddress() string
	GetTotalStTokens() *big.Int
	GetTotalTokensWithRewards() *big.Int
	Stake(amount *big.Int, exchangeRate float64) error
	Unstake(amount *big.Int, exchangeRate float64) error
	ReceiveReward(reward *big.Int) error
}

type CardanoTxsReceiver interface {
	NewUnprocessedTxs(originChainID string, txs []*indexer.Tx) error
}

type CardanoTxsDB interface {
	ClearAllTxs(chainID string) error
}

type StakingAddressDB interface {
	UpdateExchangeRate(chainID string, exchangeRate float64) error
	GetLastExchangeRate(chainID string) (float64, error)
	UpdateStakingAddress(chainID string, stakingAddress StakingAddress) error
	GetStakingAddress(chainID string, address string) (result StakingAddress, err error)
	GetAllStakingAddresses(chainID string) (result []StakingAddress, err error)
	UpdateStakingAddressAndExRate(chainID string, stakingAddress StakingAddress, exchangeRate *float64) error
}

type Database interface {
	CardanoTxsDB
	StakingAddressDB
	Init(db *bbolt.DB, smConfig *StakingManagerConfiguration)
}
