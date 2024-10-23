package core

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
)

type BridgeExpectedCardanoTxsDB interface {
	GetExpectedTxs(chainID string, priority uint8, threshold int) ([]*BridgeExpectedCardanoTx, error)
	GetAllExpectedTxs(chainID string, threshold int) ([]*BridgeExpectedCardanoTx, error)
	AddExpectedTxs(expectedTxs []*BridgeExpectedCardanoTx) error
}

type CardanoTxsDB interface {
	GetUnprocessedTxs(chainID string, priority uint8, threshold int) ([]*CardanoTx, error)
	GetAllUnprocessedTxs(chainID string, threshold int) ([]*CardanoTx, error)
	GetPendingTxs(keys [][]byte) ([]*CardanoTx, error)
	GetProcessedTx(chainID string, txHash indexer.Hash) (*ProcessedCardanoTx, error)
	AddTxs(processedTxs []*ProcessedCardanoTx, unprocessedTxs []*CardanoTx) error
	ClearAllTxs(chainID string) error
	UpdateTxs(data *CardanoUpdateTxsData) error
}

type CardanoTxsProcessorDB interface {
	CardanoTxsDB
	BridgeExpectedCardanoTxsDB
}

type Database interface {
	CardanoTxsProcessorDB
	Init(filePath string) error
	Close() error
}

type Oracle interface {
	Start() error
	Dispose() error
	ErrorCh() <-chan error
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

type BridgeDataFetcher interface {
	FetchLatestBlockPoint(chainID string) (*indexer.BlockPoint, error)
	FetchExpectedTx(chainID string) (*BridgeExpectedCardanoTx, error)
}

type BridgeSubmitter interface {
	cCore.BridgeClaimsSubmitter
	SubmitConfirmedBlocks(chainID string, blocks []*indexer.CardanoBlock) error
}
