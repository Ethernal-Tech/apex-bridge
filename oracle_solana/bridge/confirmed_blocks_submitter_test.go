package bridge

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	oracleCommon "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	solanaCore "github.com/Ethernal-Tech/apex-bridge/oracle_solana/core"
	solanaTx "github.com/Ethernal-Tech/apex-bridge/solana"
	solanaTxsStore "github.com/Ethernal-Tech/solana-infrastructure/tracker/store"
	"github.com/gagliardetto/solana-go"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func testSolanaSubmitConfig(chainID string) oracleCommon.SubmitConfig {
	return oracleCommon.SubmitConfig{
		ConfirmedBlocksThreshold:  30,
		ConfirmedBlocksSubmitTime: 10,
		EmptyBlocksThreshold: map[string]uint{
			chainID: 2,
		},
	}
}

func TestConfirmedBlocksSubmitter(t *testing.T) {
	chainID := common.ChainIDStrSolana
	testErr := fmt.Errorf("test err")

	t.Run("NewConfirmedBlocksSubmitter GetBlocksSubmitterInfo error", func(t *testing.T) {
		appConfig := &oracleCommon.AppConfig{
			Bridge: oracleCommon.BridgeConfig{SubmitConfig: testSolanaSubmitConfig(chainID)},
		}
		bridgeSubmitter := &solanaCore.SolanaBridgeSubmitterMock{}
		indexerDB := &solanaTxsStore.MockStorageHandler{}
		oracleDB := &solanaCore.SolanaTxsProcessorDBMock{}
		oracleDB.On("GetBlocksSubmitterInfo", chainID).Return(oracleCommon.BlocksSubmitterInfo{}, testErr).Once()

		_, err := NewConfirmedBlocksSubmitter(
			bridgeSubmitter, appConfig, oracleDB, indexerDB, chainID, hclog.NewNullLogger())
		require.ErrorIs(t, err, testErr)
	})

	t.Run("NewConfirmedBlocksSubmitter UpdateFromIndexerDB error", func(t *testing.T) {
		appConfig := &oracleCommon.AppConfig{
			Bridge: oracleCommon.BridgeConfig{SubmitConfig: testSolanaSubmitConfig(chainID)},
		}
		appConfig.Bridge.SubmitConfig.UpdateFromIndexerDB = true

		bridgeSubmitter := &solanaCore.SolanaBridgeSubmitterMock{}
		indexerDB := &solanaTxsStore.MockStorageHandler{}
		oracleDB := &solanaCore.SolanaTxsProcessorDBMock{}
		oracleDB.On("GetBlocksSubmitterInfo", chainID).Return(oracleCommon.BlocksSubmitterInfo{}, nil).Once()
		indexerDB.On("GetLatestBlockPoint").Return((*solanaTxsStore.BlockPoint)(nil), testErr).Once()

		_, err := NewConfirmedBlocksSubmitter(
			bridgeSubmitter, appConfig, oracleDB, indexerDB, chainID, hclog.NewNullLogger())
		require.ErrorContains(t, err, "failed to create block submitter")
		require.ErrorContains(t, err, testErr.Error())
	})

	t.Run("NewConfirmedBlocksSubmitter UpdateFromIndexerDB updates latest info", func(t *testing.T) {
		startSlot := uint64(1044)
		appConfig := &oracleCommon.AppConfig{
			Bridge: oracleCommon.BridgeConfig{SubmitConfig: testSolanaSubmitConfig(chainID)},
		}
		appConfig.Bridge.SubmitConfig.UpdateFromIndexerDB = true

		bridgeSubmitter := &solanaCore.SolanaBridgeSubmitterMock{}
		indexerDB := &solanaTxsStore.MockStorageHandler{}
		oracleDB := &solanaCore.SolanaTxsProcessorDBMock{}
		oracleDB.On("GetBlocksSubmitterInfo", chainID).Return(oracleCommon.BlocksSubmitterInfo{
			BlockNumOrSlot: startSlot - 1,
			CounterEmpty:   3,
		}, nil).Once()
		indexerDB.On("GetLatestBlockPoint").Return(&solanaTxsStore.BlockPoint{
			BlockSlot:   startSlot,
			BlockNumber: 10,
		}, nil).Once()

		bs, err := NewConfirmedBlocksSubmitter(
			bridgeSubmitter, appConfig, oracleDB, indexerDB, chainID, hclog.NewNullLogger())
		require.NoError(t, err)
		require.Equal(t, startSlot, bs.latestInfo.BlockNumOrSlot)
		require.Equal(t, 0, bs.latestInfo.CounterEmpty)
	})

	t.Run("Start ctx done", func(t *testing.T) {
		appConfig := &oracleCommon.AppConfig{
			Bridge: oracleCommon.BridgeConfig{SubmitConfig: testSolanaSubmitConfig(chainID)},
		}
		bridgeSubmitter := &solanaCore.SolanaBridgeSubmitterMock{}
		indexerDB := &solanaTxsStore.MockStorageHandler{}
		oracleDB := &solanaCore.SolanaTxsProcessorDBMock{}
		oracleDB.On("GetBlocksSubmitterInfo", chainID).Return(oracleCommon.BlocksSubmitterInfo{}, nil).Once()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		blocksSubmitter, err := NewConfirmedBlocksSubmitter(
			bridgeSubmitter, appConfig, oracleDB, indexerDB, chainID, hclog.NewNullLogger())
		require.NoError(t, err)
		blocksSubmitter.Start(ctx)
		time.Sleep(time.Millisecond * 20)
		require.Equal(t, uint64(0), blocksSubmitter.latestInfo.BlockNumOrSlot)
	})

	t.Run("execute get latest block point error", func(t *testing.T) {
		appConfig := &oracleCommon.AppConfig{
			Bridge: oracleCommon.BridgeConfig{SubmitConfig: testSolanaSubmitConfig(chainID)},
		}
		bridgeSubmitter := &solanaCore.SolanaBridgeSubmitterMock{}
		indexerDB := &solanaTxsStore.MockStorageHandler{}
		oracleDB := &solanaCore.SolanaTxsProcessorDBMock{}
		oracleDB.On("GetBlocksSubmitterInfo", chainID).Return(oracleCommon.BlocksSubmitterInfo{}, nil).Once()
		indexerDB.On("GetLatestBlockPoint").Return((*solanaTxsStore.BlockPoint)(nil), testErr).Once()

		blocksSubmitter, err := NewConfirmedBlocksSubmitter(
			bridgeSubmitter, appConfig, oracleDB, indexerDB, chainID, hclog.NewNullLogger())
		require.NoError(t, err)

		err = blocksSubmitter.execute()
		require.ErrorContains(t, err, "error getting latest block point")
	})

	t.Run("execute submit blocks error", func(t *testing.T) {
		appConfig := &oracleCommon.AppConfig{
			Bridge: oracleCommon.BridgeConfig{SubmitConfig: testSolanaSubmitConfig(chainID)},
			SolanaChains: map[string]*oracleCommon.SolanaChainConfig{
				chainID: {
					SolanaChainConfig: solanaTx.SolanaChainConfig{
						SlotRoundingThreshold: 10,
						NoBatchPeriodPercent:  0,
					},
				},
			},
		}
		bridgeSubmitter := &solanaCore.SolanaBridgeSubmitterMock{}
		indexerDB := &solanaTxsStore.MockStorageHandler{}
		oracleDB := &solanaCore.SolanaTxsProcessorDBMock{}
		oracleDB.On("GetBlocksSubmitterInfo", chainID).Return(oracleCommon.BlocksSubmitterInfo{
			BlockNumOrSlot: 1,
		}, nil).Once()

		blockHash := solana.Hash(solana.NewWallet().PublicKey())
		indexerDB.On("GetLatestBlockPoint").Return(&solanaTxsStore.BlockPoint{
			BlockSlot:   3,
			BlockHash:   blockHash,
			BlockNumber: 20,
		}, nil).Once()
		mockEmptySlots(indexerDB, blockHash, 2, 3)
		bridgeSubmitter.On("SubmitBlocks", chainID, mock.Anything).Return(testErr).Once()

		blocksSubmitter, err := NewConfirmedBlocksSubmitter(
			bridgeSubmitter, appConfig, oracleDB, indexerDB, chainID, hclog.NewNullLogger())
		require.NoError(t, err)

		err = blocksSubmitter.execute()
		require.ErrorContains(t, err, "error submitting blocks")
	})

	t.Run("execute save latest info error", func(t *testing.T) {
		appConfig := &oracleCommon.AppConfig{
			Bridge: oracleCommon.BridgeConfig{SubmitConfig: testSolanaSubmitConfig(chainID)},
			SolanaChains: map[string]*oracleCommon.SolanaChainConfig{
				chainID: {
					SolanaChainConfig: solanaTx.SolanaChainConfig{
						SlotRoundingThreshold: 10,
						NoBatchPeriodPercent:  0,
					},
				},
			},
		}
		bridgeSubmitter := &solanaCore.SolanaBridgeSubmitterMock{}
		indexerDB := &solanaTxsStore.MockStorageHandler{}
		oracleDB := &solanaCore.SolanaTxsProcessorDBMock{}
		oracleDB.On("GetBlocksSubmitterInfo", chainID).Return(oracleCommon.BlocksSubmitterInfo{
			BlockNumOrSlot: 1,
		}, nil).Once()

		blockHash := solana.Hash(solana.NewWallet().PublicKey())
		indexerDB.On("GetLatestBlockPoint").Return(&solanaTxsStore.BlockPoint{
			BlockSlot:   3,
			BlockHash:   blockHash,
			BlockNumber: 20,
		}, nil).Once()
		mockEmptySlots(indexerDB, blockHash, 2, 3)
		bridgeSubmitter.On("SubmitBlocks", chainID, mock.Anything).Return(nil).Once()
		oracleDB.On("SetBlocksSubmitterInfo", chainID, oracleCommon.BlocksSubmitterInfo{
			BlockNumOrSlot: 3,
			CounterEmpty:   0,
		}).Return(testErr).Once()

		blocksSubmitter, err := NewConfirmedBlocksSubmitter(
			bridgeSubmitter, appConfig, oracleDB, indexerDB, chainID, hclog.NewNullLogger())
		require.NoError(t, err)

		err = blocksSubmitter.execute()
		require.ErrorContains(t, err, "error saving confirmed blocks")
	})

	t.Run("execute passes and submits threshold empty block", func(t *testing.T) {
		appConfig := &oracleCommon.AppConfig{
			Bridge: oracleCommon.BridgeConfig{SubmitConfig: testSolanaSubmitConfig(chainID)},
			SolanaChains: map[string]*oracleCommon.SolanaChainConfig{
				chainID: {
					SolanaChainConfig: solanaTx.SolanaChainConfig{
						SlotRoundingThreshold: 10,
						NoBatchPeriodPercent:  0,
					},
				},
			},
		}
		bridgeSubmitter := &solanaCore.SolanaBridgeSubmitterMock{}
		indexerDB := &solanaTxsStore.MockStorageHandler{}
		oracleDB := &solanaCore.SolanaTxsProcessorDBMock{}
		oracleDB.On("GetBlocksSubmitterInfo", chainID).Return(oracleCommon.BlocksSubmitterInfo{
			BlockNumOrSlot: 1,
		}, nil).Once()

		blockHash := solana.Hash(solana.NewWallet().PublicKey())
		pendingSig, signErr := solana.NewWallet().PrivateKey.Sign([]byte("pending-tx"))
		require.NoError(t, signErr)
		indexerDB.On("GetLatestBlockPoint").Return(&solanaTxsStore.BlockPoint{
			BlockSlot:   6,
			BlockHash:   blockHash,
			BlockNumber: 99,
		}, nil).Once()
		mockEmptySlots(indexerDB, blockHash, 2, 3)
		indexerDB.On("GetEventsBySlot", uint64(4)).Return([]solanaTxsStore.EventRecord{
			{Slot: 4, TxSignature: pendingSig.String()},
		}, nil).Once()
		oracleDB.On("GetProcessedTx", oracleCommon.DBTxID{
			ChainID: chainID,
			DBKey:   pendingSig[:],
		}).Return((*solanaCore.ProcessedSolanaTx)(nil), nil).Once()

		bridgeSubmitter.On("SubmitBlocks", chainID, []eth.CardanoBlock(nil)).Return(nil).Once()
		oracleDB.On("SetBlocksSubmitterInfo", chainID, oracleCommon.BlocksSubmitterInfo{
			BlockNumOrSlot: 3,
			CounterEmpty:   0,
		}).Return(nil).Once()

		blocksSubmitter, err := NewConfirmedBlocksSubmitter(
			bridgeSubmitter, appConfig, oracleDB, indexerDB, chainID, hclog.NewNullLogger())
		require.NoError(t, err)
		require.NoError(t, blocksSubmitter.execute())
		require.Equal(t, uint64(3), blocksSubmitter.latestInfo.BlockNumOrSlot)
		require.Equal(t, 0, blocksSubmitter.latestInfo.CounterEmpty)
	})

	t.Run("getBlocksToSubmit missing threshold config", func(t *testing.T) {
		appConfig := &oracleCommon.AppConfig{
			Bridge: oracleCommon.BridgeConfig{
				SubmitConfig: oracleCommon.SubmitConfig{
					ConfirmedBlocksThreshold: 30,
					EmptyBlocksThreshold:     map[string]uint{},
				},
			},
			SolanaChains: map[string]*oracleCommon.SolanaChainConfig{
				chainID: {
					SolanaChainConfig: solanaTx.SolanaChainConfig{
						SlotRoundingThreshold: 10,
						NoBatchPeriodPercent:  0,
					},
				},
			},
		}
		bridgeSubmitter := &solanaCore.SolanaBridgeSubmitterMock{}
		indexerDB := &solanaTxsStore.MockStorageHandler{}
		oracleDB := &solanaCore.SolanaTxsProcessorDBMock{}
		oracleDB.On("GetBlocksSubmitterInfo", chainID).Return(oracleCommon.BlocksSubmitterInfo{}, nil).Once()
		indexerDB.On("GetLatestBlockPoint").Return(&solanaTxsStore.BlockPoint{
			BlockSlot:   1,
			BlockNumber: 1,
		}, nil).Once()
		indexerDB.On("GetEventsBySlot", uint64(1)).Return([]solanaTxsStore.EventRecord{}, nil).Once()

		blocksSubmitter, err := NewConfirmedBlocksSubmitter(
			bridgeSubmitter, appConfig, oracleDB, indexerDB, chainID, hclog.NewNullLogger())
		require.NoError(t, err)

		_, _, err = blocksSubmitter.getBlocksToSubmit(1)
		require.ErrorContains(t, err, "empty blocks threshold not configured")
	})

	t.Run("getBlocksToSubmit threshold too large", func(t *testing.T) {
		appConfig := &oracleCommon.AppConfig{
			Bridge: oracleCommon.BridgeConfig{
				SubmitConfig: oracleCommon.SubmitConfig{
					ConfirmedBlocksThreshold: 30,
					EmptyBlocksThreshold: map[string]uint{
						chainID: uint(math.MaxInt) + 1,
					},
				},
			},
			SolanaChains: map[string]*oracleCommon.SolanaChainConfig{
				chainID: {
					SolanaChainConfig: solanaTx.SolanaChainConfig{
						SlotRoundingThreshold: 10,
						NoBatchPeriodPercent:  0,
					},
				},
			},
		}
		bridgeSubmitter := &solanaCore.SolanaBridgeSubmitterMock{}
		indexerDB := &solanaTxsStore.MockStorageHandler{}
		oracleDB := &solanaCore.SolanaTxsProcessorDBMock{}
		oracleDB.On("GetBlocksSubmitterInfo", chainID).Return(oracleCommon.BlocksSubmitterInfo{}, nil).Once()
		indexerDB.On("GetLatestBlockPoint").Return(&solanaTxsStore.BlockPoint{
			BlockSlot:   1,
			BlockNumber: 1,
		}, nil).Once()
		indexerDB.On("GetEventsBySlot", uint64(1)).Return([]solanaTxsStore.EventRecord{}, nil).Once()

		blocksSubmitter, err := NewConfirmedBlocksSubmitter(
			bridgeSubmitter, appConfig, oracleDB, indexerDB, chainID, hclog.NewNullLogger())
		require.NoError(t, err)

		_, _, err = blocksSubmitter.getBlocksToSubmit(1)
		require.ErrorContains(t, err, "threshold too large")
	})

	t.Run("getBlocksToSubmit get events by slot error", func(t *testing.T) {
		appConfig := &oracleCommon.AppConfig{
			Bridge: oracleCommon.BridgeConfig{SubmitConfig: testSolanaSubmitConfig(chainID)},
		}
		bridgeSubmitter := &solanaCore.SolanaBridgeSubmitterMock{}
		indexerDB := &solanaTxsStore.MockStorageHandler{}
		oracleDB := &solanaCore.SolanaTxsProcessorDBMock{}
		oracleDB.On("GetBlocksSubmitterInfo", chainID).Return(oracleCommon.BlocksSubmitterInfo{}, nil).Once()
		indexerDB.On("GetLatestBlockPoint").Return(&solanaTxsStore.BlockPoint{
			BlockSlot:   3,
			BlockNumber: 2,
		}, nil).Once()
		indexerDB.On("GetEventsBySlot", uint64(1)).Return([]solanaTxsStore.EventRecord(nil), testErr).Once()

		blocksSubmitter, err := NewConfirmedBlocksSubmitter(
			bridgeSubmitter, appConfig, oracleDB, indexerDB, chainID, hclog.NewNullLogger())
		require.NoError(t, err)

		_, _, err = blocksSubmitter.getBlocksToSubmit(1)
		require.ErrorContains(t, err, "failed to get events for slot")
	})
}

func mockEmptySlots(indexerDB *solanaTxsStore.MockStorageHandler, blockHash solana.Hash, slots ...uint64) {
	for _, slot := range slots {
		indexerDB.On("GetEventsBySlot", slot).Return([]solanaTxsStore.EventRecord{}, nil).Once()
		indexerDB.On("GetBlockhashBySlot", slot).Return(blockHash, nil).Once()
	}
}
