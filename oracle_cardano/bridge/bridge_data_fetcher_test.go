package bridge

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestBridgeDataFetcher(t *testing.T) {
	t.Run("NewBridgeDataFetcher", func(t *testing.T) {
		bridgeSC := &eth.OracleBridgeSmartContractMock{}
		bridgeDataFetcher := NewBridgeDataFetcher(context.Background(), bridgeSC, hclog.NewNullLogger())

		require.NotNil(t, bridgeDataFetcher)
	})

	t.Run("FetchExpectedTx nil", func(t *testing.T) {
		bridgeSC := &eth.OracleBridgeSmartContractMock{}
		bridgeSC.On("GetRawTransactionFromLastBatch").Return(nil, nil)
		bridgeDataFetcher := NewBridgeDataFetcher(context.Background(), bridgeSC, hclog.NewNullLogger())

		require.NotNil(t, bridgeDataFetcher)

		expectedTx, err := bridgeDataFetcher.FetchExpectedTx(common.ChainIDStrPrime)
		require.NoError(t, err)
		require.Nil(t, expectedTx)
	})

	t.Run("FetchExpectedTx err", func(t *testing.T) {
		bridgeSC := &eth.OracleBridgeSmartContractMock{}
		bridgeSC.On("GetRawTransactionFromLastBatch").Return(nil, fmt.Errorf("test err"))
		bridgeDataFetcher := NewBridgeDataFetcher(context.Background(), bridgeSC, hclog.NewNullLogger())

		require.NotNil(t, bridgeDataFetcher)

		expectedTx, err := bridgeDataFetcher.FetchExpectedTx(common.ChainIDStrPrime)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to FetchExpectedTx from Bridge SC")
		require.Nil(t, expectedTx)
	})

	t.Run("FetchExpectedTx parse tx fail", func(t *testing.T) {
		bridgeSC := &eth.OracleBridgeSmartContractMock{}
		bridgeSC.On("GetRawTransactionFromLastBatch").Return([]byte{12, 33}, nil)
		bridgeDataFetcher := NewBridgeDataFetcher(context.Background(), bridgeSC, hclog.NewNullLogger())

		require.NotNil(t, bridgeDataFetcher)

		expectedTx, err := bridgeDataFetcher.FetchExpectedTx(common.ChainIDStrPrime)
		require.Error(t, err)
		require.Nil(t, expectedTx)
	})

	t.Run("FetchLatestBlockPoint nil", func(t *testing.T) {
		bridgeSC := &eth.OracleBridgeSmartContractMock{}
		bridgeSC.On("GetLastObservedBlock").Return(eth.CardanoBlock{}, nil)
		bridgeDataFetcher := NewBridgeDataFetcher(context.Background(), bridgeSC, hclog.NewNullLogger())

		require.NotNil(t, bridgeDataFetcher)

		blockPoint, err := bridgeDataFetcher.FetchLatestBlockPoint(common.ChainIDStrPrime)
		require.NoError(t, err)
		require.Nil(t, blockPoint)
	})

	t.Run("FetchLatestBlockPoint err", func(t *testing.T) {
		bridgeSC := &eth.OracleBridgeSmartContractMock{}
		bridgeSC.On("GetLastObservedBlock").Return(eth.CardanoBlock{}, fmt.Errorf("test err"))
		bridgeDataFetcher := NewBridgeDataFetcher(context.Background(), bridgeSC, hclog.NewNullLogger())

		require.NotNil(t, bridgeDataFetcher)

		blockPoint, err := bridgeDataFetcher.FetchLatestBlockPoint(common.ChainIDStrPrime)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to FetchLatestBlockPoint from Bridge SC")
		require.Nil(t, blockPoint)
	})

	t.Run("FetchLatestBlockPoint valid", func(t *testing.T) {
		bHash := indexer.Hash(common.NewHashFromHexString("FFBB"))
		bSlot := uint64(100)

		bridgeSC := &eth.OracleBridgeSmartContractMock{}
		bridgeSC.On("GetLastObservedBlock").Return(eth.CardanoBlock{
			BlockSlot: new(big.Int).SetUint64(0),
			BlockHash: bHash,
		}, error(nil)).Once()
		bridgeSC.On("GetLastObservedBlock").Return(eth.CardanoBlock{
			BlockSlot: new(big.Int).SetUint64(bSlot),
			BlockHash: bHash,
		}, error(nil))

		bridgeDataFetcher := NewBridgeDataFetcher(context.Background(), bridgeSC, hclog.NewNullLogger())

		require.NotNil(t, bridgeDataFetcher)

		blockPoint, err := bridgeDataFetcher.FetchLatestBlockPoint(common.ChainIDStrPrime)
		require.NoError(t, err)
		require.Nil(t, blockPoint)

		blockPoint, err = bridgeDataFetcher.FetchLatestBlockPoint(common.ChainIDStrPrime)
		require.NoError(t, err)
		require.NotNil(t, blockPoint)
		require.Equal(t, bHash, blockPoint.BlockHash)
		require.Equal(t, bSlot, blockPoint.BlockSlot)
	})
}
