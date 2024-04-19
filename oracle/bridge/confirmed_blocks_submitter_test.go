package bridge

import (
	"context"
	"fmt"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestConfirmedBlocksSubmitter(t *testing.T) {
	chainId := "prime"
	appConfig := &core.AppConfig{
		Bridge: core.BridgeConfig{
			SubmitConfig: core.SubmitConfig{
				ConfirmedBlocksThreshold:  10,
				ConfirmedBlocksSubmitTime: 10,
			},
		},
	}

	t.Run("NewConfirmedBlocksSubmitter 1", func(t *testing.T) {
		bridgeSubmitter := &core.BridgeSubmitterMock{}
		db := &core.CardanoTxsProcessorDbMock{}
		indexerDb := &indexer.DatabaseMock{}
		indexerDb.On("GetLatestBlockPoint").Return((*indexer.BlockPoint)(nil), fmt.Errorf("test err"))

		bs, err := NewConfirmedBlocksSubmitter(context.Background(), bridgeSubmitter, appConfig, db, indexerDb, chainId, hclog.NewNullLogger())
		require.Error(t, err)
		require.ErrorContains(t, err, "test err")
		require.Nil(t, bs)
	})

	t.Run("NewConfirmedBlocksSubmitter 2", func(t *testing.T) {
		bridgeSubmitter := &core.BridgeSubmitterMock{}
		db := &core.CardanoTxsProcessorDbMock{}
		indexerDb := &indexer.DatabaseMock{}
		indexerDb.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{}, nil)

		bs, err := NewConfirmedBlocksSubmitter(context.Background(), bridgeSubmitter, appConfig, db, indexerDb, chainId, hclog.NewNullLogger())
		require.NoError(t, err)
		require.NotNil(t, bs)
	})

	t.Run("execute 1", func(t *testing.T) {
		bridgeSubmitter := &core.BridgeSubmitterMock{}
		db := &core.CardanoTxsProcessorDbMock{}
		indexerDb := &indexer.DatabaseMock{}
		indexerDb.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{}, nil)

		bs, _ := NewConfirmedBlocksSubmitter(context.Background(), bridgeSubmitter, appConfig, db, indexerDb, chainId, hclog.NewNullLogger())

		indexerDb.On("GetConfirmedBlocksFrom", bs.latestConfirmedSlot, appConfig.Bridge.SubmitConfig.ConfirmedBlocksThreshold).Return(([]*indexer.CardanoBlock)(nil), fmt.Errorf("test err"))

		err := bs.execute()
		require.Error(t, err)
		require.ErrorContains(t, err, "error getting latest confirmed blocks")
	})

	t.Run("execute 2", func(t *testing.T) {
		bridgeSubmitter := &core.BridgeSubmitterMock{}
		db := &core.CardanoTxsProcessorDbMock{}
		indexerDb := &indexer.DatabaseMock{}
		indexerDb.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{}, nil)

		bs, _ := NewConfirmedBlocksSubmitter(context.Background(), bridgeSubmitter, appConfig, db, indexerDb, chainId, hclog.NewNullLogger())

		indexerDb.On(
			"GetConfirmedBlocksFrom", bs.latestConfirmedSlot, appConfig.Bridge.SubmitConfig.ConfirmedBlocksThreshold).Return(
			[]*indexer.CardanoBlock{}, nil)

		err := bs.execute()
		require.NoError(t, err)
	})

	t.Run("execute 3", func(t *testing.T) {
		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.On("SubmitConfirmedBlocks").Return(fmt.Errorf("test err"))

		db := &core.CardanoTxsProcessorDbMock{}
		indexerDb := &indexer.DatabaseMock{}
		indexerDb.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{}, nil)

		bs, _ := NewConfirmedBlocksSubmitter(context.Background(), bridgeSubmitter, appConfig, db, indexerDb, chainId, hclog.NewNullLogger())

		indexerDb.On(
			"GetConfirmedBlocksFrom", bs.latestConfirmedSlot, appConfig.Bridge.SubmitConfig.ConfirmedBlocksThreshold).Return(
			[]*indexer.CardanoBlock{{}}, nil)

		err := bs.execute()
		require.Error(t, err)
		require.ErrorContains(t, err, "error submitting confirmed blocks")
	})

	t.Run("execute 4", func(t *testing.T) {
		bridgeSubmitter := &core.BridgeSubmitterMock{}
		bridgeSubmitter.On("SubmitConfirmedBlocks").Return(nil)

		db := &core.CardanoTxsProcessorDbMock{}
		indexerDb := &indexer.DatabaseMock{}
		indexerDb.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{}, nil)

		bs, _ := NewConfirmedBlocksSubmitter(context.Background(), bridgeSubmitter, appConfig, db, indexerDb, chainId, hclog.NewNullLogger())

		indexerDb.On(
			"GetConfirmedBlocksFrom", bs.latestConfirmedSlot, appConfig.Bridge.SubmitConfig.ConfirmedBlocksThreshold).Return(
			[]*indexer.CardanoBlock{{}}, nil)

		err := bs.execute()
		require.NoError(t, err)
	})
}
