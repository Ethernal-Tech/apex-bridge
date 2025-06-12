package processor

import (
	"github.com/Ethernal-Tech/apex-bridge/staking/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/hashicorp/go-hclog"
)

type CardanoTxsReceiverImpl struct {
	smConfig *core.StakingManagerConfiguration
	db       core.CardanoTxsDB
	logger   hclog.Logger
}

var _ core.CardanoTxsReceiver = (*CardanoTxsReceiverImpl)(nil)

func NewCardanoTxsReceiverImpl(
	smConfig *core.StakingManagerConfiguration,
	db core.CardanoTxsDB,
	logger hclog.Logger,
) *CardanoTxsReceiverImpl {
	return &CardanoTxsReceiverImpl{
		smConfig: smConfig,
		db:       db,
		logger:   logger,
	}
}

func (r *CardanoTxsReceiverImpl) NewUnprocessedTxs(originChainID string, txs []*indexer.Tx) error {
	return nil
}
