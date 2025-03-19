package response

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
)

type BridgingRequestStateResponse struct {
	SourceChainID      string `json:"sourceChainId"`
	SourceTxHash       string `json:"sourceTxHash"`
	DestinationChainID string `json:"destinationChainId"`
	Status             string `json:"status"`
	DestinationTxHash  string `json:"destinationTxHash"`
	IsRefund           bool   `json:"isRefund"`
}

func NewBridgingRequestStateResponse(state *common.BridgingRequestState) *BridgingRequestStateResponse {
	return &BridgingRequestStateResponse{
		SourceChainID:      state.SourceChainID,
		SourceTxHash:       state.SourceTxHash.String(),
		DestinationChainID: state.DestinationChainID,
		DestinationTxHash:  state.DestinationTxHash.String(),
		Status:             state.StatusStr(),
		IsRefund:           state.IsRefund,
	}
}
