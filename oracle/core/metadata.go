package core

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"
)

type BridgingTxType string

const (
	BridgingTxTypeBridgingRequest BridgingTxType = "bridgingRequest"
	BridgingTxTypeBatchExecution  BridgingTxType = "batchExecution"
	BridgingTxTypeRefundExecution BridgingTxType = "refundExecution"
)

type BaseMetadata struct {
	BridgingTxType BridgingTxType `cbor:"type"`
}

type BaseMetadataMap struct {
	Value BaseMetadata `cbor:"1,keyasint"`
}

func UnmarshalBaseMetadata(data []byte) (*BaseMetadata, error) {
	var metadataMap BaseMetadataMap
	err := cbor.Unmarshal(data, &metadataMap)

	if err != nil {
		var metadata interface{}
		cbor.Unmarshal(data, &metadata)
		return nil, fmt.Errorf("failed to unmarshal metadata: %v", metadata)
	} else {
		return &metadataMap.Value, nil
	}
}

type BridgingRequestMetadataTransaction struct {
	Address string `cbor:"address"`
	Amount  uint64 `cbor:"amount"`
}

type BridgingRequestMetadata struct {
	BridgingTxType     BridgingTxType                       `cbor:"type"`
	DestinationChainId string                               `cbor:"destinationChainId"`
	SenderAddr         string                               `cbor:"senderAddr"`
	Transactions       []BridgingRequestMetadataTransaction `cbor:"transactions"`
}

type BridgingRequestMetadataMap struct {
	Value BridgingRequestMetadata `cbor:"1,keyasint"`
}

func UnmarshalBridgingRequestMetadata(data []byte) (*BridgingRequestMetadata, error) {
	var metadataMap BridgingRequestMetadataMap
	err := cbor.Unmarshal(data, &metadataMap)

	if err != nil {
		var metadata interface{}
		cbor.Unmarshal(data, &metadata)
		return nil, fmt.Errorf("failed to unmarshal metadata: %v", metadata)
	} else {
		return &metadataMap.Value, nil
	}
}

type BatchExecutedMetadata struct {
	BridgingTxType BridgingTxType `cbor:"type"`
	BatchNonceId   string         `cbor:"batchNonceId"`
}

type BatchExecutedMetadataMap struct {
	Value BatchExecutedMetadata `cbor:"1,keyasint"`
}

func UnmarshalBatchExecutedMetadata(data []byte) (*BatchExecutedMetadata, error) {
	var metadataMap BatchExecutedMetadataMap
	err := cbor.Unmarshal(data, &metadataMap)

	if err != nil {
		var metadata interface{}
		cbor.Unmarshal(data, &metadata)
		return nil, fmt.Errorf("failed to unmarshal metadata: %v", metadata)
	} else {
		return &metadataMap.Value, nil
	}
}

type RefundExecutedMetadata struct {
	BridgingTxType BridgingTxType `cbor:"type"`
}

type RefundExecutedMetadataMap struct {
	Value RefundExecutedMetadata `cbor:"1,keyasint"`
}

func UnmarshalRefundExecutedMetadata(data []byte) (*RefundExecutedMetadata, error) {
	var metadataMap RefundExecutedMetadataMap
	err := cbor.Unmarshal(data, &metadataMap)

	if err != nil {
		var metadata interface{}
		cbor.Unmarshal(data, &metadata)
		return nil, fmt.Errorf("failed to unmarshal metadata: %v", metadata)
	} else {
		return &metadataMap.Value, nil
	}
}
