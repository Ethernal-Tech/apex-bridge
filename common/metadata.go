package common

import (
	"encoding/json"
	"fmt"

	"github.com/fxamacker/cbor/v2"
)

type BridgingTxType string
type MetadataEncodingType string

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

type BridgingRequestMetadataTransaction struct {
	Address []string `cbor:"a" json:"a"`
	Amount  uint64   `cbor:"m" json:"m"`
}

type BridgingRequestMetadata struct {
	BridgingTxType     BridgingTxType                       `cbor:"t" json:"t"`
	DestinationChainID string                               `cbor:"d" json:"d"`
	SenderAddr         []string                             `cbor:"s" json:"s"`
	Transactions       []BridgingRequestMetadataTransaction `cbor:"tx" json:"tx"`
	BridgingFee        uint64                               `cbor:"fa" json:"fa"`
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

func MarshalMetadata[
	T BaseMetadata | BridgingRequestMetadata | RefundBridgingRequestMetadata | BatchExecutedMetadata,
](
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

func UnmarshalMetadata[
	T BaseMetadata | BridgingRequestMetadata | RefundBridgingRequestMetadata | BatchExecutedMetadata,
](
	encodingType MetadataEncodingType, data []byte,
) (
	*T, error,
) {
	unmarshalFunc, err := getUnmarshalFunc(encodingType)
	if err != nil {
		return nil, err
	}

	var metadataMap map[int]map[int]*T

	err = unmarshalFunc(data, &metadataMap)
	if err != nil {
		var metadata interface{}

		errInner := unmarshalFunc(data, &metadata)
		if errInner != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata, err: %w", err)
		} else {
			return nil, fmt.Errorf("failed to unmarshal metadata: %v, err: %w", metadata, err)
		}
	}

	for _, mapVal := range metadataMap {
		if metadata, exists := mapVal[MetadataMapKey]; exists {
			return metadata, nil
		}
	}

	return nil, fmt.Errorf("invalid metadata")
}

func MarshalMetadataMap[
	T BaseMetadata | BridgingRequestMetadata | RefundBridgingRequestMetadata | BatchExecutedMetadata,
](
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
