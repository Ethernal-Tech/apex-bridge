package core

import (
	"time"

	"github.com/stretchr/testify/mock"
)

type TxsProcessorMock struct {
	mock.Mock
}

// Start implements CardanoTxsProcessor.
func (m *TxsProcessorMock) Start() {
}

var _ TxsProcessor = (*TxsProcessorMock)(nil)

type ExpectedTxsFetcherMock struct {
	mock.Mock
}

// Start implements ExpectedTxsFetcher.
func (m *ExpectedTxsFetcherMock) Start() {
}

var _ ExpectedTxsFetcher = (*ExpectedTxsFetcherMock)(nil)

type ProtocolParamsDBMock struct {
	Params    []byte
	ExpiresAt time.Time
	SaveErr   error
	GetErr    error
}

var _ ProtocolParamsDB = (*ProtocolParamsDBMock)(nil)

func (m *ProtocolParamsDBMock) SaveProtocolParams(
	chainID string, protocolParams []byte, expiresAt time.Time,
) error {
	if m.SaveErr != nil {
		return m.SaveErr
	}

	m.Params = protocolParams
	m.ExpiresAt = expiresAt

	return nil
}

func (m *ProtocolParamsDBMock) GetProtocolParams(chainID string) ([]byte, time.Time, error) {
	if m.GetErr != nil {
		return nil, time.Time{}, m.GetErr
	}

	return m.Params, m.ExpiresAt, nil
}
