package bridge

import (
	"context"
	"fmt"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/oracle_solana/core"
	"github.com/hashicorp/go-hclog"
)

const (
	MaxRetries = 5
)

type SolanaBridgeDataFetcherImpl struct {
	ctx      context.Context
	bridgeSC eth.IOracleBridgeSmartContract
	logger   hclog.Logger
}

var _ core.SolanaBridgeDataFetcher = (*SolanaBridgeDataFetcherImpl)(nil)

func NewSolanaBridgeDataFetcher(
	ctx context.Context,
	bridgeSC eth.IOracleBridgeSmartContract,
	logger hclog.Logger,
) *SolanaBridgeDataFetcherImpl {
	return &SolanaBridgeDataFetcherImpl{
		ctx:      ctx,
		bridgeSC: bridgeSC,
		logger:   logger,
	}
}

func (df *SolanaBridgeDataFetcherImpl) FetchExpectedTx(chainID string) (*core.BridgeExpectedSolanaTx, error) {
	for retries := 1; retries <= MaxRetries; retries++ {
		lastBatchRawTx, err := df.bridgeSC.GetRawTransactionFromLastBatch(df.ctx, chainID)
		if err == nil {
			if len(lastBatchRawTx) == 0 {
				return nil, nil
			}

			// wTODO: Parse the raw transaction
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

	return nil, fmt.Errorf("failed to FetchExpectedTx from Bridge SC")
}

func (df *SolanaBridgeDataFetcherImpl) GetBatchTransactions(chainID string, batchID uint64) ([]eth.TxDataInfo, error) {
	_, txs, err := df.bridgeSC.GetBatchStatusAndTransactions(df.ctx, chainID, batchID)
	if err != nil {
		df.logger.Error("Failed to retrieve batch transactions", "chainID", chainID, "batchID", batchID, "err", err)

		return nil, err
	}

	df.logger.Info("Batch transactions retrieved", "chainID", chainID, "batchID", batchID, "txs", len(txs))

	return txs, nil
}
