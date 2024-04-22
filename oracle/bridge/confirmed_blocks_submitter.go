package bridge

import (
	"context"
	"fmt"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/hashicorp/go-hclog"
)

type ConfirmedBlocksSubmitterImpl struct {
	ctx                 context.Context
	bridgeSubmitter     core.BridgeSubmitter
	appConfig           *core.AppConfig
	chainId             string
	indexerDb           indexer.Database
	oracleDb            core.CardanoTxsDb
	logger              hclog.Logger
	latestConfirmedSlot uint64
}

var _ core.ConfirmedBlocksSubmitter = (*ConfirmedBlocksSubmitterImpl)(nil)

func NewConfirmedBlocksSubmitter(
	ctx context.Context,
	bridgeSubmitter core.BridgeSubmitter,
	appConfig *core.AppConfig,
	oracleDb core.CardanoTxsDb,
	indexerDb indexer.Database,
	chainId string,
	logger hclog.Logger,
) (*ConfirmedBlocksSubmitterImpl, error) {
	latestBlockPoint, err := indexerDb.GetLatestBlockPoint()
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
		chainId:             chainId,
		indexerDb:           indexerDb,
		oracleDb:            oracleDb,
		logger:              logger.Named("confirmed_blocks_submitter_" + chainId),
		latestConfirmedSlot: latestBlockPoint.BlockSlot,
	}, nil
}

func (bs *ConfirmedBlocksSubmitterImpl) StartSubmit() {
	go func() {
		ticker := time.NewTicker(time.Millisecond * time.Duration(bs.appConfig.Bridge.SubmitConfig.ConfirmedBlocksSubmitTime))
		defer ticker.Stop()

		for {
			select {
			case <-bs.ctx.Done():
				return
			case <-ticker.C:
				bs.execute()
			}
		}
	}()
}

func (bs *ConfirmedBlocksSubmitterImpl) execute() error {
	from := bs.latestConfirmedSlot
	if from != 0 {
		from++
	}

	blocks, err := bs.indexerDb.GetConfirmedBlocksFrom(
		from,
		bs.appConfig.Bridge.SubmitConfig.ConfirmedBlocksThreshold)
	if err != nil {
		bs.logger.Error("error getting latest confirmed blocks", "err", err)
		return fmt.Errorf("error getting latest confirmed blocks. err: %w", err)
	}

	var blockCounter = 0
	for _, block := range blocks {
		if !bs.checkIfBlockIsProcessed(block) {
			break
		}
		blockCounter++
	}

	if blockCounter == 0 {
		return nil
	}

	if err := bs.bridgeSubmitter.SubmitConfirmedBlocks(bs.chainId, blocks[:blockCounter]); err != nil {
		bs.logger.Error("error submitting confirmed blocks", "err", err)
		return fmt.Errorf("error submitting confirmed blocks. err %w", err)
	}

	bs.latestConfirmedSlot = blocks[blockCounter-1].Slot

	return nil
}

func (bs *ConfirmedBlocksSubmitterImpl) GetChainId() string {
	return bs.chainId
}

func (bs *ConfirmedBlocksSubmitterImpl) checkIfBlockIsProcessed(block *indexer.CardanoBlock) bool {
	for _, tx := range block.Txs {
		prTx, err := bs.oracleDb.GetProcessedTx(bs.chainId, tx)
		if err != nil {
			bs.logger.Error("error getting processed tx for block", "err", err)
		}

		if prTx == nil {
			return false
		}
	}

	return true
}
