package utils

import (
	"fmt"

	"slices"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

// Validate if tx inputs contain the fee address
func ValidateTxInputs(tx *core.CardanoTx, appConfig *cCore.AppConfig) error {
	_, ok := appConfig.CardanoChains[tx.OriginChainID]
	if !ok {
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
	knownTokens := make([]wallet.Token, len(chainConfig.WrappedCurrencyTokens))

	for i, tokenConfig := range chainConfig.WrappedCurrencyTokens {
		token, err := cardanotx.GetNativeTokenFromConfig(tokenConfig)
		if err != nil {
			return err
		}

		knownTokens[i] = token
	}

	zeroAddress, ok := appConfig.BridgingAddressesManager.GetPaymentAddressFromIndex(
		common.ToNumChainID(tx.OriginChainID), 0)
	if !ok {
		return fmt.Errorf("failed to get zero address from bridging address manager")
	}

	for _, out := range tx.Outputs {
		if !IsBridgingAddrForChain(appConfig, tx.OriginChainID, out.Address) &&
			out.Address != cardanoDestChainFeeAddress {
			continue
		}

		// We allow only bridging via first address with native tokens
		knownTokensForAddress := knownTokens
		if out.Address != zeroAddress {
			knownTokensForAddress = nil
		}

		if cardanotx.UtxoContainsUnknownTokens(*out, knownTokensForAddress...) {
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
		if IsBridgingAddrForChain(appConfig, tx.OriginChainID, output.Address) {
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

func IsTxDirectionAllowed(appConfing *cCore.AppConfig, srcChainID, destChainID string) error {
	_, ok := appConfing.BridgingSettings.AllowedDirections[srcChainID][destChainID]
	if !ok {
		return fmt.Errorf("transaction direction not allowed: %s -> %s", srcChainID, destChainID)
	}

	// for _, transaction := range metadata.Transactions {
	// 	if transaction.IsNativeTokenOnSource() == !allowedDirection.CurrencyBirdgingAllowed {
	// 		return fmt.Errorf("transaction direction not allowed: %s -> %s", srcChainID, destChainID)
	// 	}
	// }

	return nil
}

func IsBridgingAddrForChain(appConfig *cCore.AppConfig, chainID string, addr string) bool {
	return slices.Contains(appConfig.GetBridgingMultisigAddresses(chainID), addr)
}
