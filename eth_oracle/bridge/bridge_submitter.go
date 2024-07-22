package bridge

import (
	"context"
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/eth_oracle/core"
	oracleCore "github.com/Ethernal-Tech/apex-bridge/oracle/core"
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

func (bs *BridgeSubmitterImpl) SubmitClaims(claims *oracleCore.BridgeClaims, submitOpts *eth.SubmitOpts) error {
	err := bs.bridgeSC.SubmitClaims(bs.ctx, claims.ContractClaims, submitOpts)
	if err != nil {
		bs.logger.Error("Failed to submit claims", "claims", claims, "err", err)

		return err
	}

	bs.logger.Info("Claims submitted successfully", "claims", claims)

	return nil
}

func (bs *BridgeSubmitterImpl) SubmitConfirmedBlocks(chainID string, firstBlock uint64, lastProcessedBlock uint64,
) error {
	contractBlocks := make([]eth.CardanoBlock, 0, lastProcessedBlock-firstBlock+1)
	for blockNumber := firstBlock; blockNumber <= lastProcessedBlock; blockNumber++ {
		contractBlocks = append(contractBlocks, eth.CardanoBlock{
			BlockSlot: new(big.Int).SetUint64(blockNumber),
		})
	}

	err := bs.bridgeSC.SubmitLastObservedBlocks(bs.ctx, chainID, contractBlocks)
	if err != nil {
		bs.logger.Error("Failed to submit confirmed blocks", "for chainID", chainID, "block", lastProcessedBlock, "err", err)

		return err
	}

	bs.logger.Info("Confirmed blocks submitted successfully", "for chainID", chainID, "last block", lastProcessedBlock)

	return nil
}