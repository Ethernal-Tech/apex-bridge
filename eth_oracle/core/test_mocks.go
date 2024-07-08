package core

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	oracleCore "github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/ethgo"
	"github.com/stretchr/testify/mock"
)

type EthTxsProcessorMock struct {
	mock.Mock
	NewUnprocessedLogFn func(originChainId string, log *ethgo.Log) error
}

// NewUnprocessedTxs implements CardanoTxsProcessor.
func (m *EthTxsProcessorMock) NewUnprocessedLog(originChainID string, log *ethgo.Log) error {
	if m.NewUnprocessedLogFn != nil {
		return m.NewUnprocessedLogFn(originChainID, log)
	}

	args := m.Called(originChainID, log)

	return args.Error(0)
}

// Start implements CardanoTxsProcessor.
func (m *EthTxsProcessorMock) Start() {
}

var _ EthTxsProcessor = (*EthTxsProcessorMock)(nil)

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

func (m *EthTxsProcessorDBMock) ClearExpectedTxs(chainID string) error {
	args := m.Called(chainID)

	return args.Error(0)
}

func (m *EthTxsProcessorDBMock) MarkExpectedTxsAsProcessed(expectedTxs []*BridgeExpectedEthTx) error {
	args := m.Called(expectedTxs)

	return args.Error(0)
}

func (m *EthTxsProcessorDBMock) AddProcessedTxs(processedTxs []*ProcessedEthTx) error {
	args := m.Called(processedTxs)

	return args.Error(0)
}

func (m *EthTxsProcessorDBMock) MarkExpectedTxsAsInvalid(expectedTxs []*BridgeExpectedEthTx) error {
	args := m.Called(expectedTxs)

	return args.Error(0)
}

func (m *EthTxsProcessorDBMock) AddUnprocessedTxs(unprocessedTxs []*EthTx) error {
	args := m.Called(unprocessedTxs)

	return args.Error(0)
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

func (m *EthTxsProcessorDBMock) ClearUnprocessedTxs(chainID string) error {
	args := m.Called(chainID)

	return args.Error(0)
}

func (m *EthTxsProcessorDBMock) MarkUnprocessedTxsAsProcessed(processedTxs []*ProcessedEthTx) error {
	args := m.Called(processedTxs)

	return args.Error(0)
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

type EthTxProcessorMock struct {
	mock.Mock
	ShouldAddClaim bool
	Type           common.BridgingTxType
}

func (m *EthTxProcessorMock) GetType() common.BridgingTxType {
	if m.Type != "" {
		return m.Type
	}

	return "unspecified"
}

func (m *EthTxProcessorMock) ValidateAndAddClaim(
	claims *oracleCore.BridgeClaims, tx *EthTx, appConfig *oracleCore.AppConfig) error {
	if m.ShouldAddClaim {
		claims.BridgingRequestClaims = append(claims.BridgingRequestClaims, oracleCore.BridgingRequestClaim{})
	}

	args := m.Called(claims, tx, appConfig)

	return args.Error(0)
}

var _ EthTxProcessor = (*EthTxProcessorMock)(nil)

type EthTxFailedProcessorMock struct {
	mock.Mock
	ShouldAddClaim bool
	Type           common.BridgingTxType
}

func (m *EthTxFailedProcessorMock) GetType() common.BridgingTxType {
	if m.Type != "" {
		return m.Type
	}

	return "unspecified"
}

func (m *EthTxFailedProcessorMock) ValidateAndAddClaim(
	claims *oracleCore.BridgeClaims, tx *BridgeExpectedEthTx, appConfig *oracleCore.AppConfig,
) error {
	if m.ShouldAddClaim {
		claims.BatchExecutionFailedClaims = append(
			claims.BatchExecutionFailedClaims, oracleCore.BatchExecutionFailedClaim{BatchNonceId: 1})
	}

	args := m.Called(claims, tx, appConfig)

	return args.Error(0)
}

var _ EthTxFailedProcessor = (*EthTxFailedProcessorMock)(nil)
