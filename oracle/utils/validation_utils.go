package utils

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
)

// Validate if all tx inputs belong to the multisig address or fee address
func ValidateTxInputs(tx *core.CardanoTx, appConfig *core.AppConfig) error {
	addressesOfInterest := make(map[string]bool)

	for _, bridgingAddress := range appConfig.CardanoChains[tx.OriginChainId].BridgingAddresses {
		if bridgingAddress.ChainId == tx.OriginChainId {
			addressesOfInterest[bridgingAddress.Address] = true
			addressesOfInterest[bridgingAddress.FeeAddress] = true
		}
	}

	for _, utxo := range tx.Tx.Inputs {
		if !addressesOfInterest[utxo.Output.Address] && utxo.Output.IsUsed {
			return fmt.Errorf("unexpected address found in tx input", "address", utxo.Output.Address)
		}
	}

	return nil
}
