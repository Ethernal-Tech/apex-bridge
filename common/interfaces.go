package common

import "context"

type BridgingRequestStateUpdater interface {
	New(sourceChainID string, model *NewBridgingRequestStateModel) error
	NewMultiple(sourceChainID string, models []*NewBridgingRequestStateModel) error
	Invalid(key BridgingRequestStateKey) error
	SubmittedToBridge(key BridgingRequestStateKey, destinationChainID string) error
	IncludedInBatch(destinationChainID string, batchID uint64, txs []BridgingRequestStateKey) error
	SubmittedToDestination(destinationChainID string, batchID uint64) error
	FailedToExecuteOnDestination(destinationChainID string, batchID uint64) error
	ExecutedOnDestination(destinationChainID string, batchID uint64, destinationTxHash Hash) error
}

// ChainSpecificConfig defines the interface for chain-specific configurations
type ChainSpecificConfig interface {
	GetChainType() string
}

type IStartable interface {
	Start(context.Context)
}
