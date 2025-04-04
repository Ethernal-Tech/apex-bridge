package core

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/ethgo"
	"go.etcd.io/bbolt"
)

type BridgeExpectedEthTxsDB interface {
	GetExpectedTxs(chainID string, priority uint8, threshold int) ([]*BridgeExpectedEthTx, error)
	GetAllExpectedTxs(chainID string, threshold int) ([]*BridgeExpectedEthTx, error)
	AddExpectedTxs(expectedTxs []*BridgeExpectedEthTx) error
}

type EthTxsDB interface {
	GetUnprocessedTxs(chainID string, priority uint8, threshold int) ([]*EthTx, error)
	GetAllUnprocessedTxs(chainID string, threshold int) ([]*EthTx, error)
	GetPendingTx(entityID oCore.DBTxID) (oCore.BaseTx, error)
	GetProcessedTx(entityID oCore.DBTxID) (*ProcessedEthTx, error)
	GetProcessedTxByInnerActionTxHash(chainID string, innerActionTxHash []byte) (*ProcessedEthTx, error)
	ClearAllTxs(chainID string) error
	GetUnprocessedBatchEvents(chainID string) ([]*oCore.DBBatchInfoEvent, error)
	AddTxs(processedTxs []*ProcessedEthTx, unprocessedTxs []*EthTx) error
	UpdateTxs(data *EthUpdateTxsData) error
}

type EthTxsProcessorDB interface {
	EthTxsDB
	BridgeExpectedEthTxsDB
}

type Database interface {
	EthTxsProcessorDB
	Init(db *bbolt.DB, appConfig *oCore.AppConfig, typeRegister common.TypeRegister)
}

type Oracle interface {
	Start() error
	Dispose() error
}

type EthTxsReceiver interface {
	NewUnprocessedLog(originChainID string, log *ethgo.Log) error
}

type EthTxSuccessProcessor interface {
	GetType() common.BridgingTxType
	PreValidate(tx *EthTx, appConfig *oCore.AppConfig) error
	ValidateAndAddClaim(claims *oCore.BridgeClaims, tx *EthTx, appConfig *oCore.AppConfig) error
}

type EthTxFailedProcessor interface {
	GetType() common.BridgingTxType
	PreValidate(tx *BridgeExpectedEthTx, appConfig *oCore.AppConfig) error
	ValidateAndAddClaim(claims *oCore.BridgeClaims, tx *BridgeExpectedEthTx, appConfig *oCore.AppConfig) error
}

type EthTxSuccessRefundProcessor interface {
	EthTxSuccessProcessor

	HandleBridgingProcessorError(
		claims *oCore.BridgeClaims, tx *EthTx, appConfig *oCore.AppConfig,
		err error, errContext string) error
}

type EthChainObserver interface {
	Start() error
	Dispose() error
	GetConfig() *oCore.EthChainConfig
}

type EthBridgeDataFetcher interface {
	oCore.BridgeDataFetcher
	FetchExpectedTx(chainID string) (*BridgeExpectedEthTx, error)
}

type BridgeSubmitter interface {
	oCore.BridgeClaimsSubmitter
	SubmitConfirmedBlocks(chainID string, from uint64, to uint64) error
}

type EthConfirmedBlocksSubmitter interface {
	StartSubmit()
	GetChainID() string
}
