package core

import (
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/stretchr/testify/require"
)

func TestBridgingRequestState(t *testing.T) {
	const chainID = common.ChainIDStrPrime

	txHash := common.Hash{1, 88, 208}
	dstTxHash := common.NewHashFromHexString("0xFF")

	t.Run("NewBridgingRequestState", func(t *testing.T) {
		state := NewBridgingRequestState(chainID, txHash)
		require.NotNil(t, state)
		require.Equal(t, chainID, state.SourceChainID)
		require.Equal(t, txHash, state.SourceTxHash)
		require.Equal(t, BridgingRequestStatusDiscoveredOnSource, state.Status)
	})

	t.Run("UpdateDestChainID empty chain id", func(t *testing.T) {
		state := NewBridgingRequestState("", txHash)
		require.NoError(t, state.UpdateDestChainID(chainID))
		require.Equal(t, chainID, state.DestinationChainID)
	})

	t.Run("UpdateDestChainID not empty chain id", func(t *testing.T) {
		state := NewBridgingRequestState(common.ChainIDStrNexus, txHash)
		state.DestinationChainID = common.ChainIDStrVector
		require.Error(t, state.UpdateDestChainID(common.ChainIDStrPrime))
		require.Equal(t, common.ChainIDStrVector, state.DestinationChainID)
	})

	t.Run("IsTransitionPossible BridgingRequestStatusInvalidRequest", func(t *testing.T) {
		state := NewBridgingRequestState(common.ChainIDStrNexus, txHash)
		state.ToInvalidRequest()
		require.Error(t, state.IsTransitionPossible(BridgingRequestStatusDiscoveredOnSource))
		require.Error(t, state.IsTransitionPossible(BridgingRequestStatusExecutedOnDestination))
	})

	t.Run("IsTransitionPossible BridgingRequestStatusInvalidRequest", func(t *testing.T) {
		state := NewBridgingRequestState(common.ChainIDStrNexus, txHash)
		state.ToExecutedOnDestination(dstTxHash)
		require.Error(t, state.IsTransitionPossible(BridgingRequestStatusDiscoveredOnSource))
		require.Error(t, state.IsTransitionPossible(BridgingRequestStatusFailedToExecuteOnDestination))
		require.Error(t, state.IsTransitionPossible(BridgingRequestStatusSubmittedToBridge))
	})

	t.Run("IsTransitionPossible BridgingRequestStatusSubmittedToBridge", func(t *testing.T) {
		state := NewBridgingRequestState(common.ChainIDStrNexus, txHash)
		state.ToSubmittedToBridge()
		require.Error(t, state.IsTransitionPossible(BridgingRequestStatusDiscoveredOnSource))
		require.NoError(t, state.IsTransitionPossible(BridgingRequestStatusFailedToExecuteOnDestination))
		require.NoError(t, state.IsTransitionPossible(BridgingRequestStatusExecutedOnDestination))
	})

	t.Run("IsTransitionPossible BridgingRequestStatusIncludedInBatch", func(t *testing.T) {
		state := NewBridgingRequestState(common.ChainIDStrNexus, txHash)
		state.ToIncludedInBatch()
		require.Error(t, state.IsTransitionPossible(BridgingRequestStatusDiscoveredOnSource))
		require.NoError(t, state.IsTransitionPossible(BridgingRequestStatusFailedToExecuteOnDestination))
		require.NoError(t, state.IsTransitionPossible(BridgingRequestStatusExecutedOnDestination))
	})

	t.Run("IsTransitionPossible BridgingRequestStatusSubmittedToDestination", func(t *testing.T) {
		state := NewBridgingRequestState(common.ChainIDStrNexus, txHash)
		state.ToSubmittedToDestination()
		require.Error(t, state.IsTransitionPossible(BridgingRequestStatusDiscoveredOnSource))
		require.NoError(t, state.IsTransitionPossible(BridgingRequestStatusFailedToExecuteOnDestination))
		require.NoError(t, state.IsTransitionPossible(BridgingRequestStatusExecutedOnDestination))
	})

	t.Run("IsTransitionPossible BridgingRequestStatusDiscoveredOnSource", func(t *testing.T) {
		state := NewBridgingRequestState(common.ChainIDStrNexus, txHash)
		require.NoError(t, state.IsTransitionPossible(BridgingRequestStatusDiscoveredOnSource))
		require.NoError(t, state.IsTransitionPossible(BridgingRequestStatusFailedToExecuteOnDestination))
		require.NoError(t, state.IsTransitionPossible(BridgingRequestStatusExecutedOnDestination))
	})

	t.Run("IsTransitionPossible BridgingRequestStatusFailedToExecuteOnDestination", func(t *testing.T) {
		state := NewBridgingRequestState(common.ChainIDStrNexus, txHash)
		state.ToFailedToExecuteOnDestination()
		require.Error(t, state.IsTransitionPossible(BridgingRequestStatusDiscoveredOnSource))
		require.NoError(t, state.IsTransitionPossible(BridgingRequestStatusExecutedOnDestination))
	})
}
