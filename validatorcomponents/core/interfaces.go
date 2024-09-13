package core

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	relayerCore "github.com/Ethernal-Tech/apex-bridge/relayer/core"
)

type BridgingRequestStateDB interface {
	AddBridgingRequestState(state *BridgingRequestState) error
	UpdateBridgingRequestState(state *BridgingRequestState) error
	GetBridgingRequestState(sourceChainID string, sourceTxHash common.Hash) (*BridgingRequestState, error)
	GetBridgingRequestStatesByBatchID(destinationChainID string, batchID uint64) ([]*BridgingRequestState, error)
}

type Database interface {
	BridgingRequestStateDB
	relayerCore.Database
	Init(filePath string) error
	Close() error
}

type APIController interface {
	GetPathPrefix() string
	GetEndpoints() []*APIEndpoint
}

type API interface {
	Start()
	Dispose() error
}

type BridgingRequestStateManager interface {
	common.BridgingRequestStateUpdater

	Get(sourceChainID string, sourceTxHash common.Hash) (*BridgingRequestState, error)
	GetMultiple(sourceChainID string, sourceTxHashes []common.Hash) ([]*BridgingRequestState, error)
}

type RelayerImitator interface {
	Start()
}

type ValidatorComponents interface {
	Start() error
	Dispose() error
	ErrorCh() (<-chan error, error)
}
