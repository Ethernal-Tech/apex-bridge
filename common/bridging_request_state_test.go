package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBridgingRequestState(t *testing.T) {
	const chainID = ChainIDStrPrime

	txHash := Hash{1, 88, 208}
	dstTxHash := NewHashFromHexString("0xFF")

	t.Run("NewBridgingRequestState", func(t *testing.T) {
		state := NewBridgingRequestState(chainID, txHash, false)
		require.NotNil(t, state)
		require.Equal(t, chainID, state.SourceChainID)
		require.Equal(t, txHash, state.SourceTxHash)
		require.Equal(t, BridgingRequestStatusDiscoveredOnSource, state.Status)
	})

	t.Run("IsTransitionPossible BridgingRequestStatusInvalidRequest", func(t *testing.T) {
		state := NewBridgingRequestState(ChainIDStrNexus, txHash, false)
		state.ToInvalidRequest()
		require.Error(t, state.IsTransitionPossible(BridgingRequestStatusDiscoveredOnSource))
		require.Error(t, state.IsTransitionPossible(BridgingRequestStatusExecutedOnDestination))
	})

	t.Run("IsTransitionPossible BridgingRequestStatusInvalidRequest", func(t *testing.T) {
		state := NewBridgingRequestState(ChainIDStrNexus, txHash, false)
		state.ToExecutedOnDestination(dstTxHash)
		require.Error(t, state.IsTransitionPossible(BridgingRequestStatusDiscoveredOnSource))
		require.Error(t, state.IsTransitionPossible(BridgingRequestStatusFailedToExecuteOnDestination))
		require.Error(t, state.IsTransitionPossible(BridgingRequestStatusSubmittedToBridge))
	})

	t.Run("IsTransitionPossible BridgingRequestStatusSubmittedToBridge", func(t *testing.T) {
		state := NewBridgingRequestState(ChainIDStrNexus, txHash, false)
		state.ToSubmittedToBridge()
		require.Error(t, state.IsTransitionPossible(BridgingRequestStatusDiscoveredOnSource))
		require.NoError(t, state.IsTransitionPossible(BridgingRequestStatusFailedToExecuteOnDestination))
		require.NoError(t, state.IsTransitionPossible(BridgingRequestStatusExecutedOnDestination))
	})

	t.Run("IsTransitionPossible BridgingRequestStatusIncludedInBatch", func(t *testing.T) {
		state := NewBridgingRequestState(ChainIDStrNexus, txHash, false)
		state.ToIncludedInBatch()
		require.Error(t, state.IsTransitionPossible(BridgingRequestStatusDiscoveredOnSource))
		require.NoError(t, state.IsTransitionPossible(BridgingRequestStatusFailedToExecuteOnDestination))
		require.NoError(t, state.IsTransitionPossible(BridgingRequestStatusExecutedOnDestination))
	})

	t.Run("IsTransitionPossible BridgingRequestStatusSubmittedToDestination", func(t *testing.T) {
		state := NewBridgingRequestState(ChainIDStrNexus, txHash, false)
		state.ToSubmittedToDestination()
		require.Error(t, state.IsTransitionPossible(BridgingRequestStatusDiscoveredOnSource))
		require.NoError(t, state.IsTransitionPossible(BridgingRequestStatusFailedToExecuteOnDestination))
		require.NoError(t, state.IsTransitionPossible(BridgingRequestStatusExecutedOnDestination))
	})

	t.Run("IsTransitionPossible BridgingRequestStatusDiscoveredOnSource", func(t *testing.T) {
		state := NewBridgingRequestState(ChainIDStrNexus, txHash, false)
		require.NoError(t, state.IsTransitionPossible(BridgingRequestStatusDiscoveredOnSource))
		require.NoError(t, state.IsTransitionPossible(BridgingRequestStatusFailedToExecuteOnDestination))
		require.NoError(t, state.IsTransitionPossible(BridgingRequestStatusExecutedOnDestination))
	})

	t.Run("IsTransitionPossible BridgingRequestStatusFailedToExecuteOnDestination", func(t *testing.T) {
		state := NewBridgingRequestState(ChainIDStrNexus, txHash, false)
		state.ToFailedToExecuteOnDestination()
		require.Error(t, state.IsTransitionPossible(BridgingRequestStatusDiscoveredOnSource))
		require.NoError(t, state.IsTransitionPossible(BridgingRequestStatusExecutedOnDestination))
	})
}
