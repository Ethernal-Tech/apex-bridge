package core

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/ethgo"
)

type BridgeExpectedEthTxsDB interface {
	GetExpectedTxs(chainID string, priority uint8, threshold int) ([]*BridgeExpectedEthTx, error)
	GetAllExpectedTxs(chainID string, threshold int) ([]*BridgeExpectedEthTx, error)
	ClearExpectedTxs(chainID string) error
	AddExpectedTxs(expectedTxs []*BridgeExpectedEthTx) error
	MarkTxs(expectedInvalid, expectedProcessed []*BridgeExpectedEthTx, allProcessed []*ProcessedEthTx) error
}

type EthTxsDB interface {
	GetUnprocessedTxs(chainID string, priority uint8, threshold int) ([]*EthTx, error)
	GetAllUnprocessedTxs(chainID string, threshold int) ([]*EthTx, error)
	ClearUnprocessedTxs(chainID string) error
	GetProcessedTx(chainID string, txHash ethgo.Hash) (*ProcessedEthTx, error)
	GetProcessedTxByInnerActionTxHash(chainID string, innerActionTxHash ethgo.Hash) (*ProcessedEthTx, error)
	AddTxs(processedTxs []*ProcessedEthTx, unprocessedTxs []*EthTx) error
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

type Oracle interface {
	Start() error
	Dispose() error
}

type EthTxsReceiver interface {
	NewUnprocessedLog(originChainID string, log *ethgo.Log) error
}

type EthTxSuccessProcessor interface {
	GetType() common.BridgingTxType
	ValidateAndAddClaim(claims *oCore.BridgeClaims, tx *EthTx, appConfig *oCore.AppConfig) error
}

type EthTxFailedProcessor interface {
	GetType() common.BridgingTxType
	ValidateAndAddClaim(claims *oCore.BridgeClaims, tx *BridgeExpectedEthTx, appConfig *oCore.AppConfig) error
}

type EthChainObserver interface {
	Start() error
	Dispose() error
	GetConfig() *oCore.EthChainConfig
}

type EthBridgeDataFetcher interface {
	FetchExpectedTx(chainID string) (*BridgeExpectedEthTx, error)
}

type BridgeSubmitter interface {
	oCore.BridgeClaimsSubmitter
	SubmitConfirmedBlocks(chainID string, blocks uint64, lastBlock uint64) error
}

type EthConfirmedBlocksSubmitter interface {
	StartSubmit()
	GetChainID() string
}
