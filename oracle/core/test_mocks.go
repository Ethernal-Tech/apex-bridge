package core

import (
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/stretchr/testify/mock"
)

type CardanoTxsProcessorMock struct {
	mock.Mock
	NewUnprocessedTxsFn func(originChainId string, txs []*indexer.Tx) error
}

// NewUnprocessedTxs implements CardanoTxsProcessor.
func (m *CardanoTxsProcessorMock) NewUnprocessedTxs(originChainId string, txs []*indexer.Tx) error {
	if m.NewUnprocessedTxsFn != nil {
		return m.NewUnprocessedTxsFn(originChainId, txs)
	}

	args := m.Called()
	return args.Error(0)
}

// Start implements CardanoTxsProcessor.
func (m *CardanoTxsProcessorMock) Start() error {
	args := m.Called()
	return args.Error(0)
}

// Stop implements CardanoTxsProcessor.
func (m *CardanoTxsProcessorMock) Stop() error {
	args := m.Called()
	return args.Error(0)
}

var _ CardanoTxsProcessor = (*CardanoTxsProcessorMock)(nil)

type BridgeDataFetcherMock struct {
	mock.Mock
}

func (m *BridgeDataFetcherMock) Start() error {
	return nil
}

func (m *BridgeDataFetcherMock) Stop() error {
	return nil
}

func (m *BridgeDataFetcherMock) FetchLatestBlockPoint(chainId string) (*indexer.BlockPoint, error) {
	return nil, nil
}

var _ BridgeDataFetcher = (*BridgeDataFetcherMock)(nil)

type CardanoTxsProcessorDbMock struct {
	mock.Mock
}

func (m *CardanoTxsProcessorDbMock) AddExpectedTxs(expectedTxs []*BridgeExpectedCardanoTx) error {
	return nil
}

func (m *CardanoTxsProcessorDbMock) GetExpectedTxs(threshold int) ([]*BridgeExpectedCardanoTx, error) {
	return nil, nil
}

func (m *CardanoTxsProcessorDbMock) ClearExpectedTxs(chainId string) error {
	return nil
}

func (m *CardanoTxsProcessorDbMock) MarkExpectedTxsAsProcessed(expectedTxs []*BridgeExpectedCardanoTx) error {
	return nil
}

func (m *CardanoTxsProcessorDbMock) MarkExpectedTxsAsInvalid(expectedTxs []*BridgeExpectedCardanoTx) error {
	return nil
}

func (m *CardanoTxsProcessorDbMock) AddUnprocessedTxs(unprocessedTxs []*CardanoTx) error {
	return nil
}

func (m *CardanoTxsProcessorDbMock) GetUnprocessedTxs(threshold int) ([]*CardanoTx, error) {
	return nil, nil
}

func (m *CardanoTxsProcessorDbMock) ClearUnprocessedTxs(chainId string) error {
	return nil
}

func (m *CardanoTxsProcessorDbMock) MarkUnprocessedTxsAsProcessed(processedTxs []*ProcessedCardanoTx) error {
	return nil
}

func (m *CardanoTxsProcessorDbMock) GetProcessedTx(chainId string, txHash string) (*ProcessedCardanoTx, error) {
	return nil, nil
}

var _ CardanoTxsProcessorDb = (*CardanoTxsProcessorDbMock)(nil)

type ClaimsSubmitterMock struct {
	mock.Mock
	OnSubmitClaims func(claims *BridgeClaims)
}

// SubmitClaims implements ClaimsSubmitter.
func (m *ClaimsSubmitterMock) SubmitClaims(claims *BridgeClaims) error {
	if m.OnSubmitClaims != nil {
		m.OnSubmitClaims(claims)
	}

	args := m.Called()
	return args.Error(0)
}

// Dispose implements ClaimsSubmitter.
func (m *ClaimsSubmitterMock) Dispose() error {
	args := m.Called()
	return args.Error(0)
}

var _ ClaimsSubmitter = (*ClaimsSubmitterMock)(nil)

type CardanoTxProcessorMock struct {
	mock.Mock
	ShouldAddClaim bool
}

// IsTxRelevant implements CardanoTxProcessor.
func (m *CardanoTxProcessorMock) IsTxRelevant(tx *CardanoTx, appConfig *AppConfig) (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

// ValidateAndAddClaim implements CardanoTxProcessor.
func (m *CardanoTxProcessorMock) ValidateAndAddClaim(claims *BridgeClaims, tx *CardanoTx, appConfig *AppConfig) error {
	if m.ShouldAddClaim {
		claims.BridgingRequest = append(claims.BridgingRequest, BridgingRequestClaim{})
	}

	args := m.Called()
	return args.Error(0)
}

var _ CardanoTxProcessor = (*CardanoTxProcessorMock)(nil)

type CardanoTxFailedProcessorMock struct {
	mock.Mock
	ShouldAddClaim bool
}

// IsTxRelevant implements CardanoTxFailedProcessor.
func (m *CardanoTxFailedProcessorMock) IsTxRelevant(tx *BridgeExpectedCardanoTx, appConfig *AppConfig) (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

// ValidateAndAddClaim implements CardanoTxFailedProcessor.
func (m *CardanoTxFailedProcessorMock) ValidateAndAddClaim(claims *BridgeClaims, tx *BridgeExpectedCardanoTx, appConfig *AppConfig) error {
	if m.ShouldAddClaim {
		claims.BatchExecutionFailed = append(claims.BatchExecutionFailed, BatchExecutionFailedClaim{})
	}

	args := m.Called()
	return args.Error(0)
}

var _ CardanoTxFailedProcessor = (*CardanoTxFailedProcessorMock)(nil)
