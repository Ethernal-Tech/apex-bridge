package core

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
)

type BridgingRequestStateDB interface {
	AddBridgingRequestState(state *BridgingRequestState) error
	UpdateBridgingRequestState(state *BridgingRequestState) error
	GetBridgingRequestState(sourceChainID string, sourceTxHash common.Hash) (*BridgingRequestState, error)
}

type Database interface {
	BridgingRequestStateDB
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
	common.IStartable
}

type ValidatorComponents interface {
	Start() error
	Dispose() error
	ErrorCh() <-chan error
}
