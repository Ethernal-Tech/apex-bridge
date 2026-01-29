package core

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"go.etcd.io/bbolt"
)

type BridgeExpectedCardanoTxsDB interface {
	GetExpectedTxs(chainID string, priority uint8, threshold int) ([]*BridgeExpectedCardanoTx, error)
	GetAllExpectedTxs(chainID string, threshold int) ([]*BridgeExpectedCardanoTx, error)
	AddExpectedTxs(expectedTxs []*BridgeExpectedCardanoTx) error
}

type CardanoTxsDB interface {
	GetUnprocessedTxs(chainID string, priority uint8, threshold int) ([]*CardanoTx, error)
	GetAllUnprocessedTxs(chainID string, threshold int) ([]*CardanoTx, error)
	GetPendingTx(entityID cCore.DBTxID) (cCore.BaseTx, error)
	GetProcessedTx(entityID cCore.DBTxID) (*ProcessedCardanoTx, error)
	GetUnprocessedBatchEvents(chainID string) ([]*cCore.DBBatchInfoEvent, error)
	AddTxs(processedTxs []*ProcessedCardanoTx, unprocessedTxs []*CardanoTx) error
	ClearAllTxs(chainID string) error
	MoveProcessedExpectedTxs(chainID string) error
	UpdateTxs(data *CardanoUpdateTxsData, chainIDConverter *common.ChainIDConverter) error
}

type CardanoTxsProcessorDB interface {
	CardanoTxsDB
	BridgeExpectedCardanoTxsDB
	cCore.BlockSubmitterDB
}

type Database interface {
	CardanoTxsProcessorDB
	Init(db *bbolt.DB, appConfig *cCore.AppConfig, typeRegister common.TypeRegister)
}

type Oracle interface {
	Start() error
	Dispose() error
}

type CardanoChainObserver interface {
	Start() error
	Dispose() error
	GetConfig() *cCore.CardanoChainConfig
	ErrorCh() <-chan error
}

type CardanoTxsReceiver interface {
	NewUnprocessedTxs(originChainID string, txs []*indexer.Tx) error
}

type CardanoTxSuccessProcessor interface {
	GetType() common.BridgingTxType
	PreValidate(tx *CardanoTx, appConfig *cCore.AppConfig) error
	ValidateAndAddClaim(claims *cCore.BridgeClaims, tx *CardanoTx, appConfig *cCore.AppConfig) error
}

type CardanoTxFailedProcessor interface {
	GetType() common.BridgingTxType
	PreValidate(tx *BridgeExpectedCardanoTx, appConfig *cCore.AppConfig) error
	ValidateAndAddClaim(claims *cCore.BridgeClaims, tx *BridgeExpectedCardanoTx, appConfig *cCore.AppConfig) error
}

type CardanoTxSuccessRefundProcessor interface {
	CardanoTxSuccessProcessor

	HandleBridgingProcessorError(
		claims *cCore.BridgeClaims, tx *CardanoTx, appConfig *cCore.AppConfig,
		err error, errContext string) error

	HandleBridgingProcessorPreValidate(
		tx *CardanoTx, appConfig *cCore.AppConfig) error
}

type CardanoBridgeDataFetcher interface {
	cCore.BridgeDataFetcher
	FetchLatestBlockPoint(chainID string) (*indexer.BlockPoint, error)
	FetchExpectedTx(chainID string) (*BridgeExpectedCardanoTx, error)
}
