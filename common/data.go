package common

import "github.com/Ethernal-Tech/cardano-infrastructure/indexer"

type BridgingRequestStateKey struct {
	SourceChainID string
	SourceTxHash  indexer.Hash
}
