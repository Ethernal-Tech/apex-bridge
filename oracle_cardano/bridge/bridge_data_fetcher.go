package bridge

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/hashicorp/go-hclog"
)

const (
	MaxRetries = 5
)

type CardanoBridgeDataFetcherImpl struct {
	ctx      context.Context
	bridgeSC eth.IOracleBridgeSmartContract
	logger   hclog.Logger
}

var _ core.CardanoBridgeDataFetcher = (*CardanoBridgeDataFetcherImpl)(nil)

func NewCardanoBridgeDataFetcher(
	ctx context.Context,
	bridgeSC eth.IOracleBridgeSmartContract,
	logger hclog.Logger,
) *CardanoBridgeDataFetcherImpl {
	return &CardanoBridgeDataFetcherImpl{
		ctx:      ctx,
		bridgeSC: bridgeSC,
		logger:   logger,
	}
}

func (df *CardanoBridgeDataFetcherImpl) GetBatchTransactions(
	chainID string, batchID uint64,
) ([]eth.TxDataInfo, error) {
	txs, _, err := df.bridgeSC.GetBatchTransactions(df.ctx, chainID, batchID)
	if err != nil {
		df.logger.Error("Failed to retrieve batch transactions", "chainID", chainID, "batchID", batchID, "err", err)

		return nil, err
	}

	df.logger.Info("Batch transactions retrieved", "chainID", chainID, "batchID", batchID, "txs", len(txs))

	return txs, nil
}

func (df *CardanoBridgeDataFetcherImpl) FetchLatestBlockPoint(chainID string) (*indexer.BlockPoint, error) {
	for retries := 1; retries <= MaxRetries; retries++ {
		block, err := df.bridgeSC.GetLastObservedBlock(df.ctx, chainID)
		if err == nil {
			var blockPoint *indexer.BlockPoint

			if block.BlockSlot != nil && block.BlockSlot.BitLen() > 0 {
				blockPoint = &indexer.BlockPoint{
					BlockSlot: block.BlockSlot.Uint64(),
					BlockHash: block.BlockHash,
				}
			}

			df.logger.Debug("FetchLatestBlockPoint", "for chainID", chainID, "blockPoint", blockPoint)

			return blockPoint, nil
		} else {
			df.logger.Error("Failed to GetLastObservedBlock from Bridge SC", "err", err)
		}

		select {
		case <-df.ctx.Done():
			return nil, df.ctx.Err()
		case <-time.After(time.Millisecond * 500):
		}
	}

	df.logger.Error("Failed to FetchLatestBlockPoint from Bridge SC", "for chainID", chainID, "retries", MaxRetries)

	return nil, fmt.Errorf("failed to FetchLatestBlockPoint from Bridge SC")
}

func (df *CardanoBridgeDataFetcherImpl) FetchExpectedTx(chainID string) (*core.BridgeExpectedCardanoTx, error) {
	for retries := 1; retries <= MaxRetries; retries++ {
		lastBatchRawTx, err := df.bridgeSC.GetRawTransactionFromLastBatch(df.ctx, chainID)
		if err == nil {
			if len(lastBatchRawTx) == 0 {
				return nil, nil
			}

			info, err := common.ParseTxInfo(lastBatchRawTx, false)
			if err != nil {
				df.logger.Error("Failed to ParseTxInfo", "rawTx", hex.EncodeToString(lastBatchRawTx), "err", err)

				return nil, fmt.Errorf("failed to ParseTxInfo. err: %w", err)
			}

			expectedTx := &core.BridgeExpectedCardanoTx{
				ChainID:  chainID,
				Hash:     indexer.NewHashFromHexString(info.Hash),
				TTL:      info.TTL,
				Metadata: info.MetaData,
				Priority: 0,
			}

			df.logger.Debug("FetchExpectedTx", "for chainID", chainID, "expectedTx", expectedTx)

			return expectedTx, nil
		} else {
			df.logger.Error("Failed to GetExpectedTx from Bridge SC", "err", err)
		}

		select {
		case <-df.ctx.Done():
			return nil, df.ctx.Err()
		case <-time.After(time.Millisecond * 500):
		}
	}

	df.logger.Error("Failed to FetchExpectedTx from Bridge SC", "for chainID", chainID, "retries", MaxRetries)

	return nil, fmt.Errorf("failed to FetchExpectedTx from Bridge SC")
}
