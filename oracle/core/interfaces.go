package core

import (
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
)

type BridgeExpectedCardanoTxsDb interface {
	AddExpectedTxs(expectedTxs []*BridgeExpectedCardanoTx) error
	GetExpectedTxs(chainId string, threshold int) ([]*BridgeExpectedCardanoTx, error)
	ClearExpectedTxs(chainId string) error
	MarkExpectedTxsAsProcessed(expectedTxs []*BridgeExpectedCardanoTx) error
	MarkExpectedTxsAsInvalid(expectedTxs []*BridgeExpectedCardanoTx) error
}

type CardanoTxsDb interface {
	AddUnprocessedTxs(unprocessedTxs []*CardanoTx) error
	GetUnprocessedTxs(chainId string, threshold int) ([]*CardanoTx, error)
	ClearUnprocessedTxs(chainId string) error
	MarkUnprocessedTxsAsProcessed(processedTxs []*ProcessedCardanoTx) error
	GetProcessedTx(chainId string, txHash string) (*ProcessedCardanoTx, error)
}

type CardanoTxsProcessorDb interface {
	CardanoTxsDb
	BridgeExpectedCardanoTxsDb
}

type Database interface {
	CardanoTxsProcessorDb
	Init(filePath string) error
	Close() error
}

type Oracle interface {
	Start() error
	Stop() error
	ErrorCh() <-chan error
}

type CardanoChainObserver interface {
	Start() error
	Stop() error
	GetConfig() *CardanoChainConfig
	GetDb() indexer.Database
	ErrorCh() <-chan error
}

type CardanoTxsProcessor interface {
	NewUnprocessedTxs(originChainId string, txs []*indexer.Tx) error
	Start() error
	Stop() error
}

type CardanoTxProcessor interface {
	IsTxRelevant(tx *CardanoTx, appConfig *AppConfig) (bool, error)
	ValidateAndAddClaim(claims *BridgeClaims, tx *CardanoTx, appConfig *AppConfig) error
}

type CardanoTxFailedProcessor interface {
	IsTxRelevant(tx *BridgeExpectedCardanoTx, appConfig *AppConfig) (bool, error)
	ValidateAndAddClaim(claims *BridgeClaims, tx *BridgeExpectedCardanoTx, appConfig *AppConfig) error
}

type ExpectedTxsFetcher interface {
	Start() error
	Stop() error
}

type BridgeDataFetcher interface {
	FetchLatestBlockPoint(chainId string) (*indexer.BlockPoint, error)
	FetchExpectedTx(chainId string) (*BridgeExpectedCardanoTx, error)
	Dispose() error
}

type BridgeSubmitter interface {
	SubmitClaims(claims *BridgeClaims) error
	SubmitConfirmedBlocks(chainId string, blocks []*indexer.CardanoBlock) error
	Dispose() error
}

type ConfirmedBlocksSubmitter interface {
	StartSubmit()
	Dispose() error
	GetChainId() string
	ErrorCh() <-chan error
}
