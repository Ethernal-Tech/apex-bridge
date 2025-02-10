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
func (m *BridgingRequestStateDBMock) AddBridgingRequestState(state *core.BridgingRequestState) error {
	args := m.Called(state)

	return args.Error(0)
}

// GetBridgingRequestState implements core.BridgingRequestStateDb.
func (m *BridgingRequestStateDBMock) GetBridgingRequestState(
	sourceChainID string, sourceTxHash common.Hash,
) (*core.BridgingRequestState, error) {
	args := m.Called(sourceChainID, sourceTxHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	arg0, _ := args.Get(0).(*core.BridgingRequestState)

	return arg0, args.Error(1)
}

// UpdateBridgingRequestState implements core.BridgingRequestStateDb.
func (m *BridgingRequestStateDBMock) UpdateBridgingRequestState(state *core.BridgingRequestState) error {
	args := m.Called(state)

	return args.Error(0)
}

var _ core.BridgingRequestStateDB = (*BridgingRequestStateDBMock)(nil)
