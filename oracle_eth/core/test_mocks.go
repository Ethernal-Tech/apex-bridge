package core

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/ethgo"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/mock"
)

type EthTxsReceiverMock struct {
	mock.Mock
	NewUnprocessedLogFn func(originChainId string, log *ethgo.Log) error
}

func (m *EthTxsReceiverMock) NewUnprocessedLog(originChainID string, log *ethgo.Log) error {
	if m.NewUnprocessedLogFn != nil {
		return m.NewUnprocessedLogFn(originChainID, log)
	}

	args := m.Called(originChainID, log)

	return args.Error(0)
}

var _ EthTxsReceiver = (*EthTxsReceiverMock)(nil)

type BridgeDataFetcherMock struct {
	mock.Mock
}

// FetchExpectedTxs implements BridgeDataFetcher.
func (m *BridgeDataFetcherMock) FetchExpectedTx(chainID string) (*BridgeExpectedEthTx, error) {
	args := m.Called(chainID)
	if args.Get(0) != nil {
		arg0, _ := args.Get(0).(*BridgeExpectedEthTx)

		return arg0, args.Error(1)
	}

	return nil, args.Error(1)
}

var _ EthBridgeDataFetcher = (*BridgeDataFetcherMock)(nil)

type EthTxsProcessorDBMock struct {
	mock.Mock
}

func (m *EthTxsProcessorDBMock) AddExpectedTxs(expectedTxs []*BridgeExpectedEthTx) error {
	args := m.Called(expectedTxs)

	return args.Error(0)
}

func (m *EthTxsProcessorDBMock) GetExpectedTxs(
	chainID string, priority uint8, threshold int,
) ([]*BridgeExpectedEthTx, error) {
	args := m.Called(chainID, priority, threshold)
	if args.Get(0) != nil {
		arg0, _ := args.Get(0).([]*BridgeExpectedEthTx)

		return arg0, args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *EthTxsProcessorDBMock) GetAllExpectedTxs(
	chainID string, threshold int,
) ([]*BridgeExpectedEthTx, error) {
	args := m.Called(chainID, threshold)
	if args.Get(0) != nil {
		arg0, _ := args.Get(0).([]*BridgeExpectedEthTx)

		return arg0, args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *EthTxsProcessorDBMock) GetUnprocessedTxs(
	chainID string, priority uint8, threshold int) (
	[]*EthTx, error,
) {
	args := m.Called(chainID, priority, threshold)
	if args.Get(0) != nil {
		arg0, _ := args.Get(0).([]*EthTx)

		return arg0, args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *EthTxsProcessorDBMock) GetAllUnprocessedTxs(chainID string, threshold int) ([]*EthTx, error) {
	args := m.Called(chainID, threshold)
	if args.Get(0) != nil {
		arg0, _ := args.Get(0).([]*EthTx)

		return arg0, args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *EthTxsProcessorDBMock) GetPendingTxs(keys [][]byte) ([]*EthTx, error) {
	args := m.Called(keys)
	if args.Get(0) != nil {
		arg0, _ := args.Get(0).([]*EthTx)

		return arg0, args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *EthTxsProcessorDBMock) ClearAllTxs(chainID string) error {
	args := m.Called(chainID)

	return args.Error(0)
}

func (m *EthTxsProcessorDBMock) AddTxs(processedTxs []*ProcessedEthTx, unprocessedTxs []*EthTx) error {
	args := m.Called(processedTxs, unprocessedTxs)

	return args.Error(0)
}

func (m *EthTxsProcessorDBMock) UpdateTxs(data *EthUpdateTxsData) error {
	args := m.Called(data)

	return args.Error(0)
}

func (m *EthTxsProcessorDBMock) GetProcessedTxByInnerActionTxHash(
	chainID string, innerActionTxHash ethgo.Hash,
) (*ProcessedEthTx, error) {
	args := m.Called(chainID, innerActionTxHash)
	if args.Get(0) != nil {
		arg0, _ := args.Get(0).(*ProcessedEthTx)

		return arg0, args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *EthTxsProcessorDBMock) GetProcessedTx(
	chainID string, txHash ethgo.Hash,
) (*ProcessedEthTx, error) {
	args := m.Called(chainID, txHash)
	if args.Get(0) != nil {
		arg0, _ := args.Get(0).(*ProcessedEthTx)

		return arg0, args.Error(1)
	}

	return nil, args.Error(1)
}

var _ EthTxsProcessorDB = (*EthTxsProcessorDBMock)(nil)

type EthTxSuccessProcessorMock struct {
	mock.Mock
	ShouldAddClaim   bool
	Type             common.BridgingTxType
	ValidateError    error
	AddClaimCallback func(claims *oCore.BridgeClaims)
}

func (m *EthTxSuccessProcessorMock) GetType() common.BridgingTxType {
	if m.Type != "" {
		return m.Type
	}

	return "unspecified"
}

func (m *EthTxSuccessProcessorMock) PreValidate(tx *EthTx, appConfig *oCore.AppConfig) error {
	return m.ValidateError
}

func (m *EthTxSuccessProcessorMock) ValidateAndAddClaim(
	claims *oCore.BridgeClaims, tx *EthTx, appConfig *oCore.AppConfig) error {
	if m.AddClaimCallback != nil {
		m.AddClaimCallback(claims)
	} else if m.ShouldAddClaim {
		claims.BridgingRequestClaims = append(claims.BridgingRequestClaims, oCore.BridgingRequestClaim{})
	}

	args := m.Called(claims, tx, appConfig)

	return args.Error(0)
}

var _ EthTxSuccessProcessor = (*EthTxSuccessProcessorMock)(nil)

type EthTxFailedProcessorMock struct {
	mock.Mock
	ShouldAddClaim bool
	Type           common.BridgingTxType
	ValidateError  error
}

func (m *EthTxFailedProcessorMock) GetType() common.BridgingTxType {
	if m.Type != "" {
		return m.Type
	}

	return "unspecified"
}

func (m *EthTxFailedProcessorMock) PreValidate(tx *BridgeExpectedEthTx, appConfig *oCore.AppConfig) error {
	return m.ValidateError
}

func (m *EthTxFailedProcessorMock) ValidateAndAddClaim(
	claims *oCore.BridgeClaims, tx *BridgeExpectedEthTx, appConfig *oCore.AppConfig,
) error {
	if m.ShouldAddClaim {
		claims.BatchExecutionFailedClaims = append(
			claims.BatchExecutionFailedClaims, oCore.BatchExecutionFailedClaim{BatchNonceId: 1})
	}

	args := m.Called(claims, tx, appConfig)

	return args.Error(0)
}

var _ EthTxFailedProcessor = (*EthTxFailedProcessorMock)(nil)

type BridgeSubmitterMock struct {
	mock.Mock
	OnSubmitClaims          func(claims *oCore.BridgeClaims)
	OnSubmitConfirmedBlocks func(chainID string, from uint64, to uint64)
}

// SubmitClaims implements BridgeSubmitter.
func (m *BridgeSubmitterMock) SubmitClaims(
	claims *oCore.BridgeClaims, submitOpts *eth.SubmitOpts) (*types.Receipt, error) {
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
func (m *BridgeSubmitterMock) SubmitConfirmedBlocks(chainID string, from uint64, to uint64) error {
	if m.OnSubmitConfirmedBlocks != nil {
		m.OnSubmitConfirmedBlocks(chainID, from, to)
	}

	args := m.Called(chainID, from, to)

	return args.Error(0)
}

// Dispose implements BridgeSubmitter.
func (m *BridgeSubmitterMock) Dispose() error {
	args := m.Called()

	return args.Error(0)
}

var _ BridgeSubmitter = (*BridgeSubmitterMock)(nil)

type EventStoreMock struct {
	mock.Mock
}

func (m *EventStoreMock) GetLastProcessedBlock() (uint64, error) {
	args := m.Called()

	//nolint:forcetypeassert
	return args.Get(0).(uint64), args.Error(1)
}

func (m *EventStoreMock) GetAllLogs() ([]*ethgo.Log, error) {
	args := m.Called()

	//nolint:forcetypeassert
	return args.Get(0).([]*ethgo.Log), args.Error(1)
}

func (m *EventStoreMock) GetLog(blockNumber, logIndex uint64) (*ethgo.Log, error) {
	args := m.Called(blockNumber, logIndex)

	//nolint:forcetypeassert
	return args.Get(0).(*ethgo.Log), args.Error(1)
}

func (m *EventStoreMock) GetLogsByBlockNumber(blockNumber uint64) ([]*ethgo.Log, error) {
	args := m.Called(blockNumber)

	//nolint:forcetypeassert
	return args.Get(0).([]*ethgo.Log), args.Error(1)
}

func (m *EventStoreMock) InsertLastProcessedBlock(blockNumber uint64) error {
	args := m.Called(blockNumber)

	return args.Error(0)
}

func (m *EventStoreMock) InsertLogs(logs []*ethgo.Log) error {
	args := m.Called(logs)

	return args.Error(0)
}
