package common

import (
	"fmt"
	"math/big"
	"time"
)

type BridgingRequestStatus string // @name BridgingRequestStatus

const (
	BridgingRequestStatusDiscoveredOnSource           BridgingRequestStatus = "DiscoveredOnSource"
	BridgingRequestStatusInvalidRequest               BridgingRequestStatus = "InvalidRequest"
	BridgingRequestStatusSubmittedToBridge            BridgingRequestStatus = "SubmittedToBridge"
	BridgingRequestStatusIncludedInBatch              BridgingRequestStatus = "IncludedInBatch"
	BridgingRequestStatusSubmittedToDestination       BridgingRequestStatus = "SubmittedToDestination"
	BridgingRequestStatusFailedToExecuteOnDestination BridgingRequestStatus = "FailedToExecuteOnDestination"
	BridgingRequestStatusExecutedOnDestination        BridgingRequestStatus = "ExecutedOnDestination"

	bridgingRequestStatusRefundRequestSubmittedToBridge = "RefundRequestSubmittedToBridge"
	bridgingRequestStatusRefundSubmittedToChain         = "RefundSubmittedToChain"
	bridgingRequestStatusFailedToRefund                 = "FailedToRefund"
	bridgingRequestStatusRefundExecuted                 = "RefundExecuted"
)

// BridgingRequestStateReceiver is a single output of a bridging request, as declared in tx metadata.
type BridgingRequestStateReceiver struct {
	Address string
	Amount  *big.Int // wei
	TokenID uint16
}

// BridgingRequestStateDetails holds what was requested to be bridged. It is filled on a best effort
// basis when the request is first observed, so it is nil for refund requests and for every state
// stored before this field existed.
type BridgingRequestStateDetails struct {
	SenderAddr   string
	Receivers    []BridgingRequestStateReceiver
	Amount       *big.Int // native currency sent to the receivers on source, wei
	TokenAmount  *big.Int // token sent to the receivers on source, wei
	TokenID      uint16   // token id, zero when nothing but currency is bridged
	BridgingFee  *big.Int
	OperationFee *big.Int
}

type BridgingRequestState struct {
	SourceChainID      string
	SourceTxHash       []byte
	DestinationChainID string
	Status             BridgingRequestStatus
	DestinationTxHash  []byte
	IsRefund           bool
	// CreatedAt is assigned by the database when the state is first stored.
	// It is zero for states stored before this field existed.
	CreatedAt time.Time
	Details   *BridgingRequestStateDetails
}

func NewBridgingRequestStateDetails(
	senderAddr string, receivers []BridgingRequestStateReceiver,
	bridgingFee, operationFee *big.Int, currencyID uint16,
) *BridgingRequestStateDetails {
	details := &BridgingRequestStateDetails{
		SenderAddr:   senderAddr,
		Receivers:    receivers,
		Amount:       big.NewInt(0),
		TokenAmount:  big.NewInt(0),
		BridgingFee:  bridgingFee,
		OperationFee: operationFee,
	}

	for _, receiver := range receivers {
		if receiver.Amount == nil {
			continue
		}

		if receiver.TokenID == currencyID {
			details.Amount.Add(details.Amount, receiver.Amount)
		} else {
			details.TokenAmount.Add(details.TokenAmount, receiver.Amount)
			details.TokenID = receiver.TokenID
		}
	}

	return details
}

func (s *BridgingRequestState) ToDBKey() []byte {
	return ToBridgingRequestStateDBKey(s.SourceChainID, s.SourceTxHash)
}

func (s *BridgingRequestState) StatusStr() string {
	return BridgingRequestStateStatusStr(s.Status, s.IsRefund)
}

func BridgingRequestStateStatusStr(status BridgingRequestStatus, isRefund bool) string {
	if !isRefund {
		return string(status)
	}

	switch status {
	case BridgingRequestStatusSubmittedToBridge:
		return bridgingRequestStatusRefundRequestSubmittedToBridge
	case BridgingRequestStatusSubmittedToDestination:
		return bridgingRequestStatusRefundSubmittedToChain
	case BridgingRequestStatusFailedToExecuteOnDestination:
		return bridgingRequestStatusFailedToRefund
	case BridgingRequestStatusExecutedOnDestination:
		return bridgingRequestStatusRefundExecuted
	default:
		return string(status)
	}
}

func ToBridgingRequestStateDBKey(sourceChainID string, sourceTxHash []byte) []byte {
	return append(append([]byte(sourceChainID), '_'), sourceTxHash[:]...)
}

func NewBridgingRequestState(sourceChainID string, sourceTxHash []byte, isRefund bool) *BridgingRequestState {
	return &BridgingRequestState{
		SourceChainID: sourceChainID,
		SourceTxHash:  sourceTxHash,
		Status:        BridgingRequestStatusDiscoveredOnSource,
		IsRefund:      isRefund,
	}
}

func (s *BridgingRequestState) ToInvalidRequest() {
	s.Status = BridgingRequestStatusInvalidRequest
}

func (s *BridgingRequestState) ToSubmittedToBridge() {
	s.Status = BridgingRequestStatusSubmittedToBridge
}

func (s *BridgingRequestState) ToIncludedInBatch() {
	s.Status = BridgingRequestStatusIncludedInBatch
}

func (s *BridgingRequestState) ToSubmittedToDestination() {
	s.Status = BridgingRequestStatusSubmittedToDestination
}

func (s *BridgingRequestState) ToFailedToExecuteOnDestination() {
	s.Status = BridgingRequestStatusFailedToExecuteOnDestination
}

func (s *BridgingRequestState) ToExecutedOnDestination(destinationTxHash []byte) {
	s.Status = BridgingRequestStatusExecutedOnDestination
	s.DestinationTxHash = destinationTxHash
}

func (s *BridgingRequestState) IsTransitionPossible(newStatus BridgingRequestStatus) error {
	isInvalidTransition := false

	switch s.Status {
	case BridgingRequestStatusDiscoveredOnSource:

	case BridgingRequestStatusInvalidRequest:
		isInvalidTransition = true

	case BridgingRequestStatusSubmittedToBridge:
		isInvalidTransition = newStatus == BridgingRequestStatusDiscoveredOnSource ||
			newStatus == BridgingRequestStatusInvalidRequest

	case BridgingRequestStatusIncludedInBatch:
		isInvalidTransition = newStatus == BridgingRequestStatusDiscoveredOnSource ||
			newStatus == BridgingRequestStatusInvalidRequest ||
			newStatus == BridgingRequestStatusSubmittedToBridge

	case BridgingRequestStatusSubmittedToDestination:
		isInvalidTransition = newStatus == BridgingRequestStatusDiscoveredOnSource ||
			newStatus == BridgingRequestStatusInvalidRequest ||
			newStatus == BridgingRequestStatusSubmittedToBridge || newStatus == BridgingRequestStatusIncludedInBatch

	case BridgingRequestStatusFailedToExecuteOnDestination:
		isInvalidTransition = newStatus == BridgingRequestStatusDiscoveredOnSource

	case BridgingRequestStatusExecutedOnDestination:
		isInvalidTransition = true
	}

	if isInvalidTransition {
		return fmt.Errorf("BridgingRequestState (%s, %s) invalid transition %s -> %s",
			s.SourceChainID, s.SourceTxHash, s.StatusStr(), newStatus)
	}

	return nil
}
