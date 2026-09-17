package response

import (
	"math/big"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
)

func amountStr(amount *big.Int) string {
	if amount == nil {
		return "0"
	}

	return amount.String()
}

type BridgingRequestStateReceiverResponse struct {
	// Destination address
	Address string `json:"address"`
	// Amount in wei
	Amount string `json:"amount"`
	// Token ID
	TokenID uint16 `json:"tokenId"`
} // @name BridgingRequestStateReceiverResponse

type BridgingRequestStateDetailsResponse struct {
	// Sender address on the source chain
	SenderAddr string `json:"senderAddr"`
	// Requested outputs
	Receivers []BridgingRequestStateReceiverResponse `json:"receivers"`
	// Native currency sent to the receivers on the source chain, in wei. Fees are not included, add
	// bridgingFee and operationFee to get everything the sender paid in currency.
	Amount string `json:"amount"`
	// Token sent to the receivers on the source chain, in wei
	TokenAmount string `json:"tokenAmount"`
	// Token ID, zero when nothing but currency is bridged
	TokenID uint16 `json:"tokenId"`
	// Bridging fee in wei
	BridgingFee string `json:"bridgingFee"`
	// Operation fee in wei
	OperationFee string `json:"operationFee"`
} // @name BridgingRequestStateDetailsResponse

type BridgingRequestStateResponse struct {
	// Source chain ID
	SourceChainID string `json:"sourceChainId"`
	// Source transaction hash
	SourceTxHash string `json:"sourceTxHash"`
	// Destination chain ID
	DestinationChainID string `json:"destinationChainId"`
	// Status of bridging request
	Status common.BridgingRequestStatus `json:"status"`
	// Destination transaction hash
	DestinationTxHash string `json:"destinationTxHash"`
	// Is in refund phase
	IsRefund bool `json:"isRefund"`
	// Time the bridging request was first observed
	CreatedAt time.Time `json:"createdAt"`
	// What was requested to be bridged. Null for refund requests and for requests observed before
	// the oracle started recording details.
	Details *BridgingRequestStateDetailsResponse `json:"details"`
} // @name BridgingRequestStateResponse

func NewBridgingRequestStateResponse(state *common.BridgingRequestState) *BridgingRequestStateResponse {
	return &BridgingRequestStateResponse{
		SourceChainID:      state.SourceChainID,
		SourceTxHash:       common.TxHashBytesToString(state.SourceTxHash),
		DestinationChainID: state.DestinationChainID,
		DestinationTxHash:  common.TxHashBytesToString(state.DestinationTxHash),
		Status:             state.Status,
		IsRefund:           state.IsRefund,
		CreatedAt:          state.CreatedAt,
		Details:            newBridgingRequestStateDetailsResponse(state.Details),
	}
}

func newBridgingRequestStateDetailsResponse(
	details *common.BridgingRequestStateDetails,
) *BridgingRequestStateDetailsResponse {
	if details == nil {
		return nil
	}

	receivers := make([]BridgingRequestStateReceiverResponse, len(details.Receivers))
	for i, receiver := range details.Receivers {
		receivers[i] = BridgingRequestStateReceiverResponse{
			Address: receiver.Address,
			Amount:  amountStr(receiver.Amount),
			TokenID: receiver.TokenID,
		}
	}

	return &BridgingRequestStateDetailsResponse{
		SenderAddr:   details.SenderAddr,
		Receivers:    receivers,
		Amount:       amountStr(details.Amount),
		TokenAmount:  amountStr(details.TokenAmount),
		TokenID:      details.TokenID,
		BridgingFee:  amountStr(details.BridgingFee),
		OperationFee: amountStr(details.OperationFee),
	}
}

type BridgingRequestStatePageResponse struct {
	// Identifies the oracle database. A different value than the one seen previously means the
	// database was recreated and any stored nextFrom is meaningless.
	InstanceID string `json:"instanceId"`
	// Sync index to pass as "from" on the next call
	NextFrom uint64 `json:"nextFrom"`
	// Whether a full page was returned, meaning there may be more
	HasMore bool `json:"hasMore"`
	// Bridging request states, in the order they were observed
	Items []*BridgingRequestStateResponse `json:"items"`
} // @name BridgingRequestStatePageResponse

func NewBridgingRequestStatePageResponse(
	states []*common.BridgingRequestState, nextFrom uint64, instanceID string, limit int,
) *BridgingRequestStatePageResponse {
	items := make([]*BridgingRequestStateResponse, len(states))
	for i, state := range states {
		items[i] = NewBridgingRequestStateResponse(state)
	}

	return &BridgingRequestStatePageResponse{
		InstanceID: instanceID,
		NextFrom:   nextFrom,
		HasMore:    len(states) == limit,
		Items:      items,
	}
}
