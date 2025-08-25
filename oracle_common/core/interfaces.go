package core

import (
	"context"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/ethereum/go-ethereum/core/types"
)

type BaseTx interface {
	GetChainID() string
	GetTxHash() []byte
	UnprocessedDBKey() []byte
	SetLastTimeTried(lastTimeTried time.Time)
	IncrementSubmitTryCount()
	IncrementBatchTryCount()
	IncrementRefundTryCount()
	ToProcessed(isInvalid bool) BaseProcessedTx
	GetSubmitTryCount() uint32
	GetPriority() uint8
}

type IIsInvalid interface {
	GetIsInvalid() bool
}

type BaseProcessedTx interface {
	IIsInvalid
	GetChainID() string
	GetTxHash() []byte
	HasInnerActionTxHash() bool
	GetInnerActionTxHash() []byte
	UnprocessedDBKey() []byte
}

type BaseExpectedTx interface {
	GetChainID() string
	GetTxHash() []byte
	DBKey() []byte
	GetPriority() uint8
	GetIsInvalid() bool
	GetIsProcessed() bool
	SetProcessed()
	SetInvalid()
}

type TxsProcessor interface {
	Start()
}

type SpecificChainTxsProcessorState interface {
	GetChainType() string
	Reset()
	ProcessSavedEvents()
	RunChecks(bridgeClaims *BridgeClaims, chainID string, maxClaimsToGroup int, priority uint8, isValidatorPending bool)
	ProcessSubmitClaimsEvents(events *SubmitClaimsEvents, claims *BridgeClaims)
	UpdateBridgingRequestStates(
		bridgeClaims *BridgeClaims, bridgingRequestStateUpdater common.BridgingRequestStateUpdater)
	PersistNew()
}

type BridgeClaimsSubmitter interface {
	SubmitClaims(claims *BridgeClaims, submitOpts *eth.SubmitOpts) (*types.Receipt, error)
}

type BridgeBlocksSubmitter interface {
	SubmitBlocks(chainID string, blocks []eth.CardanoBlock) error
}

type BridgeSubmitter interface {
	BridgeClaimsSubmitter
	BridgeBlocksSubmitter
}

type BridgeDataFetcher interface {
	GetBatchTransactions(chainID string, batchID uint64) ([]eth.TxDataInfo, error)
}

type ExpectedTxsFetcher interface {
	Start()
}

type ConfirmedBlocksSubmitter interface {
	Start(ctx context.Context)
}

type BlockSubmitterDB interface {
	GetBlocksSubmitterInfo(chainID string) (BlocksSubmitterInfo, error)
	SetBlocksSubmitterInfo(chainID string, info BlocksSubmitterInfo) error
}
