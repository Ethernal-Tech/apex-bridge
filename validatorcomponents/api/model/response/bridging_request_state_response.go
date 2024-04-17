package response

import "github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"

type BridgingRequestStateResponse struct {
	SourceChainId      string                     `json:"sourceChainId"`
	SourceTxHash       string                     `json:"sourceTxHash"`
	DestinationChainId string                     `json:"destinationChainId"`
	Status             core.BridgingRequestStatus `json:"status"`
	DestinationTxHash  string                     `json:"destinationTxHash"`
}

func NewBridgingRequestStateResponse(state *core.BridgingRequestState) *BridgingRequestStateResponse {
	return &BridgingRequestStateResponse{
		SourceChainId:      state.SourceChainId,
		SourceTxHash:       state.SourceTxHash,
		DestinationChainId: state.DestinationChainId,
		DestinationTxHash:  state.DestinationTxHash,
		Status:             state.Status,
	}
}
