package core

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
)

type BridgingRequestState struct {
	SourceChainId      string
	SourceTxHash       string
	InputAddrs         []string
	DestinationChainId string
	BatchId            uint64
	Status             BridgingRequestStatus
	DestinationTxHash  string
}

func (s *BridgingRequestState) ToStrKey() string {
	return fmt.Sprintf("%v_%v", s.SourceChainId, s.SourceTxHash)
}

func (s *BridgingRequestState) ToDbKey() []byte {
	return []byte(s.ToStrKey())
}

func ToBridgingRequestStateStrKey(sourceChainId string, sourceTxHash string) string {
	return fmt.Sprintf("%v_%v", sourceChainId, sourceTxHash)
}

func ToBridgingRequestStateDbKey(sourceChainId string, sourceTxHash string) []byte {
	return []byte(ToBridgingRequestStateStrKey(sourceChainId, sourceTxHash))
}

func NewBridgingRequestState(sourceChainId string, sourceTxHash string, inputAddrs []string) *BridgingRequestState {
	return &BridgingRequestState{
		SourceChainId: sourceChainId,
		SourceTxHash:  sourceTxHash,
		InputAddrs:    inputAddrs,
		Status:        BridgingRequestStatusDiscoveredOnSource,
	}
}

func (s *BridgingRequestState) ToInvalidRequest() error {
	if s.Status != BridgingRequestStatusDiscoveredOnSource {
		return fmt.Errorf("can not change BridgingRequestState={sourceChainId: %v, sourceTxHash: %v} from %v status to %v", s.SourceChainId, s.SourceTxHash, s.Status, BridgingRequestStatusInvalidRequest)
	}

	s.Status = BridgingRequestStatusInvalidRequest

	return nil
}

func (s *BridgingRequestState) ToSubmittedToBridge(destinationChainId string) error {
	if s.Status != BridgingRequestStatusDiscoveredOnSource {
		return fmt.Errorf("can not change BridgingRequestState={sourceChainId: %v, sourceTxHash: %v} from %v status to %v", s.SourceChainId, s.SourceTxHash, s.Status, BridgingRequestStatusSubmittedToBridge)
	}

	s.Status = BridgingRequestStatusSubmittedToBridge
	s.DestinationChainId = destinationChainId

	return nil
}

func (s *BridgingRequestState) ToIncludedInBatch(batchId uint64) error {
	if s.Status != BridgingRequestStatusSubmittedToBridge {
		return fmt.Errorf("can not change BridgingRequestState={sourceChainId: %v, sourceTxHash: %v} from %v status to %v", s.SourceChainId, s.SourceTxHash, s.Status, BridgingRequestStatusIncludedInBatch)
	}

	s.Status = BridgingRequestStatusIncludedInBatch
	s.BatchId = batchId

	return nil
}

func (s *BridgingRequestState) ToSubmittedToDestination() error {
	if s.Status != BridgingRequestStatusIncludedInBatch {
		return fmt.Errorf("can not change BridgingRequestState={sourceChainId: %v, sourceTxHash: %v} from %v status to %v", s.SourceChainId, s.SourceTxHash, s.Status, BridgingRequestStatusSubmittedToDestination)
	}

	s.Status = BridgingRequestStatusSubmittedToDestination

	return nil
}

func (s *BridgingRequestState) ToFailedToExecuteOnDestination() error {
	if s.Status != BridgingRequestStatusSubmittedToDestination {
		return fmt.Errorf("can not change BridgingRequestState={sourceChainId: %v, sourceTxHash: %v} from %v status to %v", s.SourceChainId, s.SourceTxHash, s.Status, BridgingRequestStatusFailedToExecuteOnDestination)
	}

	s.Status = BridgingRequestStatusFailedToExecuteOnDestination

	return nil
}

func (s *BridgingRequestState) ToExecutedOnDestination(destinationTxHash string) error {
	if s.Status != BridgingRequestStatusSubmittedToDestination {
		return fmt.Errorf("can not change BridgingRequestState={sourceChainId: %v, sourceTxHash: %v} from %v status to %v", s.SourceChainId, s.SourceTxHash, s.Status, BridgingRequestStatusExecutedOnDestination)
	}

	s.Status = BridgingRequestStatusExecutedOnDestination
	s.DestinationTxHash = destinationTxHash

	return nil
}
