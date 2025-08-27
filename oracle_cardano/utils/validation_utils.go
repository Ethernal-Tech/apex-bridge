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

func ValidateOutputsHaveTokens(tx *core.CardanoTx, appConfig *cCore.AppConfig) error {
	chainConfig := appConfig.CardanoChains[tx.OriginChainID]

	for _, out := range tx.Outputs {
		if len(out.Tokens) > 0 && (out.Address == chainConfig.BridgingAddresses.BridgingAddress ||
			out.Address == chainConfig.BridgingAddresses.FeeAddress) {
			return fmt.Errorf("tx %s has output (%s, %d), with token count %d",
				tx.Hash, out.Address, out.Amount, len(out.Tokens))
		}
	}

	return nil
}

// Validate if there is one and only one tx output that belongs to the multisig address
// Returns found multisig output utxo
func ValidateTxOutputs(tx *core.CardanoTx, appConfig *cCore.AppConfig, allowMultiple bool) (*indexer.TxOutput, error) {
	var multisigUtxoOutput *indexer.TxOutput = nil

	chainConfig := appConfig.CardanoChains[tx.OriginChainID]

	for _, output := range tx.Tx.Outputs {
		if output.Address == chainConfig.BridgingAddresses.BridgingAddress {
			if multisigUtxoOutput == nil {
				multisigUtxoOutput = output
			} else if !allowMultiple {
				return nil, fmt.Errorf("found multiple tx outputs to the bridging address %s on %s",
					chainConfig.BridgingAddresses.BridgingAddress, tx.OriginChainID)
			}
		}
	}

	if multisigUtxoOutput == nil {
		return nil, fmt.Errorf("bridging address %s on %s not found in tx outputs",
			chainConfig.BridgingAddresses.BridgingAddress, tx.OriginChainID)
	}

	return multisigUtxoOutput, nil
}

func IsTxDirectionAllowed(appConfing *cCore.AppConfig, srcChainID, destChainID string) error {
	for _, chain := range appConfing.BridgingSettings.AllowedDirections[srcChainID] {
		if chain == destChainID {
			return nil
		}
	}

	return fmt.Errorf("transaction direction not allowed: %s -> %s", srcChainID, destChainID)
}
