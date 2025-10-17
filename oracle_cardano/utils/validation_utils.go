package utils

import (
	"fmt"

	"slices"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/sendtx"
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
	originChainID := common.ToNumChainID(tx.OriginChainID)

	cardanoDestChainFeeAddress := appConfig.GetFeeMultisigAddress(tx.OriginChainID)

	knownTokens, err := resolveKnownTokens(chainConfig.NativeTokens)
	if err != nil {
		return err
	}

	zeroAddress, ok := appConfig.BridgingAddressesManager.GetFirstIndexAddress(originChainID)
	if !ok {
		return fmt.Errorf("failed to get zero address from bridging address manager")
	}

	rewardZeroAddress, _ := appConfig.RewardBridgingAddressesManager.GetFirstIndexAddress(originChainID)

	for _, out := range tx.Outputs {
		isFeeAddress := out.Address == cardanoDestChainFeeAddress
		addressIdx, isBridgingAddr := GetBridgingAddressIndex(appConfig, tx.OriginChainID, out.Address)

		if !isBridgingAddr && !isFeeAddress {
			continue
		}

		knownTokensForAddress, err := getAllowedTokensForAddress(
			out.Address, addressIdx, knownTokens, chainConfig.StakedToken, zeroAddress, rewardZeroAddress,
		)

		if err != nil {
			return err
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
	for _, chain := range appConfing.BridgingSettings.AllowedDirections[srcChainID] {
		if chain == destChainID {
			return nil
		}
	}

	return fmt.Errorf("transaction direction not allowed: %s -> %s", srcChainID, destChainID)
}

func IsBridgingAddrForChain(appConfig *cCore.AppConfig, chainID string, addr string) bool {
	return slices.Contains(appConfig.GetBridgingMultisigAddresses(chainID), addr)
}

func GetBridgingAddressIndex(appConfig *cCore.AppConfig, chainID string, addr string) (uint8, bool) {
	chainIDNum := common.ToNumChainID(chainID)

	if idx, ok := appConfig.BridgingAddressesManager.GetPaymentAddressIndex(chainIDNum, addr); ok {
		return idx, true
	}

	return appConfig.RewardBridgingAddressesManager.GetPaymentAddressIndex(chainIDNum, addr)
}

func resolveKnownTokens(nativeTokens []sendtx.TokenExchangeConfig) ([]wallet.Token, error) {
	knownTokens := make([]wallet.Token, len(nativeTokens))

	for i, tokenConfig := range nativeTokens {
		token, err := cardanotx.GetNativeTokenFromConfig(tokenConfig)
		if err != nil {
			return nil, err
		}

		knownTokens[i] = token
	}

	return knownTokens, nil
}

func getAllowedTokensForAddress(
	address string,
	addressIdx uint8,
	knownTokens []wallet.Token,
	stakedToken sendtx.TokenExchangeConfig,
	zeroAddress string,
	rewardZeroAddress string,
) ([]wallet.Token, error) {
	if addressIdx >= common.FirstRewardBridgingAddressIndex {
		if rewardZeroAddress == "" {
			return nil, fmt.Errorf("reward zero address doesn't exist")
		}

		if address != rewardZeroAddress {
			return nil, nil // Only reward-zero address can bridge with tokens
		}

		stakedToken, err := cardanotx.GetNativeTokenFromConfig(stakedToken)
		if err != nil {
			return nil, err
		}

		return []wallet.Token{stakedToken}, nil
	}

	if address != zeroAddress {
		return nil, nil // Only zero-index address can bridge with tokens
	}

	return knownTokens, nil
}
