package cardanotx

import (
	"fmt"

	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

func IsValidOutputAddress(addr string, networkID wallet.CardanoNetworkType) bool {
	cardAddr, err := wallet.NewCardanoAddressFromString(addr)

	return err == nil && cardAddr.GetInfo().AddressType != wallet.RewardAddress &&
		cardAddr.GetInfo().Network == networkID
}

func UtxoContainsUnknownTokens(txOut indexer.TxOutput, knownTokens ...wallet.Token) bool {
	knownTokensMap := make(map[string]bool, len(knownTokens))

	for _, t := range knownTokens {
		knownTokensMap[t.String()] = true
	}

	for _, token := range txOut.Tokens {
		if _, exists := knownTokensMap[token.TokenName()]; !exists {
			return true
		}
	}

	return false
}

func GetKnownTokens(cardanoConfig *CardanoChainConfig) ([]wallet.Token, error) {
	knownTokens := make([]wallet.Token, len(cardanoConfig.NativeTokens))

	for i, tokenConfig := range cardanoConfig.NativeTokens {
		token, err := GetNativeTokenFromConfig(tokenConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve native tokens from config: %w", err)
		}

		knownTokens[i] = token
	}

	return knownTokens, nil
}

func GetTokenAmount(utxo *indexer.TxOutput, tokenName string) uint64 {
	if tokenName == wallet.AdaTokenName {
		return utxo.Amount
	}

	for _, token := range utxo.Tokens {
		if token.TokenName() == tokenName {
			return token.Amount
		}
	}

	return 0
}
