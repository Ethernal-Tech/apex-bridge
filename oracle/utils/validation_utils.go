package utils

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
)

// Validate if all tx inputs belong to the multisig address or fee address
func ValidateTxInputs(tx *core.CardanoTx, appConfig *core.AppConfig) error {
	foundBridgingAddress := false
	foundFeeAddress := false
	for _, chainConfig := range appConfig.CardanoChains {
		if chainConfig.ChainId == tx.OriginChainId {
			for _, utxo := range tx.Tx.Inputs {
				if utxo.Output.Address == chainConfig.BridgingAddresses.BridgingAddress && !foundBridgingAddress {
					foundBridgingAddress = true
				} else if utxo.Output.Address == chainConfig.BridgingAddresses.FeeAddress && !foundFeeAddress {
					foundFeeAddress = true
				} else {
					return fmt.Errorf("unexpected address found in tx input", "address", utxo.Output.Address)
				}
			}
			break
		}
	}

	if !foundBridgingAddress {
		return fmt.Errorf("bridging address not found in tx inptus")
	}

	if !foundFeeAddress {
		return fmt.Errorf("fee address not found in tx intpus")
	}

	return nil
}
