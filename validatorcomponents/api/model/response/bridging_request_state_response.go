package response

import "github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"

type BridgingRequestStateResponse struct {
	SourceChainID      string                     `json:"sourceChainId"`
	SourceTxHash       string                     `json:"sourceTxHash"`
	DestinationChainID string                     `json:"destinationChainId"`
	Status             core.BridgingRequestStatus `json:"status"`
	DestinationTxHash  string                     `json:"destinationTxHash"`
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
