package bridge

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/eth"
	oracleCommon "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	ethCore "github.com/Ethernal-Tech/apex-bridge/oracle_eth/core"
	"github.com/Ethernal-Tech/ethgo"

	eventTrackerStore "github.com/Ethernal-Tech/blockchain-event-tracker/store"
	"github.com/hashicorp/go-hclog"
)

type ConfirmedBlocksSubmitterImpl struct {
	bridgeSubmitter oracleCommon.BridgeBlocksSubmitter
	appConfig       *oracleCommon.AppConfig
	chainID         string
	oracleDB        ethCore.EthTxsProcessorDB
	indexerDB       eventTrackerStore.EventTrackerStore
	latestInfo      oracleCommon.BlocksSubmitterInfo
	logger          hclog.Logger
}

var _ oracleCommon.ConfirmedBlocksSubmitter = (*ConfirmedBlocksSubmitterImpl)(nil)

func NewConfirmedBlocksSubmitter(
	bridgeSubmitter oracleCommon.BridgeBlocksSubmitter,
	appConfig *oracleCommon.AppConfig,
	oracleDB ethCore.EthTxsProcessorDB,
	indexerDB eventTrackerStore.EventTrackerStore,
	chainID string,
	logger hclog.Logger,
) (*ConfirmedBlocksSubmitterImpl, error) {
	latestInfo, err := oracleDB.GetBlocksSubmitterInfo(chainID)
	if err != nil {
		return nil, err
	}

	if config := appConfig.EthChains[chainID]; config != nil && config.StartBlockNumber > latestInfo.BlockNumOrSlot {
		latestInfo.BlockNumOrSlot = config.StartBlockNumber
		latestInfo.CounterEmpty = 0
	}

	if appConfig.Bridge.SubmitConfig.UpdateFromIndexerDB {
		blockNum, err := indexerDB.GetLastProcessedBlock()
		if err != nil {
			return nil, fmt.Errorf("failed to create block submitter for %s: %w", chainID, err)
		}

		if latestInfo.BlockNumOrSlot < blockNum {
			latestInfo.BlockNumOrSlot = blockNum
			latestInfo.CounterEmpty = 0
		}
	}

	return &ConfirmedBlocksSubmitterImpl{
		bridgeSubmitter: bridgeSubmitter,
		appConfig:       appConfig,
		chainID:         chainID,
		oracleDB:        oracleDB,
		indexerDB:       indexerDB,
		latestInfo:      latestInfo,
		logger:          logger.Named("confirmed_blocks_submitter_" + chainID),
	}, nil
}

func (bs *ConfirmedBlocksSubmitterImpl) Start(ctx context.Context) {
	waitTime := time.Millisecond * time.Duration(bs.appConfig.Bridge.SubmitConfig.ConfirmedBlocksSubmitTime)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(waitTime):
				if err := bs.execute(); err != nil {
					bs.logger.Error("error while executing", "chainID", bs.chainID, "err", err)
				}
			}
		}
	}()
}

func (bs *ConfirmedBlocksSubmitterImpl) execute() error {
	from := bs.latestInfo.BlockNumOrSlot
	if from != 0 {
		from++
	}

	blocksToSubmit, latestInfo, err := bs.getBlocksToSubmit(from)
	if err != nil {
		return err
	}

	if err := bs.bridgeSubmitter.SubmitBlocks(bs.chainID, blocksToSubmit); err != nil {
		return fmt.Errorf("error submitting blocks: %w", err)
	}

	if err := bs.oracleDB.SetBlocksSubmitterInfo(bs.chainID, latestInfo); err != nil {
		return fmt.Errorf("error saving confirmed blocks. err %w", err)
	}

	bs.latestInfo = latestInfo

	bs.logger.Info("Submitted confirmed blocks",
		"chainID", bs.chainID, "block", bs.latestInfo.BlockNumOrSlot, "counter", bs.latestInfo.CounterEmpty)

	return nil
}

func (bs *ConfirmedBlocksSubmitterImpl) getBlocksToSubmit(from uint64) (
	result []eth.CardanoBlock, latestInfo oracleCommon.BlocksSubmitterInfo, err error,
) {
	bs.logger.Debug("Executing", "chainID", bs.chainID, "from", from)

	latestInfo = bs.latestInfo

	lastProcessedBlock, err := bs.indexerDB.GetLastProcessedBlock()
	if err != nil {
		return result, latestInfo, fmt.Errorf("error getting blocks: %w", err)
	}

	if lastProcessedBlock < from {
		return result, latestInfo, nil
	}

	//nolint:gosec
	to := min(lastProcessedBlock, from+uint64(bs.appConfig.Bridge.SubmitConfig.ConfirmedBlocksThreshold)-1)

	for blockNum := from; blockNum <= to; blockNum++ {
		logs, err := bs.indexerDB.GetLogsByBlockNumber(blockNum)
		if err != nil {
			return result, latestInfo, fmt.Errorf("failed to get logs for block %d: %w", blockNum, err)
		}

		if len(logs) == 0 {
			latestInfo.CounterEmpty++
			// add empty block only if threshold is reached
			if latestInfo.CounterEmpty < bs.appConfig.Bridge.SubmitConfig.EmptyBlocksThreshold {
				continue
			}
		} else {
			allProccessed, err := bs.checkIfBlockIsProcessed(logs)
			if err != nil {
				return result, latestInfo, err
			} else if !allProccessed {
				latestInfo.CounterEmpty = 0

				break // do not process any more block until this block is fully processed
			}
		}

		latestInfo.CounterEmpty = 0
		latestInfo.BlockNumOrSlot = blockNum

		result = append(result, eth.CardanoBlock{
			BlockSlot: new(big.Int).SetUint64(blockNum),
		})
	}

	return result, latestInfo, nil
}

func (bs *ConfirmedBlocksSubmitterImpl) checkIfBlockIsProcessed(
	logs []*ethgo.Log,
) (bool, error) {
	for _, tx := range logs {
		prTx, err := bs.oracleDB.GetProcessedTx(oracleCommon.DBTxID{
			ChainID: bs.chainID,
			DBKey:   tx.TransactionHash[:],
		})
		if err != nil {
			return false, fmt.Errorf("failed to check if txs %s is processed: %w", tx.TransactionHash, err)
		}

		if prTx == nil {
			return false, nil
		}
	}

	return true, nil
}
