package common

import (
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/stretchr/testify/mock"
)

type BridgingRequestStateUpdaterMock struct {
	mock.Mock
	ReturnNil bool
}

var _ BridgingRequestStateUpdater = (*BridgingRequestStateUpdaterMock)(nil)

// New implements BridgingRequestStateUpdater.
func (m *BridgingRequestStateUpdaterMock) New(sourceChainId string, tx *indexer.Tx) error {
	if m.ReturnNil {
		return nil
	}

	args := m.Called()
	return args.Error(0)
}

// NewMultiple implements BridgingRequestStateUpdater.
func (m *BridgingRequestStateUpdaterMock) NewMultiple(sourceChainId string, txs []*indexer.Tx) error {
	if m.ReturnNil {
		return nil
	}

	args := m.Called()
	return args.Error(0)
}

// Invalid implements BridgingRequestStateUpdater.
func (m *BridgingRequestStateUpdaterMock) Invalid(key BridgingRequestStateKey) error {
	if m.ReturnNil {
		return nil
	}

	args := m.Called()
	return args.Error(0)
}

// SubmittedToBridge implements BridgingRequestStateUpdater.
func (m *BridgingRequestStateUpdaterMock) SubmittedToBridge(
	key BridgingRequestStateKey, destinationChainId string,
) error {
	if m.ReturnNil {
		return nil
	}

	args := m.Called()
	return args.Error(0)
}

// IncludedInBatch implements BridgingRequestStateUpdater.
func (m *BridgingRequestStateUpdaterMock) IncludedInBatch(destinationChainId string, batchId uint64, txs []BridgingRequestStateKey) error {
	if m.ReturnNil {
		return nil
	}

	args := m.Called()
	return args.Error(0)
}

// SubmittedToDestination implements BridgingRequestStateUpdater.
func (m *BridgingRequestStateUpdaterMock) SubmittedToDestination(destinationChainId string, batchId uint64) error {
	if m.ReturnNil {
		return nil
	}

	args := m.Called()
	return args.Error(0)
}

// FailedToExecuteOnDestination implements BridgingRequestStateUpdater.
func (m *BridgingRequestStateUpdaterMock) FailedToExecuteOnDestination(destinationChainId string, batchId uint64) error {
	if m.ReturnNil {
		return nil
	}

	args := m.Called()
	return args.Error(0)
}

// ExecutedOnDestination implements BridgingRequestStateUpdater.
func (m *BridgingRequestStateUpdaterMock) ExecutedOnDestination(destinationChainId string, batchId uint64, destinationTxHash string) error {
	if m.ReturnNil {
		return nil
	}

	args := m.Called()
	return args.Error(0)
}
