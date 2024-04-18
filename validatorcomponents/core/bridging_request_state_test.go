package core

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBridgingRequestState(t *testing.T) {

	t.Run("NewBridgingRequestState", func(t *testing.T) {
		chainId := "prime"
		txHash := "0xtest"

		state := NewBridgingRequestState(chainId, txHash, nil)
		require.NotNil(t, state)
		require.Equal(t, chainId, state.SourceChainId)
		require.Equal(t, txHash, state.SourceTxHash)
		require.Equal(t, BridgingRequestStatusDiscoveredOnSource, state.Status)
	})

	t.Run("state changes from BridgingRequestStatusDiscoveredOnSource", func(t *testing.T) {
		chainId := "prime"
		txHash := "0xtest"

		state := NewBridgingRequestState(chainId, txHash, nil)
		require.NotNil(t, state)

		err := state.ToIncludedInBatch(1)
		require.Error(t, err)
		require.ErrorContains(t, err, "can not change BridgingRequestState")

		err = state.ToSubmittedToDestination()
		require.Error(t, err)
		require.ErrorContains(t, err, "can not change BridgingRequestState")

		err = state.ToFailedToExecuteOnDestination()
		require.Error(t, err)
		require.ErrorContains(t, err, "can not change BridgingRequestState")

		err = state.ToExecutedOnDestination("0xdest")
		require.Error(t, err)
		require.ErrorContains(t, err, "can not change BridgingRequestState")

		err = state.ToInvalidRequest()
		require.NoError(t, err)
		require.Equal(t, BridgingRequestStatusInvalidRequest, state.Status)

		state = NewBridgingRequestState(chainId, txHash, nil)
		err = state.ToSubmittedToBridge("test")
		require.NoError(t, err)
		require.Equal(t, "test", state.DestinationChainId)
		require.Equal(t, BridgingRequestStatusSubmittedToBridge, state.Status)
	})

	t.Run("state changes from BridgingRequestStatusInvalidRequest", func(t *testing.T) {
		chainId := "prime"
		txHash := "0xtest"

		state := NewBridgingRequestState(chainId, txHash, nil)
		err := state.ToInvalidRequest()
		require.NoError(t, err)

		err = state.ToInvalidRequest()
		require.Error(t, err)
		require.ErrorContains(t, err, "can not change BridgingRequestState")

		err = state.ToSubmittedToBridge("test")
		require.Error(t, err)
		require.ErrorContains(t, err, "can not change BridgingRequestState")

		err = state.ToIncludedInBatch(1)
		require.Error(t, err)
		require.ErrorContains(t, err, "can not change BridgingRequestState")

		err = state.ToSubmittedToDestination()
		require.Error(t, err)
		require.ErrorContains(t, err, "can not change BridgingRequestState")

		err = state.ToFailedToExecuteOnDestination()
		require.Error(t, err)
		require.ErrorContains(t, err, "can not change BridgingRequestState")

		err = state.ToExecutedOnDestination("0xdest")
		require.Error(t, err)
		require.ErrorContains(t, err, "can not change BridgingRequestState")
	})

	t.Run("state changes from BridgingRequestStatusSubmittedToBridge", func(t *testing.T) {
		chainId := "prime"
		txHash := "0xtest"

		state := NewBridgingRequestState(chainId, txHash, nil)
		err := state.ToSubmittedToBridge("test")
		require.NoError(t, err)

		err = state.ToInvalidRequest()
		require.Error(t, err)
		require.ErrorContains(t, err, "can not change BridgingRequestState")

		err = state.ToSubmittedToBridge("test")
		require.Error(t, err)
		require.ErrorContains(t, err, "can not change BridgingRequestState")

		err = state.ToSubmittedToDestination()
		require.Error(t, err)
		require.ErrorContains(t, err, "can not change BridgingRequestState")

		err = state.ToFailedToExecuteOnDestination()
		require.Error(t, err)
		require.ErrorContains(t, err, "can not change BridgingRequestState")

		err = state.ToExecutedOnDestination("0xdest")
		require.Error(t, err)
		require.ErrorContains(t, err, "can not change BridgingRequestState")

		err = state.ToIncludedInBatch(1)
		require.NoError(t, err)
		require.Equal(t, BridgingRequestStatusIncludedInBatch, state.Status)
	})

	t.Run("state changes from BridgingRequestStatusIncludedInBatch", func(t *testing.T) {
		chainId := "prime"
		txHash := "0xtest"

		state := NewBridgingRequestState(chainId, txHash, nil)
		err := state.ToSubmittedToBridge("test")
		require.NoError(t, err)
		err = state.ToIncludedInBatch(1)
		require.NoError(t, err)

		err = state.ToInvalidRequest()
		require.Error(t, err)
		require.ErrorContains(t, err, "can not change BridgingRequestState")

		err = state.ToSubmittedToBridge("test")
		require.Error(t, err)
		require.ErrorContains(t, err, "can not change BridgingRequestState")

		err = state.ToIncludedInBatch(1)
		require.Error(t, err)
		require.ErrorContains(t, err, "can not change BridgingRequestState")

		err = state.ToFailedToExecuteOnDestination()
		require.Error(t, err)
		require.ErrorContains(t, err, "can not change BridgingRequestState")

		err = state.ToExecutedOnDestination("0xdest")
		require.Error(t, err)
		require.ErrorContains(t, err, "can not change BridgingRequestState")

		err = state.ToSubmittedToDestination()
		require.NoError(t, err)
		require.Equal(t, BridgingRequestStatusSubmittedToDestination, state.Status)
	})

	t.Run("state changes from BridgingRequestStatusSubmittedToDestination", func(t *testing.T) {
		chainId := "prime"
		txHash := "0xtest"

		state := NewBridgingRequestState(chainId, txHash, nil)
		err := state.ToSubmittedToBridge("test")
		require.NoError(t, err)
		err = state.ToIncludedInBatch(1)
		require.NoError(t, err)
		err = state.ToSubmittedToDestination()
		require.NoError(t, err)

		err = state.ToInvalidRequest()
		require.Error(t, err)
		require.ErrorContains(t, err, "can not change BridgingRequestState")

		err = state.ToSubmittedToBridge("test")
		require.Error(t, err)
		require.ErrorContains(t, err, "can not change BridgingRequestState")

		err = state.ToIncludedInBatch(1)
		require.Error(t, err)
		require.ErrorContains(t, err, "can not change BridgingRequestState")

		err = state.ToSubmittedToDestination()
		require.Error(t, err)
		require.ErrorContains(t, err, "can not change BridgingRequestState")

		err = state.ToFailedToExecuteOnDestination()
		require.NoError(t, err)
		require.Equal(t, BridgingRequestStatusFailedToExecuteOnDestination, state.Status)

		state = NewBridgingRequestState(chainId, txHash, nil)
		err = state.ToSubmittedToBridge("test")
		require.NoError(t, err)
		err = state.ToIncludedInBatch(1)
		require.NoError(t, err)
		err = state.ToSubmittedToDestination()
		require.NoError(t, err)

		err = state.ToExecutedOnDestination("0xdest")
		require.NoError(t, err)
		require.Equal(t, BridgingRequestStatusExecutedOnDestination, state.Status)
	})

	t.Run("state changes from BridgingRequestStatusFailedToExecuteOnDestination", func(t *testing.T) {
		chainId := "prime"
		txHash := "0xtest"

		state := NewBridgingRequestState(chainId, txHash, nil)
		err := state.ToSubmittedToBridge("test")
		require.NoError(t, err)
		err = state.ToIncludedInBatch(1)
		require.NoError(t, err)
		err = state.ToSubmittedToDestination()
		require.NoError(t, err)
		err = state.ToFailedToExecuteOnDestination()
		require.NoError(t, err)

		err = state.ToInvalidRequest()
		require.Error(t, err)
		require.ErrorContains(t, err, "can not change BridgingRequestState")

		err = state.ToSubmittedToBridge("test")
		require.Error(t, err)
		require.ErrorContains(t, err, "can not change BridgingRequestState")

		err = state.ToIncludedInBatch(1)
		require.Error(t, err)
		require.ErrorContains(t, err, "can not change BridgingRequestState")

		err = state.ToSubmittedToDestination()
		require.Error(t, err)
		require.ErrorContains(t, err, "can not change BridgingRequestState")

		err = state.ToFailedToExecuteOnDestination()
		require.Error(t, err)
		require.ErrorContains(t, err, "can not change BridgingRequestState")

		err = state.ToExecutedOnDestination("0xdest")
		require.Error(t, err)
		require.ErrorContains(t, err, "can not change BridgingRequestState")
	})

	t.Run("state changes from BridgingRequestStatusExecutedOnDestination", func(t *testing.T) {
		chainId := "prime"
		txHash := "0xtest"

		state := NewBridgingRequestState(chainId, txHash, nil)
		err := state.ToSubmittedToBridge("test")
		require.NoError(t, err)
		err = state.ToIncludedInBatch(1)
		require.NoError(t, err)
		err = state.ToSubmittedToDestination()
		require.NoError(t, err)
		err = state.ToExecutedOnDestination("0xdest")
		require.NoError(t, err)

		err = state.ToInvalidRequest()
		require.Error(t, err)
		require.ErrorContains(t, err, "can not change BridgingRequestState")

		err = state.ToSubmittedToBridge("test")
		require.Error(t, err)
		require.ErrorContains(t, err, "can not change BridgingRequestState")

		err = state.ToIncludedInBatch(1)
		require.Error(t, err)
		require.ErrorContains(t, err, "can not change BridgingRequestState")

		err = state.ToSubmittedToDestination()
		require.Error(t, err)
		require.ErrorContains(t, err, "can not change BridgingRequestState")

		err = state.ToFailedToExecuteOnDestination()
		require.Error(t, err)
		require.ErrorContains(t, err, "can not change BridgingRequestState")

		err = state.ToExecutedOnDestination("0xdest")
		require.Error(t, err)
		require.ErrorContains(t, err, "can not change BridgingRequestState")
	})
}
