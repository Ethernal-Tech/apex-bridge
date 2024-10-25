package core

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/ethereum/go-ethereum/core/types"
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

func (m *CardanoTxsProcessorDBMock) GetPendingTxs(keys [][]byte) ([]*CardanoTx, error) {
	args := m.Called(keys)
	if args.Get(0) != nil {
		arg0, _ := args.Get(0).([]*CardanoTx)

		return arg0, args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *CardanoTxsProcessorDBMock) ClearAllTxs(chainID string) error {
	args := m.Called(chainID)

	return args.Error(0)
}

func (m *CardanoTxsProcessorDBMock) AddTxs(processedTxs []*ProcessedCardanoTx, unprocessedTxs []*CardanoTx) error {
	args := m.Called(processedTxs, unprocessedTxs)

	return args.Error(0)
}

func (m *CardanoTxsProcessorDBMock) UpdateTxs(data *CardanoUpdateTxsData) error {
	args := m.Called(data)

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
	OnSubmitClaims          func(claims *cCore.BridgeClaims)
	OnSubmitConfirmedBlocks func(chainID string, blocks []*indexer.CardanoBlock)
}

// SubmitClaims implements BridgeSubmitter.
func (m *BridgeSubmitterMock) SubmitClaims(
	claims *cCore.BridgeClaims, submitOpts *eth.SubmitOpts) (*types.Receipt, error) {
	if m.OnSubmitClaims != nil {
		m.OnSubmitClaims(claims)
	}

	args := m.Called(claims, submitOpts)
	if args.Get(0) != nil {
		arg0, _ := args.Get(0).(*types.Receipt)

		return arg0, args.Error(1)
	}

	return nil, args.Error(1)
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
	ShouldAddClaim   bool
	Type             common.BridgingTxType
	ValidateError    error
	AddClaimCallback func(claims *cCore.BridgeClaims)
}

// GetType implements CardanoTxProcessor.
func (m *CardanoTxSuccessProcessorMock) GetType() common.BridgingTxType {
	if m.Type != "" {
		return m.Type
	}

	return "unspecified"
}

func (m *CardanoTxSuccessProcessorMock) PreValidate(tx *CardanoTx, appConfig *cCore.AppConfig) error {
	return m.ValidateError
}

// ValidateAndAddClaim implements CardanoTxProcessor.
func (m *CardanoTxSuccessProcessorMock) ValidateAndAddClaim(
	claims *cCore.BridgeClaims, tx *CardanoTx, appConfig *cCore.AppConfig) error {
	if m.AddClaimCallback != nil {
		m.AddClaimCallback(claims)
	} else if m.ShouldAddClaim {
		claims.BridgingRequestClaims = append(claims.BridgingRequestClaims, cCore.BridgingRequestClaim{})
	}

	args := m.Called(claims, tx, appConfig)

	return args.Error(0)
}

var _ CardanoTxSuccessProcessor = (*CardanoTxSuccessProcessorMock)(nil)

type CardanoTxFailedProcessorMock struct {
	mock.Mock
	ShouldAddClaim bool
	Type           common.BridgingTxType
	ValidateError  error
}

// GetType implements CardanoTxProcessor.
func (m *CardanoTxFailedProcessorMock) GetType() common.BridgingTxType {
	if m.Type != "" {
		return m.Type
	}

	return "unspecified"
}

func (m *CardanoTxFailedProcessorMock) PreValidate(tx *BridgeExpectedCardanoTx, appConfig *cCore.AppConfig) error {
	return m.ValidateError
}

// ValidateAndAddClaim implements CardanoTxFailedProcessor.
func (m *CardanoTxFailedProcessorMock) ValidateAndAddClaim(
	claims *cCore.BridgeClaims, tx *BridgeExpectedCardanoTx, appConfig *cCore.AppConfig,
) error {
	if m.ShouldAddClaim {
		claims.BatchExecutionFailedClaims = append(
			claims.BatchExecutionFailedClaims, cCore.BatchExecutionFailedClaim{BatchNonceId: 1})
	}

	args := m.Called(claims, tx, appConfig)

	return args.Error(0)
}

var _ CardanoTxFailedProcessor = (*CardanoTxFailedProcessorMock)(nil)
