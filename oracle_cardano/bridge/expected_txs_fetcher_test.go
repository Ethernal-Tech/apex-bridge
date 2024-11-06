package bridge

import (
	"context"
	"fmt"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestExpectedTxsFetcher(t *testing.T) {
	appConfig := &cCore.AppConfig{
		CardanoChains: map[string]*cCore.CardanoChainConfig{
			common.ChainIDStrPrime: {},
		},
	}

	t.Run("NewBridgeDataFetcher", func(t *testing.T) {
		bridgeDataFetcher := &core.CardanoBridgeDataFetcherMock{}
		expectedTxsFetcher := NewExpectedTxsFetcher(context.Background(), bridgeDataFetcher, appConfig, &core.CardanoTxsProcessorDBMock{}, hclog.NewNullLogger())

		require.NotNil(t, expectedTxsFetcher)
	})

	t.Run("fetchData nil", func(t *testing.T) {
		bridgeDataFetcher := &core.CardanoBridgeDataFetcherMock{}
		bridgeDataFetcher.On("FetchExpectedTx", mock.Anything).Return(nil, nil)

		db := &core.CardanoTxsProcessorDBMock{}
		db.On("GetAllExpectedTxs", mock.Anything, mock.Anything).Return(nil, nil)
		db.On("AddExpectedTxs", mock.Anything).Return(nil)

		expectedTxsFetcher := NewExpectedTxsFetcher(context.Background(), bridgeDataFetcher, appConfig, db, hclog.NewNullLogger())
		require.NotNil(t, expectedTxsFetcher)

		err := expectedTxsFetcher.fetchData()
		require.NoError(t, err)
	})

	t.Run("fetchData err", func(t *testing.T) {
		bridgeDataFetcher := &core.CardanoBridgeDataFetcherMock{}
		bridgeDataFetcher.On("FetchExpectedTx", mock.Anything).Return(&core.BridgeExpectedCardanoTx{}, nil)

		db := &core.CardanoTxsProcessorDBMock{}
		db.On("GetAllExpectedTxs", mock.Anything, mock.Anything).Return(nil, nil)
		db.On("AddExpectedTxs", mock.Anything).Return(fmt.Errorf("test err"))

		expectedTxsFetcher := NewExpectedTxsFetcher(context.Background(), bridgeDataFetcher, appConfig, db, hclog.NewNullLogger())
		require.NotNil(t, expectedTxsFetcher)

		err := expectedTxsFetcher.fetchData()
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to add expected txs")
	})
}
