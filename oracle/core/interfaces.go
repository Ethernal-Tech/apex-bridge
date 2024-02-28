package core

import (
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
)

type CardanoBlockProcessorDb interface {
	AddUnprocessedTxs(unprocessedTxs []*CardanoTx) error
	GetUnprocessedTxs(threshold int) ([]*CardanoTx, error)
	MarkTxsAsProcessed(processedTxs []*CardanoTx) error
}

type Database interface {
	CardanoBlockProcessorDb
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
	ErrorCh() <-chan error
}

type CardanoBlockProcessor interface {
	NewUnprocessedTxs(originChainId string, txs []*indexer.Tx) error
	Start() error
	Stop() error
}

type CardanoTxProcessor interface {
	IsTxRelevant(tx *CardanoTx, appConfig *AppConfig) (bool, error)
	ValidateAndAddClaim(claims *BridgeClaims, tx *CardanoTx, appConfig *AppConfig) error
}

type ClaimsSubmitter interface {
	SubmitClaims(claims *BridgeClaims) error
}
