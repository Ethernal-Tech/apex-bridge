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

func (m *BridgeDataFetcherMock) FetchLatestBlockPoint(chainId string) (*indexer.BlockPoint, error) {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).(*indexer.BlockPoint), args.Error(1)
	}

	return nil, args.Error(1)
}

// FetchExpectedTxs implements BridgeDataFetcher.
func (m *BridgeDataFetcherMock) FetchExpectedTx(chainId string) (*BridgeExpectedCardanoTx, error) {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).(*BridgeExpectedCardanoTx), args.Error(1)
	}

	return nil, args.Error(1)
}

// Dispose implements BridgeDataFetcher.
func (m *BridgeDataFetcherMock) Dispose() error {
	args := m.Called()
	return args.Error(0)
}

var _ BridgeDataFetcher = (*BridgeDataFetcherMock)(nil)

type ExpectedTxsFetcherMock struct {
	mock.Mock
}

// Start implements ExpectedTxsFetcher.
func (m *ExpectedTxsFetcherMock) Start() error {
	args := m.Called()
	return args.Error(0)
}

// Stop implements ExpectedTxsFetcher.
func (m *ExpectedTxsFetcherMock) Stop() error {
	args := m.Called()
	return args.Error(0)
}

var _ ExpectedTxsFetcher = (*ExpectedTxsFetcherMock)(nil)

type CardanoTxsProcessorDbMock struct {
	mock.Mock
}

func (m *CardanoTxsProcessorDbMock) AddExpectedTxs(expectedTxs []*BridgeExpectedCardanoTx) error {
	args := m.Called()
	return args.Error(0)
}

func (m *CardanoTxsProcessorDbMock) GetExpectedTxs(chainId string, threshold int) ([]*BridgeExpectedCardanoTx, error) {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).([]*BridgeExpectedCardanoTx), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *CardanoTxsProcessorDbMock) ClearExpectedTxs(chainId string) error {
	args := m.Called()
	return args.Error(0)
}

func (m *CardanoTxsProcessorDbMock) MarkExpectedTxsAsProcessed(expectedTxs []*BridgeExpectedCardanoTx) error {
	args := m.Called()
	return args.Error(0)
}

func (m *CardanoTxsProcessorDbMock) MarkExpectedTxsAsInvalid(expectedTxs []*BridgeExpectedCardanoTx) error {
	args := m.Called()
	return args.Error(0)
}

func (m *CardanoTxsProcessorDbMock) AddUnprocessedTxs(unprocessedTxs []*CardanoTx) error {
	args := m.Called()
	return args.Error(0)
}

func (m *CardanoTxsProcessorDbMock) GetUnprocessedTxs(chainId string, threshold int) ([]*CardanoTx, error) {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).([]*CardanoTx), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *CardanoTxsProcessorDbMock) ClearUnprocessedTxs(chainId string) error {
	args := m.Called()
	return args.Error(0)
}

func (m *CardanoTxsProcessorDbMock) MarkUnprocessedTxsAsProcessed(processedTxs []*ProcessedCardanoTx) error {
	args := m.Called()
	return args.Error(0)
}

func (m *CardanoTxsProcessorDbMock) GetProcessedTx(chainId string, txHash string) (*ProcessedCardanoTx, error) {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).(*ProcessedCardanoTx), args.Error(1)
	}
	return nil, args.Error(1)
}

var _ CardanoTxsProcessorDb = (*CardanoTxsProcessorDbMock)(nil)

type BridgeSubmitterMock struct {
	mock.Mock
	OnSubmitClaims          func(claims *BridgeClaims)
	OnSubmitConfirmedBlocks func(chainId string, blocks []*indexer.CardanoBlock)
}

// SubmitClaims implements BridgeSubmitter.
func (m *BridgeSubmitterMock) SubmitClaims(claims *BridgeClaims) error {
	if m.OnSubmitClaims != nil {
		m.OnSubmitClaims(claims)
	}

	args := m.Called()
	return args.Error(0)
}

// SubmitConfirmedBlocks implements BridgeSubmitter.
func (m *BridgeSubmitterMock) SubmitConfirmedBlocks(chainId string, blocks []*indexer.CardanoBlock) error {
	if m.OnSubmitConfirmedBlocks != nil {
		m.OnSubmitConfirmedBlocks(chainId, blocks)
	}

	args := m.Called()
	return args.Error(0)
}

// Dispose implements BridgeSubmitter.
func (m *BridgeSubmitterMock) Dispose() error {
	args := m.Called()
	return args.Error(0)
}

var _ BridgeSubmitter = (*BridgeSubmitterMock)(nil)

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
		claims.BridgingRequestClaims = append(claims.BridgingRequestClaims, BridgingRequestClaim{})
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
		claims.BatchExecutionFailedClaims = append(claims.BatchExecutionFailedClaims, BatchExecutionFailedClaim{})
	}

	args := m.Called()
	return args.Error(0)
}

var _ CardanoTxFailedProcessor = (*CardanoTxFailedProcessorMock)(nil)
