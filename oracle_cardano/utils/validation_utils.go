package utils

import (
	"fmt"

	"slices"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
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

func ValidateOutputsHaveUnknownTokens(tx *core.CardanoTx, appConfig *cCore.AppConfig, isHotWallet bool) error {
	chainConfig := appConfig.CardanoChains[tx.OriginChainID]
	cardanoDestChainFeeAddress := appConfig.GetFeeMultisigAddress(tx.OriginChainID)

	var (
		knownTokens []wallet.Token
		err         error
	)

	if isHotWallet {
		knownTokens, err = cardanotx.GetWrappedTokens(&chainConfig.CardanoChainConfig)
	} else {
		knownTokens, err = cardanotx.GetKnownTokens(&chainConfig.CardanoChainConfig)
	}

	if err != nil {
		return fmt.Errorf("failed to get known tokens from chain config: %w", err)
	}

	zeroAddress, ok := appConfig.BridgingAddressesManager.GetPaymentAddressFromIndex(
		appConfig.ChainIDConverter.ToChainIDNum(tx.OriginChainID), 0)
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

// Validate if there sufficient amount of tx outputs for bridging request
// If validateTreasury is set to true, it will also validate treasury output existence and amount
// Returns found multisig output utxo
func ValidateTxOutputs(
	tx *core.CardanoTx, appConfig *cCore.AppConfig, allowMultiple bool, validateTreasury bool) (*indexer.TxOutput, error) {
	var (
		multisigUtxoOutput  *indexer.TxOutput = nil
		treasuryUtxoOutput  *indexer.TxOutput = nil
		foundMultisigOutput                   = false
		foundTreasuryOutput                   = false
	)

	for _, output := range tx.Tx.Outputs {
		if IsBridgingAddrForChain(appConfig, tx.OriginChainID, output.Address) {
			if multisigUtxoOutput == nil {
				foundMultisigOutput = true
				multisigUtxoOutput = output
			} else if !allowMultiple {
				return nil, fmt.Errorf("found multiple tx outputs to the bridging addresses on %s", tx.OriginChainID)
			}
		} else if IsTreasuryAddrForChain(appConfig, tx.OriginChainID, output.Address) {
			foundTreasuryOutput = true
			treasuryUtxoOutput = output
		}
	}

	if !foundMultisigOutput {
		return nil, fmt.Errorf("none of bridging addresses on %s found in tx outputs", tx.OriginChainID)
	}

	if validateTreasury && foundTreasuryOutput { //nolint:gocritic
		if treasuryUtxoOutput == nil {
			return nil, fmt.Errorf("treasury output on %s is not found in tx outputs", tx.OriginChainID)
		}

		if appConfig.CardanoChains[tx.OriginChainID].MinOperationFee > treasuryUtxoOutput.Amount {
			return nil, fmt.Errorf("treasury output amount %d is less than minimum operation fee %d on %s",
				treasuryUtxoOutput.Amount, appConfig.CardanoChains[tx.OriginChainID].MinOperationFee, tx.OriginChainID)
		}
	} else if validateTreasury && !foundTreasuryOutput {
		return nil, fmt.Errorf("treasury addresses on %s is not found in tx outputs", tx.OriginChainID)
	} else if !validateTreasury && foundTreasuryOutput {
		return nil, fmt.Errorf("treasury addresses on %s is found in tx outputs, but it shouldn't be there", tx.OriginChainID)
	}

	return multisigUtxoOutput, nil
}

func IsBridgingAddrForChain(appConfig *cCore.AppConfig, chainID string, addr string) bool {
	return slices.Contains(appConfig.GetBridgingMultisigAddresses(chainID), addr)
}

func IsTreasuryAddrForChain(appConfig *cCore.AppConfig, chainID string, addr string) bool {
	return appConfig.GetTreasuryAddress(chainID) == addr
}
