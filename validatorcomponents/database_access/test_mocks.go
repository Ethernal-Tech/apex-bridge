package databaseaccess

import (
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/stretchr/testify/mock"
)

type BridgingRequestStateDBMock struct {
	mock.Mock
}

// AddBridgingRequestState implements core.BridgingRequestStateDb.
func (m *BridgingRequestStateDBMock) AddBridgingRequestState(state *core.BridgingRequestState) error {
	args := m.Called()

	return args.Error(0)
}

// GetBridgingRequestState implements core.BridgingRequestStateDb.
func (m *BridgingRequestStateDBMock) GetBridgingRequestState(
	sourceChainID string, sourceTxHash string,
) (
	*core.BridgingRequestState, error,
) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	arg0, _ := args.Get(0).(*core.BridgingRequestState)

	return arg0, args.Error(1)
}

// GetBridgingRequestStatesByBatchID implements core.BridgingRequestStateDb.
func (m *BridgingRequestStateDBMock) GetBridgingRequestStatesByBatchID(
	destinationChainID string, batchID uint64,
) (
	[]*core.BridgingRequestState, error,
) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	arg0, _ := args.Get(0).([]*core.BridgingRequestState)

	return arg0, args.Error(1)
}

// GetUserBridgingRequestStates implements core.BridgingRequestStateDb.
func (m *BridgingRequestStateDBMock) GetUserBridgingRequestStates(
	sourceChainID string, userAddr string,
) (
	[]*core.BridgingRequestState, error,
) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	arg0, _ := args.Get(0).([]*core.BridgingRequestState)

	return arg0, args.Error(1)
}

// UpdateBridgingRequestState implements core.BridgingRequestStateDb.
func (m *BridgingRequestStateDBMock) UpdateBridgingRequestState(state *core.BridgingRequestState) error {
	args := m.Called()

	return args.Error(0)
}

var _ core.BridgingRequestStateDB = (*BridgingRequestStateDBMock)(nil)
