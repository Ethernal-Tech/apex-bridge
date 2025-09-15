package core

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
)

type BridgingRequestStateDB interface {
	AddBridgingRequestState(state *common.BridgingRequestState) error
	UpdateBridgingRequestState(state *common.BridgingRequestState) error
	GetBridgingRequestState(sourceChainID string, sourceTxHash common.Hash) (*common.BridgingRequestState, error)
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

	Get(sourceChainID string, sourceTxHash common.Hash) (*common.BridgingRequestState, error)
	GetMultiple(sourceChainID string, sourceTxHashes []common.Hash) ([]*common.BridgingRequestState, error)
}

type RelayerImitator interface {
	common.IStartable
}

type ValidatorComponents interface {
	Start() error
	Dispose() error
}
