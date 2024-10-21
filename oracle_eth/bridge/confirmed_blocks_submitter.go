package bridge

import (
	"context"
	"fmt"
	"time"

	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	eth_core "github.com/Ethernal-Tech/apex-bridge/oracle_eth/core"

	eventTrackerStore "github.com/Ethernal-Tech/blockchain-event-tracker/store"
	"github.com/hashicorp/go-hclog"
)

type ConfirmedBlocksSubmitterImpl struct {
	ctx               context.Context
	bridgeSubmitter   eth_core.BridgeSubmitter
	appConfig         *oCore.AppConfig
	chainID           string
	indexerDB         eventTrackerStore.EventTrackerStore
	logger            hclog.Logger
	latestBlockNumber uint64
}

var _ oCore.ConfirmedBlocksSubmitter = (*ConfirmedBlocksSubmitterImpl)(nil)

func NewConfirmedBlocksSubmitter(
	ctx context.Context,
	bridgeSubmitter eth_core.BridgeSubmitter,
	appConfig *oCore.AppConfig,
	indexerDB eventTrackerStore.EventTrackerStore,
	chainID string,
	logger hclog.Logger,
) (*ConfirmedBlocksSubmitterImpl, error) {
	latestBlockPoint, err := indexerDB.GetLastProcessedBlock()
	if err != nil {
		return nil, err
	}

	return &ConfirmedBlocksSubmitterImpl{
		ctx:               ctx,
		bridgeSubmitter:   bridgeSubmitter,
		appConfig:         appConfig,
		chainID:           chainID,
		indexerDB:         indexerDB,
		logger:            logger.Named("confirmed_blocks_submitter_" + chainID),
		latestBlockNumber: latestBlockPoint,
	}, nil
}

func (bs *ConfirmedBlocksSubmitterImpl) StartSubmit() {
	waitTime := time.Millisecond * time.Duration(bs.appConfig.Bridge.SubmitConfig.ConfirmedBlocksSubmitTime)

	go func() {
		for {
			select {
			case <-bs.ctx.Done():
				return
			case <-time.After(waitTime):
				err := bs.execute()
				if err != nil {
					bs.logger.Error("error while executing", "err", err)
				}
			}
		}
	}()
}

func (bs *ConfirmedBlocksSubmitterImpl) execute() error {
	from := bs.latestBlockNumber
	if from != 0 {
		from++
	}

	bs.logger.Debug("Executing ConfirmedBlocksSubmitterImpl", "chainID", bs.chainID, "from block", from)

	lastProcessedBlock, err := bs.indexerDB.GetLastProcessedBlock()
	if err != nil {
		bs.logger.Error("error getting latest confirmed blocks", "err", err)

		return fmt.Errorf("error getting latest confirmed blocks. err: %w", err)
	}

	to := lastProcessedBlock
	if from > to {
		return nil
	}

	//nolint:gosec
	maxBlock := from + uint64(bs.appConfig.Bridge.SubmitConfig.ConfirmedBlocksThreshold)
	if to > maxBlock {
		to = maxBlock
	}

	if err := bs.bridgeSubmitter.SubmitConfirmedBlocks(bs.chainID, from, to); err != nil {
		bs.logger.Error("error submitting confirmed blocks", "err", err)

		return fmt.Errorf("error submitting confirmed blocks. err %w", err)
	}

	bs.latestBlockNumber = to
	bs.logger.Info("Submitted confirmed blocks", "chainID", bs.chainID, "latestBlockNumber", bs.latestBlockNumber)

	return nil
}

func (bs *ConfirmedBlocksSubmitterImpl) GetChainID() string {
	return bs.chainID
}
