package common

import "context"

type BridgingRequestStateUpdater interface {
	New(srcChainID string, model *NewBridgingRequestStateModel) error
	NewMultiple(srcChainID string, models []*NewBridgingRequestStateModel) error
	Invalid(key BridgingRequestStateKey) error
	SubmittedToBridge(key BridgingRequestStateKey, dstChainID string) error
	IncludedInBatch(txs []BridgingRequestStateKey, dstChainID string) error
	SubmittedToDestination(txs []BridgingRequestStateKey) error
	FailedToExecuteOnDestination(txs []BridgingRequestStateKey) error
	ExecutedOnDestination(txs []BridgingRequestStateKey, dstTxHash Hash) error
}

// ChainSpecificConfig defines the interface for chain-specific configurations
type ChainSpecificConfig interface {
	GetChainType() string
}

type IStartable interface {
	Start(context.Context)
}
