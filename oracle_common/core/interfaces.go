package core

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/ethereum/go-ethereum/core/types"
)

type TxsProcessor interface {
	Start()
}

type SpecificChainTxsProcessorState interface {
	GetChainType() string
	Reset()
	RunChecks(bridgeClaims *BridgeClaims, chainID string, maxClaimsToGroup int, priority uint8)
	ProcessSubmitClaimsEvents(events *SubmitClaimsEvents, claims *BridgeClaims)
	PersistNew(bridgeClaims *BridgeClaims, bridgingRequestStateUpdater common.BridgingRequestStateUpdater)
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
