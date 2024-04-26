package common

import "github.com/Ethernal-Tech/cardano-infrastructure/indexer"

type BridgingRequestStateUpdater interface {
	New(sourceChainID string, tx *indexer.Tx) error
	NewMultiple(sourceChainID string, txs []*indexer.Tx) error
	Invalid(key BridgingRequestStateKey) error
	SubmittedToBridge(key BridgingRequestStateKey, destinationChainID string) error
	IncludedInBatch(destinationChainID string, batchID uint64, txs []BridgingRequestStateKey) error
	SubmittedToDestination(destinationChainID string, batchID uint64) error
	FailedToExecuteOnDestination(destinationChainID string, batchID uint64) error
	ExecutedOnDestination(destinationChainID string, batchID uint64, destinationTxHash string) error
}

// ChainSpecificConfig defines the interface for chain-specific configurations
type ChainSpecificConfig interface {
	GetChainType() string
}
