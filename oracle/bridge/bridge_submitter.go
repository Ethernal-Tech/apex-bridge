package bridge

import (
	"context"

	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/hashicorp/go-hclog"
)

type BridgeSubmitterImpl struct {
	bridgeSC  eth.IOracleBridgeSmartContract
	logger    hclog.Logger
	ctx       context.Context
	cancelCtx context.CancelFunc
}

var _ core.BridgeSubmitter = (*BridgeSubmitterImpl)(nil)

func NewBridgeSubmitter(
	bridgeSC eth.IOracleBridgeSmartContract,
	logger hclog.Logger,
) *BridgeSubmitterImpl {
	ctx, cancelCtx := context.WithCancel(context.Background())
	return &BridgeSubmitterImpl{
		bridgeSC:  bridgeSC,
		logger:    logger,
		ctx:       ctx,
		cancelCtx: cancelCtx,
	}
}

func (bs *BridgeSubmitterImpl) SubmitClaims(claims *core.BridgeClaims) error {
	err := bs.bridgeSC.SubmitClaims(bs.ctx, claims.ContractClaims)
	if err != nil {
		bs.logger.Error("Failed to submit claims", "err", err)
		return err
	}

	bs.logger.Info("Claims submitted successfully")
	return nil
}

func (bs *BridgeSubmitterImpl) SubmitConfirmedBlocks(chainId string, blocks []*indexer.CardanoBlock) error {
	contractBlocks := make([]eth.CardanoBlock, 0, len(blocks))
	for _, bl := range blocks {
		contractBlocks = append(contractBlocks, eth.CardanoBlock{
			BlockHash: bl.Hash,
			BlockSlot: bl.Slot,
		})
	}

	err := bs.bridgeSC.SubmitLastObservableBlocks(bs.ctx, chainId, contractBlocks)

	return err
}

func (bs *BridgeSubmitterImpl) Dispose() error {
	bs.cancelCtx()

	return nil
}
