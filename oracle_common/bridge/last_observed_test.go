package bridge

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestLastObserved(t *testing.T) {
	t.Run("get last observed block success", func(t *testing.T) {
		// 1. Setup the Mock
		bridgeSC := eth.OracleBridgeSmartContractMock{}
		expectedSlot := big.NewInt(42)

		// Create the block structure that the SC returns
		mockBlockReturn := &eth.CardanoBlock{
			BlockSlot: expectedSlot,
		}

		// Mock expectations:
		// Expect GetLastObservedBlock to be called with any Context and specific chainID
		// Return the mock block and no error
		bridgeSC.On("GetLastObservedBlock", mock.Anything, common.ChainIDStrPrime).Return(mockBlockReturn, nil)

		// 2. Initialize the struct
		lastObserved := NewLastObserved(context.Background(), &bridgeSC, hclog.NewNullLogger())
		require.NotNil(t, lastObserved)

		// 3. Execute the method
		resultSlot, err := lastObserved.GetLastObservedBlock(common.ChainIDStrPrime)

		// 4. Assertions
		require.NoError(t, err)
		require.NotNil(t, resultSlot)
		require.Equal(t, expectedSlot, resultSlot)

		// Verify the mock was called as expected
		bridgeSC.AssertExpectations(t)
	})

	t.Run("get last observed block failure", func(t *testing.T) {
		// 1. Setup the Mock
		bridgeSC := eth.OracleBridgeSmartContractMock{}
		expectedError := errors.New("failed to fetch block")

		// Mock expectations:
		// Return nil block and an error
		bridgeSC.On("GetLastObservedBlock", mock.Anything, common.ChainIDStrPrime).Return(nil, expectedError)

		// 2. Initialize the struct
		lastObserved := NewLastObserved(context.Background(), &bridgeSC, hclog.NewNullLogger())

		// 3. Execute the method
		resultSlot, err := lastObserved.GetLastObservedBlock(common.ChainIDStrPrime)

		// 4. Assertions
		require.Error(t, err)
		require.Nil(t, resultSlot)
		require.Equal(t, expectedError, err)

		bridgeSC.AssertExpectations(t)
	})
}
