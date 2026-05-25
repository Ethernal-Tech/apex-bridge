package bridge

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_solana/core"
	sendtx "github.com/Ethernal-Tech/solana-infrastructure/sendtx"
	solanaTxsStore "github.com/Ethernal-Tech/solana-infrastructure/tracker/store"
	"github.com/gagliardetto/solana-go"
	"github.com/hashicorp/go-hclog"
)

const (
	MaxRetries = 5

	// Solana slot target is 400ms
	// TTL for solana tx is caped to 150 blocks
	// source: https://solana.com/developers/guides/advanced/confirmation#how-does-transaction-expiration-work
	// we add 50 blocks on top of the ttl to be safe

	TTLOfset = 200
)

type SolanaBridgeDataFetcherImpl struct {
	ctx        context.Context
	bridgeSC   eth.IOracleBridgeSmartContract
	logger     hclog.Logger
	indexerDbs map[string]solanaTxsStore.StorageHandler
	appConfig  *oCore.AppConfig
}

var _ core.SolanaBridgeDataFetcher = (*SolanaBridgeDataFetcherImpl)(nil)

func NewSolanaBridgeDataFetcher(
	ctx context.Context,
	bridgeSC eth.IOracleBridgeSmartContract,
	indexerDbs map[string]solanaTxsStore.StorageHandler,
	logger hclog.Logger,
	appConfig *oCore.AppConfig,
) *SolanaBridgeDataFetcherImpl {
	return &SolanaBridgeDataFetcherImpl{
		ctx:        ctx,
		bridgeSC:   bridgeSC,
		indexerDbs: indexerDbs,
		logger:     logger,
		appConfig:  appConfig,
	}
}

func (df *SolanaBridgeDataFetcherImpl) FetchExpectedTx(chainID string) (*core.BridgeExpectedSolanaTx, error) {
	for retries := 1; retries <= MaxRetries; retries++ {
		lastBatchRawTx, err := df.bridgeSC.GetRawTransactionFromLastBatch(df.ctx, chainID)
		if err == nil {
			if len(lastBatchRawTx) == 0 {
				return nil, nil
			}

			var payload sendtx.SolanaPayload

			err := payload.Unmarshal(lastBatchRawTx)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
			}

			expectedTxMetadata := core.BatchExecutedSolMetadata{
				BridgingTxType: common.BridgingTxTypeBatchExecution,
				BatchNonceID:   payload.BatchID,
			}

			txMetadata, err := core.MarshalSolMetadata(expectedTxMetadata)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal metadata. err: %w", err)
			}

			// 2. TTL
			indexerDB := df.indexerDbs[chainID]
			if indexerDB == nil {
				return nil, fmt.Errorf("indexer db not found for chainID: %s", chainID)
			}

			if len(payload.Blockhash) != 32 || payload.Blockhash == [32]byte{} {
				return nil, fmt.Errorf("failed to convert blockhash to solana.Hash: empty blockhash")
			}

			blockhash := solana.HashFromBytes(payload.Blockhash[:])

			blockNumber, err := indexerDB.GetBlockNumberByBlockhash(blockhash)
			if err != nil {
				return nil, fmt.Errorf("failed to get block by hash. err: %w", err)
			}

			hash := sha256.Sum256(lastBatchRawTx)

			expectedTx := &core.BridgeExpectedSolanaTx{
				ChainID:  chainID,
				Hash:     hash,
				TTL:      blockNumber + TTLOfset + df.appConfig.SolanaChains[chainID].TTLNumberInc,
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
