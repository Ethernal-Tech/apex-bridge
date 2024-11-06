package core

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
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
)

type BridgingRequestState struct {
	SourceChainID      string
	SourceTxHash       common.Hash
	DestinationChainID string
	BatchID            uint64
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

func (s *BridgingRequestState) ToInvalidRequest() error {
	if s.Status != BridgingRequestStatusDiscoveredOnSource {
		return fmt.Errorf("can not change BridgingRequestState={sourceChainId: %v, sourceTxHash: %v} from %v status to %v",
			s.SourceChainID, s.SourceTxHash, s.Status, BridgingRequestStatusInvalidRequest)
	}

	s.Status = BridgingRequestStatusInvalidRequest

	return nil
}

func (s *BridgingRequestState) ToSubmittedToBridge(destinationChainID string) error {
	if s.Status != BridgingRequestStatusDiscoveredOnSource &&
		s.Status != BridgingRequestStatusFailedToExecuteOnDestination {
		return fmt.Errorf("can not change BridgingRequestState={sourceChainId: %v, sourceTxHash: %v} from %v status to %v",
			s.SourceChainID, s.SourceTxHash, s.Status, BridgingRequestStatusSubmittedToBridge)
	}

	s.Status = BridgingRequestStatusSubmittedToBridge
	s.DestinationChainID = destinationChainID
	s.BatchID = 0

	return nil
}

func (s *BridgingRequestState) ToIncludedInBatch(batchID uint64) error {
	if s.Status != BridgingRequestStatusDiscoveredOnSource &&
		s.Status != BridgingRequestStatusSubmittedToBridge &&
		s.Status != BridgingRequestStatusFailedToExecuteOnDestination {
		return fmt.Errorf("can not change BridgingRequestState={sourceChainId: %v, sourceTxHash: %v} from %v status to %v",
			s.SourceChainID, s.SourceTxHash, s.Status, BridgingRequestStatusIncludedInBatch)
	}

	s.Status = BridgingRequestStatusIncludedInBatch
	s.BatchID = batchID

	return nil
}

func (s *BridgingRequestState) ToSubmittedToDestination() error {
	if s.Status != BridgingRequestStatusSubmittedToBridge &&
		s.Status != BridgingRequestStatusIncludedInBatch {
		return fmt.Errorf("can not change BridgingRequestState={sourceChainId: %v, sourceTxHash: %v} from %v status to %v",
			s.SourceChainID, s.SourceTxHash, s.Status, BridgingRequestStatusSubmittedToDestination)
	}

	s.Status = BridgingRequestStatusSubmittedToDestination

	return nil
}

func (s *BridgingRequestState) ToFailedToExecuteOnDestination() error {
	if s.Status != BridgingRequestStatusIncludedInBatch &&
		s.Status != BridgingRequestStatusSubmittedToDestination {
		return fmt.Errorf("can not change BridgingRequestState={sourceChainId: %v, sourceTxHash: %v} from %v status to %v",
			s.SourceChainID, s.SourceTxHash, s.Status, BridgingRequestStatusFailedToExecuteOnDestination)
	}

	s.Status = BridgingRequestStatusFailedToExecuteOnDestination

	return nil
}

func (s *BridgingRequestState) ToExecutedOnDestination(destinationTxHash common.Hash) error {
	if s.Status != BridgingRequestStatusIncludedInBatch &&
		s.Status != BridgingRequestStatusSubmittedToDestination {
		return fmt.Errorf("can not change BridgingRequestState={sourceChainId: %v, sourceTxHash: %v} from %v status to %v",
			s.SourceChainID, s.SourceTxHash, s.Status, BridgingRequestStatusExecutedOnDestination)
	}

	s.Status = BridgingRequestStatusExecutedOnDestination
	s.DestinationTxHash = destinationTxHash

	return nil
}
