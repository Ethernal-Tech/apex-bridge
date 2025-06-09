package bridge

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/eth"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hashicorp/go-hclog"
)

type BridgeSubmitterImpl struct {
	ctx      context.Context
	bridgeSC eth.IOracleBridgeSmartContract
	logger   hclog.Logger
}

var _ oCore.BridgeBlocksSubmitter = (*BridgeSubmitterImpl)(nil)

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
	claims *oCore.BridgeClaims, submitOpts *eth.SubmitOpts) (*types.Receipt, error) {
	receipt, err := bs.bridgeSC.SubmitClaims(bs.ctx, claims.ContractClaims, submitOpts)
	if err != nil {
		bs.logger.Error("Failed to submit claims", "claims", claims, "err", err)

		return nil, err
	}

	bs.logger.Info("Claims submitted successfully", "claims", claims)

	return receipt, nil
}

func (bs *BridgeSubmitterImpl) SubmitBlocks(chainID string, blocks []eth.CardanoBlock) error {
	var latestSlot *big.Int
	if len(blocks) > 0 {
		latestSlot = blocks[len(blocks)-1].BlockSlot
	}

	if bs.logger.IsDebug() {
		slots := make([]string, len(blocks))
		for i, x := range blocks {
			slots[i] = x.BlockSlot.String()
		}

		bs.logger.Info("Submitting blocks", "chainID", chainID, "slots", strings.Join(slots, ", "))
	}

	err := bs.bridgeSC.SubmitLastObservedBlocks(bs.ctx, chainID, blocks)
	if err != nil {
		return fmt.Errorf("failed to submit blocks for %s (%d): %w", chainID, latestSlot, err)
	}

	bs.logger.Info("Blocks has been submitted successfully", "chainID", chainID, "latestBlock", latestSlot)

	return nil
}
