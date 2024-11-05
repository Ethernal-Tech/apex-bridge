package core

import (
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
	IncrementTryCount()
	IncrementBatchFailedCount()
	ToProcessed(isInvalid bool) BaseProcessedTx
	GetTryCount() uint32
	GetPriority() uint8
}

type BaseProcessedTx interface {
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
	RunChecks(bridgeClaims *BridgeClaims, chainID string, maxClaimsToGroup int, priority uint8)
	ProcessSubmitClaimsEvents(events *SubmitClaimsEvents, claims *BridgeClaims)
	UpdateBridgingRequestStates(bridgeClaims *BridgeClaims, bridgingRequestStateUpdater common.BridgingRequestStateUpdater)
	PersistNew()
}

type BridgeClaimsSubmitter interface {
	SubmitClaims(claims *BridgeClaims, submitOpts *eth.SubmitOpts) (*types.Receipt, error)
}

type ExpectedTxsFetcher interface {
	Start()
}

type ConfirmedBlocksSubmitter interface {
	StartSubmit()
	GetChainID() string
}
