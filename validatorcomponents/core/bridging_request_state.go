package core

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
)

type BridgingRequestStatus string //@name BridgingRequestStatus

const (
	BridgingRequestStatusDiscoveredOnSource           BridgingRequestStatus = "DiscoveredOnSource"
	BridgingRequestStatusInvalidRequest               BridgingRequestStatus = "InvalidRequest"
	BridgingRequestStatusSubmittedToBridge            BridgingRequestStatus = "SubmittedToBridge"
	BridgingRequestStatusIncludedInBatch              BridgingRequestStatus = "IncludedInBatch"
	BridgingRequestStatusSubmittedToDestination       BridgingRequestStatus = "SubmittedToDestination"
	BridgingRequestStatusFailedToExecuteOnDestination BridgingRequestStatus = "FailedToExecuteOnDestination"
	BridgingRequestStatusExecutedOnDestination        BridgingRequestStatus = "ExecutedOnDestination"
)

type BridgingRequestState struct {
	SourceChainID      string
	SourceTxHash       common.Hash
	DestinationChainID string
	Status             BridgingRequestStatus
	DestinationTxHash  common.Hash
}

func (s *BridgingRequestState) ToDBKey() []byte {
	return ToBridgingRequestStateDBKey(s.SourceChainID, s.SourceTxHash)
}

func ToBridgingRequestStateDBKey(sourceChainID string, sourceTxHash common.Hash) []byte {
	return append(append([]byte(sourceChainID), '_'), sourceTxHash[:]...)
}

func NewBridgingRequestState(sourceChainID string, sourceTxHash common.Hash) *BridgingRequestState {
	return &BridgingRequestState{
		SourceChainID: sourceChainID,
		SourceTxHash:  sourceTxHash,
		Status:        BridgingRequestStatusDiscoveredOnSource,
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

func (s *BridgingRequestState) ToExecutedOnDestination(destinationTxHash common.Hash) {
	s.Status = BridgingRequestStatusExecutedOnDestination
	s.DestinationTxHash = destinationTxHash
}

func (s *BridgingRequestState) UpdateDestChainID(chainID string) error {
	if s.DestinationChainID == "" {
		s.DestinationChainID = chainID
	}

	if s.DestinationChainID != chainID {
		return fmt.Errorf("destination chain not equal %s != %s for (%s, %s)",
			s.DestinationChainID, chainID, s.SourceChainID, s.SourceTxHash)
	}

	return nil
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
			s.SourceChainID, s.SourceTxHash, s.Status, newStatus)
	}

	return nil
}
