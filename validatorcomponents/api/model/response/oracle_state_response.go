package response

import (
	"encoding/hex"

	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
)

type OracleStateResponse struct {
	ChainID   string                         `json:"chainID"`
	Utxos     []oCore.CardanoChainConfigUtxo `json:"utxos"`
	BlockSlot uint64                         `json:"slot"`
	BlockHash string                         `json:"hash"`
}

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
