package core

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/common"
)

type BridgingRequestEthMetadataTransaction struct {
	Address string   `json:"a"`
	Amount  *big.Int `json:"m"`
}

type BaseEthMetadata struct {
	BridgingTxType common.BridgingTxType `json:"t"`
}

type BridgingRequestEthMetadata struct {
	BridgingTxType     common.BridgingTxType                   `json:"t"`
	DestinationChainID string                                  `json:"d"`
	SenderAddr         string                                  `json:"s"`
	Transactions       []BridgingRequestEthMetadataTransaction `json:"tx"`
	BridgingFee        *big.Int                                `json:"fa"`
}

type BatchExecutedEthMetadata struct {
	BridgingTxType common.BridgingTxType `json:"t"`
	BatchNonceID   uint64                `json:"n"`
}

func MarshalEthMetadata[
	T BaseEthMetadata | BridgingRequestEthMetadata | BatchExecutedEthMetadata,
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

func UnmarshalEthMetadata[
	T BaseEthMetadata | BridgingRequestEthMetadata | BatchExecutedEthMetadata,
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
