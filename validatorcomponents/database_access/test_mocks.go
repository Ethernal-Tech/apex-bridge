package databaseaccess

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/stretchr/testify/mock"
)

type BridgingRequestStateDBMock struct {
	mock.Mock
}

// AddBridgingRequestState implements core.BridgingRequestStateDb.
func (m *BridgingRequestStateDBMock) AddBridgingRequestState(state *common.BridgingRequestState) error {
	args := m.Called(state)

	return args.Error(0)
}

// GetBridgingRequestState implements core.BridgingRequestStateDb.
func (m *BridgingRequestStateDBMock) GetBridgingRequestState(
	sourceChainID string, sourceTxHash []byte,
) (*common.BridgingRequestState, error) {
	args := m.Called(sourceChainID, sourceTxHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	arg0, _ := args.Get(0).(*common.BridgingRequestState)

	return arg0, args.Error(1)
}

// UpdateBridgingRequestState implements core.BridgingRequestStateDb.
func (m *BridgingRequestStateDBMock) UpdateBridgingRequestState(state *common.BridgingRequestState) error {
	args := m.Called(state)

	return args.Error(0)
}

// GetBridgingRequestStatesPage implements core.BridgingRequestStateDb.
func (m *BridgingRequestStateDBMock) GetBridgingRequestStatesPage(
	fromSyncIndex uint64, limit int,
) ([]*common.BridgingRequestState, uint64, error) {
	args := m.Called(fromSyncIndex, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}

	arg0, _ := args.Get(0).([]*common.BridgingRequestState)
	arg1, _ := args.Get(1).(uint64)

	return arg0, arg1, args.Error(2)
}

// GetSyncInstanceID implements core.BridgingRequestStateDb.
func (m *BridgingRequestStateDBMock) GetSyncInstanceID() (string, error) {
	args := m.Called()

	return args.String(0), args.Error(1)
}

var _ core.BridgingRequestStateDB = (*BridgingRequestStateDBMock)(nil)
