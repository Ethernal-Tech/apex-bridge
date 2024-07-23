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
	ctx               context.Context
	bridgeSubmitter   eth_core.BridgeSubmitter
	appConfig         *core.AppConfig
	chainID           string
	indexerDB         eventTrackerStore.EventTrackerStore
	oracleDB          eth_core.EthTxsDB
	logger            hclog.Logger
	latestBlockNumber uint64
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
		ctx:               ctx,
		bridgeSubmitter:   bridgeSubmitter,
		appConfig:         appConfig,
		chainID:           chainID,
		indexerDB:         indexerDB,
		oracleDB:          oracleDB,
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

	var blockCounter = uint64(0)

	bs.logger.Debug("Checking if blocks are processed", "chainID", bs.chainID, "from block", from,
		"to last block", lastProcessedBlock)

	for blockIdx := from; blockIdx <= lastProcessedBlock; blockIdx++ {
		if !bs.checkIfBlockIsProcessed(blockIdx) {
			break
		}

		blockCounter++
	}

	if blockCounter == 0 {
		bs.logger.Debug("No new processed blocks", "chainID", bs.chainID)

		return nil
	}

	if err := bs.bridgeSubmitter.SubmitConfirmedBlocks(bs.chainID, from, blockCounter); err != nil {
		bs.logger.Error("error submitting confirmed blocks", "err", err)

		return fmt.Errorf("error submitting confirmed blocks. err %w", err)
	}

	bs.latestBlockNumber = from + blockCounter - 1
	bs.logger.Info("Submitted confirmed blocks", "chainID", bs.chainID, "latestBlockNumber", bs.latestBlockNumber)

	return nil
}

func (bs *ConfirmedBlocksSubmitterImpl) GetChainID() string {
	return bs.chainID
}

func (bs *ConfirmedBlocksSubmitterImpl) checkIfBlockIsProcessed(blockNumber uint64) bool {
	logs, err := bs.indexerDB.GetLogsByBlockNumber(blockNumber)
	if err != nil {
		bs.logger.Error("error getting logs for", "blockNumber", blockNumber, "err", err)

		return false
	}

	for _, log := range logs {
		prTx, err := bs.oracleDB.GetProcessedTx(bs.chainID, log.TransactionHash)
		if err != nil {
			bs.logger.Error("error getting processed log for block", "err", err)
		}

		if prTx == nil {
			return false
		}
	}

	return true
}
