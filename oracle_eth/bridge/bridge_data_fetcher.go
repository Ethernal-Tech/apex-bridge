package bridge

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/batcher/batcher"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/oracle_eth/core"
	"github.com/Ethernal-Tech/ethgo"
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

func (df *EthBridgeDataFetcherImpl) GetBatchTransactions(
	chainID string, batchID uint64,
) ([]eth.TxDataInfo, error) {
	_, txs, err := df.bridgeSC.GetBatchStatusAndTransactions(df.ctx, chainID, batchID)
	if err != nil {
		df.logger.Error("Failed to retrieve batch transactions", "chainID", chainID, "batchID", batchID, "err", err)

		return nil, err
	}

	df.logger.Info("Batch transactions retrieved", "chainID", chainID, "batchID", batchID, "txs", len(txs))

	return txs, nil
}

func (df *EthBridgeDataFetcherImpl) FetchExpectedTx(chainID string) (*core.BridgeExpectedEthTx, error) {
	for retries := 1; retries <= MaxRetries; retries++ {
		lastBatchRawTx, batchType, err := df.bridgeSC.GetRawTransactionAndBatchTypeFromLastBatch(df.ctx, chainID)
		if err == nil {
			if len(lastBatchRawTx) == 0 {
				return nil, nil
			}

			var txData struct {
				BatchNonceID uint64
				TTL          uint64
			}

			if batchType == uint8(batcher.Normal) {
				tx, err := eth.NewEVMSmartContractTransaction(lastBatchRawTx)
				if err != nil {
					df.logger.Error("Failed to parse evm tx", "rawTx", hex.EncodeToString(lastBatchRawTx), "err", err)

					return nil, fmt.Errorf("failed to parse evm tx. err: %w", err)
				}

				txData.BatchNonceID = tx.BatchNonceID
				txData.TTL = tx.TTL
			} else {
				tx, err := eth.NewEVMValidatorSetChangeTransaction(lastBatchRawTx)
				if err != nil {
					df.logger.Error("Failed to parse validator set change evm tx", "rawTx", hex.EncodeToString(lastBatchRawTx), "err", err)

					return nil, fmt.Errorf("failed to parse evm tx. err: %w", err)
				}

				txData.BatchNonceID = tx.BatchNonceID
				txData.TTL = tx.TTL.Uint64()
			}

			txHash, err := common.Keccak256(lastBatchRawTx)
			if err != nil {
				return nil, fmt.Errorf("failed to create txHash. err: %w", err)
			}

			expectedTxMetadata := core.BatchExecutedEthMetadata{
				BridgingTxType: common.BridgingTxTypeBatchExecution,
				BatchNonceID:   txData.BatchNonceID,
			}

			txMetadata, err := core.MarshalEthMetadata(expectedTxMetadata)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal metadata. err: %w", err)
			}

			expectedTx := &core.BridgeExpectedEthTx{
				ChainID:  chainID,
				Hash:     ethgo.BytesToHash(txHash),
				TTL:      txData.TTL,
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
