package bridge

import (
	"context"
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/eth"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/hashicorp/go-hclog"
)

type LastObservedImpl struct {
	bridgeSC eth.IOracleBridgeSmartContract
	logger   hclog.Logger
}

var _ oCore.LastBlockObsvervedTracker = (*LastObservedImpl)(nil)

func NewLastObserved(
	bridgeSC eth.IOracleBridgeSmartContract,
	logger hclog.Logger,
) *LastObservedImpl {
	return &LastObservedImpl{
		bridgeSC: bridgeSC,
		logger:   logger,
	}
}

func (lo *LastObservedImpl) GetLastObservedBlock(ctx context.Context, sourceChain string) (*big.Int, error) {
	lastObservedBlock, err := lo.bridgeSC.GetLastObservedBlock(ctx, sourceChain)
	if err != nil {
		lo.logger.Error("Failed to get last observed block", "chain", sourceChain, "err", err)

		return nil, err
	}

	lo.logger.Debug("Retrieved last observed block", "chain", sourceChain, "slot", lastObservedBlock.BlockSlot)

	return lastObservedBlock.BlockSlot, nil
}
