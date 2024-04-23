package common

import (
	"encoding/json"
	"fmt"

	"github.com/fxamacker/cbor/v2"
)

type BridgingTxType string
type MetadataEncodingType string

const (
	BridgingTxTypeBridgingRequest BridgingTxType = "bridgingRequest"
	BridgingTxTypeBatchExecution  BridgingTxType = "batchExecution"
	BridgingTxTypeRefundExecution BridgingTxType = "refundExecution"

	MetadataEncodingTypeJson MetadataEncodingType = "json"
	MetadataEncodingTypeCbor MetadataEncodingType = "cbor"

	MetadataMapKey = 1
)

type BaseMetadata struct {
	BridgingTxType BridgingTxType `cbor:"type" json:"type"`
}

type BridgingRequestMetadataTransaction struct {
	Address string `cbor:"address" json:"address"`
	Amount  uint64 `cbor:"amount" json:"amount"`
}

type BridgingRequestMetadata struct {
	BridgingTxType     BridgingTxType                       `cbor:"type"`
	DestinationChainId string                               `cbor:"destinationChainId"`
	SenderAddr         string                               `cbor:"senderAddr"`
	Transactions       []BridgingRequestMetadataTransaction `cbor:"transactions"`
}

type BatchExecutedMetadata struct {
	BridgingTxType BridgingTxType `cbor:"type" json:"type"`
	BatchNonceId   uint64         `cbor:"batchNonceId" json:"batchNonceId"`
}

type RefundExecutedMetadata struct {
	BridgingTxType BridgingTxType `cbor:"type" json:"type"`
}

type marshalFunc = func(v any) ([]byte, error)

func getMarshalFunc(encodingType MetadataEncodingType) (marshalFunc, error) {
	if encodingType == MetadataEncodingTypeJson {
		return json.Marshal, nil
	} else if encodingType == MetadataEncodingTypeCbor {
		return cbor.Marshal, nil
	}

	return nil, fmt.Errorf("unsupported metadata encoding type")
}

type unmarshalFunc = func(data []byte, v interface{}) error

func getUnmarshalFunc(encodingType MetadataEncodingType) (unmarshalFunc, error) {
	if encodingType == MetadataEncodingTypeJson {
		return json.Unmarshal, nil
	} else if encodingType == MetadataEncodingTypeCbor {
		return cbor.Unmarshal, nil
	}

	return nil, fmt.Errorf("unsupported metadata encoding type")
}

func MarshalMetadata[
	T BaseMetadata | BridgingRequestMetadata | BatchExecutedMetadata | RefundExecutedMetadata,
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
	T BaseMetadata | BridgingRequestMetadata | BatchExecutedMetadata | RefundExecutedMetadata,
](
	encodingType MetadataEncodingType, data []byte,
) (
	*T, error,
) {
	unmarshalFunc, err := getUnmarshalFunc(encodingType)
	if err != nil {
		return nil, err
	}

	var metadataMap map[int]T

	err = unmarshalFunc(data, &metadataMap)
	if err != nil {
		var metadata interface{}
		unmarshalFunc(data, &metadata)
		return nil, fmt.Errorf("failed to unmarshal metadata: %v, err: %w", metadata, err)
	}

	metadata := metadataMap[MetadataMapKey]
	return &metadata, nil
}
