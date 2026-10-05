package common

import (
	"testing"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/require"
)

func TestResourceLimits(t *testing.T) {
	// deeply nested cbor: 10k nested arrays
	deep := make([]byte, 0, 10000)
	for i := 0; i < 10000; i++ {
		deep = append(deep, 0x81) // array(1)
	}

	deep = append(deep, 0x00)

	start := time.Now().UTC()
	_, err := UnmarshalMetadata[BridgingRequestMetadata](MetadataEncodingTypeCbor, deep)
	t.Logf("10k-deep nesting: err=%v elapsed=%s", err != nil, time.Since(start))
	require.Error(t, err)

	// huge declared array length with no payload (cbor length bomb)
	bomb := []byte{0x9b, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	start = time.Now().UTC()
	_, err = UnmarshalMetadata[BridgingRequestMetadata](MetadataEncodingTypeCbor, bomb)
	t.Logf("length bomb: err=%v elapsed=%s", err != nil, time.Since(start))
	require.Error(t, err)

	// huge declared map length
	mapBomb := []byte{0xbb, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	start = time.Now().UTC()
	_, err = UnmarshalMetadata[BridgingRequestMetadata](MetadataEncodingTypeCbor, mapBomb)
	t.Logf("map length bomb: err=%v elapsed=%s", err != nil, time.Since(start))
	require.Error(t, err)
}

func FuzzUnmarshalMetadata(f *testing.F) {
	seed, _ := cbor.Marshal(cbor.Tag{Number: 259, Content: map[uint64]interface{}{
		0: map[uint64]BaseMetadata{1: {BridgingTxType: "bridge"}},
	}})
	f.Add(seed)
	f.Add([]byte{0xa1, 0x01, 0xa1, 0x01, 0x00})
	f.Add([]byte{0x82, 0xa0, 0x80})

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = UnmarshalMetadata[BaseMetadata](MetadataEncodingTypeCbor, data)
		_, _ = UnmarshalMetadata[BridgingRequestMetadata](MetadataEncodingTypeCbor, data)
		_, _ = UnmarshalMetadata[BatchExecutedMetadata](MetadataEncodingTypeCbor, data)
	})
}
