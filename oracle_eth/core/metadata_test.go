package core

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEthMetadata(t *testing.T) {
	t.Run("Marshal BaseMetadata", func(t *testing.T) {
		result, err := MarshalEthMetadata(BaseEthMetadata{BridgingTxType: "test"})

		require.NoError(t, err)
		require.NotNil(t, result)
	})

	t.Run("Unmarshal BaseMetadata", func(t *testing.T) {
		result, err := MarshalEthMetadata(BaseEthMetadata{BridgingTxType: "test"})
		require.NoError(t, err)
		require.NotNil(t, result)

		metadata, err := UnmarshalEthMetadata[BaseEthMetadata](result)
		require.NoError(t, err)
		require.NotNil(t, metadata)
	})

	t.Run("Marshal BridgingRequestMetadata", func(t *testing.T) {
		result, err := MarshalEthMetadata(BridgingRequestEthMetadata{BridgingTxType: "test"})

		require.NoError(t, err)
		require.NotNil(t, result)
	})

	t.Run("Unmarshal BridgingRequestMetadata", func(t *testing.T) {
		result, err := MarshalEthMetadata(BridgingRequestEthMetadata{BridgingTxType: "test"})
		require.NoError(t, err)
		require.NotNil(t, result)

		metadata, err := UnmarshalEthMetadata[BridgingRequestEthMetadata](result)
		require.NoError(t, err)
		require.NotNil(t, metadata)
	})

	t.Run("Marshal BatchExecutedMetadata", func(t *testing.T) {
		result, err := MarshalEthMetadata(BatchExecutedEthMetadata{BridgingTxType: "test"})

		require.NoError(t, err)
		require.NotNil(t, result)
	})

	t.Run("Unmarshal BatchExecutedMetadata", func(t *testing.T) {
		result, err := MarshalEthMetadata(BatchExecutedEthMetadata{BatchNonceID: 245})
		require.NoError(t, err)
		require.NotNil(t, result)

		metadata, err := UnmarshalEthMetadata[BatchExecutedEthMetadata](result)
		require.NoError(t, err)
		require.NotNil(t, metadata)
		require.Equal(t, uint64(245), metadata.BatchNonceID)
	})
}
