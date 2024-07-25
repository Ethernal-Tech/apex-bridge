package response

import (
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
)

type OracleStateResponse struct {
	ChainID string                                   `json:"chainID"`
	Utxos   map[string][]core.CardanoChainConfigUtxo `json:"utxos"`
	Point   *indexer.BlockPoint                      `json:"point"`
}

func NewOracleStateResponse(
	chainID string, utxos map[string][]core.CardanoChainConfigUtxo, point *indexer.BlockPoint,
) *OracleStateResponse {
	return &OracleStateResponse{
		ChainID: chainID,
		Utxos:   utxos,
		Point:   point,
	}
}
