package common

import (
	"fmt"

	"github.com/stretchr/testify/mock"
)

func SimulateRealMetadata[
	T BaseMetadata | BridgingRequestMetadata | BatchExecutedMetadata | RefundExecutedMetadata | BridgingRequestMetadataV1,
](
	encodingType MetadataEncodingType, metadata T,
) (
	[]byte, error,
) {
	marshalFunc, err := getMarshalFunc(encodingType)
	if err != nil {
		return nil, err
	}

	result, err := marshalFunc(map[int]map[int]T{1: {MetadataMapKey: metadata}})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %v, err: %w", metadata, err)
	}

	return result, nil
}

type BridgingRequestStateUpdaterMock struct {
	mock.Mock
	ReturnNil bool
}

var _ BridgingRequestStateUpdater = (*BridgingRequestStateUpdaterMock)(nil)

// New implements BridgingRequestStateUpdater.
func (m *BridgingRequestStateUpdaterMock) New(sourceChainID string, tx *NewBridgingRequestStateModel) error {
	if m.ReturnNil {
		return nil
	}

	args := m.Called()

	return args.Error(0)
}

// NewMultiple implements BridgingRequestStateUpdater.
func (m *BridgingRequestStateUpdaterMock) NewMultiple(sourceChainID string, txs []*NewBridgingRequestStateModel) error {
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
	key BridgingRequestStateKey, destinationChainID string,
) error {
	if m.ReturnNil {
		return nil
	}

	args := m.Called()

	return args.Error(0)
}

// IncludedInBatch implements BridgingRequestStateUpdater.
func (m *BridgingRequestStateUpdaterMock) IncludedInBatch(
	destinationChainID string, batchID uint64, txs []BridgingRequestStateKey,
) error {
	if m.ReturnNil {
		return nil
	}

	args := m.Called()

	return args.Error(0)
}

// SubmittedToDestination implements BridgingRequestStateUpdater.
func (m *BridgingRequestStateUpdaterMock) SubmittedToDestination(destinationChainID string, batchID uint64) error {
	if m.ReturnNil {
		return nil
	}

	args := m.Called()

	return args.Error(0)
}

// FailedToExecuteOnDestination implements BridgingRequestStateUpdater.
func (m *BridgingRequestStateUpdaterMock) FailedToExecuteOnDestination(
	destinationChainID string, batchID uint64,
) error {
	if m.ReturnNil {
		return nil
	}

	args := m.Called()

	return args.Error(0)
}

// ExecutedOnDestination implements BridgingRequestStateUpdater.
func (m *BridgingRequestStateUpdaterMock) ExecutedOnDestination(
	destinationChainID string, batchID uint64, destinationTxHash Hash,
) error {
	if m.ReturnNil {
		return nil
	}

	args := m.Called()

	return args.Error(0)
}
