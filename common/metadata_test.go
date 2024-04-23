package common

import (
	"testing"

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
		result, err := MarshalMetadata[BaseMetadata](MetadataEncodingTypeJson, BaseMetadata{BridgingTxType: "test"})
		require.NoError(t, err)
		require.NotNil(t, result)

		metadata, err := UnmarshalMetadata[BaseMetadata]("invalid", result)
		require.Error(t, err)
		require.ErrorContains(t, err, "unsupported metadata encoding type")
		require.Nil(t, metadata)
	})

	t.Run("Json Marshal BaseMetadata", func(t *testing.T) {
		result, err := MarshalMetadata[BaseMetadata](MetadataEncodingTypeJson, BaseMetadata{BridgingTxType: "test"})

		require.NoError(t, err)
		require.NotNil(t, result)
	})

	t.Run("Json Unmarshal BaseMetadata", func(t *testing.T) {
		result, err := MarshalMetadata[BaseMetadata](MetadataEncodingTypeJson, BaseMetadata{BridgingTxType: "test"})
		require.NoError(t, err)
		require.NotNil(t, result)

		metadata, err := UnmarshalMetadata[BaseMetadata](MetadataEncodingTypeJson, result)
		require.NoError(t, err)
		require.NotNil(t, metadata)
	})

	t.Run("Cbor Marshal BaseMetadata", func(t *testing.T) {
		result, err := MarshalMetadata[BaseMetadata](MetadataEncodingTypeCbor, BaseMetadata{BridgingTxType: "test"})

		require.NoError(t, err)
		require.NotNil(t, result)
	})

	t.Run("Cbor Unmarshal BaseMetadata", func(t *testing.T) {
		result, err := MarshalMetadata[BaseMetadata](MetadataEncodingTypeCbor, BaseMetadata{BridgingTxType: "test"})
		require.NoError(t, err)
		require.NotNil(t, result)

		metadata, err := UnmarshalMetadata[BaseMetadata](MetadataEncodingTypeCbor, result)
		require.NoError(t, err)
		require.NotNil(t, metadata)
	})

	t.Run("Json Marshal BridgingRequestMetadata", func(t *testing.T) {
		result, err := MarshalMetadata[BridgingRequestMetadata](MetadataEncodingTypeJson, BridgingRequestMetadata{BridgingTxType: "test"})

		require.NoError(t, err)
		require.NotNil(t, result)
	})

	t.Run("Json Unmarshal BridgingRequestMetadata", func(t *testing.T) {
		result, err := MarshalMetadata[BridgingRequestMetadata](MetadataEncodingTypeJson, BridgingRequestMetadata{BridgingTxType: "test"})
		require.NoError(t, err)
		require.NotNil(t, result)

		metadata, err := UnmarshalMetadata[BridgingRequestMetadata](MetadataEncodingTypeJson, result)
		require.NoError(t, err)
		require.NotNil(t, metadata)
	})

	t.Run("Cbor Marshal BridgingRequestMetadata", func(t *testing.T) {
		result, err := MarshalMetadata[BridgingRequestMetadata](MetadataEncodingTypeCbor, BridgingRequestMetadata{BridgingTxType: "test"})

		require.NoError(t, err)
		require.NotNil(t, result)
	})

	t.Run("Cbor Unmarshal BridgingRequestMetadata", func(t *testing.T) {
		result, err := MarshalMetadata[BridgingRequestMetadata](MetadataEncodingTypeCbor, BridgingRequestMetadata{BridgingTxType: "test"})
		require.NoError(t, err)
		require.NotNil(t, result)

		metadata, err := UnmarshalMetadata[BridgingRequestMetadata](MetadataEncodingTypeCbor, result)
		require.NoError(t, err)
		require.NotNil(t, metadata)
	})

	t.Run("Json Marshal BatchExecutedMetadata", func(t *testing.T) {
		result, err := MarshalMetadata[BatchExecutedMetadata](MetadataEncodingTypeJson, BatchExecutedMetadata{BridgingTxType: "test"})

		require.NoError(t, err)
		require.NotNil(t, result)
	})

	t.Run("Json Unmarshal BatchExecutedMetadata", func(t *testing.T) {
		result, err := MarshalMetadata[BatchExecutedMetadata](MetadataEncodingTypeJson, BatchExecutedMetadata{BatchNonceId: 245})
		require.NoError(t, err)
		require.NotNil(t, result)

		metadata, err := UnmarshalMetadata[BatchExecutedMetadata](MetadataEncodingTypeJson, result)
		require.NoError(t, err)
		require.NotNil(t, metadata)
		require.Equal(t, uint64(245), metadata.BatchNonceId)
	})

	t.Run("Cbor Marshal BatchExecutedMetadata", func(t *testing.T) {
		result, err := MarshalMetadata[BatchExecutedMetadata](MetadataEncodingTypeCbor, BatchExecutedMetadata{BridgingTxType: "test"})

		require.NoError(t, err)
		require.NotNil(t, result)
	})

	t.Run("Cbor Unmarshal BatchExecutedMetadata", func(t *testing.T) {
		result, err := MarshalMetadata[BatchExecutedMetadata](MetadataEncodingTypeCbor, BatchExecutedMetadata{BridgingTxType: "test"})
		require.NoError(t, err)
		require.NotNil(t, result)

		metadata, err := UnmarshalMetadata[BatchExecutedMetadata](MetadataEncodingTypeCbor, result)
		require.NoError(t, err)
		require.NotNil(t, metadata)
		require.Equal(t, BridgingTxType("test"), metadata.BridgingTxType)
	})

	t.Run("Json Marshal RefundExecutedMetadata", func(t *testing.T) {
		result, err := MarshalMetadata[RefundExecutedMetadata](MetadataEncodingTypeJson, RefundExecutedMetadata{BridgingTxType: "test"})

		require.NoError(t, err)
		require.NotNil(t, result)
	})

	t.Run("Json Unmarshal RefundExecutedMetadata", func(t *testing.T) {
		result, err := MarshalMetadata[RefundExecutedMetadata](MetadataEncodingTypeJson, RefundExecutedMetadata{BridgingTxType: "test"})
		require.NoError(t, err)
		require.NotNil(t, result)

		metadata, err := UnmarshalMetadata[RefundExecutedMetadata](MetadataEncodingTypeJson, result)
		require.NoError(t, err)
		require.NotNil(t, metadata)
	})

	t.Run("Cbor Marshal RefundExecutedMetadata", func(t *testing.T) {
		result, err := MarshalMetadata[RefundExecutedMetadata](MetadataEncodingTypeCbor, RefundExecutedMetadata{BridgingTxType: "test"})

		require.NoError(t, err)
		require.NotNil(t, result)
	})

	t.Run("Cbor Unmarshal RefundExecutedMetadata", func(t *testing.T) {
		result, err := MarshalMetadata[RefundExecutedMetadata](MetadataEncodingTypeCbor, RefundExecutedMetadata{BridgingTxType: "test"})
		require.NoError(t, err)
		require.NotNil(t, result)

		metadata, err := UnmarshalMetadata[RefundExecutedMetadata](MetadataEncodingTypeCbor, result)
		require.NoError(t, err)
		require.NotNil(t, metadata)
	})
}
