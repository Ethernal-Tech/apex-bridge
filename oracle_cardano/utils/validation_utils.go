package utils

import (
	"fmt"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

func ValidateTxInputs(tx *core.CardanoTx, metadata *common.BatchExecutedMetadata, appConfig *cCore.AppConfig) error {
	chainConfig := appConfig.CardanoChains[tx.OriginChainID]
	if chainConfig == nil {
		return fmt.Errorf("unsupported chain id found in tx. chain id: %v", tx.OriginChainID)
	}

	if metadata.IsStakeDelegation == 1 {
		return validateStakeDelTxInputs(tx, chainConfig.BridgingAddresses.FeeAddress)
	}

	return validateTxInputs(tx, chainConfig.BridgingAddresses.BridgingAddress, chainConfig.BridgingAddresses.FeeAddress)
}

func ValidateOutputsHaveUnknownTokens(tx *core.CardanoTx, appConfig *cCore.AppConfig) error {
	chainConfig := appConfig.CardanoChains[tx.OriginChainID]

	for _, out := range tx.Outputs {
		if out.Address != chainConfig.BridgingAddresses.BridgingAddress &&
			out.Address != chainConfig.BridgingAddresses.FeeAddress {
			continue
		}

		knownTokens := make([]wallet.Token, len(chainConfig.NativeTokens))

		for i, tokenConfig := range chainConfig.NativeTokens {
			token, err := cardanotx.GetNativeTokenFromConfig(tokenConfig)
			if err != nil {
				return err
			}

			knownTokens[i] = token
		}

		if cardanotx.UtxoContainsUnknownTokens(*out, knownTokens...) {
			return fmt.Errorf("tx %s has output (%s, %d), with some unknown tokens",
				tx.Hash, out.Address, out.Amount)
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

// Validate if all tx inputs belong to the multisig address or fee address
func validateTxInputs(tx *core.CardanoTx, bridgingAddress string, feeAddress string) error {
	foundBridgingAddress := false
	foundFeeAddress := false

	for _, utxo := range tx.Tx.Inputs {
		switch utxo.Output.Address {
		case bridgingAddress:
			foundBridgingAddress = true
		case feeAddress:
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

// Validate if all tx inputs belong to the fee address
func validateStakeDelTxInputs(tx *core.CardanoTx, feeAddress string) error {
	if len(tx.Tx.Inputs) == 0 {
		return fmt.Errorf("no inputs found in transaction")
	}

	for _, utxo := range tx.Tx.Inputs {
		if utxo.Output.Address != feeAddress {
			return fmt.Errorf("unexpected address found in tx input. address: %v", utxo.Output.Address)
		}
	}

	return nil
}
