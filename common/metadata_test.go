package common

import (
	"math"
	"testing"

	"github.com/Ethernal-Tech/cardano-infrastructure/sendtx"
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

	t.Run("Cbor Unmarshal BC bridging request", func(t *testing.T) {
		feeAmount := uint64(1)
		result, err := SimulateRealMetadata(MetadataEncodingTypeCbor, BridgingRequestMetadataBC{
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

	t.Run("Json Unmarshal BC bridging request", func(t *testing.T) {
		feeAmount := uint64(1)
		result, err := SimulateRealMetadata(MetadataEncodingTypeJSON, BridgingRequestMetadataBC{
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

func TestUnmarshalMetadataAuxiliaryDataEncodings(t *testing.T) {
	metadata := BridgingRequestMetadata{
		BridgingTxType:     "bridge",
		DestinationChainID: "cardano",
		SenderAddr:         []string{"addr1_chunk_one", "addr1_chunk_two"},
		Transactions: []sendtx.BridgingRequestMetadataTransaction{
			{Address: []string{"addr1_dst"}, Amount: 312974944, TokenID: 1},
		},
		BridgingFee:  4000000,
		OperationFee: 0,
	}

	// metadata map: { label => metadatum }, the payload every envelope wraps
	metadataMap, err := cbor.Marshal(map[uint64]BridgingRequestMetadata{MetadataMapKey: metadata})
	require.NoError(t, err)

	raw := cbor.RawMessage(metadataMap)
	scripts := []interface{}{}

	envelopes := map[string]interface{}{
		"shelley":                     raw,
		"shelley-ma":                  []interface{}{raw, scripts},
		"alonzo tagged":               cbor.Tag{Number: 259, Content: map[uint64]interface{}{0: raw}},
		"alonzo tagged, with scripts": cbor.Tag{Number: 259, Content: map[uint64]interface{}{0: raw, 1: scripts}},
	}

	for name, envelope := range envelopes {
		t.Run(name, func(t *testing.T) {
			auxData, err := cbor.Marshal(envelope)
			require.NoError(t, err)

			decoded, err := UnmarshalMetadata[BridgingRequestMetadata](MetadataEncodingTypeCbor, auxData)
			require.NoError(t, err)
			require.Equal(t, metadata, *decoded)

			// the tx type sniff the processor selection does must work over the same bytes
			base, err := UnmarshalMetadata[BaseMetadata](MetadataEncodingTypeCbor, auxData)
			require.NoError(t, err)
			require.Equal(t, BridgingTxTypeBridgingRequest, base.BridgingTxType)
		})
	}

	t.Run("an unrelated label alongside ours is ignored", func(t *testing.T) {
		// e.g. a CIP-20 message: it must not make the bridging tx undecodable
		withMemo, err := cbor.Marshal(cbor.Tag{Number: 259, Content: map[uint64]interface{}{
			0: map[uint64]interface{}{
				MetadataMapKey: metadata,
				674:            map[string]interface{}{"msg": []string{"hello"}},
			},
		}})
		require.NoError(t, err)

		decoded, err := UnmarshalMetadata[BridgingRequestMetadata](MetadataEncodingTypeCbor, withMemo)
		require.NoError(t, err)
		require.Equal(t, metadata, *decoded)
	})

	t.Run("untagged {0: metadata} is not a legal envelope", func(t *testing.T) {
		// the alonzo form is always tagged, so this is a shelley metadata map whose
		// label 0 happens to hold a map - reading it as auxiliary_data would let one
		// tx mean different things to different observers
		untagged, err := cbor.Marshal(map[uint64]interface{}{0: raw})
		require.NoError(t, err)

		decoded, err := UnmarshalMetadata[BridgingRequestMetadata](MetadataEncodingTypeCbor, untagged)
		require.Error(t, err)
		require.Nil(t, decoded)
	})

	t.Run("metadata under another label is still rejected", func(t *testing.T) {
		other, err := cbor.Marshal(map[uint64]interface{}{
			0: map[uint64]BridgingRequestMetadata{674: metadata},
		})
		require.NoError(t, err)

		decoded, err := UnmarshalMetadata[BridgingRequestMetadata](MetadataEncodingTypeCbor, other)
		require.Error(t, err)
		require.Nil(t, decoded)
	})

	t.Run("garbage is still rejected", func(t *testing.T) {
		decoded, err := UnmarshalMetadata[BridgingRequestMetadata](
			MetadataEncodingTypeCbor, []byte{0x01, 0x02, 0x03})
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to unmarshal metadata")
		require.Nil(t, decoded)
	})

	t.Run("empty auxiliary_data array is still rejected", func(t *testing.T) {
		empty, err := cbor.Marshal([]interface{}{})
		require.NoError(t, err)

		decoded, err := UnmarshalMetadata[BridgingRequestMetadata](MetadataEncodingTypeCbor, empty)
		require.Error(t, err)
		require.Nil(t, decoded)
	})
}

// TestUnmarshalMetadataAdversarialAuxiliaryData covers auxiliary_data a sender can
// craft to make one transaction read differently here than it does to other observers
// of the chain, or differently across validators.
func TestUnmarshalMetadataAdversarialAuxiliaryData(t *testing.T) {
	mk := func(dest string, amount uint64) BridgingRequestMetadata {
		return BridgingRequestMetadata{
			BridgingTxType:     "bridge",
			DestinationChainID: dest,
			SenderAddr:         []string{"sender"},
			Transactions: []sendtx.BridgingRequestMetadataTransaction{
				{Address: []string{dest + "_addr"}, Amount: amount, TokenID: 1},
			},
			BridgingFee: 4000000,
		}
	}

	declared, decoy := mk("nexus", 1000), mk("cardano", 999999999)

	t.Run("no label 1 means rejected, however many other labels nest one", func(t *testing.T) {
		// both labels nest a {1: M} that looks like a metadata map. Reading either as
		// auxiliary_data would pick one at random, so neither may be read at all.
		data, err := cbor.Marshal(map[uint64]interface{}{
			0: map[uint64]BridgingRequestMetadata{MetadataMapKey: decoy},
			7: map[uint64]BridgingRequestMetadata{MetadataMapKey: declared},
		})
		require.NoError(t, err)

		for i := 0; i < 50; i++ {
			metadata, err := UnmarshalMetadata[BridgingRequestMetadata](MetadataEncodingTypeCbor, data)
			require.Error(t, err)
			require.Nil(t, metadata)
		}
	})

	t.Run("label 0 cannot shadow the real label 1", func(t *testing.T) {
		data, err := cbor.Marshal(map[uint64]interface{}{
			0:              map[uint64]BridgingRequestMetadata{MetadataMapKey: decoy},
			MetadataMapKey: declared,
		})
		require.NoError(t, err)

		for i := 0; i < 50; i++ {
			metadata, err := UnmarshalMetadata[BridgingRequestMetadata](MetadataEncodingTypeCbor, data)
			require.NoError(t, err)
			require.Equal(t, declared, *metadata, "label 0 was read instead of label 1")
		}
	})

	t.Run("duplicate label 1 is rejected rather than last-wins", func(t *testing.T) {
		first, err := cbor.Marshal(declared)
		require.NoError(t, err)

		second, err := cbor.Marshal(decoy)
		require.NoError(t, err)

		// map(2) { 1: declared, 1: decoy } - go maps cannot express this
		data := append([]byte{0xa2, 0x01}, first...)
		data = append(append(data, 0x01), second...)

		metadata, err := UnmarshalMetadata[BridgingRequestMetadata](MetadataEncodingTypeCbor, data)
		require.Error(t, err)
		require.ErrorContains(t, err, "duplicate map key")
		require.Nil(t, metadata)
	})

	t.Run("processor selection and claim building agree on one payload", func(t *testing.T) {
		// the oracle decodes the same bytes twice, as BaseMetadata to pick a processor
		// and again as the payload type. Both must resolve the same metadatum.
		data, err := cbor.Marshal(cbor.Tag{Number: alonzoAuxiliaryDataTag,
			Content: map[uint64]interface{}{
				0: map[uint64]BridgingRequestMetadata{MetadataMapKey: declared},
			}})
		require.NoError(t, err)

		base, err := UnmarshalMetadata[BaseMetadata](MetadataEncodingTypeCbor, data)
		require.NoError(t, err)

		full, err := UnmarshalMetadata[BridgingRequestMetadata](MetadataEncodingTypeCbor, data)
		require.NoError(t, err)

		require.Equal(t, BridgingTxTypeBridgingRequest, base.BridgingTxType)
		require.EqualValues(t, base.BridgingTxType, full.BridgingTxType)
	})

	t.Run("a deeply nested native script does not hide the metadata", func(t *testing.T) {
		// native_script is recursive and cardano-ledger applies no depth counter, so a
		// script the node accepts can nest far past a decoder default of 32. The nesting
		// counter is per message, so an over-deep script next to the metadata would make
		// the whole envelope undecodable and drop an otherwise valid bridging request.
		script := mustMarshal(t, []interface{}{0, []byte("keyhash")})
		for range 500 {
			script = mustMarshal(t, []interface{}{1, []cbor.RawMessage{script}})
		}

		metadataMap := mustMarshal(t, map[uint64]BridgingRequestMetadata{MetadataMapKey: declared})
		scripts := mustMarshal(t, []cbor.RawMessage{script})

		for name, data := range map[string]cbor.RawMessage{
			"shelley-ma": mustMarshal(t, []cbor.RawMessage{metadataMap, scripts}),
			"alonzo": mustMarshal(t, cbor.Tag{Number: alonzoAuxiliaryDataTag,
				Content: map[uint64]cbor.RawMessage{0: metadataMap, 1: scripts}}),
		} {
			t.Run(name, func(t *testing.T) {
				metadata, err := UnmarshalMetadata[BridgingRequestMetadata](MetadataEncodingTypeCbor, data)
				require.NoError(t, err)
				require.Equal(t, declared, *metadata)
			})
		}
	})

	t.Run("a metadata label above MaxInt64 does not hide the metadata", func(t *testing.T) {
		// transaction_metadatum_label is a cbor uint, so the full uint64 range is on
		// chain-valid. Decoding labels into a signed type would overflow on a label
		// anyone can attach and take the bridging request down with it.
		data := mustMarshal(t, map[uint64]cbor.RawMessage{
			MetadataMapKey:    mustMarshal(t, declared),
			math.MaxUint64:    mustMarshal(t, "unrelated"),
			math.MaxInt64 + 1: mustMarshal(t, "unrelated"),
		})

		metadata, err := UnmarshalMetadata[BridgingRequestMetadata](MetadataEncodingTypeCbor, data)
		require.NoError(t, err)
		require.Equal(t, declared, *metadata)
	})

	t.Run("unknown auxiliary_data tag is rejected", func(t *testing.T) {
		data, err := cbor.Marshal(cbor.Tag{Number: 42, Content: map[uint64]interface{}{
			0: map[uint64]BridgingRequestMetadata{MetadataMapKey: decoy},
		}})
		require.NoError(t, err)

		metadata, err := UnmarshalMetadata[BridgingRequestMetadata](MetadataEncodingTypeCbor, data)
		require.Error(t, err)
		require.Nil(t, metadata)
	})

	t.Run("NUL in a string is kept rather than ending it", func(t *testing.T) {
		// NUL is valid UTF-8, so the ledger accepts it in metadata text. Ending a string
		// at it would read "nexus\x00..." as the configured "nexus".
		withNUL := BridgingRequestMetadata{
			BridgingTxType:     "bridge\x00",
			DestinationChainID: "nexus\x00cardano",
			SenderAddr:         []string{"\x00", "sender\x00"},
			Transactions: []sendtx.BridgingRequestMetadataTransaction{
				{Address: []string{"\x00nexus_addr", "\x00"}, Amount: 1000, TokenID: 1},
			},
			BridgingFee: 4000000,
		}

		data := mustMarshal(t, cbor.Tag{Number: alonzoAuxiliaryDataTag,
			Content: map[uint64]interface{}{
				0: map[uint64]BridgingRequestMetadata{MetadataMapKey: withNUL},
			}})

		metadata, err := UnmarshalMetadata[BridgingRequestMetadata](MetadataEncodingTypeCbor, data)
		require.NoError(t, err)
		require.Equal(t, withNUL, *metadata)

		base, err := UnmarshalMetadata[BaseMetadata](MetadataEncodingTypeCbor, data)
		require.NoError(t, err)
		require.Equal(t, withNUL.BridgingTxType, sendtx.BridgingRequestType(base.BridgingTxType))
	})

	t.Run("a key that differs only by NUL is another key", func(t *testing.T) {
		// a reader that ends strings at NUL sees "d\x00" as a second "d"
		data := mustMarshal(t, map[uint64]interface{}{
			MetadataMapKey: map[string]interface{}{"t": "bridge", "d": "nexus", "d\x00": "cardano"},
		})

		metadata, err := UnmarshalMetadata[BridgingRequestMetadata](MetadataEncodingTypeCbor, data)
		require.NoError(t, err)
		require.Equal(t, "nexus", metadata.DestinationChainID)

		data = mustMarshal(t, map[uint64]interface{}{
			MetadataMapKey: map[string]interface{}{"t\x00": "bridge", "d\x00": "nexus"},
		})

		metadata, err = UnmarshalMetadata[BridgingRequestMetadata](MetadataEncodingTypeCbor, data)
		require.NoError(t, err)
		require.Empty(t, metadata.BridgingTxType)
		require.Empty(t, metadata.DestinationChainID)
	})
}

func mustMarshal(t *testing.T, v interface{}) cbor.RawMessage {
	t.Helper()

	data, err := cbor.Marshal(v)
	require.NoError(t, err)

	return data
}
