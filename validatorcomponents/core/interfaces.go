package core

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	relayerCore "github.com/Ethernal-Tech/apex-bridge/relayer/core"
)

type BridgingRequestStateDb interface {
	AddBridgingRequestState(state *BridgingRequestState) error
	UpdateBridgingRequestState(state *BridgingRequestState) error
	GetBridgingRequestState(sourceChainId string, sourceTxHash string) (*BridgingRequestState, error)
	GetBridgingRequestStatesByBatchId(destinationChainId string, batchId uint64) ([]*BridgingRequestState, error)
	GetUserBridgingRequestStates(sourceChainId string, userAddr string) ([]*BridgingRequestState, error)
}

type Database interface {
	BridgingRequestStateDb
	relayerCore.Database
	Init(filePath string) error
	Close() error
}

type ApiController interface {
	GetPathPrefix() string
	GetEndpoints() []*ApiEndpoint
}

type Api interface {
	Start() error
	Dispose() error
}

type BridgingRequestStateManager interface {
	common.BridgingRequestStateUpdater

	Get(sourceChainId string, sourceTxHash string) (*BridgingRequestState, error)
	GetAllForUser(sourceChainId string, userAddr string) ([]*BridgingRequestState, error)
}

type RelayerImitator interface {
	Start()
}

type ValidatorComponents interface {
	Start() error
	Dispose() error
	ErrorCh() <-chan error
}
