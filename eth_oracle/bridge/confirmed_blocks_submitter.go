package bridge

import (
	"context"
	"fmt"
	"time"

	eth_core "github.com/Ethernal-Tech/apex-bridge/eth_oracle/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"

	eventTrackerStore "github.com/Ethernal-Tech/blockchain-event-tracker/store"
	"github.com/hashicorp/go-hclog"
)

type ConfirmedBlocksSubmitterImpl struct {
	ctx                 context.Context
	bridgeSubmitter     eth_core.BridgeSubmitter
	appConfig           *core.AppConfig
	chainID             string
	indexerDB           eventTrackerStore.EventTrackerStore
	oracleDB            eth_core.EthTxsDB
	logger              hclog.Logger
	latestConfirmedSlot uint64
}

var _ core.ConfirmedBlocksSubmitter = (*ConfirmedBlocksSubmitterImpl)(nil)

func NewConfirmedBlocksSubmitter(
	ctx context.Context,
	bridgeSubmitter eth_core.BridgeSubmitter,
	appConfig *core.AppConfig,
	oracleDB eth_core.EthTxsDB,
	indexerDB eventTrackerStore.EventTrackerStore,
	chainID string,
	logger hclog.Logger,
) (*ConfirmedBlocksSubmitterImpl, error) {
	latestBlockPoint, err := indexerDB.GetLastProcessedBlock()
	if err != nil {
		return nil, err
	}

	return &ConfirmedBlocksSubmitterImpl{
		ctx:                 ctx,
		bridgeSubmitter:     bridgeSubmitter,
		appConfig:           appConfig,
		chainID:             chainID,
		indexerDB:           indexerDB,
		oracleDB:            oracleDB,
		logger:              logger.Named("confirmed_blocks_submitter_" + chainID),
		latestConfirmedSlot: latestBlockPoint,
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
	from := bs.latestConfirmedSlot
	if from != 0 {
		from++
	}

	bs.logger.Debug("Executing ConfirmedBlocksSubmitterImpl", "chainID", bs.chainID, "from slot", from)

	lastProcessedBlock, err := bs.indexerDB.GetLastProcessedBlock()
	if err != nil {
		bs.logger.Error("error getting latest confirmed blocks", "err", err)

		return fmt.Errorf("error getting latest confirmed blocks. err: %w", err)
	}

	if lastProcessedBlock < from {
		bs.logger.Debug("No new processed blocks", "chainID", bs.chainID)

		return nil
	}

	if err := bs.bridgeSubmitter.SubmitConfirmedBlocks(bs.chainID, from, lastProcessedBlock); err != nil {
		bs.logger.Error("error submitting confirmed blocks", "err", err)

		return fmt.Errorf("error submitting confirmed blocks. err %w", err)
	}

	bs.latestConfirmedSlot = lastProcessedBlock
	bs.logger.Info("Submitted confirmed blocks", "chainID", bs.chainID, "latestConfirmedSlot", bs.latestConfirmedSlot)

	return nil
}

func (bs *ConfirmedBlocksSubmitterImpl) GetChainID() string {
	return bs.chainID
}