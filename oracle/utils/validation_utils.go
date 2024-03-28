package utils

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
)

// Validate if all tx inputs belong to the multisig address or fee address
func ValidateTxInputs(tx *core.CardanoTx, appConfig *core.AppConfig) error {
	foundBridgingAddress := false
	foundFeeAddress := false
	for _, chainConfig := range appConfig.CardanoChains {
		if chainConfig.ChainId == tx.OriginChainId {
			for _, utxo := range tx.Tx.Inputs {
				if utxo.Output.Address == chainConfig.BridgingAddresses.BridgingAddress {
					foundBridgingAddress = true
				} else if utxo.Output.Address == chainConfig.BridgingAddresses.FeeAddress {
					foundFeeAddress = true
				} else {
					return fmt.Errorf("unexpected address found in tx input. address: %v", utxo.Output.Address)
				}
			}
			break
		}
	}

	if !foundBridgingAddress {
		return fmt.Errorf("bridging address not found in tx inputs")
	}

	if !foundFeeAddress {
		return fmt.Errorf("fee address not found in tx inputs")
	}

	return nil
}

// Validate if there is one and only one tx output that belongs to the multisig address
// Returns found multisig output utxo
func ValidateTxOutputs(tx *core.CardanoTx, appConfig *core.AppConfig) (*indexer.TxOutput, error) {
	var multisigUtxoOutput *indexer.TxOutput = nil
	for _, chainConfig := range appConfig.CardanoChains {
		if chainConfig.ChainId == tx.OriginChainId {
			for _, utxo := range tx.Tx.Outputs {
				if utxo.Address == chainConfig.BridgingAddresses.BridgingAddress {
					if multisigUtxoOutput == nil {
						multisigUtxoOutput = utxo
					} else {
						return nil, fmt.Errorf("found multiple utxos to the bridging address on origin")
					}
				}
			}
			break
		}
	}

	if multisigUtxoOutput == nil {
		return nil, fmt.Errorf("bridging address on origin not found in utxos")
	}

	return multisigUtxoOutput, nil
}
