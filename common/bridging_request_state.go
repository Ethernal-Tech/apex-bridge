package common

import (
	"fmt"
)

type BridgingRequestStatus string

const (
	BridgingRequestStatusDiscoveredOnSource           BridgingRequestStatus = "DiscoveredOnSource"
	BridgingRequestStatusInvalidRequest               BridgingRequestStatus = "InvalidRequest"
	BridgingRequestStatusSubmittedToBridge            BridgingRequestStatus = "SubmittedToBridge"
	BridgingRequestStatusIncludedInBatch              BridgingRequestStatus = "IncludedInBatch"
	BridgingRequestStatusSubmittedToDestination       BridgingRequestStatus = "SubmittedToDestination"
	BridgingRequestStatusFailedToExecuteOnDestination BridgingRequestStatus = "FailedToExecuteOnDestination"
	BridgingRequestStatusExecutedOnDestination        BridgingRequestStatus = "ExecutedOnDestination"

	BridgingRequestStatusRefundRequestSubmittedToBridge = "RefundRequestSubmittedToBridge"
	BridgingRequestStatusRefundSubmittedToChain         = "RefundSubmittedToChain"
	BridgingRequestStatusFailedToRefund                 = "FailedToRefund"
	BridgingRequestStatusRefundExecuted                 = "RefundExecuted"
)

type BridgingRequestState struct {
	SourceChainID      string
	SourceTxHash       Hash
	DestinationChainID string
	Status             BridgingRequestStatus
	DestinationTxHash  Hash
	IsRefund           bool
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
		return BridgingRequestStatusRefundRequestSubmittedToBridge
	case BridgingRequestStatusSubmittedToDestination:
		return BridgingRequestStatusRefundSubmittedToChain
	case BridgingRequestStatusFailedToExecuteOnDestination:
		return BridgingRequestStatusFailedToRefund
	case BridgingRequestStatusExecutedOnDestination:
		return BridgingRequestStatusRefundExecuted
	default:
		return string(status)
	}
}

func ToBridgingRequestStateDBKey(sourceChainID string, sourceTxHash Hash) []byte {
	return append(append([]byte(sourceChainID), '_'), sourceTxHash[:]...)
}

func NewBridgingRequestState(sourceChainID string, sourceTxHash Hash, isRefund bool) *BridgingRequestState {
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

func (s *BridgingRequestState) ToExecutedOnDestination(destinationTxHash Hash) {
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
