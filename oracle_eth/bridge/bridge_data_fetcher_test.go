package bridge

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/oracle_eth/core"
	"github.com/Ethernal-Tech/ethgo"
	goEthCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestEthBridgeDataFetcher(t *testing.T) {
	t.Run("NewBridgeDataFetcher", func(t *testing.T) {
		bridgeSC := &eth.OracleBridgeSmartContractMock{}
		bridgeDataFetcher := NewEthBridgeDataFetcher(context.Background(), bridgeSC, hclog.NewNullLogger())

		require.NotNil(t, bridgeDataFetcher)
	})

	t.Run("GetBatchTransactions err", func(t *testing.T) {
		bridgeSC := &eth.OracleBridgeSmartContractMock{}
		bridgeSC.On("GetBatchTransactions", mock.Anything, mock.Anything, mock.Anything).
			Return(nil, uint8(0), fmt.Errorf("test err"))

		bridgeDataFetcher := NewEthBridgeDataFetcher(context.Background(), bridgeSC, hclog.NewNullLogger())
		require.NotNil(t, bridgeDataFetcher)

		_, err := bridgeDataFetcher.GetBatchTransactions(common.ChainIDStrPrime, 1)
		require.Error(t, err)
		require.ErrorContains(t, err, "test err")
	})

	t.Run("GetBatchTransactions valid", func(t *testing.T) {
		bridgeSC := &eth.OracleBridgeSmartContractMock{}
		bridgeSC.On("GetBatchTransactions", mock.Anything, mock.Anything, mock.Anything).
			Return([]eth.TxDataInfo{{}}, uint8(0), nil)

		bridgeDataFetcher := NewEthBridgeDataFetcher(context.Background(), bridgeSC, hclog.NewNullLogger())
		require.NotNil(t, bridgeDataFetcher)

		batchTxs, err := bridgeDataFetcher.GetBatchTransactions(common.ChainIDStrPrime, 1)
		require.NoError(t, err)
		require.Len(t, batchTxs, 1)
	})

	t.Run("FetchExpectedTx nil", func(t *testing.T) {
		bridgeSC := &eth.OracleBridgeSmartContractMock{}
		bridgeSC.On("GetRawTransactionFromLastBatch").Return(nil, nil)
		bridgeDataFetcher := NewEthBridgeDataFetcher(context.Background(), bridgeSC, hclog.NewNullLogger())

		require.NotNil(t, bridgeDataFetcher)

		expectedTx, err := bridgeDataFetcher.FetchExpectedTx(common.ChainIDStrNexus)
		require.NoError(t, err)
		require.Nil(t, expectedTx)
	})

	t.Run("FetchExpectedTx err", func(t *testing.T) {
		bridgeSC := &eth.OracleBridgeSmartContractMock{}
		bridgeSC.On("GetRawTransactionFromLastBatch").Return(nil, fmt.Errorf("test err"))
		bridgeDataFetcher := NewEthBridgeDataFetcher(context.Background(), bridgeSC, hclog.NewNullLogger())

		require.NotNil(t, bridgeDataFetcher)

		expectedTx, err := bridgeDataFetcher.FetchExpectedTx(common.ChainIDStrNexus)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to FetchExpectedTx from Bridge SC")
		require.Nil(t, expectedTx)
	})

	t.Run("FetchExpectedTx parse tx fail", func(t *testing.T) {
		bridgeSC := &eth.OracleBridgeSmartContractMock{}
		bridgeSC.On("GetRawTransactionFromLastBatch").Return([]byte{12, 33}, nil)
		bridgeDataFetcher := NewEthBridgeDataFetcher(context.Background(), bridgeSC, hclog.NewNullLogger())

		require.NotNil(t, bridgeDataFetcher)

		expectedTx, err := bridgeDataFetcher.FetchExpectedTx(common.ChainIDStrNexus)
		require.Error(t, err)
		require.Nil(t, expectedTx)
	})

	t.Run("FetchExpectedTx success", func(t *testing.T) {
		bridgeSC := &eth.OracleBridgeSmartContractMock{}

		ethTx := eth.EVMSmartContractTransaction{
			BatchNonceID: 1,
			TTL:          2,
			FeeAmount:    big.NewInt(1),
			Receivers: []eth.EVMSmartContractTransactionReceiver{
				{
					Address: goEthCommon.Address([20]byte{1, 2}),
					Amount:  big.NewInt(1),
				},
			},
		}

		ethTxBytes, err := ethTx.Pack()
		require.NoError(t, err)

		txHash, err := common.Keccak256(ethTxBytes)
		require.NoError(t, err)

		bridgeSC.On("GetRawTransactionFromLastBatch").Return(ethTxBytes, nil)
		bridgeDataFetcher := NewEthBridgeDataFetcher(context.Background(), bridgeSC, hclog.NewNullLogger())

		require.NotNil(t, bridgeDataFetcher)

		expectedTx, err := bridgeDataFetcher.FetchExpectedTx(common.ChainIDStrNexus)
		require.NoError(t, err)
		require.NotNil(t, expectedTx)
		require.Equal(t, ethTx.TTL, expectedTx.TTL)
		require.Equal(t, ethgo.BytesToHash(txHash), expectedTx.Hash)

		metadata, err := core.UnmarshalEthMetadata[core.BatchExecutedEthMetadata](expectedTx.Metadata)
		require.NoError(t, err)
		require.NotNil(t, metadata)
		require.Equal(t, ethTx.BatchNonceID, metadata.BatchNonceID)
		require.Equal(t, common.BridgingTxTypeBatchExecution, metadata.BridgingTxType)
	})
}
