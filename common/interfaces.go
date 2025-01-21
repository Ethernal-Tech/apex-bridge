package common

import "context"

type BridgingRequestStateUpdater interface {
	New(srcChainID string, model *NewBridgingRequestStateModel) error
	NewMultiple(srcChainID string, models []*NewBridgingRequestStateModel) error
	Invalid(key BridgingRequestStateKey) error
	SubmittedToBridge(key BridgingRequestStateKey, dstChainID string) error
	IncludedInBatch(txs []BridgingRequestStateKey, dstChainID string) error
	SubmittedToDestination(txs []BridgingRequestStateKey, dstChainID string) error
	FailedToExecuteOnDestination(txs []BridgingRequestStateKey, dstChainID string) error
	ExecutedOnDestination(txs []BridgingRequestStateKey, dstTxHash Hash, dstChainID string) error
}

// ChainSpecificConfig defines the interface for chain-specific configurations
type ChainSpecificConfig interface {
	GetChainType() string
}

type IStartable interface {
	Start(context.Context)
}
