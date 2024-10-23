package utils

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
)

// Validate if all tx inputs belong to the multisig address or fee address
func ValidateTxInputs(tx *core.CardanoTx, appConfig *cCore.AppConfig) error {
	foundBridgingAddress := false
	foundFeeAddress := false

	chainConfig := appConfig.CardanoChains[tx.OriginChainID]
	if chainConfig == nil {
		return fmt.Errorf("unsupported chain id found in tx. chain id: %v", tx.OriginChainID)
	}

	for _, utxo := range tx.Tx.Inputs {
		switch utxo.Output.Address {
		case chainConfig.BridgingAddresses.BridgingAddress:
			foundBridgingAddress = true
		case chainConfig.BridgingAddresses.FeeAddress:
			foundFeeAddress = true
		default:
			return fmt.Errorf("unexpected address found in tx input. address: %v", utxo.Output.Address)
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
func ValidateTxOutputs(tx *core.CardanoTx, appConfig *cCore.AppConfig, allowMultiple bool) (*indexer.TxOutput, error) {
	var multisigUtxoOutput *indexer.TxOutput = nil

	chainConfig := appConfig.CardanoChains[tx.OriginChainID]
	if chainConfig == nil {
		return nil, fmt.Errorf("unsupported chain id found in tx. chain id: %v", tx.OriginChainID)
	}

	for _, utxo := range tx.Tx.Outputs {
		if utxo.Address == chainConfig.BridgingAddresses.BridgingAddress {
			if multisigUtxoOutput == nil {
				multisigUtxoOutput = utxo
			} else if !allowMultiple {
				return nil, fmt.Errorf("found multiple utxos to the bridging address on origin")
			}
		}
	}

	if multisigUtxoOutput == nil {
		return nil, fmt.Errorf("bridging address on origin not found in utxos")
	}

	return multisigUtxoOutput, nil
}
