package utils

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
)

// Validate if all tx inputs belong to the multisig address or fee address
func ValidateTxInputs(tx *core.CardanoTx, appConfig *core.AppConfig) error {

	for _, utxo := range tx.Tx.Inputs {
		if (utxo.Output.Address != appConfig.CardanoChains[tx.OriginChainId].BridgingAddresses.BridgingAddress &&
			utxo.Output.Address != appConfig.CardanoChains[tx.OriginChainId].BridgingAddresses.FeeAddress) && utxo.Output.IsUsed {
			return fmt.Errorf("unexpected address found in tx input", "address", utxo.Output.Address)
		}
	}

	return nil
}
