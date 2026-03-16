package core

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/common"
)

type BridgingRequestSolMetadataTransaction struct {
	Address string   `json:"a"`
	Amount  *big.Int `json:"m"`
	TokenID uint16   `json:"t"`
}

type BaseSolMetadata struct {
	BridgingTxType common.BridgingTxType `json:"t"`
}

type BridgingRequestSolMetadata struct {
	BridgingTxType     common.BridgingTxType                   `json:"t"`
	DestinationChainID string                                  `json:"d"`
	SenderAddr         string                                  `json:"s"`
	Transactions       []BridgingRequestSolMetadataTransaction `json:"tx"`
	BridgingFee        uint64                                  `json:"fa"`
	OperationFee       uint64                                  `json:"of"`
}

type RefundBridgingRequestSolMetadata struct {
	BridgingTxType     common.BridgingTxType                   `json:"t"`
	SenderAddr         string                                  `json:"s"`
	DestinationChainID string                                  `json:"d"`
	Transactions       []BridgingRequestSolMetadataTransaction `json:"tx"`
}

type BatchExecutedSolMetadata struct {
	BridgingTxType common.BridgingTxType `json:"t"`
	BatchNonceID   uint64                `json:"n"`
}

func MarshalSolMetadata[
	T BaseSolMetadata | BridgingRequestSolMetadata | RefundBridgingRequestSolMetadata | BatchExecutedSolMetadata,
](
	metadata T,
) (
	[]byte, error,
) {
	result, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %v, err: %w", metadata, err)
	}

	return result, nil
}

func UnmarshalSolMetadata[
	T BaseSolMetadata | BridgingRequestSolMetadata | RefundBridgingRequestSolMetadata | BatchExecutedSolMetadata,
](
	data []byte,
) (
	*T, error,
) {
	var metadata *T

	err := json.Unmarshal(data, &metadata)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}
