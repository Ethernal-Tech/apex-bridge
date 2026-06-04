package bridge

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/eth"
	oracleCommon "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	solanaCore "github.com/Ethernal-Tech/apex-bridge/oracle_solana/core"
	solanaTxsStore "github.com/Ethernal-Tech/solana-infrastructure/tracker/store"
	"github.com/gagliardetto/solana-go"
	"github.com/hashicorp/go-hclog"
)

type ConfirmedBlocksSubmitterImpl struct {
	bridgeSubmitter oracleCommon.BridgeBlocksSubmitter
	appConfig       *oracleCommon.AppConfig
	chainID         string
	oracleDB        solanaCore.SolanaTxsProcessorDB
	indexerDB       solanaTxsStore.StorageHandler
	latestInfo      oracleCommon.BlocksSubmitterInfo
	logger          hclog.Logger
}

var _ oracleCommon.ConfirmedBlocksSubmitter = (*ConfirmedBlocksSubmitterImpl)(nil)

func NewConfirmedBlocksSubmitter(
	bridgeSubmitter oracleCommon.BridgeBlocksSubmitter,
	appConfig *oracleCommon.AppConfig,
	oracleDB solanaCore.SolanaTxsProcessorDB,
	indexerDB solanaTxsStore.StorageHandler,
	chainID string,
	logger hclog.Logger,
) (*ConfirmedBlocksSubmitterImpl, error) {
	latestInfo, err := oracleDB.GetBlocksSubmitterInfo(chainID)
	if err != nil {
		return nil, err
	}

	if appConfig.Bridge.SubmitConfig.UpdateFromIndexerDB {
		latestBlockPoint, err := indexerDB.GetLatestBlockPoint()
		if err != nil {
			return nil, fmt.Errorf("failed to create block submitter for %s: %w", chainID, err)
		}

		if latestBlockPoint.BlockNumber > 0 {
			latestInfo.BlockNumOrSlot = latestBlockPoint.BlockNumber
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
		"chainID", bs.chainID, "slot", bs.latestInfo.BlockNumOrSlot, "counter", bs.latestInfo.CounterEmpty)

	return nil
}

func (bs *ConfirmedBlocksSubmitterImpl) getBlocksToSubmit(from uint64) (
	result []eth.CardanoBlock, latestInfo oracleCommon.BlocksSubmitterInfo, err error,
) {
	bs.logger.Debug("Executing", "chainID", bs.chainID, "from", from)

	latestInfo = bs.latestInfo

	latestBlockPoint, err := bs.indexerDB.GetLatestBlockPoint()
	if err != nil {
		return result, latestInfo, fmt.Errorf("error getting latest block point: %w", err)
	}

	if latestBlockPoint == nil {
		return result, latestInfo, nil
	}

	if latestBlockPoint.BlockNumber == 0 {
		return result, latestInfo, nil
	}

	if latestBlockPoint.BlockNumber < from {
		return result, latestInfo, nil
	}

	//nolint:gosec
	to := min(latestBlockPoint.BlockNumber, from+uint64(bs.appConfig.Bridge.SubmitConfig.ConfirmedBlocksThreshold)-1)

	for blockNum := from; blockNum <= to; blockNum++ {
		events, err := bs.indexerDB.GetEventsByBlockNumber(blockNum)
		if err != nil {
			return result, latestInfo, fmt.Errorf("failed to get events for slot %d: %w", blockNum, err)
		}

		if len(events) == 0 {
			latestInfo.BlockNumOrSlot = blockNum
			latestInfo.CounterEmpty++

			threshold, ok := bs.appConfig.Bridge.SubmitConfig.EmptyBlocksThreshold[bs.chainID]
			if !ok {
				return result, latestInfo, fmt.Errorf("empty blocks threshold not configured for chain: %s", bs.chainID)
			}

			if threshold > uint(math.MaxInt) {
				return result, latestInfo, fmt.Errorf("threshold too large: %d", threshold)
			}

			if latestInfo.CounterEmpty < int(threshold) {
				continue
			}
		} else {
			allProcessed, err := bs.checkIfBlockIsProcessed(events)
			if err != nil {
				return result, latestInfo, err
			} else if !allProcessed {
				latestInfo.CounterEmpty = 0

				break
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
	events []solanaTxsStore.EventRecord,
) (bool, error) {
	seen := make(map[string]struct{}, len(events))

	for _, event := range events {
		if _, exists := seen[event.TxSignature]; exists {
			continue
		}

		seen[event.TxSignature] = struct{}{}

		txSignature, err := solana.SignatureFromBase58(event.TxSignature)
		if err != nil {
			return false, fmt.Errorf("failed to parse tx signature %s: %w", event.TxSignature, err)
		}

		prTx, err := bs.oracleDB.GetProcessedTx(oracleCommon.DBTxID{
			ChainID: bs.chainID,
			DBKey:   txSignature[:],
		})
		if err != nil {
			return false, fmt.Errorf("failed to check if tx %s is processed: %w", event.TxSignature, err)
		}

		if prTx == nil {
			return false, nil
		}
	}

	return true, nil
}
