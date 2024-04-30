package response

import (
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
)

type CardanoIndexerTxResponse = indexer.Tx

func NewCardanoIndexerTxResponse(tx *indexer.Tx) *CardanoIndexerTxResponse {
	return tx
}
