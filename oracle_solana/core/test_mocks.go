package core

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/solana-infrastructure/tracker"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/mock"
)

type SolanaTxsReceiverMock struct {
	mock.Mock
	NewUnprocessedEventFn func(originChainID string, event tracker.EventNotification) error
}

var _ SolanaTxsReceiver = (*SolanaTxsReceiverMock)(nil)

func (m *SolanaTxsReceiverMock) NewUnprocessedEvent(originChainID string, event tracker.EventNotification) error {
	if m.NewUnprocessedEventFn != nil {
		return m.NewUnprocessedEventFn(originChainID, event)
	}

	args := m.Called(originChainID, event)

	return args.Error(0)
}

type SolanaBridgeDataFetcherMock struct {
	mock.Mock
}

func (m *SolanaBridgeDataFetcherMock) GetBatchTransactions(
	chainID string, batchID uint64,
) ([]eth.TxDataInfo, error) {
	args := m.Called(chainID, batchID)

	return args.Get(0).([]eth.TxDataInfo), args.Error(1) //nolint
}

func (m *SolanaBridgeDataFetcherMock) FetchExpectedTx(chainID string) (*BridgeExpectedSolanaTx, error) {
	args := m.Called(chainID)
	if args.Get(0) != nil {
		arg0, _ := args.Get(0).(*BridgeExpectedSolanaTx)

		return arg0, args.Error(1)
	}

	return nil, args.Error(1)
}

var _ SolanaBridgeDataFetcher = (*SolanaBridgeDataFetcherMock)(nil)

type SolanaTxsProcessorDBMock struct {
	mock.Mock
}

func (m *SolanaTxsProcessorDBMock) GetBlocksSubmitterInfo(chainID string) (oCore.BlocksSubmitterInfo, error) {
	args := m.Called(chainID)

	return args.Get(0).(oCore.BlocksSubmitterInfo), args.Error(1) //nolint
}

func (m *SolanaTxsProcessorDBMock) SetBlocksSubmitterInfo(chainID string, info oCore.BlocksSubmitterInfo) error {
	args := m.Called(chainID, info)

	return args.Error(0)
}

func (m *SolanaTxsProcessorDBMock) GetUnprocessedBatchEvents(chainID string) ([]*oCore.DBBatchInfoEvent, error) {
	args := m.Called(chainID)
	if args.Get(0) != nil {
		return args.Get(0).([]*oCore.DBBatchInfoEvent), args.Error(1) //nolint
	}

	return nil, args.Error(1)
}

func (m *SolanaTxsProcessorDBMock) AddExpectedTxs(expectedTxs []*BridgeExpectedSolanaTx) error {
	args := m.Called(expectedTxs)

	return args.Error(0)
}

func (m *SolanaTxsProcessorDBMock) GetExpectedTxs(
	chainID string, priority uint8, threshold int,
) ([]*BridgeExpectedSolanaTx, error) {
	args := m.Called(chainID, priority, threshold)
	if args.Get(0) != nil {
		return args.Get(0).([]*BridgeExpectedSolanaTx), args.Error(1) //nolint
	}

	return nil, args.Error(1)
}

func (m *SolanaTxsProcessorDBMock) GetAllExpectedTxs(
	chainID string, threshold int,
) ([]*BridgeExpectedSolanaTx, error) {
	args := m.Called(chainID, threshold)
	if args.Get(0) != nil {
		return args.Get(0).([]*BridgeExpectedSolanaTx), args.Error(1) //nolint
	}

	return nil, args.Error(1)
}

func (m *SolanaTxsProcessorDBMock) GetUnprocessedTxs(
	chainID string, priority uint8, threshold int,
) ([]*SolanaTx, error) {
	args := m.Called(chainID, priority, threshold)
	if args.Get(0) != nil {
		return args.Get(0).([]*SolanaTx), args.Error(1) //nolint
	}

	return nil, args.Error(1)
}

func (m *SolanaTxsProcessorDBMock) GetAllUnprocessedTxs(chainID string, threshold int) ([]*SolanaTx, error) {
	args := m.Called(chainID, threshold)
	if args.Get(0) != nil {
		return args.Get(0).([]*SolanaTx), args.Error(1) //nolint
	}

	return nil, args.Error(1)
}

func (m *SolanaTxsProcessorDBMock) GetPendingTx(entityID oCore.DBTxID) (oCore.BaseTx, error) {
	args := m.Called(entityID)
	if args.Get(0) != nil {
		return args.Get(0).(oCore.BaseTx), args.Error(1) //nolint
	}

	return nil, args.Error(1)
}

func (m *SolanaTxsProcessorDBMock) GetProcessedTx(entityID oCore.DBTxID) (*ProcessedSolanaTx, error) {
	args := m.Called(entityID)
	if args.Get(0) != nil {
		return args.Get(0).(*ProcessedSolanaTx), args.Error(1) //nolint
	}

	return nil, args.Error(1)
}

func (m *SolanaTxsProcessorDBMock) GetProcessedTxByInnerActionTxHash(
	chainID string, innerActionTxHash []byte,
) (*ProcessedSolanaTx, error) {
	args := m.Called(chainID, innerActionTxHash)
	if args.Get(0) != nil {
		return args.Get(0).(*ProcessedSolanaTx), args.Error(1) //nolint
	}

	return nil, args.Error(1)
}

func (m *SolanaTxsProcessorDBMock) ClearAllTxs(chainID string) error {
	args := m.Called(chainID)

	return args.Error(0)
}

func (m *SolanaTxsProcessorDBMock) MoveProcessedExpectedTxs(chainID string) error {
	args := m.Called(chainID)

	return args.Error(0)
}

func (m *SolanaTxsProcessorDBMock) AddTxs(processedTxs []*ProcessedSolanaTx, unprocessedTxs []*SolanaTx) error {
	args := m.Called(processedTxs, unprocessedTxs)

	return args.Error(0)
}

func (m *SolanaTxsProcessorDBMock) UpdateTxs(
	data *SolanaUpdateTxsData, chainIDConverter *common.ChainIDConverter,
) error {
	args := m.Called(data, chainIDConverter)

	return args.Error(0)
}

var _ SolanaTxsProcessorDB = (*SolanaTxsProcessorDBMock)(nil)

type SolanaTxSuccessProcessorMock struct {
	mock.Mock
	ShouldAddClaim   bool
	Type             common.BridgingTxType
	ValidateError    error
	AddClaimCallback func(claims *oCore.BridgeClaims)
}

func (m *SolanaTxSuccessProcessorMock) GetType() common.BridgingTxType {
	if m.Type != "" {
		return m.Type
	}

	return "unspecified"
}

func (m *SolanaTxSuccessProcessorMock) PreValidate(tx *SolanaTx, appConfig *oCore.AppConfig) error {
	return m.ValidateError
}

func (m *SolanaTxSuccessProcessorMock) ValidateAndAddClaim(
	claims *oCore.BridgeClaims, tx *SolanaTx, appConfig *oCore.AppConfig,
) error {
	if m.AddClaimCallback != nil {
		m.AddClaimCallback(claims)
	} else if m.ShouldAddClaim {
		claims.BridgingRequestClaims = append(claims.BridgingRequestClaims, oCore.BridgingRequestClaim{
			SourceChainId:           appConfig.ChainIDConverter.ToChainIDNum(tx.OriginChainID),
			ObservedTransactionHash: tx.TxSignature[:],
		})
	}

	args := m.Called(claims, tx, appConfig)

	return args.Error(0)
}

var _ SolanaTxSuccessProcessor = (*SolanaTxSuccessProcessorMock)(nil)

type SolanaTxFailedProcessorMock struct {
	mock.Mock
	ShouldAddClaim bool
	Type           common.BridgingTxType
	ValidateError  error
}

func (m *SolanaTxFailedProcessorMock) GetType() common.BridgingTxType {
	if m.Type != "" {
		return m.Type
	}

	return "unspecified"
}

func (m *SolanaTxFailedProcessorMock) PreValidate(tx *BridgeExpectedSolanaTx, appConfig *oCore.AppConfig) error {
	return m.ValidateError
}

func (m *SolanaTxFailedProcessorMock) ValidateAndAddClaim(
	claims *oCore.BridgeClaims, tx *BridgeExpectedSolanaTx, appConfig *oCore.AppConfig,
) error {
	if m.ShouldAddClaim {
		claims.BatchExecutionFailedClaims = append(
			claims.BatchExecutionFailedClaims, oCore.BatchExecutionFailedClaim{
				BatchNonceId:            1,
				ChainId:                 0,
				ObservedTransactionHash: tx.Hash[:],
			})
	}

	args := m.Called(claims, tx, appConfig)

	return args.Error(0)
}

var _ SolanaTxFailedProcessor = (*SolanaTxFailedProcessorMock)(nil)

type SolanaBridgeSubmitterMock struct {
	mock.Mock
	OnSubmitClaims func(claims *oCore.BridgeClaims) (*types.Receipt, error)
}

func (m *SolanaBridgeSubmitterMock) SubmitClaims(
	claims *oCore.BridgeClaims, submitOpts *eth.SubmitOpts,
) (*types.Receipt, error) {
	if m.OnSubmitClaims != nil {
		return m.OnSubmitClaims(claims)
	}

	args := m.Called(claims, submitOpts)
	if args.Get(0) != nil {
		arg0, _ := args.Get(0).(*types.Receipt)

		return arg0, args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *SolanaBridgeSubmitterMock) SubmitBlocks(chainID string, blocks []eth.CardanoBlock) error {
	args := m.Called(chainID, blocks)

	return args.Error(0)
}

func (m *SolanaBridgeSubmitterMock) Dispose() error {
	args := m.Called()

	return args.Error(0)
}

var _ oCore.BridgeClaimsSubmitter = (*SolanaBridgeSubmitterMock)(nil)
