package bridge

import (
	"context"
	"fmt"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/eth_oracle/core"
	"github.com/hashicorp/go-hclog"
)

const (
	MaxRetries = 5
)

type EthBridgeDataFetcherImpl struct {
	ctx      context.Context
	bridgeSC eth.IOracleBridgeSmartContract
	logger   hclog.Logger
}

var _ core.EthBridgeDataFetcher = (*EthBridgeDataFetcherImpl)(nil)

func NewEthBridgeDataFetcher(
	ctx context.Context,
	bridgeSC eth.IOracleBridgeSmartContract,
	logger hclog.Logger,
) *EthBridgeDataFetcherImpl {
	return &EthBridgeDataFetcherImpl{
		ctx:      ctx,
		bridgeSC: bridgeSC,
		logger:   logger,
	}
}

func (df *EthBridgeDataFetcherImpl) FetchExpectedTx(chainID string) (*core.BridgeExpectedEthTx, error) {
	for retries := 1; retries <= MaxRetries; retries++ {
		lastBatchRawTx, err := df.bridgeSC.GetRawTransactionFromLastBatch(df.ctx, chainID)
		if err == nil {
			if len(lastBatchRawTx) == 0 {
				return nil, nil
			}
			// a TODO: finish this
			/*
				tx, err := indexer.ParseTxInfo(lastBatchRawTx)
				if err != nil {
					df.logger.Error("Failed to ParseTxInfo", "rawTx", hex.EncodeToString(lastBatchRawTx), "err", err)

					return nil, fmt.Errorf("failed to ParseTxInfo. err: %w", err)
				}

				expectedTx := &core.BridgeExpectedEthTx{
					ChainID:  chainID,
					Hash:     indexer.NewHashFromHexString(tx.Hash),
					TTL:      tx.TTL,
					Metadata: tx.MetaData,
					Priority: 0,
				}

				df.logger.Debug("FetchExpectedTx", "for chainID", chainID, "expectedTx", expectedTx)

				return expectedTx, nil
			*/

			return nil, nil
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
