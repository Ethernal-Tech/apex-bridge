package common

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBridgingRequestState(t *testing.T) {
	const chainID = ChainIDStrPrime

	txHash := []byte{1, 88, 208}
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
		state.ToExecutedOnDestination(dstTxHash[:])
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

func TestNewBridgingRequestStateDetails(t *testing.T) {
	const (
		currencyID = uint16(1)
		wrappedID  = uint16(2)
		otherID    = uint16(3)
	)

	t.Run("sums currency and wrapped token separately", func(t *testing.T) {
		details := NewBridgingRequestStateDetails("sender", []BridgingRequestStateReceiver{
			{Address: "addr1", Amount: big.NewInt(10), TokenID: currencyID},
			{Address: "addr2", Amount: big.NewInt(5), TokenID: currencyID},
			{Address: "addr3", Amount: big.NewInt(70), TokenID: wrappedID},
		}, big.NewInt(1), big.NewInt(2), currencyID)

		require.Equal(t, "sender", details.SenderAddr)
		require.Len(t, details.Receivers, 3)
		require.Equal(t, big.NewInt(15), details.Amount)
		require.Equal(t, big.NewInt(70), details.TokenAmount)
		require.Equal(t, wrappedID, details.TokenID)
		require.Equal(t, big.NewInt(1), details.BridgingFee)
		require.Equal(t, big.NewInt(2), details.OperationFee)
	})

	t.Run("token id stays zero when only currency is bridged", func(t *testing.T) {
		details := NewBridgingRequestStateDetails("sender", []BridgingRequestStateReceiver{
			{Address: "addr1", Amount: big.NewInt(10), TokenID: currencyID},
		}, nil, nil, currencyID)

		require.Equal(t, big.NewInt(10), details.Amount)
		require.Equal(t, big.NewInt(0), details.TokenAmount)
		require.Zero(t, details.TokenID)
	})

	// a colored coin is not the chain currency and is not flagged IsWrappedCurrency, but it is still
	// bridged and has to reach the history
	t.Run("counts a token that is not the wrapped currency", func(t *testing.T) {
		details := NewBridgingRequestStateDetails("sender", []BridgingRequestStateReceiver{
			{Address: "addr1", Amount: big.NewInt(10), TokenID: otherID},
		}, nil, nil, currencyID)

		require.Equal(t, big.NewInt(0), details.Amount)
		require.Equal(t, big.NewInt(10), details.TokenAmount)
		require.Equal(t, otherID, details.TokenID)
	})

	t.Run("skips receivers with no amount", func(t *testing.T) {
		details := NewBridgingRequestStateDetails("sender", []BridgingRequestStateReceiver{
			{Address: "addr1", Amount: nil, TokenID: currencyID},
			{Address: "addr2", Amount: nil, TokenID: otherID},
		}, nil, nil, currencyID)

		require.Equal(t, big.NewInt(0), details.Amount)
		require.Equal(t, big.NewInt(0), details.TokenAmount)
		require.Zero(t, details.TokenID)
	})
}
