package utils

import (
	"fmt"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

// Validate if tx inputs contain the fee address
func ValidateTxInputs(tx *core.CardanoTx, appConfig *cCore.AppConfig) error {
	chainConfig := appConfig.CardanoChains[tx.OriginChainID]
	if chainConfig == nil {
		return fmt.Errorf("unsupported chain id found in tx. chain id: %v", tx.OriginChainID)
	}

	cardanoDestChainFeeAddress := appConfig.GetFeeMultisigAddress(tx.OriginChainID)

	for _, utxo := range tx.Tx.Inputs {
		if utxo.Output.Address == cardanoDestChainFeeAddress {
			return nil
		}
	}

	return fmt.Errorf("fee address not found in tx inputs")
}

func ValidateOutputsHaveUnknownTokens(tx *core.CardanoTx, appConfig *cCore.AppConfig) error {
	chainConfig := appConfig.CardanoChains[tx.OriginChainID]
	cardanoDestChainFeeAddress := appConfig.GetFeeMultisigAddress(tx.OriginChainID)

	for _, out := range tx.Outputs {
		if !CheckBridgingAddrForChain(appConfig, tx.OriginChainID, out.Address) &&
			out.Address != cardanoDestChainFeeAddress {
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

	for _, output := range tx.Tx.Outputs {
		if CheckBridgingAddrForChain(appConfig, tx.OriginChainID, output.Address) {
			if multisigUtxoOutput == nil {
				multisigUtxoOutput = output
			} else if !allowMultiple {
				return nil, fmt.Errorf("found multiple tx outputs to the bridging addresses on %s", tx.OriginChainID)
			}
		}
	}

	if multisigUtxoOutput == nil {
		return nil, fmt.Errorf("none of bridging addresses on %s found in tx outputs", tx.OriginChainID)
	}

	return multisigUtxoOutput, nil
}

func CheckBridgingAddrForChain(appConfig *cCore.AppConfig, chainID string, addr string) bool {
	for _, bridgingAddr := range appConfig.GetBridgingMultisigAddresses(chainID) {
		if bridgingAddr == addr {
			return true
		}
	}

	return false
}
