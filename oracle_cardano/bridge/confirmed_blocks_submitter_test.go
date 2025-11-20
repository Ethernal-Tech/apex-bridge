package bridge

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	oracleCommon "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/validatorobserver"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/ethgo"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestConfirmedBlocksSubmitter(t *testing.T) {
	chainID := common.ChainIDStrPrime
	appConfig := &oracleCommon.AppConfig{
		Bridge: oracleCommon.BridgeConfig{
			SubmitConfig: oracleCommon.SubmitConfig{
				ConfirmedBlocksThreshold:  30,
				ConfirmedBlocksSubmitTime: 10,
				EmptyBlocksThreshold: map[string]uint{
					common.ChainIDStrPrime:  3,
					common.ChainIDStrVector: 3,
					common.ChainIDStrNexus:  3,
				},
			},
		},
	}

	bridgeSubmitter := &core.BridgeSubmitterMock{}
	indexerDB := &indexer.DatabaseMock{}
	oracleDB := &core.CardanoTxsProcessorDBMock{}
	vsObserver := &validatorobserver.ValidatorSetObserverMock{}
	testErr := fmt.Errorf("test err")

	t.Run("NewConfirmedBlocksSubmitter GetBlocksSubmitterInfo error", func(t *testing.T) {
		oracleDB.On("GetBlocksSubmitterInfo", chainID).Return(oracleCommon.BlocksSubmitterInfo{}, testErr).Once()

		_, err := NewConfirmedBlocksSubmitter(
			bridgeSubmitter, appConfig, oracleDB, indexerDB, chainID, vsObserver, hclog.NewNullLogger())
		require.ErrorIs(t, err, testErr)
	})

	t.Run("NewConfirmedBlocksSubmitter Start from chain config", func(t *testing.T) {
		const startSlot = uint64(1044)

		appConfig.CardanoChains = map[string]*oracleCommon.CardanoChainConfig{
			chainID: {
				StartSlot: startSlot,
			},
		}

		defer func() {
			appConfig.CardanoChains = nil
		}()

		oracleDB.On("GetBlocksSubmitterInfo", chainID).Return(oracleCommon.BlocksSubmitterInfo{}, nil).Once()

		bs, err := NewConfirmedBlocksSubmitter(
			bridgeSubmitter, appConfig, oracleDB, indexerDB, chainID, vsObserver, hclog.NewNullLogger())

		require.NoError(t, err)
		require.Equal(t, startSlot, bs.latestInfo.BlockNumOrSlot)
	})

	t.Run("NewConfirmedBlocksSubmitter UpdateFromIndexerDB", func(t *testing.T) {
		const startSlot = uint64(1044)

		appConfig.Bridge.SubmitConfig.UpdateFromIndexerDB = true

		defer func() {
			appConfig.Bridge.SubmitConfig.UpdateFromIndexerDB = false
		}()

		oracleDB.On("GetBlocksSubmitterInfo", chainID).Return(oracleCommon.BlocksSubmitterInfo{BlockNumOrSlot: startSlot - 1}, nil).Once()
		indexerDB.On("GetLatestBlockPoint").Return(&indexer.BlockPoint{BlockSlot: startSlot}, nil).Once()

		bs, err := NewConfirmedBlocksSubmitter(
			bridgeSubmitter, appConfig, oracleDB, indexerDB, chainID, vsObserver, hclog.NewNullLogger())

		require.NoError(t, err)
		require.Equal(t, startSlot, bs.latestInfo.BlockNumOrSlot)
	})

	t.Run("Start ctx done", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		cancel()

		oracleDB.On("GetBlocksSubmitterInfo", chainID).Return(oracleCommon.BlocksSubmitterInfo{}, nil).Once()

		blocksSubmitter, err := NewConfirmedBlocksSubmitter(
			bridgeSubmitter, appConfig, oracleDB, indexerDB, chainID, vsObserver, hclog.NewNullLogger())
		require.NoError(t, err)

		blocksSubmitter.Start(ctx)

		time.Sleep(time.Second)

		require.Equal(t, uint64(0), blocksSubmitter.latestInfo.BlockNumOrSlot)
	})

	t.Run("Execute", func(t *testing.T) {
		hashes := [6]ethgo.Hash{
			ethgo.HexToHash("F1"), ethgo.HexToHash("F2"), ethgo.HexToHash("F3"),
			ethgo.HexToHash("F4"), ethgo.HexToHash("F5"), ethgo.HexToHash("F6"),
		}

		oracleDB.On("GetBlocksSubmitterInfo", chainID).Return(oracleCommon.BlocksSubmitterInfo{BlockNumOrSlot: 1}, nil).Once()
		indexerDB.On("GetConfirmedBlocksFrom", uint64(2), appConfig.Bridge.SubmitConfig.ConfirmedBlocksThreshold).Return([]*indexer.CardanoBlock{
			{
				Slot: 2,
			},
			{
				Slot: 3,
			},
			{
				Slot: 4, // this
				Hash: indexer.Hash(hashes[3]),
			},
			{
				Slot: 5,
			},
			{
				Slot: 6,
				Hash: indexer.Hash(hashes[4]),
				Txs:  []indexer.Hash{indexer.Hash(hashes[0]), indexer.Hash(hashes[1])}, // this
			},
			{
				Slot: 7,
			},
			{
				Slot: 8,
			},
			{
				Slot: 9, // this
				Hash: indexer.Hash(hashes[5]),
			},
			{
				Slot: 10,
			},
			{
				Slot: 11,
				Txs:  []indexer.Hash{indexer.Hash(hashes[2])}, // quit
			},
			{
				Slot: 12,
			},
			{
				Slot: 13,
			},
			{
				Slot: 14,
			},
		}, nil).Once()

		submittedBlocks := []eth.CardanoBlock{
			{BlockSlot: big.NewInt(4), BlockHash: indexer.Hash(hashes[3])},
			{BlockSlot: big.NewInt(6), BlockHash: indexer.Hash(hashes[4])},
			{BlockSlot: big.NewInt(9), BlockHash: indexer.Hash(hashes[5])},
		}

		oracleDB.On("GetProcessedTx", oracleCommon.DBTxID{ChainID: chainID, DBKey: hashes[0].Bytes()}).Return(&core.ProcessedCardanoTx{}, nil).Once()
		oracleDB.On("GetProcessedTx", oracleCommon.DBTxID{ChainID: chainID, DBKey: hashes[1].Bytes()}).Return(&core.ProcessedCardanoTx{}, nil).Once()
		oracleDB.On("GetProcessedTx", oracleCommon.DBTxID{ChainID: chainID, DBKey: hashes[2].Bytes()}).Return((*core.ProcessedCardanoTx)(nil), nil).Once()

		oracleDB.On("SetBlocksSubmitterInfo", chainID, oracleCommon.BlocksSubmitterInfo{
			BlockNumOrSlot: 10,
			CounterEmpty:   0,
		}).Return(nil).Once()
		bridgeSubmitter.On("SubmitBlocks", chainID, submittedBlocks).Return(nil).Once()

		blocksSubmitter, err := NewConfirmedBlocksSubmitter(
			bridgeSubmitter, appConfig, oracleDB, indexerDB, chainID, vsObserver, hclog.NewNullLogger())
		require.NoError(t, err)

		vsObserver.On("IsValidatorSetPending").Return(true).Once()

		require.NoError(t, blocksSubmitter.execute())

		require.Equal(t, uint64(1), blocksSubmitter.latestInfo.BlockNumOrSlot)
		require.Equal(t, 0, blocksSubmitter.latestInfo.CounterEmpty)

		vsObserver.On("IsValidatorSetPending").Return(false).Once()

		require.NoError(t, blocksSubmitter.execute())

		require.Equal(t, uint64(10), blocksSubmitter.latestInfo.BlockNumOrSlot)
		require.Equal(t, 0, blocksSubmitter.latestInfo.CounterEmpty)
	})

	bridgeSubmitter.AssertExpectations(t)
	oracleDB.AssertExpectations(t)
	indexerDB.AssertExpectations(t)
}
