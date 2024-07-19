package response

import "github.com/Ethernal-Tech/cardano-infrastructure/indexer"

type OracleStateResponse struct {
	ChainID string                              `json:"chainID"`
	Utxos   map[string][]*indexer.TxInputOutput `json:"utxos"`
	Point   *indexer.BlockPoint                 `json:"point"`
}

func NewOracleStateResponse(
	chainID string, utxos map[string][]*indexer.TxInputOutput, point *indexer.BlockPoint,
) *OracleStateResponse {
	return &OracleStateResponse{
		ChainID: chainID,
		Utxos:   utxos,
		Point:   point,
	}
}
