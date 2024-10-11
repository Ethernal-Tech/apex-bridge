package core

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
)

type BridgeExpectedCardanoTxsDB interface {
	AddExpectedTxs(expectedTxs []*BridgeExpectedCardanoTx) error
	GetExpectedTxs(chainID string, priority uint8, threshold int) ([]*BridgeExpectedCardanoTx, error)
	GetAllExpectedTxs(chainID string, threshold int) ([]*BridgeExpectedCardanoTx, error)
	ClearExpectedTxs(chainID string) error
	MarkExpectedTxsAsProcessed(expectedTxs []*BridgeExpectedCardanoTx) error
	MarkExpectedTxsAsInvalid(expectedTxs []*BridgeExpectedCardanoTx) error
}

type CardanoTxsDB interface {
	AddUnprocessedTxs(unprocessedTxs []*CardanoTx) error
	GetUnprocessedTxs(chainID string, priority uint8, threshold int) ([]*CardanoTx, error)
	GetAllUnprocessedTxs(chainID string, threshold int) ([]*CardanoTx, error)
	ClearUnprocessedTxs(chainID string) error
	MarkUnprocessedTxsAsProcessed(processedTxs []*ProcessedCardanoTx) error
	AddProcessedTxs(processedTxs []*ProcessedCardanoTx) error
	GetProcessedTx(chainID string, txHash indexer.Hash) (*ProcessedCardanoTx, error)
}

type BalanceStatesDB interface {
	AddChainBalance(chainID string, height uint64, balance string) error
	GetChainBalance(chainID string, height uint64) (*ChainBalance, error)
	GetAllChainBalances(chainID string, threshold int) ([]*ChainBalance, error)
	GetLastChainBalances(chainID string, threshold int) ([]*ChainBalance, error)
}

type CardanoTxsProcessorDB interface {
	CardanoTxsDB
	BridgeExpectedCardanoTxsDB
}

type Database interface {
	CardanoTxsProcessorDB
	BalanceStatesDB
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
	GetType() common.BridgingTxType
	ValidateAndAddClaim(claims *BridgeClaims, tx *CardanoTx, appConfig *AppConfig) error
}

type CardanoTxFailedProcessor interface {
	GetType() common.BridgingTxType
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
	SubmitClaims(claims *BridgeClaims, submitOpts *eth.SubmitOpts) error
	SubmitConfirmedBlocks(chainID string, blocks []*indexer.CardanoBlock) error
}

type ConfirmedBlocksSubmitter interface {
	StartSubmit()
	GetChainID() string
}

type ChainBalanceFetcher interface {
	NewUnprocessedTxs(originChainID string, txs []*indexer.Tx) error
	// NewUnprocessedEvmTxs(originChainID string, txs []*indexer.Tx) error
	Start()
}
