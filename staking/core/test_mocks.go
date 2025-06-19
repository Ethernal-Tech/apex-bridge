package core

import (
	ocCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/stretchr/testify/mock"
)

type CardanoChainObserverMock struct {
	mock.Mock
}

func (m *CardanoChainObserverMock) Start() error {
	args := m.Called()

	return args.Error(0)
}

func (m *CardanoChainObserverMock) Dispose() error {
	args := m.Called()

	return args.Error(0)
}

func (m *CardanoChainObserverMock) GetConfig() ocCore.ChainConfigReader {
	args := m.Called()
	return args.Get(0).(ocCore.ChainConfigReader) //nolint
}

func (m *CardanoChainObserverMock) ErrorCh() <-chan error {
	args := m.Called()
	return args.Get(0).(chan error) //nolint
}
