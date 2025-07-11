package response

import (
	"encoding/hex"

	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
)

type OracleStateResponse struct {
	// Chain ID
	ChainID string `json:"chainID"`
	// Unspent transaction outputs
	Utxos []oCore.CardanoChainConfigUtxo `json:"utxos"`
	// Latest block slot
	BlockSlot uint64 `json:"slot"`
	// Latest block hash
	BlockHash string `json:"hash"`
} // @name OracleStateResponse

func NewOracleStateResponse(
	chainID string, utxos []oCore.CardanoChainConfigUtxo, slot uint64, hash [32]byte,
) *OracleStateResponse {
	return &OracleStateResponse{
		ChainID:   chainID,
		Utxos:     utxos,
		BlockSlot: slot,
		BlockHash: hex.EncodeToString(hash[:]),
	}
}
