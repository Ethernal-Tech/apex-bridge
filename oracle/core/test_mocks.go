package core

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/stretchr/testify/mock"
)

type CardanoTxsReceiverMock struct {
	mock.Mock
	NewUnprocessedTxsFn func(originChainId string, txs []*indexer.Tx) error
}

var _ CardanoTxsReceiver = (*CardanoTxsReceiverMock)(nil)

// NewUnprocessedTxs implements CardanoTxsProcessor.
func (m *CardanoTxsReceiverMock) NewUnprocessedTxs(originChainID string, txs []*indexer.Tx) error {
	if m.NewUnprocessedTxsFn != nil {
		return m.NewUnprocessedTxsFn(originChainID, txs)
	}

	args := m.Called(originChainID, txs)

	return args.Error(0)
}

type TxsProcessorMock struct {
	mock.Mock
}

// Start implements CardanoTxsProcessor.
func (m *TxsProcessorMock) Start() {
}

var _ TxsProcessor = (*TxsProcessorMock)(nil)

type BridgeDataFetcherMock struct {
	mock.Mock
}

func (m *BridgeDataFetcherMock) FetchLatestBlockPoint(chainID string) (*indexer.BlockPoint, error) {
	args := m.Called(chainID)
	if args.Get(0) != nil {
		arg0, _ := args.Get(0).(*indexer.BlockPoint)

		return arg0, args.Error(1)
	}

	return nil, args.Error(1)
}

// FetchExpectedTxs implements BridgeDataFetcher.
func (m *BridgeDataFetcherMock) FetchExpectedTx(chainID string) (*BridgeExpectedCardanoTx, error) {
	args := m.Called(chainID)
	if args.Get(0) != nil {
		arg0, _ := args.Get(0).(*BridgeExpectedCardanoTx)

		return arg0, args.Error(1)
	}

	return nil, args.Error(1)
}

var _ BridgeDataFetcher = (*BridgeDataFetcherMock)(nil)

type ExpectedTxsFetcherMock struct {
	mock.Mock
}

// Start implements ExpectedTxsFetcher.
func (m *ExpectedTxsFetcherMock) Start() {
}

var _ ExpectedTxsFetcher = (*ExpectedTxsFetcherMock)(nil)

type CardanoTxsProcessorDBMock struct {
	mock.Mock
}

func (m *CardanoTxsProcessorDBMock) AddExpectedTxs(expectedTxs []*BridgeExpectedCardanoTx) error {
	args := m.Called(expectedTxs)

	return args.Error(0)
}

func (m *CardanoTxsProcessorDBMock) GetExpectedTxs(
	chainID string, priority uint8, threshold int,
) ([]*BridgeExpectedCardanoTx, error) {
	args := m.Called(chainID, priority, threshold)
	if args.Get(0) != nil {
		arg0, _ := args.Get(0).([]*BridgeExpectedCardanoTx)

		return arg0, args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *CardanoTxsProcessorDBMock) GetAllExpectedTxs(
	chainID string, threshold int,
) ([]*BridgeExpectedCardanoTx, error) {
	args := m.Called(chainID, threshold)
	if args.Get(0) != nil {
		arg0, _ := args.Get(0).([]*BridgeExpectedCardanoTx)

		return arg0, args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *CardanoTxsProcessorDBMock) ClearExpectedTxs(chainID string) error {
	args := m.Called(chainID)

	return args.Error(0)
}

func (m *CardanoTxsProcessorDBMock) MarkExpectedTxsAsProcessed(expectedTxs []*BridgeExpectedCardanoTx) error {
	args := m.Called(expectedTxs)

	return args.Error(0)
}

// AddProcessedTxs implements CardanoTxsProcessorDB.
func (m *CardanoTxsProcessorDBMock) AddProcessedTxs(processedTxs []*ProcessedCardanoTx) error {
	args := m.Called(processedTxs)

	return args.Error(0)
}

func (m *CardanoTxsProcessorDBMock) MarkExpectedTxsAsInvalid(expectedTxs []*BridgeExpectedCardanoTx) error {
	args := m.Called(expectedTxs)

	return args.Error(0)
}

func (m *CardanoTxsProcessorDBMock) AddUnprocessedTxs(unprocessedTxs []*CardanoTx) error {
	args := m.Called(unprocessedTxs)

	return args.Error(0)
}

func (m *CardanoTxsProcessorDBMock) GetUnprocessedTxs(
	chainID string, priority uint8, threshold int) (
	[]*CardanoTx, error,
) {
	args := m.Called(chainID, priority, threshold)
	if args.Get(0) != nil {
		arg0, _ := args.Get(0).([]*CardanoTx)

		return arg0, args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *CardanoTxsProcessorDBMock) GetAllUnprocessedTxs(chainID string, threshold int) ([]*CardanoTx, error) {
	args := m.Called(chainID, threshold)
	if args.Get(0) != nil {
		arg0, _ := args.Get(0).([]*CardanoTx)

		return arg0, args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *CardanoTxsProcessorDBMock) ClearUnprocessedTxs(chainID string) error {
	args := m.Called(chainID)

	return args.Error(0)
}

func (m *CardanoTxsProcessorDBMock) MarkUnprocessedTxsAsProcessed(processedTxs []*ProcessedCardanoTx) error {
	args := m.Called(processedTxs)

	return args.Error(0)
}

func (m *CardanoTxsProcessorDBMock) GetProcessedTx(
	chainID string, txHash indexer.Hash,
) (*ProcessedCardanoTx, error) {
	args := m.Called(chainID, txHash)
	if args.Get(0) != nil {
		arg0, _ := args.Get(0).(*ProcessedCardanoTx)

		return arg0, args.Error(1)
	}

	return nil, args.Error(1)
}

var _ CardanoTxsProcessorDB = (*CardanoTxsProcessorDBMock)(nil)

type BridgeSubmitterMock struct {
	mock.Mock
	OnSubmitClaims          func(claims *BridgeClaims)
	OnSubmitConfirmedBlocks func(chainID string, blocks []*indexer.CardanoBlock)
}

// SubmitClaims implements BridgeSubmitter.
func (m *BridgeSubmitterMock) SubmitClaims(claims *BridgeClaims, submitOpts *eth.SubmitOpts) error {
	if m.OnSubmitClaims != nil {
		m.OnSubmitClaims(claims)
	}

	args := m.Called(claims, submitOpts)

	return args.Error(0)
}

// SubmitConfirmedBlocks implements BridgeSubmitter.
func (m *BridgeSubmitterMock) SubmitConfirmedBlocks(chainID string, blocks []*indexer.CardanoBlock) error {
	if m.OnSubmitConfirmedBlocks != nil {
		m.OnSubmitConfirmedBlocks(chainID, blocks)
	}

	args := m.Called(chainID, blocks)

	return args.Error(0)
}

// Dispose implements BridgeSubmitter.
func (m *BridgeSubmitterMock) Dispose() error {
	args := m.Called()

	return args.Error(0)
}

var _ BridgeSubmitter = (*BridgeSubmitterMock)(nil)

type CardanoTxSuccessProcessorMock struct {
	mock.Mock
	ShouldAddClaim bool
	Type           common.BridgingTxType
}

// GetType implements CardanoTxProcessor.
func (m *CardanoTxSuccessProcessorMock) GetType() common.BridgingTxType {
	if m.Type != "" {
		return m.Type
	}

	return "unspecified"
}

// ValidateAndAddClaim implements CardanoTxProcessor.
func (m *CardanoTxSuccessProcessorMock) ValidateAndAddClaim(
	claims *BridgeClaims, tx *CardanoTx, appConfig *AppConfig) error {
	if m.ShouldAddClaim {
		claims.BridgingRequestClaims = append(claims.BridgingRequestClaims, BridgingRequestClaim{})
	}

	args := m.Called(claims, tx, appConfig)

	return args.Error(0)
}

var _ CardanoTxSuccessProcessor = (*CardanoTxSuccessProcessorMock)(nil)

type CardanoTxFailedProcessorMock struct {
	mock.Mock
	ShouldAddClaim bool
	Type           common.BridgingTxType
}

// GetType implements CardanoTxProcessor.
func (m *CardanoTxFailedProcessorMock) GetType() common.BridgingTxType {
	if m.Type != "" {
		return m.Type
	}

	return "unspecified"
}

// ValidateAndAddClaim implements CardanoTxFailedProcessor.
func (m *CardanoTxFailedProcessorMock) ValidateAndAddClaim(
	claims *BridgeClaims, tx *BridgeExpectedCardanoTx, appConfig *AppConfig,
) error {
	if m.ShouldAddClaim {
		claims.BatchExecutionFailedClaims = append(
			claims.BatchExecutionFailedClaims, BatchExecutionFailedClaim{BatchNonceId: 1})
	}

	args := m.Called(claims, tx, appConfig)

	return args.Error(0)
}

var _ CardanoTxFailedProcessor = (*CardanoTxFailedProcessorMock)(nil)
