package bridge

import (
	"fmt"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestBridgeDataFetcher(t *testing.T) {
	t.Run("NewBridgeDataFetcher", func(t *testing.T) {
		bridgeSC := &eth.OracleBridgeSmartContractMock{}
		bridgeDataFetcher := NewBridgeDataFetcher(bridgeSC, hclog.NewNullLogger())

		require.NotNil(t, bridgeDataFetcher)
	})

	t.Run("Dispose", func(t *testing.T) {
		bridgeSC := &eth.OracleBridgeSmartContractMock{}
		bridgeDataFetcher := NewBridgeDataFetcher(bridgeSC, hclog.NewNullLogger())

		require.NotNil(t, bridgeDataFetcher)
		err := bridgeDataFetcher.Dispose()
		require.NoError(t, err)
	})

	t.Run("FetchExpectedTx nil", func(t *testing.T) {
		bridgeSC := &eth.OracleBridgeSmartContractMock{}
		bridgeSC.On("GetRawTransactionFromLastBatch").Return(nil, nil)
		bridgeDataFetcher := NewBridgeDataFetcher(bridgeSC, hclog.NewNullLogger())

		require.NotNil(t, bridgeDataFetcher)

		expectedTx, err := bridgeDataFetcher.FetchExpectedTx("prime")
		require.NoError(t, err)
		require.Nil(t, expectedTx)
	})

	t.Run("FetchExpectedTx err", func(t *testing.T) {
		bridgeSC := &eth.OracleBridgeSmartContractMock{}
		bridgeSC.On("GetRawTransactionFromLastBatch").Return(nil, fmt.Errorf("test err"))
		bridgeDataFetcher := NewBridgeDataFetcher(bridgeSC, hclog.NewNullLogger())

		require.NotNil(t, bridgeDataFetcher)

		expectedTx, err := bridgeDataFetcher.FetchExpectedTx("prime")
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to FetchExpectedTx from Bridge SC")
		require.Nil(t, expectedTx)
	})

	t.Run("FetchExpectedTx parse tx fail", func(t *testing.T) {
		bridgeSC := &eth.OracleBridgeSmartContractMock{}
		bridgeSC.On("GetRawTransactionFromLastBatch").Return(&eth.LastBatchRawTx{}, nil)
		bridgeDataFetcher := NewBridgeDataFetcher(bridgeSC, hclog.NewNullLogger())

		require.NotNil(t, bridgeDataFetcher)

		expectedTx, err := bridgeDataFetcher.FetchExpectedTx("prime")
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to ParseTxInfo")
		require.Nil(t, expectedTx)
	})

	t.Run("FetchLatestBlockPoint nil", func(t *testing.T) {
		bridgeSC := &eth.OracleBridgeSmartContractMock{}
		bridgeSC.On("GetLastObservedBlock").Return(nil, nil)
		bridgeDataFetcher := NewBridgeDataFetcher(bridgeSC, hclog.NewNullLogger())

		require.NotNil(t, bridgeDataFetcher)

		blockPoint, err := bridgeDataFetcher.FetchLatestBlockPoint("prime")
		require.NoError(t, err)
		require.Nil(t, blockPoint)
	})

	t.Run("FetchLatestBlockPoint err", func(t *testing.T) {
		bridgeSC := &eth.OracleBridgeSmartContractMock{}
		bridgeSC.On("GetLastObservedBlock").Return(nil, fmt.Errorf("test err"))
		bridgeDataFetcher := NewBridgeDataFetcher(bridgeSC, hclog.NewNullLogger())

		require.NotNil(t, bridgeDataFetcher)

		blockPoint, err := bridgeDataFetcher.FetchLatestBlockPoint("prime")
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to FetchLatestBlockPoint from Bridge SC")
		require.Nil(t, blockPoint)
	})

	t.Run("FetchLatestBlockPoint valid", func(t *testing.T) {
		bridgeSC := &eth.OracleBridgeSmartContractMock{}
		bridgeSC.On("GetLastObservedBlock").Return(&eth.CardanoBlock{}, nil)
		bridgeDataFetcher := NewBridgeDataFetcher(bridgeSC, hclog.NewNullLogger())

		require.NotNil(t, bridgeDataFetcher)

		blockPoint, err := bridgeDataFetcher.FetchLatestBlockPoint("prime")
		require.NoError(t, err)
		require.NotNil(t, blockPoint)
	})
}
