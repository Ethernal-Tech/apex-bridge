package common

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Ethernal-Tech/cardano-infrastructure/sendtx"
	"github.com/fxamacker/cbor/v2"
)

type BridgingTxType string
type MetadataEncodingType string
type BridgingRequestMetadata sendtx.BridgingRequestMetadata

const (
	BridgingTxTypeBridgingRequest BridgingTxType = "bridge"
	BridgingTxTypeBatchExecution  BridgingTxType = "batch"

	TxTypeRefundRequest BridgingTxType = "refund"
	TxTypeHotWalletFund BridgingTxType = "fund"

	MetadataEncodingTypeJSON MetadataEncodingType = "json"
	MetadataEncodingTypeCbor MetadataEncodingType = "cbor"

	MetadataMapKey = 1
)

type BaseMetadata struct {
	BridgingTxType BridgingTxType `cbor:"t" json:"t"`
}

type RefundBridgingRequestMetadata struct {
	BridgingTxType BridgingTxType `cbor:"t" json:"t"`
	SenderAddr     []string       `cbor:"s" json:"s"`
}

type BatchExecutedMetadata struct {
	BridgingTxType BridgingTxType `cbor:"t" json:"t"`
	BatchNonceID   uint64         `cbor:"n" json:"n"`
	IsFeeOnlyTx    uint8          `cbor:"f" json:"f"`
}

// metadataTypes is the set of metadata payloads this package marshals and unmarshals.
type metadataTypes interface {
	BaseMetadata | BridgingRequestMetadata |
		RefundBridgingRequestMetadata | BatchExecutedMetadata
}

// alonzoAuxiliaryDataTag is the cbor tag wrapping auxiliary_data from alonzo onwards.
const alonzoAuxiliaryDataTag = 259

// Reject duplicate map keys. Transaction metadata is user
// controlled, and a decoder that quietly keeps the last of a duplicated key decides
// something a stricter decoder elsewhere would reject or read the other way round.
var getAuxiliaryDataDecMode = sync.OnceValues(func() (cbor.DecMode, error) {
	return cbor.DecOptions{
		DupMapKey: cbor.DupMapKeyEnforcedAPF,
		// auxiliary_data may carry native scripts alongside the metadata, and those
		// nest arbitrarily deep. cardano-ledger applies no depth counter, so the only
		// bound is max_tx_size; the default of 32 is far below it. The counter
		// is per message, so one deep script makes the whole envelope - metadata
		// included - undecodable, and the bridging request is dropped as invalid. This
		// is the same limit, for the same reason, as the gouroboros fork go.mod points
		// at. 65535 is the maximum fxamacker allows and stays finite because decoding
		// recurses on the Go stack.
		MaxNestedLevels: 65535,
	}.DecMode()
})

type marshalFunc = func(v any) ([]byte, error)

func getMarshalFunc(encodingType MetadataEncodingType) (marshalFunc, error) {
	if encodingType == MetadataEncodingTypeJSON {
		return json.Marshal, nil
	} else if encodingType == MetadataEncodingTypeCbor {
		return cbor.Marshal, nil
	}

	return nil, fmt.Errorf("unsupported metadata encoding type")
}

type unmarshalFunc = func(data []byte, v interface{}) error

func getUnmarshalFunc(encodingType MetadataEncodingType) (unmarshalFunc, error) {
	if encodingType == MetadataEncodingTypeJSON {
		return json.Unmarshal, nil
	} else if encodingType == MetadataEncodingTypeCbor {
		return cbor.Unmarshal, nil
	}

	return nil, fmt.Errorf("unsupported metadata encoding type")
}

func MarshalMetadata[T metadataTypes](
	encodingType MetadataEncodingType, metadata T,
) (
	[]byte, error,
) {
	marshalFunc, err := getMarshalFunc(encodingType)
	if err != nil {
		return nil, err
	}

	result, err := marshalFunc(map[int]T{
		MetadataMapKey: metadata,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %v, err: %w", metadata, err)
	}

	return result, nil
}

func UnmarshalMetadata[T metadataTypes](
	encodingType MetadataEncodingType, data []byte,
) (
	*T, error,
) {
	unmarshalFunc, err := getUnmarshalFunc(encodingType)
	if err != nil {
		return nil, err
	}

	if encodingType == MetadataEncodingTypeCbor {
		return unmarshalAuxiliaryData[T](data)
	}

	var metadataMap map[int]map[int]*T

	err = unmarshalFunc(data, &metadataMap)
	if err != nil {
		var metadata interface{}

		errInner := unmarshalFunc(data, &metadata)
		if errInner != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata, err: %w", err)
		}

		return nil, fmt.Errorf("failed to unmarshal metadata: %v, err: %w", metadata, err)
	}

	for _, mapVal := range metadataMap {
		if metadata, exists := mapVal[MetadataMapKey]; exists {
			return metadata, nil
		}
	}

	return nil, fmt.Errorf("invalid metadata")
}

// unmarshalAuxiliaryData resolves a cardano transaction's auxiliary_data envelope and
// decodes the metadatum stored under MetadataMapKey.
//
// Babbage accepts exactly three encodings, and the transaction builder picks - none of
// them means the transaction is malformed:
//
//	metadata                                                        ; shelley
//	[ metadata, [* native_script] ]                                 ; shelley-ma
//	#6.259({ ?0: metadata, ?1: [* native_script], ?2: .., ?3: .. }) ; alonzo and later
//
// where metadata is the { label => metadatum } map.
//
// Resolution has to be exact rather than best-effort. The alonzo form is always
// tagged, so an untagged map is the metadata map itself and its key 0 is a label, not
// an auxiliary_data field. Treating the two interchangeably lets a transaction read
// one way here and another way to every other observer of the chain, and lets a
// single transaction carry two payloads that different validators may disagree on.
func unmarshalAuxiliaryData[T metadataTypes](data []byte) (*T, error) {
	decMode, err := getAuxiliaryDataDecMode()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize CBOR decoder mode: %w", err)
	}

	metadataMap := cbor.RawMessage(data)

	var tagged cbor.RawTag

	var scripted []cbor.RawMessage

	switch {
	case decMode.Unmarshal(data, &tagged) == nil:
		if tagged.Number != alonzoAuxiliaryDataTag {
			return nil, fmt.Errorf("unexpected auxiliary_data tag: %d", tagged.Number)
		}

		var fields map[uint64]cbor.RawMessage
		if err := decMode.Unmarshal(tagged.Content, &fields); err != nil {
			return nil, fmt.Errorf("failed to unmarshal auxiliary_data, err: %w", err)
		}

		var exists bool
		if metadataMap, exists = fields[0]; !exists {
			return nil, fmt.Errorf("invalid metadata")
		}
	case decMode.Unmarshal(data, &scripted) == nil:
		if len(scripted) == 0 {
			return nil, fmt.Errorf("invalid metadata")
		}

		metadataMap = scripted[0]
	}

	// only MetadataMapKey is decoded into T; other labels are left alone, so an
	// unrelated label such as a CIP-20 message cannot make a bridging tx undecodable
	var labels map[uint64]cbor.RawMessage
	if err := decMode.Unmarshal(metadataMap, &labels); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata, err: %w", err)
	}

	raw, exists := labels[MetadataMapKey]
	if !exists {
		return nil, fmt.Errorf("invalid metadata")
	}

	var metadata T
	if err := decMode.Unmarshal(raw, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %v, err: %w", raw, err)
	}

	return &metadata, nil
}

func MarshalMetadataMap[T metadataTypes](
	encodingType MetadataEncodingType, metadata T,
) (
	[]byte, error,
) {
	marshalFunc, err := getMarshalFunc(encodingType)
	if err != nil {
		return nil, err
	}

	result, err := marshalFunc(map[int]map[int]T{1: {MetadataMapKey: metadata}})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %v, err: %w", metadata, err)
	}

	return result, nil
}
