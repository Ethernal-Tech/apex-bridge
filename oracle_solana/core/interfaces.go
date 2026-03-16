package core

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/solana-infrastructure/tracker"
	"go.etcd.io/bbolt"
)

type SolanaChainObserver interface {
	Start() error
	Dispose() error
	GetConfig() *oCore.SolanaChainConfig
}

type Oracle interface {
	Start() error
	Dispose() error
}

type SolanaTxsDB interface {
	GetUnprocessedTxs(chainID string, priority uint8, threshold int) ([]*SolanaTx, error)
	GetAllUnprocessedTxs(chainID string, threshold int) ([]*SolanaTx, error)
	GetPendingTx(entityID oCore.DBTxID) (oCore.BaseTx, error)
	GetProcessedTx(entityID oCore.DBTxID) (*ProcessedSolanaTx, error)
	GetProcessedTxByInnerActionTxHash(chainID string, innerActionTxHash []byte) (*ProcessedSolanaTx, error)
	ClearAllTxs(chainID string) error
	MoveProcessedExpectedTxs(chainID string) error
	GetUnprocessedBatchEvents(chainID string) ([]*oCore.DBBatchInfoEvent, error)
	AddTxs(processedTxs []*ProcessedSolanaTx, unprocessedTxs []*SolanaTx) error
	UpdateTxs(data *SolanaUpdateTxsData, chainIDConverter *common.ChainIDConverter) error
}

type SolanaTxsProcessorDB interface {
	SolanaTxsDB
	BridgeExpectedSolanaTxsDB
	oCore.BlockSubmitterDB
}

type Database interface {
	SolanaTxsProcessorDB
	Init(db *bbolt.DB, appConfig *oCore.AppConfig, typeRegister common.TypeRegister)
}

type SolanaTxsReceiver interface {
	NewUnprocessedEvent(originChainID string, event tracker.EventNotification) error
}

type SolanaTxsProcessor interface {
}

type SolanaBridgeDataFetcher interface {
	oCore.BridgeDataFetcher
	FetchExpectedTx(chainID string) (*BridgeExpectedSolanaTx, error)
}

type BridgeExpectedSolanaTxsDB interface {
	GetExpectedTxs(chainID string, priority uint8, threshold int) ([]*BridgeExpectedSolanaTx, error)
	GetAllExpectedTxs(chainID string, threshold int) ([]*BridgeExpectedSolanaTx, error)
	AddExpectedTxs(expectedTxs []*BridgeExpectedSolanaTx) error
}

type SolanaTxSuccessProcessor interface {
	GetType() common.BridgingTxType
	PreValidate(tx *SolanaTx, appConfig *oCore.AppConfig) error
	ValidateAndAddClaim(claims *oCore.BridgeClaims, tx *SolanaTx, appConfig *oCore.AppConfig) error
}

type SolanaTxFailedProcessor interface {
	GetType() common.BridgingTxType
	PreValidate(tx *BridgeExpectedSolanaTx, appConfig *oCore.AppConfig) error
	ValidateAndAddClaim(claims *oCore.BridgeClaims, tx *BridgeExpectedSolanaTx, appConfig *oCore.AppConfig) error
}
