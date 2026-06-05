package bridge

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	oracleCommon "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	solanaCore "github.com/Ethernal-Tech/apex-bridge/oracle_solana/core"
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
		indexerDB.On("GetLatestFinalizedBlockNumber").Return((uint64)(0), testErr).Once()

		_, err := NewConfirmedBlocksSubmitter(
			bridgeSubmitter, appConfig, oracleDB, indexerDB, chainID, hclog.NewNullLogger())
		require.ErrorContains(t, err, "failed to create block submitter")
		require.ErrorContains(t, err, testErr.Error())
	})

	t.Run("NewConfirmedBlocksSubmitter UpdateFromIndexerDB updates latest info", func(t *testing.T) {
		startBlockNum := uint64(1044)
		appConfig := &oracleCommon.AppConfig{
			Bridge: oracleCommon.BridgeConfig{SubmitConfig: testSolanaSubmitConfig(chainID)},
		}
		appConfig.Bridge.SubmitConfig.UpdateFromIndexerDB = true

		bridgeSubmitter := &solanaCore.SolanaBridgeSubmitterMock{}
		indexerDB := &solanaTxsStore.MockStorageHandler{}
		oracleDB := &solanaCore.SolanaTxsProcessorDBMock{}
		oracleDB.On("GetBlocksSubmitterInfo", chainID).Return(oracleCommon.BlocksSubmitterInfo{
			BlockNumOrSlot: startBlockNum - 1,
			CounterEmpty:   3,
		}, nil).Once()
		indexerDB.On("GetLatestFinalizedBlockNumber").Return(startBlockNum, nil).Once()

		bs, err := NewConfirmedBlocksSubmitter(
			bridgeSubmitter, appConfig, oracleDB, indexerDB, chainID, hclog.NewNullLogger())
		require.NoError(t, err)
		require.Equal(t, startBlockNum, bs.latestInfo.BlockNumOrSlot)
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
		indexerDB.On("GetLatestFinalizedBlockNumber").Return((uint64)(0), testErr).Once()

		blocksSubmitter, err := NewConfirmedBlocksSubmitter(
			bridgeSubmitter, appConfig, oracleDB, indexerDB, chainID, hclog.NewNullLogger())
		require.NoError(t, err)

		err = blocksSubmitter.execute()
		require.ErrorContains(t, err, "error getting latest finalized block number: test err")
	})

	t.Run("execute submit blocks error", func(t *testing.T) {
		appConfig := &oracleCommon.AppConfig{
			Bridge: oracleCommon.BridgeConfig{SubmitConfig: testSolanaSubmitConfig(chainID)},
		}
		bridgeSubmitter := &solanaCore.SolanaBridgeSubmitterMock{}
		indexerDB := &solanaTxsStore.MockStorageHandler{}
		oracleDB := &solanaCore.SolanaTxsProcessorDBMock{}
		oracleDB.On("GetBlocksSubmitterInfo", chainID).Return(oracleCommon.BlocksSubmitterInfo{
			BlockNumOrSlot: 1,
		}, nil).Once()

		indexerDB.On("GetLatestFinalizedBlockNumber").Return(uint64(20), nil).Once()
		mockEmptyBlocks(indexerDB, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20)
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
		}
		bridgeSubmitter := &solanaCore.SolanaBridgeSubmitterMock{}
		indexerDB := &solanaTxsStore.MockStorageHandler{}
		oracleDB := &solanaCore.SolanaTxsProcessorDBMock{}
		oracleDB.On("GetBlocksSubmitterInfo", chainID).Return(oracleCommon.BlocksSubmitterInfo{
			BlockNumOrSlot: 1,
		}, nil).Once()

		indexerDB.On("GetLatestFinalizedBlockNumber").Return(uint64(3), nil).Once()
		mockEmptyBlocks(indexerDB, 2, 3)
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
		}
		bridgeSubmitter := &solanaCore.SolanaBridgeSubmitterMock{}
		indexerDB := &solanaTxsStore.MockStorageHandler{}
		oracleDB := &solanaCore.SolanaTxsProcessorDBMock{}
		oracleDB.On("GetBlocksSubmitterInfo", chainID).Return(oracleCommon.BlocksSubmitterInfo{
			BlockNumOrSlot: 1,
		}, nil).Once()

		pendingSig, signErr := solana.NewWallet().PrivateKey.Sign([]byte("pending-tx"))
		require.NoError(t, signErr)
		indexerDB.On("GetLatestFinalizedBlockNumber").Return(uint64(6), nil).Once()
		mockEmptyBlocks(indexerDB, 2, 3)
		indexerDB.On("GetEventsByBlockNumber", uint64(4)).Return([]solanaTxsStore.EventRecord{
			{Slot: 4, TxSignature: pendingSig.String()},
		}, nil).Once()
		oracleDB.On("GetProcessedTx", oracleCommon.DBTxID{
			ChainID: chainID,
			DBKey:   pendingSig[:],
		}).Return((*solanaCore.ProcessedSolanaTx)(nil), nil).Once()

		bridgeSubmitter.On("SubmitBlocks", chainID, []eth.CardanoBlock{
			{BlockSlot: big.NewInt(3)},
		}).Return(nil).Once()
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
		}
		bridgeSubmitter := &solanaCore.SolanaBridgeSubmitterMock{}
		indexerDB := &solanaTxsStore.MockStorageHandler{}
		oracleDB := &solanaCore.SolanaTxsProcessorDBMock{}
		oracleDB.On("GetBlocksSubmitterInfo", chainID).Return(oracleCommon.BlocksSubmitterInfo{}, nil).Once()
		indexerDB.On("GetLatestFinalizedBlockNumber").Return(uint64(1), nil).Once()
		indexerDB.On("GetEventsByBlockNumber", uint64(1)).Return([]solanaTxsStore.EventRecord{}, nil).Once()

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
		}
		bridgeSubmitter := &solanaCore.SolanaBridgeSubmitterMock{}
		indexerDB := &solanaTxsStore.MockStorageHandler{}
		oracleDB := &solanaCore.SolanaTxsProcessorDBMock{}
		oracleDB.On("GetBlocksSubmitterInfo", chainID).Return(oracleCommon.BlocksSubmitterInfo{}, nil).Once()
		indexerDB.On("GetLatestFinalizedBlockNumber").Return(uint64(1), nil).Once()
		indexerDB.On("GetEventsByBlockNumber", uint64(1)).Return([]solanaTxsStore.EventRecord{}, nil).Once()

		blocksSubmitter, err := NewConfirmedBlocksSubmitter(
			bridgeSubmitter, appConfig, oracleDB, indexerDB, chainID, hclog.NewNullLogger())
		require.NoError(t, err)

		_, _, err = blocksSubmitter.getBlocksToSubmit(1)
		require.ErrorContains(t, err, "threshold too large")
	})

	t.Run("getBlocksToSubmit get events by block number error", func(t *testing.T) {
		appConfig := &oracleCommon.AppConfig{
			Bridge: oracleCommon.BridgeConfig{SubmitConfig: testSolanaSubmitConfig(chainID)},
		}
		bridgeSubmitter := &solanaCore.SolanaBridgeSubmitterMock{}
		indexerDB := &solanaTxsStore.MockStorageHandler{}
		oracleDB := &solanaCore.SolanaTxsProcessorDBMock{}
		oracleDB.On("GetBlocksSubmitterInfo", chainID).Return(oracleCommon.BlocksSubmitterInfo{}, nil).Once()
		indexerDB.On("GetLatestFinalizedBlockNumber").Return(uint64(2), nil).Once()
		indexerDB.On("GetEventsByBlockNumber", uint64(1)).Return([]solanaTxsStore.EventRecord(nil), testErr).Once()

		blocksSubmitter, err := NewConfirmedBlocksSubmitter(
			bridgeSubmitter, appConfig, oracleDB, indexerDB, chainID, hclog.NewNullLogger())
		require.NoError(t, err)

		_, _, err = blocksSubmitter.getBlocksToSubmit(1)
		require.ErrorContains(t, err, "failed to get events for slot")
	})
}

func mockEmptyBlocks(indexerDB *solanaTxsStore.MockStorageHandler, blockNums ...uint64) {
	for _, blockNum := range blockNums {
		indexerDB.On("GetEventsByBlockNumber", blockNum).Return([]solanaTxsStore.EventRecord{}, nil).Once()
	}
}
