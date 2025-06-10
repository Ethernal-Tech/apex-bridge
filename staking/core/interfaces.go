package core

import (
	"context"

	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"go.etcd.io/bbolt"
)

type StakingManager interface {
	Start()
}

type StakingComponent interface {
	Start(ctx context.Context)
}

type CardanoChainObserver interface {
	Start()
	Dispose() error
	GetConfig() *ChainConfig
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
