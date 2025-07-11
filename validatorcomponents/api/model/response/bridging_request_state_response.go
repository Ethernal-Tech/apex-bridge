package response

import (
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
)

type BridgingRequestStateResponse struct {
	// Source chain ID
	SourceChainID string `json:"sourceChainId"`
	// Source transaction hash
	SourceTxHash string `json:"sourceTxHash"`
	// Destination chain ID
	DestinationChainID string `json:"destinationChainId"`
	// Status of bridging request
	Status core.BridgingRequestStatus `json:"status"`
	// Destination transaction hash
	DestinationTxHash string `json:"destinationTxHash"`
} // @name BridgingRequestStateResponse

func NewBridgingRequestStateResponse(state *core.BridgingRequestState) *BridgingRequestStateResponse {
	return &BridgingRequestStateResponse{
		SourceChainID:      state.SourceChainID,
		SourceTxHash:       state.SourceTxHash.String(),
		DestinationChainID: state.DestinationChainID,
		DestinationTxHash:  state.DestinationTxHash.String(),
		Status:             state.Status,
	}
}
