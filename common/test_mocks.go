package common

import (
	"fmt"

	"github.com/stretchr/testify/mock"
)

func SimulateRealMetadata[
	T BaseMetadata | BridgingRequestMetadata | BatchExecutedMetadata,
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

	args := m.Called(sourceChainID, tx)

	return args.Error(0)
}

// NewMultiple implements BridgingRequestStateUpdater.
func (m *BridgingRequestStateUpdaterMock) NewMultiple(sourceChainID string, txs []*NewBridgingRequestStateModel) error {
	if m.ReturnNil {
		return nil
	}

	args := m.Called(sourceChainID, txs)

	return args.Error(0)
}

// Invalid implements BridgingRequestStateUpdater.
func (m *BridgingRequestStateUpdaterMock) Invalid(key BridgingRequestStateKey) error {
	if m.ReturnNil {
		return nil
	}

	args := m.Called(key)

	return args.Error(0)
}

// SubmittedToBridge implements BridgingRequestStateUpdater.
func (m *BridgingRequestStateUpdaterMock) SubmittedToBridge(
	key BridgingRequestStateKey, dstChainID string,
) error {
	if m.ReturnNil {
		return nil
	}

	args := m.Called(key, dstChainID)

	return args.Error(0)
}

// IncludedInBatch implements BridgingRequestStateUpdater.
func (m *BridgingRequestStateUpdaterMock) IncludedInBatch(
	txs []BridgingRequestStateKey, dstChainID string,
) error {
	if m.ReturnNil {
		return nil
	}

	args := m.Called(txs, dstChainID)

	return args.Error(0)
}

// SubmittedToDestination implements BridgingRequestStateUpdater.
func (m *BridgingRequestStateUpdaterMock) SubmittedToDestination(
	txs []BridgingRequestStateKey, dstChainID string,
) error {
	if m.ReturnNil {
		return nil
	}

	args := m.Called(txs, dstChainID)

	return args.Error(0)
}

// FailedToExecuteOnDestination implements BridgingRequestStateUpdater.
func (m *BridgingRequestStateUpdaterMock) FailedToExecuteOnDestination(
	txs []BridgingRequestStateKey, dstChainID string,
) error {
	if m.ReturnNil {
		return nil
	}

	args := m.Called(txs, dstChainID)

	return args.Error(0)
}

// ExecutedOnDestination implements BridgingRequestStateUpdater.
func (m *BridgingRequestStateUpdaterMock) ExecutedOnDestination(
	txs []BridgingRequestStateKey, destinationTxHash Hash, dstChainID string,
) error {
	if m.ReturnNil {
		return nil
	}

	args := m.Called(txs, destinationTxHash, dstChainID)

	return args.Error(0)
}
