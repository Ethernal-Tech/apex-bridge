package bridge

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/eth_oracle/core"
	"github.com/Ethernal-Tech/ethgo"
	"github.com/hashicorp/go-hclog"
	// 	"github.com/Ethernal-Tech/ethgo"
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

			tx, err := eth.NewEVMSmartContractTransaction(lastBatchRawTx)
			if err != nil {
				df.logger.Error("Failed to parse evm tx", "rawTx", hex.EncodeToString(lastBatchRawTx), "err", err)

				return nil, fmt.Errorf("failed to parse evm tx. err: %w", err)
			}

			txHash, err := common.Keccak256(lastBatchRawTx)
			if err != nil {
				return nil, fmt.Errorf("failed to create txHash. err: %w", err)
			}

			expectedTxMetada := common.BatchExecutedMetadata{
				BridgingTxType: common.BridgingTxTypeBatchExecution,
				BatchNonceID:   tx.BatchNonceID,
			}

			txMetadata, err := common.MarshalMetadata(common.MetadataEncodingTypeJSON, expectedTxMetada)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal metadata. err: %w", err)
			}

			expectedTx := &core.BridgeExpectedEthTx{
				ChainID:  chainID,
				Hash:     ethgo.BytesToHash(txHash),
				TTL:      tx.TTL,
				Metadata: txMetadata,
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
