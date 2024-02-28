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

type BridgingRequestMetadata struct {
	BridgingTxType BridgingTxType `cbor:"type"`
	ChainId        string         `cbor:"chainId"`
	SenderAddr     string         `cbor:"senderAddr"`
	Transactions   []struct {
		Address string `cbor:"address"`
		Amount  uint64 `cbor:"amount"`
	}
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
	OriginChainId  string         `cbor:"chainId"`
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
