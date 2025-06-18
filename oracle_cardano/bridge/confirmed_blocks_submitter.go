package bridge

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	oracleCommon "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/hashicorp/go-hclog"
)

type ConfirmedBlocksSubmitterImpl struct {
	bridgeSubmitter oracleCommon.BridgeBlocksSubmitter
	appConfig       *oracleCommon.AppConfig
	chainID         string
	oracleDB        core.CardanoTxsProcessorDB
	indexerDB       indexer.Database
	latestInfo      oracleCommon.BlocksSubmitterInfo
	logger          hclog.Logger
}

var _ oracleCommon.ConfirmedBlocksSubmitter = (*ConfirmedBlocksSubmitterImpl)(nil)

func NewConfirmedBlocksSubmitter(
	bridgeSubmitter oracleCommon.BridgeBlocksSubmitter,
	appConfig *oracleCommon.AppConfig,
	oracleDB core.CardanoTxsProcessorDB,
	indexerDB indexer.Database,
	chainID string,
	logger hclog.Logger,
) (*ConfirmedBlocksSubmitterImpl, error) {
	latestInfo, err := oracleDB.GetBlocksSubmitterInfo(chainID)
	if err != nil {
		return nil, err
	}

	if config := appConfig.CardanoChains[chainID]; config != nil && config.StartSlot > latestInfo.BlockNumOrSlot {
		latestInfo.BlockNumOrSlot = config.StartSlot
		latestInfo.CounterEmpty = 0
	}

	if appConfig.Bridge.SubmitConfig.UpdateFromIndexerDB {
		blockPoint, err := indexerDB.GetLatestBlockPoint()
		if err != nil {
			return nil, fmt.Errorf("failed to create block submitter for %s: %w", chainID, err)
		}

		if latestInfo.BlockNumOrSlot < blockPoint.BlockSlot {
			latestInfo.BlockNumOrSlot = blockPoint.BlockSlot
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
		return fmt.Errorf("error saving info: %w", err)
	}

	bs.latestInfo = latestInfo

	bs.logger.Info("Submitted confirmed blocks",
		"chainID", bs.chainID, "slot", bs.latestInfo.BlockNumOrSlot, "counter", bs.latestInfo.CounterEmpty)

	return nil
}

func (bs *ConfirmedBlocksSubmitterImpl) getBlocksToSubmit(from uint64) (
	result []eth.CardanoBlock, latestInfo oracleCommon.BlocksSubmitterInfo, err error,
) {
	bs.logger.Debug("Executing", "chainID", bs.chainID, "from", from)

	latestInfo = bs.latestInfo

	blocksToSubmit, err := bs.indexerDB.GetConfirmedBlocksFrom(
		from,
		bs.appConfig.Bridge.SubmitConfig.ConfirmedBlocksThreshold)
	if err != nil {
		return result, latestInfo,
			fmt.Errorf("error getting blocks: %w", err)
	}

	for _, block := range blocksToSubmit {
		if len(block.Txs) == 0 {
			latestInfo.BlockNumOrSlot = block.Slot
			latestInfo.CounterEmpty++

			threshold, ok := bs.appConfig.Bridge.SubmitConfig.EmptyBlocksThreshold[bs.chainID]
			if !ok {
				return result, latestInfo, fmt.Errorf("empty blocks threshold not configured for chain: %s", bs.chainID)
			}

			// add empty block only if threshold is reached
			if latestInfo.CounterEmpty < threshold {
				continue
			}
		} else {
			allProccessed, err := bs.checkIfBlockIsProcessed(block.Txs)
			if err != nil {
				return result, latestInfo, err
			} else if !allProccessed {
				latestInfo.CounterEmpty = 0

				break // do not process any more block until this block is fully processed
			}
		}

		latestInfo.CounterEmpty = 0
		latestInfo.BlockNumOrSlot = block.Slot

		result = append(result, eth.CardanoBlock{
			BlockSlot: new(big.Int).SetUint64(block.Slot),
			BlockHash: block.Hash,
		})
	}

	return result, latestInfo, nil
}

func (bs *ConfirmedBlocksSubmitterImpl) checkIfBlockIsProcessed(
	txsHashes []indexer.Hash,
) (bool, error) {
	for _, hash := range txsHashes {
		prTx, err := bs.oracleDB.GetProcessedTx(oracleCommon.DBTxID{
			ChainID: bs.chainID,
			DBKey:   hash[:],
		})
		if err != nil {
			return false, fmt.Errorf("failed to check if txs %s is processed: %w", hash, err)
		}

		if prTx == nil {
			return false, nil
		}
	}

	return true, nil
}
