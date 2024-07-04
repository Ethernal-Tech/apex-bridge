package core

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	oracleCore "github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/ethgo"
)

type BridgeExpectedEthTxsDB interface {
	AddExpectedTxs(expectedTxs []*BridgeExpectedEthTx) error
	GetExpectedTxs(chainID string, priority uint8, threshold int) ([]*BridgeExpectedEthTx, error)
	GetAllExpectedTxs(chainID string, threshold int) ([]*BridgeExpectedEthTx, error)
	ClearExpectedTxs(chainID string) error
	MarkExpectedTxsAsProcessed(expectedTxs []*BridgeExpectedEthTx) error
	MarkExpectedTxsAsInvalid(expectedTxs []*BridgeExpectedEthTx) error
}

type EthTxsDB interface {
	AddUnprocessedTxs(unprocessedTxs []*EthTx) error
	GetUnprocessedTxs(chainID string, priority uint8, threshold int) ([]*EthTx, error)
	GetAllUnprocessedTxs(chainID string, threshold int) ([]*EthTx, error)
	ClearUnprocessedTxs(chainID string) error
	MarkUnprocessedTxsAsProcessed(processedTxs []*ProcessedEthTx) error
	AddProcessedTxs(processedTxs []*ProcessedEthTx) error
	GetProcessedTx(chainID string, txHash ethgo.Hash) (*ProcessedEthTx, error)
}

type EthTxsProcessorDB interface {
	EthTxsDB
	BridgeExpectedEthTxsDB
}

type Database interface {
	EthTxsProcessorDB
	Init(filePath string) error
	Close() error
}

type EthTxsProcessor interface {
	NewUnprocessedLog(originChainID string, log *ethgo.Log) error
	Start()
}

type EthTxProcessor interface {
	GetType() common.BridgingTxType
	ValidateAndAddClaim(claims *oracleCore.BridgeClaims, tx *EthTx, appConfig *oracleCore.AppConfig) error
}

type EthTxFailedProcessor interface {
	GetType() common.BridgingTxType
	ValidateAndAddClaim(claims *oracleCore.BridgeClaims, tx *BridgeExpectedEthTx, appConfig *oracleCore.AppConfig) error
}

type EthChainObserver interface {
	Start() error
	Dispose() error
	GetConfig() *oracleCore.EthChainConfig
	ErrorCh() <-chan error
}
