package bridge

import (
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

	t.Run("FetchExpectedTx", func(t *testing.T) {
		bridgeSC := &eth.OracleBridgeSmartContractMock{}
		bridgeSC.On("GetExpectedTx").Return("", nil)
		bridgeDataFetcher := NewBridgeDataFetcher(bridgeSC, hclog.NewNullLogger())

		require.NotNil(t, bridgeDataFetcher)

		expectedTx, err := bridgeDataFetcher.FetchExpectedTx("prime")
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to ParseTxInfo")
		require.Nil(t, expectedTx)
	})

	t.Run("FetchLatestBlockPoint", func(t *testing.T) {
		bridgeSC := &eth.OracleBridgeSmartContractMock{}
		bridgeSC.On("GetLastObservedBlock").Return(nil, nil)
		bridgeDataFetcher := NewBridgeDataFetcher(bridgeSC, hclog.NewNullLogger())

		require.NotNil(t, bridgeDataFetcher)

		blockPoint, err := bridgeDataFetcher.FetchLatestBlockPoint("prime")
		require.NoError(t, err)
		require.Nil(t, blockPoint)
	})
}
