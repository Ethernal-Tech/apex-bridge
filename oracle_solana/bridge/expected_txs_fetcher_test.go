package bridge

import (
	"context"
	"fmt"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_solana/core"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestExpectedTxsFetcher(t *testing.T) {
	appConfig := &oCore.AppConfig{
		SolanaChains: map[string]*oCore.SolanaChainConfig{
			common.ChainIDStrSolana: {},
		},
	}

	t.Run("NewExpectedTxsFetcher", func(t *testing.T) {
		bridgeDataFetcher := &core.SolanaBridgeDataFetcherMock{}
		db := &core.SolanaTxsProcessorDBMock{}
		fetcher := NewExpectedTxsFetcher(context.Background(), bridgeDataFetcher, appConfig, db, hclog.NewNullLogger())

		require.NotNil(t, fetcher)
	})

	t.Run("fetchData nil", func(t *testing.T) {
		bridgeDataFetcher := &core.SolanaBridgeDataFetcherMock{}
		bridgeDataFetcher.On("FetchExpectedTx", mock.Anything).Return(nil, nil)

		db := &core.SolanaTxsProcessorDBMock{}
		db.On("GetAllExpectedTxs", mock.Anything, mock.Anything).Return(nil, nil)
		db.On("AddExpectedTxs", mock.Anything).Return(nil)

		fetcher := NewExpectedTxsFetcher(context.Background(), bridgeDataFetcher, appConfig, db, hclog.NewNullLogger())
		require.NotNil(t, fetcher)

		err := fetcher.fetchData()
		require.NoError(t, err)
	})

	t.Run("fetchData err from db AddExpectedTxs", func(t *testing.T) {
		bridgeDataFetcher := &core.SolanaBridgeDataFetcherMock{}
		bridgeDataFetcher.On("FetchExpectedTx", mock.Anything).Return(&core.BridgeExpectedSolanaTx{}, nil)

		db := &core.SolanaTxsProcessorDBMock{}
		db.On("GetAllExpectedTxs", mock.Anything, mock.Anything).Return(nil, nil)
		db.On("AddExpectedTxs", mock.Anything).Return(fmt.Errorf("test err"))

		fetcher := NewExpectedTxsFetcher(context.Background(), bridgeDataFetcher, appConfig, db, hclog.NewNullLogger())
		require.NotNil(t, fetcher)

		err := fetcher.fetchData()
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to add expected txs")
	})

	t.Run("fetchData success", func(t *testing.T) {
		bridgeDataFetcher := &core.SolanaBridgeDataFetcherMock{}
		bridgeDataFetcher.On("FetchExpectedTx", mock.Anything).Return(&core.BridgeExpectedSolanaTx{}, nil)

		db := &core.SolanaTxsProcessorDBMock{}
		db.On("GetAllExpectedTxs", mock.Anything, mock.Anything).Return(nil, nil)
		db.On("AddExpectedTxs", mock.Anything).Return(nil)

		fetcher := NewExpectedTxsFetcher(context.Background(), bridgeDataFetcher, appConfig, db, hclog.NewNullLogger())
		require.NotNil(t, fetcher)

		err := fetcher.fetchData()
		require.NoError(t, err)
	})

	t.Run("fetchData skips chain with existing expected txs", func(t *testing.T) {
		bridgeDataFetcher := &core.SolanaBridgeDataFetcherMock{}

		db := &core.SolanaTxsProcessorDBMock{}
		db.On("GetAllExpectedTxs", mock.Anything, mock.Anything).
			Return([]*core.BridgeExpectedSolanaTx{{ChainID: common.ChainIDStrSolana}}, nil)

		fetcher := NewExpectedTxsFetcher(context.Background(), bridgeDataFetcher, appConfig, db, hclog.NewNullLogger())
		require.NotNil(t, fetcher)

		err := fetcher.fetchData()
		require.NoError(t, err)
	})

	t.Run("fetchData handles GetAllExpectedTxs error", func(t *testing.T) {
		bridgeDataFetcher := &core.SolanaBridgeDataFetcherMock{}

		db := &core.SolanaTxsProcessorDBMock{}
		db.On("GetAllExpectedTxs", mock.Anything, mock.Anything).
			Return(nil, fmt.Errorf("db error"))

		fetcher := NewExpectedTxsFetcher(context.Background(), bridgeDataFetcher, appConfig, db, hclog.NewNullLogger())
		require.NotNil(t, fetcher)

		err := fetcher.fetchData()
		require.NoError(t, err)
	})

	t.Run("fetchData handles FetchExpectedTx error", func(t *testing.T) {
		bridgeDataFetcher := &core.SolanaBridgeDataFetcherMock{}
		bridgeDataFetcher.On("FetchExpectedTx", mock.Anything).Return(nil, fmt.Errorf("fetch error"))

		db := &core.SolanaTxsProcessorDBMock{}
		db.On("GetAllExpectedTxs", mock.Anything, mock.Anything).Return(nil, nil)

		fetcher := NewExpectedTxsFetcher(context.Background(), bridgeDataFetcher, appConfig, db, hclog.NewNullLogger())
		require.NotNil(t, fetcher)

		err := fetcher.fetchData()
		require.NoError(t, err)
	})
}
