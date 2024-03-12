package core

import (
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
)

type BridgeExpectedCardanoTxsDb interface {
	AddExpectedTxs(expectedTxs []*BridgeExpectedCardanoTx) error
	GetExpectedTxs(threshold int) ([]*BridgeExpectedCardanoTx, error)
	MarkExpectedTxsAsProcessed(expectedTxs []*BridgeExpectedCardanoTx) error
	MarkExpectedTxsAsInvalid(expectedTxs []*BridgeExpectedCardanoTx) error
}

type UnprocessedCardanoTxsDb interface {
	AddUnprocessedTxs(unprocessedTxs []*CardanoTx) error
	GetUnprocessedTxs(threshold int) ([]*CardanoTx, error)
	MarkUnprocessedTxsAsProcessed(processedTxs []*CardanoTx) error
}

type CardanoTxsProcessorDb interface {
	UnprocessedCardanoTxsDb
	BridgeExpectedCardanoTxsDb

	AddInvalidTxHashes(invalidTxHashes []string) error
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
	GetConfig() CardanoChainConfig
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

type BridgeDataFetcher interface {
	Start() error
	Stop() error
}

type ClaimsSubmitter interface {
	SubmitClaims(claims *BridgeClaims) error
	Dispose() error
}
