package bridge

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestLastObserved(t *testing.T) {
	t.Run("get last observed block success", func(t *testing.T) {
		bridgeSC := &eth.OracleBridgeSmartContractMock{}
		expectedSlot := big.NewInt(42)

		mockBlockReturn := eth.CardanoBlock{
			BlockSlot: expectedSlot,
		}

		bridgeSC.On("GetLastObservedBlock").Return(mockBlockReturn, nil)

		lastObserved := NewLastObserved(bridgeSC, hclog.NewNullLogger())
		require.NotNil(t, lastObserved)
		resultSlot, err := lastObserved.GetLastObservedBlock(context.Background(), common.ChainIDStrPrime)

		require.NoError(t, err)
		require.NotNil(t, resultSlot)
		require.Equal(t, expectedSlot, resultSlot)

		bridgeSC.AssertExpectations(t)
	})

	t.Run("get last observed block failure", func(t *testing.T) {
		bridgeSC := eth.OracleBridgeSmartContractMock{}
		expectedError := errors.New("failed to fetch block")

		bridgeSC.On("GetLastObservedBlock").Return(eth.CardanoBlock{}, expectedError)

		lastObserved := NewLastObserved(&bridgeSC, hclog.NewNullLogger())

		resultSlot, err := lastObserved.GetLastObservedBlock(context.Background(), common.ChainIDStrPrime)

		require.Error(t, err)
		require.Nil(t, resultSlot)
		require.Equal(t, expectedError, err)

		bridgeSC.AssertExpectations(t)
	})
}
