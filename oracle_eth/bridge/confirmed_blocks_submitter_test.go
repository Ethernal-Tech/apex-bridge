package bridge

import (
	"context"
	"fmt"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	eth_core "github.com/Ethernal-Tech/apex-bridge/oracle_eth/core"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestConfirmedBlocksSubmitter(t *testing.T) {
	chainID := common.ChainIDStrPrime
	appConfig := &oCore.AppConfig{
		Bridge: oCore.BridgeConfig{
			SubmitConfig: oCore.SubmitConfig{
				ConfirmedBlocksThreshold:  10,
				ConfirmedBlocksSubmitTime: 10,
			},
		},
	}

	t.Run("NewConfirmedBlocksSubmitter 1", func(t *testing.T) {
		bridgeSubmitter := &eth_core.BridgeSubmitterMock{}
		indexerDB := &eth_core.EventStoreMock{}

		indexerDB.On("GetLastProcessedBlock").Return(uint64(0), fmt.Errorf("test err"))

		bs, err := NewConfirmedBlocksSubmitter(context.Background(), bridgeSubmitter, appConfig, indexerDB, chainID, hclog.NewNullLogger())
		require.Error(t, err)
		require.ErrorContains(t, err, "test err")
		require.Nil(t, bs)
	})

	t.Run("NewConfirmedBlocksSubmitter 2", func(t *testing.T) {
		bridgeSubmitter := &eth_core.BridgeSubmitterMock{}
		indexerDB := &eth_core.EventStoreMock{}

		indexerDB.On("GetLastProcessedBlock").Return(uint64(10), nil)

		bs, err := NewConfirmedBlocksSubmitter(context.Background(), bridgeSubmitter, appConfig, indexerDB, chainID, hclog.NewNullLogger())
		require.NoError(t, err)
		require.NotNil(t, bs)
	})

	t.Run("execute 1", func(t *testing.T) {
		bridgeSubmitter := &eth_core.BridgeSubmitterMock{}
		indexerDB := &eth_core.EventStoreMock{}
		indexerDB.On("GetLastProcessedBlock").Return(uint64(10), nil).Once()

		bs, _ := NewConfirmedBlocksSubmitter(context.Background(), bridgeSubmitter, appConfig, indexerDB, chainID, hclog.NewNullLogger())

		indexerDB.On("GetLastProcessedBlock").Return(uint64(0), fmt.Errorf("test err"))

		err := bs.execute()
		require.Error(t, err)
		require.ErrorContains(t, err, "error getting latest confirmed blocks")
	})

	t.Run("execute 2", func(t *testing.T) {
		bridgeSubmitter := &eth_core.BridgeSubmitterMock{}
		bridgeSubmitter.On("SubmitConfirmedBlocks", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("test err"))

		indexerDB := &eth_core.EventStoreMock{}
		indexerDB.On("GetLastProcessedBlock").Return(uint64(10), nil).Once()

		bs, _ := NewConfirmedBlocksSubmitter(context.Background(), bridgeSubmitter, appConfig, indexerDB, chainID, hclog.NewNullLogger())

		indexerDB.On("GetLastProcessedBlock").Return(uint64(11), nil)

		err := bs.execute()
		require.Error(t, err)
		require.ErrorContains(t, err, "test err")
	})

	t.Run("execute 3", func(t *testing.T) {
		bridgeSubmitter := &eth_core.BridgeSubmitterMock{}
		bridgeSubmitter.On("SubmitConfirmedBlocks", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		indexerDB := &eth_core.EventStoreMock{}
		indexerDB.On("GetLastProcessedBlock").Return(uint64(10), nil).Once()

		bs, _ := NewConfirmedBlocksSubmitter(context.Background(), bridgeSubmitter, appConfig, indexerDB, chainID, hclog.NewNullLogger())

		indexerDB.On("GetLastProcessedBlock").Return(uint64(11), nil)

		err := bs.execute()
		require.NoError(t, err)
	})
}
