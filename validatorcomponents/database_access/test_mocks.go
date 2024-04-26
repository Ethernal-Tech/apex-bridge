package database_access

import (
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/stretchr/testify/mock"
)

type BridgingRequestStateDbMock struct {
	mock.Mock
}

// AddBridgingRequestState implements core.BridgingRequestStateDb.
func (m *BridgingRequestStateDbMock) AddBridgingRequestState(state *core.BridgingRequestState) error {
	args := m.Called()
	return args.Error(0)
}

// GetBridgingRequestState implements core.BridgingRequestStateDb.
func (m *BridgingRequestStateDbMock) GetBridgingRequestState(
	sourceChainId string, sourceTxHash string,
) (
	*core.BridgingRequestState, error,
) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*core.BridgingRequestState), args.Error(1)
}

// GetBridgingRequestStatesByBatchId implements core.BridgingRequestStateDb.
func (m *BridgingRequestStateDbMock) GetBridgingRequestStatesByBatchId(
	destinationChainId string, batchId uint64,
) (
	[]*core.BridgingRequestState, error,
) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	arg0 := args.Get(0).([]*core.BridgingRequestState)
	return arg0, args.Error(1)
}

// GetUserBridgingRequestStates implements core.BridgingRequestStateDb.
func (m *BridgingRequestStateDbMock) GetUserBridgingRequestStates(
	sourceChainId string, userAddr string,
) (
	[]*core.BridgingRequestState, error,
) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]*core.BridgingRequestState), args.Error(1)
}

// UpdateBridgingRequestState implements core.BridgingRequestStateDb.
func (m *BridgingRequestStateDbMock) UpdateBridgingRequestState(state *core.BridgingRequestState) error {
	args := m.Called()
	return args.Error(0)
}

var _ core.BridgingRequestStateDb = (*BridgingRequestStateDbMock)(nil)
