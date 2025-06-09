package bridge

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	oracleCommon "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	ethCore "github.com/Ethernal-Tech/apex-bridge/oracle_eth/core"
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
				EmptyBlocksThreshold:      4,
			},
		},
	}

	bridgeSubmitter := &ethCore.BridgeSubmitterMock{}
	indexerDB := &ethCore.EventStoreMock{}
	oracleDB := &ethCore.EthTxsProcessorDBMock{}
	testErr := fmt.Errorf("test err")

	t.Run("NewConfirmedBlocksSubmitter GetBlocksSubmitterInfo error", func(t *testing.T) {
		oracleDB.On("GetBlocksSubmitterInfo", chainID).Return(oracleCommon.BlocksSubmitterInfo{}, testErr).Once()

		_, err := NewConfirmedBlocksSubmitter(
			bridgeSubmitter, appConfig, oracleDB, indexerDB, chainID, hclog.NewNullLogger())
		require.ErrorIs(t, err, testErr)
	})

	t.Run("Start ctx done", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		cancel()

		oracleDB.On("GetBlocksSubmitterInfo", chainID).Return(oracleCommon.BlocksSubmitterInfo{}, nil).Once()

		blocksSubmitter, err := NewConfirmedBlocksSubmitter(
			bridgeSubmitter, appConfig, oracleDB, indexerDB, chainID, hclog.NewNullLogger())
		require.NoError(t, err)

		blocksSubmitter.Start(ctx)

		time.Sleep(time.Second)

		require.Equal(t, uint64(0), blocksSubmitter.latestInfo.BlockNumOrSlot)
	})

	t.Run("Execute", func(t *testing.T) {
		oracleDB.On("GetBlocksSubmitterInfo", chainID).Return(oracleCommon.BlocksSubmitterInfo{}, nil).Once()
		indexerDB.On("GetLastProcessedBlock").Return(uint64(20), nil).Once()

		hashes := [6]ethgo.Hash{
			ethgo.HexToHash("F1"), ethgo.HexToHash("F2"), ethgo.HexToHash("F3"),
			ethgo.HexToHash("F4"), ethgo.HexToHash("F5"), ethgo.HexToHash("F6"),
		}

		for i := uint64(0); i <= 15; i++ {
			logs := []*ethgo.Log(nil)

			switch i {
			case 6:
				logs = []*ethgo.Log{
					{TransactionHash: hashes[0]},
					{TransactionHash: hashes[1]},
				}
			case 8:
				logs = []*ethgo.Log{
					{TransactionHash: hashes[2]},
				}
			case 15:
				logs = []*ethgo.Log{
					{TransactionHash: hashes[3]},
					{TransactionHash: hashes[4]},
					{TransactionHash: hashes[5]},
				}
			}

			indexerDB.On("GetLogsByBlockNumber", i).Return(logs, nil).Once()
		}

		submittedBlocks := []eth.CardanoBlock{
			{BlockSlot: big.NewInt(3)}, {BlockSlot: big.NewInt(6)}, {BlockSlot: big.NewInt(8)}, {BlockSlot: big.NewInt(12)},
		}

		oracleDB.On("GetProcessedTx", oracleCommon.DBTxID{ChainID: chainID, DBKey: hashes[0].Bytes()}).Return(&ethCore.ProcessedEthTx{}, nil).Once()
		oracleDB.On("GetProcessedTx", oracleCommon.DBTxID{ChainID: chainID, DBKey: hashes[1].Bytes()}).Return(&ethCore.ProcessedEthTx{}, nil).Once()
		oracleDB.On("GetProcessedTx", oracleCommon.DBTxID{ChainID: chainID, DBKey: hashes[2].Bytes()}).Return(&ethCore.ProcessedEthTx{}, nil).Once()
		oracleDB.On("GetProcessedTx", oracleCommon.DBTxID{ChainID: chainID, DBKey: hashes[3].Bytes()}).Return(&ethCore.ProcessedEthTx{}, nil).Once()
		oracleDB.On("GetProcessedTx", oracleCommon.DBTxID{ChainID: chainID, DBKey: hashes[4].Bytes()}).Return((*ethCore.ProcessedEthTx)(nil), nil).Once()

		oracleDB.On("SetBlocksSubmitterInfo", chainID, oracleCommon.BlocksSubmitterInfo{
			BlockNumOrSlot: 12,
			CounterEmpty:   0,
		}).Return(nil).Once()
		bridgeSubmitter.On("SubmitBlocks", chainID, submittedBlocks).Return(nil).Once()

		blocksSubmitter, err := NewConfirmedBlocksSubmitter(
			bridgeSubmitter, appConfig, oracleDB, indexerDB, chainID, hclog.NewNullLogger())
		require.NoError(t, err)

		require.NoError(t, blocksSubmitter.execute())

		require.Equal(t, uint64(12), blocksSubmitter.latestInfo.BlockNumOrSlot)
		require.Equal(t, 0, blocksSubmitter.latestInfo.CounterEmpty)
	})

	bridgeSubmitter.AssertExpectations(t)
	oracleDB.AssertExpectations(t)
	indexerDB.AssertExpectations(t)
}
