package core

import (
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
)

type BridgeExpectedCardanoTxsDB interface {
	AddExpectedTxs(expectedTxs []*BridgeExpectedCardanoTx) error
	GetExpectedTxs(chainID string, threshold int) ([]*BridgeExpectedCardanoTx, error)
	ClearExpectedTxs(chainID string) error
	MarkExpectedTxsAsProcessed(expectedTxs []*BridgeExpectedCardanoTx) error
	MarkExpectedTxsAsInvalid(expectedTxs []*BridgeExpectedCardanoTx) error
}

type CardanoTxsDB interface {
	AddUnprocessedTxs(unprocessedTxs []*CardanoTx) error
	GetUnprocessedTxs(chainID string, threshold int) ([]*CardanoTx, error)
	ClearUnprocessedTxs(chainID string) error
	MarkUnprocessedTxsAsProcessed(processedTxs []*ProcessedCardanoTx) error
	GetProcessedTx(chainID string, txHash string) (*ProcessedCardanoTx, error)
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
	GetConfig() *CardanoChainConfig
	ErrorCh() <-chan error
}

type CardanoTxsProcessor interface {
	NewUnprocessedTxs(originChainID string, txs []*indexer.Tx) error
	Start()
}

type CardanoTxProcessor interface {
	GetType() TxProcessorType
	IsTxRelevant(tx *CardanoTx) (bool, error)
	ValidateAndAddClaim(claims *BridgeClaims, tx *CardanoTx, appConfig *AppConfig) error
}

type CardanoTxFailedProcessor interface {
	GetType() TxProcessorType
	IsTxRelevant(tx *BridgeExpectedCardanoTx) (bool, error)
	ValidateAndAddClaim(claims *BridgeClaims, tx *BridgeExpectedCardanoTx, appConfig *AppConfig) error
}

type ExpectedTxsFetcher interface {
	Start()
}

type BridgeDataFetcher interface {
	FetchLatestBlockPoint(chainID string) (*indexer.BlockPoint, error)
	FetchExpectedTx(chainID string) (*BridgeExpectedCardanoTx, error)
}

type BridgeSubmitter interface {
	SubmitClaims(claims *BridgeClaims) error
	SubmitConfirmedBlocks(chainID string, blocks []*indexer.CardanoBlock) error
}

type ConfirmedBlocksSubmitter interface {
	StartSubmit()
	GetChainID() string
}
