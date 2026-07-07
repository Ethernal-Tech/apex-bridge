package core

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSolMetadata(t *testing.T) {
	t.Run("Marshal BaseMetadata", func(t *testing.T) {
		result, err := MarshalSolMetadata(BaseSolMetadata{BridgingTxType: "test"})

		require.NoError(t, err)
		require.NotNil(t, result)
	})

	t.Run("Unmarshal BaseMetadata", func(t *testing.T) {
		result, err := MarshalSolMetadata(BaseSolMetadata{BridgingTxType: "test"})
		require.NoError(t, err)
		require.NotNil(t, result)

		metadata, err := UnmarshalSolMetadata[BaseSolMetadata](result)
		require.NoError(t, err)
		require.NotNil(t, metadata)
	})

	t.Run("Marshal BridgingRequestMetadata", func(t *testing.T) {
		result, err := MarshalSolMetadata(BridgingRequestSolMetadata{BridgingTxType: "test"})

		require.NoError(t, err)
		require.NotNil(t, result)
	})

	t.Run("Unmarshal BridgingRequestMetadata", func(t *testing.T) {
		result, err := MarshalSolMetadata(BridgingRequestSolMetadata{BridgingTxType: "test"})
		require.NoError(t, err)
		require.NotNil(t, result)

		metadata, err := UnmarshalSolMetadata[BridgingRequestSolMetadata](result)
		require.NoError(t, err)
		require.NotNil(t, metadata)
	})

	t.Run("Marshal BatchExecutedMetadata", func(t *testing.T) {
		result, err := MarshalSolMetadata(BatchExecutedSolMetadata{BridgingTxType: "test"})

		require.NoError(t, err)
		require.NotNil(t, result)
	})

	t.Run("Unmarshal BatchExecutedMetadata", func(t *testing.T) {
		result, err := MarshalSolMetadata(BatchExecutedSolMetadata{BatchNonceID: 245})
		require.NoError(t, err)
		require.NotNil(t, result)

		metadata, err := UnmarshalSolMetadata[BatchExecutedSolMetadata](result)
		require.NoError(t, err)
		require.NotNil(t, metadata)
		require.Equal(t, uint64(245), metadata.BatchNonceID)
	})
}
