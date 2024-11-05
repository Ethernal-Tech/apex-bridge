package bridge

import (
	"context"
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hashicorp/go-hclog"
)

type BridgeSubmitterImpl struct {
	ctx      context.Context
	bridgeSC eth.IOracleBridgeSmartContract
	logger   hclog.Logger
}

var _ core.BridgeSubmitter = (*BridgeSubmitterImpl)(nil)

func NewBridgeSubmitter(
	ctx context.Context,
	bridgeSC eth.IOracleBridgeSmartContract,
	logger hclog.Logger,
) *BridgeSubmitterImpl {
	return &BridgeSubmitterImpl{
		ctx:      ctx,
		bridgeSC: bridgeSC,
		logger:   logger,
	}
}

func (bs *BridgeSubmitterImpl) SubmitClaims(
	claims *cCore.BridgeClaims, submitOpts *eth.SubmitOpts) (*types.Receipt, error) {
	receipt, err := bs.bridgeSC.SubmitClaims(bs.ctx, claims.ContractClaims, submitOpts)
	if err != nil {
		bs.logger.Error("Failed to submit claims", "claims", claims, "err", err)

		return nil, err
	}

	bs.logger.Info("Claims submitted successfully", "claims", claims)

	return receipt, nil
}

func (bs *BridgeSubmitterImpl) SubmitConfirmedBlocks(chainID string, blocks []*indexer.CardanoBlock) error {
	contractBlocks := make([]eth.CardanoBlock, 0, len(blocks))
	for _, bl := range blocks {
		contractBlocks = append(contractBlocks, eth.CardanoBlock{
			BlockHash: bl.Hash,
			BlockSlot: new(big.Int).SetUint64(bl.Slot),
		})
	}

	err := bs.bridgeSC.SubmitLastObservedBlocks(bs.ctx, chainID, contractBlocks)
	if err != nil {
		bs.logger.Error("Failed to submit confirmed blocks", "for chainID", chainID, "blocks", blocks, "err", err)

		return err
	}

	bs.logger.Info("Confirmed blocks submitted successfully", "for chainID", chainID, "blocks", blocks)

	return nil
}
