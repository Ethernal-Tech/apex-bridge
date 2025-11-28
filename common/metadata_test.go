package common

import (
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/require"
)

func TestMetadata(t *testing.T) {
	t.Run("Json Marshal BaseMetadata unsupported encoding", func(t *testing.T) {
		result, err := MarshalMetadata[BaseMetadata]("invalid", BaseMetadata{BridgingTxType: "test"})

		require.Error(t, err)
		require.ErrorContains(t, err, "unsupported metadata encoding type")
		require.Nil(t, result)
	})

	t.Run("Json Unmarshal BaseMetadata  unsupported encoding", func(t *testing.T) {
		result, err := MarshalMetadata[BaseMetadata](MetadataEncodingTypeJSON, BaseMetadata{BridgingTxType: "test"})
		require.NoError(t, err)
		require.NotNil(t, result)

		metadata, err := UnmarshalMetadata[BaseMetadata]("invalid", result)
		require.Error(t, err)
		require.ErrorContains(t, err, "unsupported metadata encoding type")
		require.Nil(t, metadata)
	})

	t.Run("Json Marshal BaseMetadata", func(t *testing.T) {
		result, err := MarshalMetadata[BaseMetadata](MetadataEncodingTypeJSON, BaseMetadata{BridgingTxType: "test"})

		require.NoError(t, err)
		require.NotNil(t, result)
	})

	t.Run("Json Unmarshal BaseMetadata", func(t *testing.T) {
		result, err := SimulateRealMetadata(MetadataEncodingTypeJSON, BaseMetadata{BridgingTxType: "test"})
		require.NoError(t, err)
		require.NotNil(t, result)

		metadata, err := UnmarshalMetadata[BaseMetadata](MetadataEncodingTypeJSON, result)
		require.NoError(t, err)
		require.NotNil(t, metadata)
	})

	t.Run("Cbor Marshal BaseMetadata", func(t *testing.T) {
		result, err := MarshalMetadata[BaseMetadata](MetadataEncodingTypeCbor, BaseMetadata{BridgingTxType: "test"})

		require.NoError(t, err)
		require.NotNil(t, result)
	})

	t.Run("Cbor Unmarshal BaseMetadata", func(t *testing.T) {
		result, err := SimulateRealMetadata(MetadataEncodingTypeCbor, BaseMetadata{BridgingTxType: "test"})
		require.NoError(t, err)
		require.NotNil(t, result)

		metadata, err := UnmarshalMetadata[BaseMetadata](MetadataEncodingTypeCbor, result)
		require.NoError(t, err)
		require.NotNil(t, metadata)
	})

	t.Run("Json Marshal BridgingRequestMetadata", func(t *testing.T) {
		result, err := MarshalMetadata[BridgingRequestMetadata](
			MetadataEncodingTypeJSON, BridgingRequestMetadata{BridgingTxType: "test"})

		require.NoError(t, err)
		require.NotNil(t, result)
	})

	t.Run("Json Unmarshal BridgingRequestMetadata", func(t *testing.T) {
		result, err := SimulateRealMetadata(
			MetadataEncodingTypeJSON, BridgingRequestMetadata{BridgingTxType: "test"})
		require.NoError(t, err)
		require.NotNil(t, result)

		metadata, err := UnmarshalMetadata[BridgingRequestMetadata](MetadataEncodingTypeJSON, result)
		require.NoError(t, err)
		require.NotNil(t, metadata)
	})

	t.Run("Cbor Marshal BridgingRequestMetadata", func(t *testing.T) {
		result, err := MarshalMetadata[BridgingRequestMetadata](
			MetadataEncodingTypeCbor, BridgingRequestMetadata{BridgingTxType: "test"})

		require.NoError(t, err)
		require.NotNil(t, result)
	})

	t.Run("Cbor Unmarshal BridgingRequestMetadata", func(t *testing.T) {
		result, err := SimulateRealMetadata(
			MetadataEncodingTypeCbor, BridgingRequestMetadata{BridgingTxType: "test"})
		require.NoError(t, err)
		require.NotNil(t, result)

		metadata, err := UnmarshalMetadata[BridgingRequestMetadata](MetadataEncodingTypeCbor, result)
		require.NoError(t, err)
		require.NotNil(t, metadata)
	})

	t.Run("Json Marshal BatchExecutedMetadata", func(t *testing.T) {
		result, err := MarshalMetadata[BatchExecutedMetadata](MetadataEncodingTypeJSON, BatchExecutedMetadata{BridgingTxType: "test"})

		require.NoError(t, err)
		require.NotNil(t, result)
	})

	t.Run("Json Unmarshal BatchExecutedMetadata", func(t *testing.T) {
		result, err := SimulateRealMetadata(MetadataEncodingTypeJSON, BatchExecutedMetadata{BatchNonceID: 245})
		require.NoError(t, err)
		require.NotNil(t, result)

		metadata, err := UnmarshalMetadata[BatchExecutedMetadata](MetadataEncodingTypeJSON, result)
		require.NoError(t, err)
		require.NotNil(t, metadata)
		require.Equal(t, uint64(245), metadata.BatchNonceID)
	})

	t.Run("Cbor Marshal BatchExecutedMetadata", func(t *testing.T) {
		result, err := MarshalMetadata[BatchExecutedMetadata](MetadataEncodingTypeCbor, BatchExecutedMetadata{BridgingTxType: "test"})

		require.NoError(t, err)
		require.NotNil(t, result)
	})

	t.Run("Cbor Unmarshal BatchExecutedMetadata", func(t *testing.T) {
		result, err := SimulateRealMetadata(MetadataEncodingTypeCbor, BatchExecutedMetadata{BridgingTxType: "test"})
		require.NoError(t, err)
		require.NotNil(t, result)

		metadata, err := UnmarshalMetadata[BatchExecutedMetadata](MetadataEncodingTypeCbor, result)
		require.NoError(t, err)
		require.NotNil(t, metadata)
		require.Equal(t, BridgingTxType("test"), metadata.BridgingTxType)
	})

	t.Run("Cbor Unmarshal V1 obsolete bridging request", func(t *testing.T) {
		feeAmount := uint64(1)
		result, err := SimulateRealMetadata(MetadataEncodingTypeCbor, BridgingRequestMetadataV1{
			BridgingTxType: "test",
			BridgingFee:    feeAmount,
		})
		require.NoError(t, err)
		require.NotNil(t, result)

		metadata, err := UnmarshalMetadata[BridgingRequestMetadata](MetadataEncodingTypeCbor, result)
		require.NoError(t, err)
		require.NotNil(t, metadata)
		require.Equal(t, feeAmount, metadata.BridgingFee)
	})

	t.Run("Json Unmarshal V1 obsolete bridging request", func(t *testing.T) {
		feeAmount := uint64(1)
		result, err := SimulateRealMetadata(MetadataEncodingTypeJSON, BridgingRequestMetadataV1{
			BridgingTxType: "test",
			BridgingFee:    feeAmount,
		})
		require.NoError(t, err)
		require.NotNil(t, result)

		metadata, err := UnmarshalMetadata[BridgingRequestMetadata](MetadataEncodingTypeJSON, result)
		require.NoError(t, err)
		require.NotNil(t, metadata)
		require.Equal(t, feeAmount, metadata.BridgingFee)
	})

	t.Run("Cbor Unmarshal Well structured, but invalid", func(t *testing.T) {
		metadataRaw := map[int]interface{}{
			0: map[int]interface{}{
				0: map[string]interface{}{
					"user_id": "2", "source": "asfsadad",
				},
			},
		}

		result, err := cbor.Marshal(metadataRaw)
		require.NoError(t, err)

		metadata, err := UnmarshalMetadata[BaseMetadata](MetadataEncodingTypeCbor, result)
		require.Error(t, err)
		require.ErrorContains(t, err, "invalid metadata")
		require.Nil(t, metadata)
	})
}
