package core

import (
	"context"
	"math/big"

	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"go.etcd.io/bbolt"
)

type StakingManager interface {
	Start()
	GetExchangeRate(chainID string) (float64, error)
	ChooseStakeAddrForStaking(chainID string, amount *big.Int) (string, error)
	ChooseStakeAddrForUnstaking(chainID string, amount *big.Int) (map[string]*big.Int, error)
	Stake(chainID string, amount *big.Int, stakingAddress string) error
	Unstake(chainID string, amount *big.Int, stakingAddress string) error
	ReceiveReward(chainID string, reward *big.Int, stakingAddress string) error
}

type StakingComponent interface {
	Start(ctx context.Context) error
	GetExchangeRate() float64
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

type Database interface {
	CardanoTxsDB
	Init(db *bbolt.DB, smConfig *StakingManagerConfiguration)
}
