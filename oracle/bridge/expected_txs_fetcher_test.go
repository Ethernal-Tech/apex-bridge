package bridge

import (
	"context"
	"fmt"
	"testing"

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
		expectedTxsFetcher := NewExpectedTxsFetcher(context.Background(), bridgeDataFetcher, appConfig, &core.CardanoTxsProcessorDbMock{}, hclog.NewNullLogger())

		require.NotNil(t, expectedTxsFetcher)
	})

	t.Run("fetchData nil", func(t *testing.T) {
		bridgeDataFetcher := &core.BridgeDataFetcherMock{}
		bridgeDataFetcher.On("FetchExpectedTx").Return(nil, nil)
		db := &core.CardanoTxsProcessorDbMock{}
		db.On("GetExpectedTxs").Return(nil, nil)
		db.On("AddExpectedTxs").Return(nil)

		expectedTxsFetcher := NewExpectedTxsFetcher(context.Background(), bridgeDataFetcher, appConfig, db, hclog.NewNullLogger())
		require.NotNil(t, expectedTxsFetcher)

		err := expectedTxsFetcher.fetchData()
		require.NoError(t, err)
	})

	t.Run("fetchData err", func(t *testing.T) {
		bridgeDataFetcher := &core.BridgeDataFetcherMock{}
		bridgeDataFetcher.On("FetchExpectedTx").Return(&core.BridgeExpectedCardanoTx{}, nil)
		db := &core.CardanoTxsProcessorDbMock{}
		db.On("GetExpectedTxs").Return(nil, nil)
		db.On("AddExpectedTxs").Return(fmt.Errorf("test err"))

		expectedTxsFetcher := NewExpectedTxsFetcher(context.Background(), bridgeDataFetcher, appConfig, db, hclog.NewNullLogger())
		require.NotNil(t, expectedTxsFetcher)

		err := expectedTxsFetcher.fetchData()
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to add expected txs")
	})
}
