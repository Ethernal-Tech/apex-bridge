package bridge

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/hashicorp/go-hclog"
)

const (
	MaxRetries = 5
)

type BridgeDataFetcherImpl struct {
	ctx      context.Context
	bridgeSC eth.IOracleBridgeSmartContract
	logger   hclog.Logger
}

var _ core.BridgeDataFetcher = (*BridgeDataFetcherImpl)(nil)

func NewBridgeDataFetcher(
	ctx context.Context,
	bridgeSC eth.IOracleBridgeSmartContract,
	logger hclog.Logger,
) *BridgeDataFetcherImpl {
	return &BridgeDataFetcherImpl{
		ctx:      ctx,
		bridgeSC: bridgeSC,
		logger:   logger,
	}
}

func (df *BridgeDataFetcherImpl) FetchLatestBlockPoint(chainID string) (*indexer.BlockPoint, error) {
	for retries := 1; retries <= MaxRetries; retries++ {
		block, err := df.bridgeSC.GetLastObservedBlock(df.ctx, chainID)
		if err == nil {
			var blockPoint *indexer.BlockPoint

			if block != nil {
				hash, _ := hex.DecodeString(block.BlockHash)
				blockPoint = &indexer.BlockPoint{
					BlockSlot: block.BlockSlot,
					BlockHash: hash,
				}
			}

			return blockPoint, nil
		} else {
			df.logger.Error("Failed to GetLastObservedBlock from Bridge SC", "err", err)
		}

		time.Sleep(time.Millisecond * 500)
	}

	df.logger.Info("Failed to FetchLatestBlockPoint from Bridge SC", "retries", MaxRetries)

	return nil, fmt.Errorf("failed to FetchLatestBlockPoint from Bridge SC")
}

func (df *BridgeDataFetcherImpl) FetchExpectedTx(chainID string) (*core.BridgeExpectedCardanoTx, error) {
	for retries := 1; retries <= MaxRetries; retries++ {
		lastBatchRawTx, err := df.bridgeSC.GetRawTransactionFromLastBatch(df.ctx, chainID)
		if err == nil {
			if lastBatchRawTx == nil {
				return nil, nil
			}

			rawTx, err := hex.DecodeString(lastBatchRawTx.RawTx)
			if err != nil {
				df.logger.Error("Failed to decode rawTx string", "rawTx", lastBatchRawTx.RawTx, "err", err)

				return nil, fmt.Errorf("failed to decode rawTx string. rawTx: %v. err: %w", lastBatchRawTx.RawTx, err)
			}

			tx, err := indexer.ParseTxInfo(rawTx)
			if err != nil {
				df.logger.Error("Failed to ParseTxInfo", "rawTx", lastBatchRawTx.RawTx, "err", err)

				return nil, fmt.Errorf("failed to ParseTxInfo. err: %w", err)
			}

			expectedTx := &core.BridgeExpectedCardanoTx{
				ChainID:  chainID,
				Hash:     tx.Hash,
				TTL:      tx.TTL,
				Metadata: tx.MetaData,
			}

			return expectedTx, nil
		} else {
			df.logger.Error("Failed to GetExpectedTx from Bridge SC", "err", err)
		}

		time.Sleep(time.Millisecond * 500)
	}

	df.logger.Info("Failed to FetchExpectedTx from Bridge SC", "retries", MaxRetries)

	return nil, fmt.Errorf("failed to FetchExpectedTx from Bridge SC")
}
