package bridge

import (
	"context"
	"fmt"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSolanaBridgeDataFetcher(t *testing.T) {
	t.Run("NewSolanaBridgeDataFetcher", func(t *testing.T) {
		bridgeSC := &eth.OracleBridgeSmartContractMock{}
		fetcher := NewSolanaBridgeDataFetcher(context.Background(), bridgeSC, hclog.NewNullLogger())

		require.NotNil(t, fetcher)
	})

	t.Run("GetBatchTransactions err", func(t *testing.T) {
		bridgeSC := &eth.OracleBridgeSmartContractMock{}
		bridgeSC.On("GetBatchStatusAndTransactions", mock.Anything, mock.Anything, mock.Anything).
			Return(uint8(0), nil, fmt.Errorf("test err"))

		fetcher := NewSolanaBridgeDataFetcher(context.Background(), bridgeSC, hclog.NewNullLogger())
		require.NotNil(t, fetcher)

		_, err := fetcher.GetBatchTransactions(common.ChainIDStrSolana, 1)
		require.Error(t, err)
		require.ErrorContains(t, err, "test err")
	})

	t.Run("GetBatchTransactions valid", func(t *testing.T) {
		bridgeSC := &eth.OracleBridgeSmartContractMock{}
		bridgeSC.On("GetBatchStatusAndTransactions", mock.Anything, mock.Anything, mock.Anything).
			Return(uint8(0), []eth.TxDataInfo{{}}, nil)

		fetcher := NewSolanaBridgeDataFetcher(context.Background(), bridgeSC, hclog.NewNullLogger())
		require.NotNil(t, fetcher)

		batchTxs, err := fetcher.GetBatchTransactions(common.ChainIDStrSolana, 1)
		require.NoError(t, err)
		require.Len(t, batchTxs, 1)
	})

	t.Run("FetchExpectedTx nil raw tx", func(t *testing.T) {
		bridgeSC := &eth.OracleBridgeSmartContractMock{}
		bridgeSC.On("GetRawTransactionFromLastBatch").Return(nil, nil)

		fetcher := NewSolanaBridgeDataFetcher(context.Background(), bridgeSC, hclog.NewNullLogger())
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

		fetcher := NewSolanaBridgeDataFetcher(ctx, bridgeSC, hclog.NewNullLogger())
		require.NotNil(t, fetcher)

		expectedTx, err := fetcher.FetchExpectedTx(common.ChainIDStrSolana)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to FetchExpectedTx from Bridge SC")
		require.Nil(t, expectedTx)
	})

	t.Run("FetchExpectedTx with empty raw tx returns nil", func(t *testing.T) {
		bridgeSC := &eth.OracleBridgeSmartContractMock{}
		bridgeSC.On("GetRawTransactionFromLastBatch").Return([]byte{}, nil)

		fetcher := NewSolanaBridgeDataFetcher(context.Background(), bridgeSC, hclog.NewNullLogger())
		require.NotNil(t, fetcher)

		expectedTx, err := fetcher.FetchExpectedTx(common.ChainIDStrSolana)
		require.NoError(t, err)
		require.Nil(t, expectedTx)
	})
}
