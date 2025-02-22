package bridge

import (
	"context"
	"fmt"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_eth/core"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestExpectedTxsFetcher(t *testing.T) {
	appConfig := &oCore.AppConfig{
		EthChains: map[string]*oCore.EthChainConfig{
			common.ChainIDStrNexus: {},
		},
	}

	t.Run("NewBridgeDataFetcher", func(t *testing.T) {
		bridgeDataFetcher := &core.EthBridgeDataFetcherMock{}
		expectedTxsFetcher := NewExpectedTxsFetcher(context.Background(), bridgeDataFetcher, appConfig, &core.EthTxsProcessorDBMock{}, hclog.NewNullLogger())

		require.NotNil(t, expectedTxsFetcher)
	})

	t.Run("fetchData nil", func(t *testing.T) {
		bridgeDataFetcher := &core.EthBridgeDataFetcherMock{}
		bridgeDataFetcher.On("FetchExpectedTx", mock.Anything).Return(nil, nil)

		db := &core.EthTxsProcessorDBMock{}
		db.On("GetAllExpectedTxs", mock.Anything, mock.Anything).Return(nil, nil)
		db.On("AddExpectedTxs", mock.Anything).Return(nil)

		expectedTxsFetcher := NewExpectedTxsFetcher(context.Background(), bridgeDataFetcher, appConfig, db, hclog.NewNullLogger())
		require.NotNil(t, expectedTxsFetcher)

		err := expectedTxsFetcher.fetchData()
		require.NoError(t, err)
	})

	t.Run("fetchData err", func(t *testing.T) {
		bridgeDataFetcher := &core.EthBridgeDataFetcherMock{}
		bridgeDataFetcher.On("FetchExpectedTx", mock.Anything).Return(&core.BridgeExpectedEthTx{}, nil)

		db := &core.EthTxsProcessorDBMock{}
		db.On("GetAllExpectedTxs", mock.Anything, mock.Anything).Return(nil, nil)
		db.On("AddExpectedTxs", mock.Anything).Return(fmt.Errorf("test err"))

		expectedTxsFetcher := NewExpectedTxsFetcher(context.Background(), bridgeDataFetcher, appConfig, db, hclog.NewNullLogger())
		require.NotNil(t, expectedTxsFetcher)

		err := expectedTxsFetcher.fetchData()
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to add expected txs")
	})

	t.Run("fetchData success", func(t *testing.T) {
		bridgeDataFetcher := &core.EthBridgeDataFetcherMock{}
		bridgeDataFetcher.On("FetchExpectedTx", mock.Anything).Return(&core.BridgeExpectedEthTx{}, nil)

		db := &core.EthTxsProcessorDBMock{}
		db.On("GetAllExpectedTxs", mock.Anything, mock.Anything).Return(nil, nil)
		db.On("AddExpectedTxs", mock.Anything).Return(nil)

		expectedTxsFetcher := NewExpectedTxsFetcher(context.Background(), bridgeDataFetcher, appConfig, db, hclog.NewNullLogger())
		require.NotNil(t, expectedTxsFetcher)

		err := expectedTxsFetcher.fetchData()
		require.NoError(t, err)
	})
}
