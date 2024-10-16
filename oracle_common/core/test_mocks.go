package core

import "github.com/stretchr/testify/mock"

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
