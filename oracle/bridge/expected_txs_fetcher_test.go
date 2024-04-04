package bridge

import (
	"testing"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestExpectedTxsFetcher(t *testing.T) {
	appConfig := &core.AppConfig{
		CardanoChains: map[string]*core.CardanoChainConfig{
			"prime": {},
		},
	}

	t.Run("NewBridgeDataFetcher", func(t *testing.T) {
		bridgeDataFetcher := &core.BridgeDataFetcherMock{}
		expectedTxsFetcher := NewExpectedTxsFetcher(bridgeDataFetcher, appConfig, &core.CardanoTxsProcessorDbMock{}, hclog.NewNullLogger())

		require.NotNil(t, expectedTxsFetcher)
	})

	t.Run("Stop", func(t *testing.T) {
		bridgeDataFetcher := &core.BridgeDataFetcherMock{}
		bridgeDataFetcher.On("FetchExpectedTx").Return(nil, nil)
		db := &core.CardanoTxsProcessorDbMock{}
		db.On("GetExpectedTxs").Return(nil, nil)
		db.On("AddExpectedTxs").Return(nil)

		expectedTxsFetcher := NewExpectedTxsFetcher(bridgeDataFetcher, appConfig, db, hclog.NewNullLogger())
		require.NotNil(t, expectedTxsFetcher)

		go expectedTxsFetcher.Start()
		time.Sleep(100 * time.Millisecond)
		err := expectedTxsFetcher.Stop()
		require.NoError(t, err)
	})

	t.Run("fetchData", func(t *testing.T) {
		bridgeDataFetcher := &core.BridgeDataFetcherMock{}
		bridgeDataFetcher.On("FetchExpectedTx").Return(nil, nil)
		db := &core.CardanoTxsProcessorDbMock{}
		db.On("GetExpectedTxs").Return(nil, nil)
		db.On("AddExpectedTxs").Return(nil)

		expectedTxsFetcher := NewExpectedTxsFetcher(bridgeDataFetcher, appConfig, db, hclog.NewNullLogger())
		require.NotNil(t, expectedTxsFetcher)

		err := expectedTxsFetcher.fetchData()
		require.NoError(t, err)
	})
}
