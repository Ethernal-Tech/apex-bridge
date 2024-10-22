package bridge

import (
	"context"
	"fmt"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestConfirmedBlocksSubmitter(t *testing.T) {
	chainID := common.ChainIDStrPrime
	appConfig := &cCore.AppConfig{
		Bridge: cCore.BridgeConfig{
			SubmitConfig: cCore.SubmitConfig{
				ConfirmedBlocksThreshold:  10,
				ConfirmedBlocksSubmitTime: 10,
			},
		},
	}

	t.Run("NewConfirmedBlocksSubmitter 1", func(t *testing.T) {
		bridgeSubmitter := &core.BridgeSubmitterMock{}
		indexerDB := &indexer.DatabaseMock{}
		indexerDB.On("GetLatestBlockPoint").Return((*indexer.BlockPoint)(nil), fmt.Errorf("test err"))

		bs, err := NewConfirmedBlocksSubmitter(context.Background(), bridgeSubmitter, appConfig, indexerDB, chainID, hclog.NewNullLogger())
		require.Error(t, err)
		require.ErrorContains(t, err, "test err")
		require.Nil(t, bs)
	})

	t.Run("NewConfirmedBlocksSubmitter 2", func(t *testing.T) {
		bridgeSubmitter := &core.BridgeSubmitterMock{}
		indexerDB := &indexer.DatabaseMock{}
		indexerDB.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{}, nil)

		bs, err := NewConfirmedBlocksSubmitter(context.Background(), bridgeSubmitter, appConfig, indexerDB, chainID, hclog.NewNullLogger())
		require.NoError(t, err)
		require.NotNil(t, bs)
	})

	t.Run("execute 1", func(t *testing.T) {
		bridgeSubmitter := &core.BridgeSubmitterMock{}
		indexerDB := &indexer.DatabaseMock{}
		indexerDB.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{}, nil)

		bs, _ := NewConfirmedBlocksSubmitter(context.Background(), bridgeSubmitter, appConfig, indexerDB, chainID, hclog.NewNullLogger())

		indexerDB.On("GetConfirmedBlocksFrom", bs.latestConfirmedSlot, appConfig.Bridge.SubmitConfig.ConfirmedBlocksThreshold).Return(([]*indexer.CardanoBlock)(nil), fmt.Errorf("test err"))

		err := bs.execute()
		require.Error(t, err)
		require.ErrorContains(t, err, "error getting latest confirmed blocks")
	})

	t.Run("execute 2", func(t *testing.T) {
		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.On("SubmitConfirmedBlocks", mock.Anything, mock.Anything).Return(fmt.Errorf("test err"))

		indexerDB := &indexer.DatabaseMock{}
		indexerDB.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{}, nil)

		bs, _ := NewConfirmedBlocksSubmitter(context.Background(), bridgeSubmitter, appConfig, indexerDB, chainID, hclog.NewNullLogger())

		indexerDB.On(
			"GetConfirmedBlocksFrom", bs.latestConfirmedSlot, appConfig.Bridge.SubmitConfig.ConfirmedBlocksThreshold).Return(
			[]*indexer.CardanoBlock{{}}, nil)

		err := bs.execute()
		require.Error(t, err)
		require.ErrorContains(t, err, "error submitting confirmed blocks")
	})

	t.Run("execute 3", func(t *testing.T) {
		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.On("SubmitConfirmedBlocks", mock.Anything, mock.Anything).Return(nil)

		indexerDB := &indexer.DatabaseMock{}
		indexerDB.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{}, nil)

		bs, _ := NewConfirmedBlocksSubmitter(context.Background(), bridgeSubmitter, appConfig, indexerDB, chainID, hclog.NewNullLogger())

		indexerDB.On(
			"GetConfirmedBlocksFrom", bs.latestConfirmedSlot, appConfig.Bridge.SubmitConfig.ConfirmedBlocksThreshold).Return(
			[]*indexer.CardanoBlock{{}}, nil)

		err := bs.execute()
		require.NoError(t, err)
	})
}
