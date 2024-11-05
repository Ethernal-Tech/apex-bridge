package bridge

import (
	"context"
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/eth"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_eth/core"
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
	claims *oCore.BridgeClaims, submitOpts *eth.SubmitOpts) (*types.Receipt, error) {
	receipt, err := bs.bridgeSC.SubmitClaims(bs.ctx, claims.ContractClaims, submitOpts)
	if err != nil {
		bs.logger.Error("Failed to submit claims", "claims", claims, "err", err)

		return nil, err
	}

	bs.logger.Info("Claims submitted successfully", "claims", claims)

	return receipt, nil
}

func (bs *BridgeSubmitterImpl) GetBatchTransactions(
	chainID string, batchID uint64,
) ([]eth.TxDataInfo, error) {
	txs, err := bs.bridgeSC.GetBatchTransactions(bs.ctx, chainID, batchID)
	if err != nil {
		bs.logger.Error("Failed to retrieve batch transactions", "chainID", chainID, "batchID", batchID, "err", err)

		return nil, err
	}

	bs.logger.Info("Batch transactions retrieved", "chainID", chainID, "batchID", batchID, "txs", len(txs))

	return txs, nil
}

func (bs *BridgeSubmitterImpl) SubmitConfirmedBlocks(chainID string, from uint64, to uint64,
) error {
	contractBlocks := make([]eth.CardanoBlock, 0, to-from+1)
	for blockNumber := from; blockNumber <= to; blockNumber++ {
		contractBlocks = append(contractBlocks, eth.CardanoBlock{
			BlockSlot: new(big.Int).SetUint64(blockNumber),
		})
	}

	err := bs.bridgeSC.SubmitLastObservedBlocks(bs.ctx, chainID, contractBlocks)
	if err != nil {
		bs.logger.Error("Failed to submit confirmed blocks", "for chainID", "chainID", chainID, "from block", from,
			"to block", to, "err", err)

		return err
	}

	bs.logger.Info("Confirmed blocks submitted successfully", "for chainID", chainID, "from block", from,
		"to block", to)

	return nil
}
