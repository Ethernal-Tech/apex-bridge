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

func CalculateMinUtxoCurrencyAmount(
	cardanoCliBinary string, protocolParams []byte,
	addr string, txInputOutputs []*indexer.TxInputOutput, txOutputs []wallet.TxOutput,
) (uint64, error) {
	sumMap := subtractTxOutputsFromSumMap(GetSumMapFromTxInputOutput(txInputOutputs), txOutputs)

	tokens, err := wallet.GetTokensFromSumMap(sumMap)
	if err != nil {
		return 0, err
	}

	txBuilder, err := wallet.NewTxBuilder(cardanoCliBinary)
	if err != nil {
		return 0, err
	}

	defer txBuilder.Dispose()

	minUtxo, err := txBuilder.SetProtocolParameters(protocolParams).CalculateMinUtxo(wallet.TxOutput{
		Addr:   addr,
		Amount: sumMap[wallet.AdaTokenName],
		Tokens: tokens,
	})
	if err != nil {
		return 0, err
	}

	return minUtxo, nil
}

func subtractTxOutputsFromSumMap(
	sumMap map[string]uint64, txOutputs []wallet.TxOutput,
) map[string]uint64 {
	updateTokenInMap := func(tokenName string, amount uint64) {
		if existingAmount, exists := sumMap[tokenName]; exists {
			if existingAmount > amount {
				sumMap[tokenName] = existingAmount - amount
			} else {
				delete(sumMap, tokenName)
			}
		}
	}

	for _, out := range txOutputs {
		updateTokenInMap(wallet.AdaTokenName, out.Amount)

		for _, token := range out.Tokens {
			updateTokenInMap(token.TokenName(), token.Amount)
		}
	}

	return sumMap
}

func GetSumMapFromTxInputOutput(utxos []*indexer.TxInputOutput) map[string]uint64 {
	totalSum := map[string]uint64{}

	for _, utxo := range utxos {
		totalSum[wallet.AdaTokenName] += utxo.Output.Amount

		for _, token := range utxo.Output.Tokens {
			totalSum[token.TokenName()] += token.Amount
		}
	}

	return totalSum
}

func FilterOutUtxosWithUnknownTokens(
	utxos []*indexer.TxInputOutput, excludingTokens ...wallet.Token,
) []*indexer.TxInputOutput {
	result := make([]*indexer.TxInputOutput, 0, len(utxos))

	for _, utxo := range utxos {
		if !UtxoContainsUnknownTokens(utxo.Output, excludingTokens...) {
			result = append(result, utxo)
		}
	}

	return result
}
