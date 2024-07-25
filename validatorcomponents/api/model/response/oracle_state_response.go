package response

import (
	"encoding/hex"

	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
)

type OracleStateResponse struct {
	ChainID   string                        `json:"chainID"`
	Utxos     []core.CardanoChainConfigUtxo `json:"utxos"`
	BlockSlot uint64                        `json:"slot"`
	BlockHash string                        `json:"hash"`
}

func NewOracleStateResponse(
	chainID string, utxos []core.CardanoChainConfigUtxo, slot uint64, hash [32]byte,
) *OracleStateResponse {
	return &OracleStateResponse{
		ChainID:   chainID,
		Utxos:     utxos,
		BlockSlot: slot,
		BlockHash: hex.EncodeToString(hash[:]),
	}
}
