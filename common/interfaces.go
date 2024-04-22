package common

import "github.com/Ethernal-Tech/cardano-infrastructure/indexer"

type BridgingRequestStateUpdater interface {
	New(sourceChainId string, tx *indexer.Tx) error
	NewMultiple(sourceChainId string, txs []*indexer.Tx) error
	Invalid(key BridgingRequestStateKey) error
	SubmittedToBridge(key BridgingRequestStateKey, destinationChainId string) error
	IncludedInBatch(destinationChainId string, batchId uint64, txs []BridgingRequestStateKey) error
	SubmittedToDestination(destinationChainId string, batchId uint64) error
	FailedToExecuteOnDestination(destinationChainId string, batchId uint64) error
	ExecutedOnDestination(destinationChainId string, batchId uint64, destinationTxHash string) error
}

// ChainSpecificConfig defines the interface for chain-specific configurations
type ChainSpecificConfig interface {
	GetChainType() string
}
