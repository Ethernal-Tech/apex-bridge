package response

import (
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
)

type BridgingRequestStateResponse struct {
	SourceChainID      string                     `json:"sourceChainId"`
	SourceTxHash       indexer.Hash               `json:"sourceTxHash"`
	DestinationChainID string                     `json:"destinationChainId"`
	Status             core.BridgingRequestStatus `json:"status"`
	DestinationTxHash  indexer.Hash               `json:"destinationTxHash"`
}

func NewBridgingRequestStateResponse(state *core.BridgingRequestState) *BridgingRequestStateResponse {
	return &BridgingRequestStateResponse{
		SourceChainID:      state.SourceChainID,
		SourceTxHash:       state.SourceTxHash,
		DestinationChainID: state.DestinationChainID,
		DestinationTxHash:  state.DestinationTxHash,
		Status:             state.Status,
	}
}
