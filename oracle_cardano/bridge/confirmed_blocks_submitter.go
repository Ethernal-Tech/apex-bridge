package bridge

import (
	"context"
	"fmt"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/hashicorp/go-hclog"
)

type ConfirmedBlocksSubmitterImpl struct {
	ctx                 context.Context
	bridgeSubmitter     core.BridgeSubmitter
	appConfig           *cCore.AppConfig
	chainID             string
	indexerDB           indexer.Database
	oracleDB            core.CardanoTxsDB
	logger              hclog.Logger
	latestConfirmedSlot uint64
}

var _ cCore.ConfirmedBlocksSubmitter = (*ConfirmedBlocksSubmitterImpl)(nil)

func NewConfirmedBlocksSubmitter(
	ctx context.Context,
	bridgeSubmitter core.BridgeSubmitter,
	appConfig *cCore.AppConfig,
	oracleDB core.CardanoTxsDB,
	indexerDB indexer.Database,
	chainID string,
	logger hclog.Logger,
) (*ConfirmedBlocksSubmitterImpl, error) {
	latestBlockPoint, err := indexerDB.GetLatestBlockPoint()
	if err != nil {
		return nil, err
	}

	if latestBlockPoint == nil {
		latestBlockPoint = &indexer.BlockPoint{}
	}

	return &ConfirmedBlocksSubmitterImpl{
		ctx:                 ctx,
		bridgeSubmitter:     bridgeSubmitter,
		appConfig:           appConfig,
		chainID:             chainID,
		indexerDB:           indexerDB,
		oracleDB:            oracleDB,
		logger:              logger.Named("confirmed_blocks_submitter_" + chainID),
		latestConfirmedSlot: latestBlockPoint.BlockSlot,
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

	blocks, err := bs.indexerDB.GetConfirmedBlocksFrom(
		from,
		bs.appConfig.Bridge.SubmitConfig.ConfirmedBlocksThreshold)
	if err != nil {
		bs.logger.Error("error getting latest confirmed blocks", "err", err)

		return fmt.Errorf("error getting latest confirmed blocks. err: %w", err)
	}

	var blockCounter = 0

	bs.logger.Debug("Checking if blocks processed", "chainID", bs.chainID, "blocks", blocks)

	for _, block := range blocks {
		if !bs.checkIfBlockIsProcessed(block) {
			break
		}

		blockCounter++
	}

	if blockCounter == 0 {
		bs.logger.Debug("No new processed blocks", "chainID", bs.chainID)

		return nil
	}

	blocksToSubmit := blocks[:blockCounter]

	bs.logger.Debug("Submitting blocks", "chainID", bs.chainID, "blocks", blocksToSubmit)

	if err := bs.bridgeSubmitter.SubmitConfirmedBlocks(bs.chainID, blocksToSubmit); err != nil {
		bs.logger.Error("error submitting confirmed blocks", "err", err)

		return fmt.Errorf("error submitting confirmed blocks. err %w", err)
	}

	bs.latestConfirmedSlot = blocks[blockCounter-1].Slot
	bs.logger.Info("Submitted confirmed blocks", "chainID", bs.chainID, "latestConfirmedSlot", bs.latestConfirmedSlot)

	return nil
}

func (bs *ConfirmedBlocksSubmitterImpl) GetChainID() string {
	return bs.chainID
}

func (bs *ConfirmedBlocksSubmitterImpl) checkIfBlockIsProcessed(block *indexer.CardanoBlock) bool {
	for _, tx := range block.Txs {
		prTx, err := bs.oracleDB.GetProcessedTx(bs.chainID, tx)
		if err != nil {
			bs.logger.Error("error getting processed tx for block", "err", err)
		}

		if prTx == nil {
			return false
		}
	}

	return true
}