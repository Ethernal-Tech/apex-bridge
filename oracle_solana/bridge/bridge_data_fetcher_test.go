package bridge

import (
	"context"
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_solana/core"
	sol "github.com/Ethernal-Tech/apex-bridge/solana"
	sendtx "github.com/Ethernal-Tech/solana-infrastructure/sendtx"
	solanaTxsStore "github.com/Ethernal-Tech/solana-infrastructure/tracker/store"
	"github.com/gagliardetto/solana-go"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSolanaBridgeDataFetcher(t *testing.T) {
	emptyIndexerDbs := map[string]solanaTxsStore.StorageHandler{}

	appConfig := &oCore.AppConfig{
		SolanaChains: map[string]*oCore.SolanaChainConfig{
			common.ChainIDStrSolana: {
				SolanaChainConfig: sol.SolanaChainConfig{
					TTLNumberInc: 0,
				},
			},
		},
	}

	t.Run("NewSolanaBridgeDataFetcher", func(t *testing.T) {
		bridgeSC := &eth.OracleBridgeSmartContractMock{}
		fetcher := NewSolanaBridgeDataFetcher(context.Background(), bridgeSC, emptyIndexerDbs, hclog.NewNullLogger(), appConfig)

		require.NotNil(t, fetcher)
	})

	t.Run("GetBatchTransactions err", func(t *testing.T) {
		bridgeSC := &eth.OracleBridgeSmartContractMock{}
		bridgeSC.On("GetBatchStatusAndTransactions", mock.Anything, mock.Anything, mock.Anything).
			Return(uint8(0), nil, fmt.Errorf("test err"))

		fetcher := NewSolanaBridgeDataFetcher(context.Background(), bridgeSC, emptyIndexerDbs, hclog.NewNullLogger(), appConfig)
		require.NotNil(t, fetcher)

		_, err := fetcher.GetBatchTransactions(common.ChainIDStrSolana, 1)
		require.Error(t, err)
		require.ErrorContains(t, err, "test err")
	})

	t.Run("GetBatchTransactions valid", func(t *testing.T) {
		bridgeSC := &eth.OracleBridgeSmartContractMock{}
		bridgeSC.On("GetBatchStatusAndTransactions", mock.Anything, mock.Anything, mock.Anything).
			Return(uint8(0), []eth.TxDataInfo{{}}, nil)

		fetcher := NewSolanaBridgeDataFetcher(context.Background(), bridgeSC, emptyIndexerDbs, hclog.NewNullLogger(), appConfig)
		require.NotNil(t, fetcher)

		batchTxs, err := fetcher.GetBatchTransactions(common.ChainIDStrSolana, 1)
		require.NoError(t, err)
		require.Len(t, batchTxs, 1)
	})

	t.Run("FetchExpectedTx nil raw tx", func(t *testing.T) {
		bridgeSC := &eth.OracleBridgeSmartContractMock{}
		bridgeSC.On("GetRawTransactionFromLastBatch").Return(nil, nil)

		fetcher := NewSolanaBridgeDataFetcher(context.Background(), bridgeSC, emptyIndexerDbs, hclog.NewNullLogger(), appConfig)
		require.NotNil(t, fetcher)

		expectedTx, err := fetcher.FetchExpectedTx(common.ChainIDStrSolana)
		require.NoError(t, err)
		require.Nil(t, expectedTx)
	})

	t.Run("FetchExpectedTx error retries exhausted", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		bridgeSC := &eth.OracleBridgeSmartContractMock{}
		bridgeSC.On("GetRawTransactionFromLastBatch").Return(nil, fmt.Errorf("sc error"))

		fetcher := NewSolanaBridgeDataFetcher(ctx, bridgeSC, emptyIndexerDbs, hclog.NewNullLogger(), appConfig)
		require.NotNil(t, fetcher)

		expectedTx, err := fetcher.FetchExpectedTx(common.ChainIDStrSolana)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to FetchExpectedTx from Bridge SC")
		require.Nil(t, expectedTx)
	})

	t.Run("FetchExpectedTx with empty raw tx returns nil", func(t *testing.T) {
		bridgeSC := &eth.OracleBridgeSmartContractMock{}
		bridgeSC.On("GetRawTransactionFromLastBatch").Return([]byte{}, nil)

		fetcher := NewSolanaBridgeDataFetcher(context.Background(), bridgeSC, emptyIndexerDbs, hclog.NewNullLogger(), appConfig)
		require.NotNil(t, fetcher)

		expectedTx, err := fetcher.FetchExpectedTx(common.ChainIDStrSolana)
		require.NoError(t, err)
		require.Nil(t, expectedTx)
	})

	t.Run("FetchExpectedTx payload unmarshal error", func(t *testing.T) {
		rawPayload := []byte{1, 2, 3}

		bridgeSC := &eth.OracleBridgeSmartContractMock{}
		bridgeSC.On("GetRawTransactionFromLastBatch").Return(rawPayload, nil)

		fetcher := NewSolanaBridgeDataFetcher(context.Background(), bridgeSC, emptyIndexerDbs, hclog.NewNullLogger(), appConfig)
		require.NotNil(t, fetcher)

		expectedTx, err := fetcher.FetchExpectedTx(common.ChainIDStrSolana)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to unmarshal payload")
		require.Nil(t, expectedTx)
	})

	t.Run("FetchExpectedTx indexer db is nil", func(t *testing.T) {
		chainID := common.ChainIDStrSolana
		batchID := uint64(42)
		blockNumber := uint64(1_000)
		blockhash := solana.Hash(solana.NewWallet().PublicKey())

		payload := sendtx.SolanaPayload{
			Blockhash: blockhash,
			BatchID:   batchID,
		}
		rawPayload, err := payload.Marshal()
		require.NoError(t, err)

		bridgeSC := &eth.OracleBridgeSmartContractMock{}
		bridgeSC.On("GetRawTransactionFromLastBatch").
			Return(rawPayload, nil).
			Once()

		indexerDB := &solanaTxsStore.MockStorageHandler{}
		indexerDB.On("GetBlockNumberByBlockhash", blockhash).
			Return(blockNumber, nil).
			Once()

		indexerDbs := map[string]solanaTxsStore.StorageHandler{
			chainID: nil, // nil indexer db
		}

		fetcher := NewSolanaBridgeDataFetcher(context.Background(), bridgeSC, indexerDbs, hclog.NewNullLogger(), appConfig)
		require.NotNil(t, fetcher)

		expectedTx, err := fetcher.FetchExpectedTx(chainID)
		require.Error(t, err)
		require.ErrorContains(t, err, "indexer db not found for chainID")
		require.Nil(t, expectedTx)
	})

	t.Run("FetchExpectedTx blockhash conversion error", func(t *testing.T) {
		chainID := common.ChainIDStrSolana
		batchID := uint64(42)

		payload := sendtx.SolanaPayload{
			Blockhash: [32]byte{},
			BatchID:   batchID,
		}
		rawPayload, err := payload.Marshal()
		require.NoError(t, err)

		bridgeSC := &eth.OracleBridgeSmartContractMock{}
		bridgeSC.On("GetRawTransactionFromLastBatch").
			Return(rawPayload, nil).
			Once()

		indexerDB := &solanaTxsStore.MockStorageHandler{}

		indexerDbs := map[string]solanaTxsStore.StorageHandler{
			chainID: indexerDB,
		}

		fetcher := NewSolanaBridgeDataFetcher(context.Background(), bridgeSC, indexerDbs, hclog.NewNullLogger(), appConfig)
		require.NotNil(t, fetcher)

		expectedTx, err := fetcher.FetchExpectedTx(chainID)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to convert blockhash to solana.Hash")
		require.Nil(t, expectedTx)
	})

	t.Run("FetchExpectedTx get block number error", func(t *testing.T) {
		chainID := common.ChainIDStrSolana
		batchID := uint64(42)
		blockhash := solana.Hash(solana.NewWallet().PublicKey())

		payload := sendtx.SolanaPayload{
			Blockhash: blockhash,
			BatchID:   batchID,
		}
		rawPayload, err := payload.Marshal()
		require.NoError(t, err)

		bridgeSC := &eth.OracleBridgeSmartContractMock{}
		bridgeSC.On("GetRawTransactionFromLastBatch").
			Return(rawPayload, nil).
			Once()

		indexerDB := &solanaTxsStore.MockStorageHandler{}
		indexerDB.On("GetBlockNumberByBlockhash", blockhash).
			Return(uint64(0), fmt.Errorf("test err")).
			Once()

		indexerDbs := map[string]solanaTxsStore.StorageHandler{
			chainID: indexerDB,
		}

		fetcher := NewSolanaBridgeDataFetcher(context.Background(), bridgeSC, indexerDbs, hclog.NewNullLogger(), appConfig)
		require.NotNil(t, fetcher)

		expectedTx, err := fetcher.FetchExpectedTx(chainID)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to get block by hash")
		require.Nil(t, expectedTx)
	})

	t.Run("FetchExpectedTx passes", func(t *testing.T) {
		chainID := common.ChainIDStrSolana
		batchID := uint64(42)
		blockNumber := uint64(1_000)
		blockhash := solana.Hash(solana.NewWallet().PublicKey())

		payload := sendtx.SolanaPayload{
			Blockhash: blockhash,
			BatchID:   batchID,
		}
		rawPayload, err := payload.Marshal()
		require.NoError(t, err)

		bridgeSC := &eth.OracleBridgeSmartContractMock{}
		bridgeSC.On("GetRawTransactionFromLastBatch").
			Return(rawPayload, nil).
			Once()

		indexerDB := &solanaTxsStore.MockStorageHandler{}
		indexerDB.On("GetBlockNumberByBlockhash", blockhash).
			Return(blockNumber, nil).
			Once()

		indexerDbs := map[string]solanaTxsStore.StorageHandler{
			chainID: indexerDB,
		}

		fetcher := NewSolanaBridgeDataFetcher(context.Background(), bridgeSC, indexerDbs, hclog.NewNullLogger(), appConfig)
		require.NotNil(t, fetcher)

		expectedTx, err := fetcher.FetchExpectedTx(chainID)
		require.NoError(t, err)
		require.NotNil(t, expectedTx)

		expectedMetadata, err := core.MarshalSolMetadata(core.BatchExecutedSolMetadata{
			BridgingTxType: common.BridgingTxTypeBatchExecution,
			BatchNonceID:   batchID,
		})
		require.NoError(t, err)

		expectedHash := sha256.Sum256(rawPayload)

		require.Equal(t, chainID, expectedTx.ChainID)
		require.Equal(t, expectedHash, expectedTx.Hash)
		require.Equal(t, blockNumber+TTLOfset, expectedTx.TTL)
		require.Equal(t, expectedMetadata, expectedTx.Metadata)
		require.EqualValues(t, 0, expectedTx.Priority)

		bridgeSC.AssertExpectations(t)
		indexerDB.AssertExpectations(t)
	})
}
