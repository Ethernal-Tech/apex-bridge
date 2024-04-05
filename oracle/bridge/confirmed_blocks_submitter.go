package bridge

import (
	"context"
	"fmt"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle/utils"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	indexerDb "github.com/Ethernal-Tech/cardano-infrastructure/indexer/db"
	"github.com/hashicorp/go-hclog"
)

type ConfirmedBlocksSubmitterImpl struct {
	bridgeSubmitter core.BridgeSubmitter
	appConfig       *core.AppConfig
	chainId         string
	indexerDb       indexer.Database
	oracleDb        core.CardanoTxsDb
	logger          hclog.Logger
	ctx             context.Context
	cancelCtx       context.CancelFunc

	latestConfirmedSlot uint64
	errorCh             chan error
}

var _ core.ConfirmedBlocksSubmitter = (*ConfirmedBlocksSubmitterImpl)(nil)

func NewConfirmedBlocksSubmitter(
	bridgeSubmitter core.BridgeSubmitter,
	appConfig *core.AppConfig,
	oracleDb core.CardanoTxsDb,
	chainId string,
	logger hclog.Logger,
) (*ConfirmedBlocksSubmitterImpl, error) {
	if err := utils.CreateDirectoryIfNotExists(appConfig.Settings.DbsPath); err != nil {
		return nil, err
	}

	indexerDb, err := indexerDb.NewDatabaseInit("", appConfig.Settings.DbsPath+chainId+".db")
	if err != nil {
		return nil, err
	}

	latestBlockPoint, err := indexerDb.GetLatestBlockPoint()
	if err != nil {
		return nil, err
	}
	if latestBlockPoint == nil {
		latestBlockPoint = &indexer.BlockPoint{}
	}

	ctx, cancelCtx := context.WithCancel(context.Background())

	return &ConfirmedBlocksSubmitterImpl{
		bridgeSubmitter: bridgeSubmitter,
		appConfig:       appConfig,
		chainId:         chainId,
		indexerDb:       indexerDb,
		oracleDb:        oracleDb,
		logger:          logger,
		ctx:             ctx,
		cancelCtx:       cancelCtx,

		latestConfirmedSlot: latestBlockPoint.BlockSlot,
		errorCh:             make(chan error, 1),
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
				from := bs.latestConfirmedSlot
				if from != 0 {
					from++
				}

				blocks, err := bs.indexerDb.GetConfirmedBlocksFrom(
					from,
					bs.appConfig.Bridge.SubmitConfig.ConfirmedBlocksThreshhold)
				if err != nil {
					bs.errorCh <- fmt.Errorf("error getting latest confirmed blocks err: %v", err)
				}

				var blockCounter = 0
				for _, block := range blocks {
					if !bs.checkIfBlockIsProcessed(block) {
						break
					}
					blockCounter++
				}

				if blockCounter == 0 {
					continue
				}

				if err := bs.bridgeSubmitter.SubmitConfirmedBlocks(bs.chainId, blocks[:blockCounter]); err != nil {
					bs.errorCh <- fmt.Errorf("error submitting confirmed blocks: %v", err)
					continue
				}
				bs.latestConfirmedSlot = blocks[blockCounter-1].Slot
			}
		}
	}()
}

func (bs *ConfirmedBlocksSubmitterImpl) Dispose() error {
	bs.cancelCtx()
	close(bs.errorCh)

	return nil
}

func (bs *ConfirmedBlocksSubmitterImpl) ErrorCh() <-chan error {
	return bs.errorCh
}

func (bs *ConfirmedBlocksSubmitterImpl) GetChainId() string {
	return bs.chainId
}

func (bs *ConfirmedBlocksSubmitterImpl) checkIfBlockIsProcessed(block *indexer.CardanoBlock) bool {
	if len(block.Txs) == 0 {
		return true
	}

	for _, tx := range block.Txs {
		prTx, err := bs.oracleDb.GetProcessedTx(bs.chainId, tx)
		if err != nil {
			bs.errorCh <- fmt.Errorf("error getting processed tx for block err: %v", err)
		}

		if prTx == nil {
			return false
		}
	}

	return true
}
