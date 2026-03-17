package bridge

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/oracle_solana/core"
	solanaTxsStore "github.com/Ethernal-Tech/solana-infrastructure/tracker/store"
	wallet "github.com/Ethernal-Tech/solana-infrastructure/wallet"
	binary "github.com/gagliardetto/binary"
	"github.com/hashicorp/go-hclog"
)

const (
	MaxRetries = 5

	// Solana slot target is 400ms
	// TTL for solana tx is caped to 150 blocks ~ 60-90 seconds
	// we take 150 seconds as TTL offset => 150s/400ms = 375 slots

	// source: https://solana.com/developers/guides/advanced/confirmation#how-does-transaction-expiration-work
	TTLOfset = 375
)

type SolanaBridgeDataFetcherImpl struct {
	ctx        context.Context
	bridgeSC   eth.IOracleBridgeSmartContract
	logger     hclog.Logger
	indexerDbs map[string]solanaTxsStore.StorageHandler
}

var _ core.SolanaBridgeDataFetcher = (*SolanaBridgeDataFetcherImpl)(nil)

func NewSolanaBridgeDataFetcher(
	ctx context.Context,
	bridgeSC eth.IOracleBridgeSmartContract,
	indexerDbs map[string]solanaTxsStore.StorageHandler,
	logger hclog.Logger,
) *SolanaBridgeDataFetcherImpl {
	return &SolanaBridgeDataFetcherImpl{
		ctx:        ctx,
		bridgeSC:   bridgeSC,
		indexerDbs: indexerDbs,
		logger:     logger,
	}
}

func (df *SolanaBridgeDataFetcherImpl) FetchExpectedTx(chainID string) (*core.BridgeExpectedSolanaTx, error) {
	for retries := 1; retries <= MaxRetries; retries++ {
		lastBatchRawTx, err := df.bridgeSC.GetRawTransactionFromLastBatch(df.ctx, chainID)
		if err == nil {
			if len(lastBatchRawTx) == 0 {
				return nil, nil
			}

			unmarshaledTx, err := wallet.UnmarshalTransaction(lastBatchRawTx)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal transaction. err: %w", err)
			}

			if len(unmarshaledTx.Message.Instructions) != 1 {
				return nil, fmt.Errorf("expected 1 instruction, got %d", len(unmarshaledTx.Message.Instructions))
			}

			// 1. Batch ID
			decoder := binary.NewBorshDecoder(unmarshaledTx.Message.Instructions[0].Data[8:])

			var batchNonceID uint64

			err = decoder.Decode(&batchNonceID)
			if err != nil {
				return nil, fmt.Errorf("failed to decode batchID. err: %w", err)
			}

			expectedTxMetadata := core.BatchExecutedSolMetadata{
				BridgingTxType: common.BridgingTxTypeBatchExecution,
				BatchNonceID:   batchNonceID,
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

			slot, err := indexerDB.GetSlotByBlockhash(unmarshaledTx.Message.RecentBlockhash)
			if err != nil {
				return nil, fmt.Errorf("failed to get block by hash. err: %w", err)
			}

			hash := sha256.Sum256(lastBatchRawTx)

			expectedTx := &core.BridgeExpectedSolanaTx{
				ChainID:  chainID,
				Hash:     hash,
				TTL:      slot + TTLOfset,
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
