package cardanotx

import (
	"bytes"
	"fmt"
	"slices"

	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

// IsValidReceiverAddress validates an address that a batch pays to. The oracles use it on user
// supplied metadata and the batcher uses it again before building outputs, so both apply the same rules.
//
// Byron addresses are rejected by header before GetInfo is called: the cardano-infrastructure
// Byron parser panics on a malformed address root, and the ledger refuses any tx running a
// Plutus script that has a Byron output (ByronTxOutInContext), so such a receiver fails the whole batch.
//
// The address must also be in canonical form (see isCanonicalShelleyAddress), because the
// cardano-infrastructure parser accepts encodings that the ledger's strict address decoding rejects.
func IsValidReceiverAddress(addr string, networkID wallet.CardanoNetworkType) bool {
	cardAddr, err := wallet.NewCardanoAddressFromString(addr)
	if err != nil || wallet.GetAddressTypeFromHeader(cardAddr.GetBytes()[0]) == wallet.ByronAddress {
		return false
	}

	info := cardAddr.GetInfo()
	if info.AddressType == wallet.RewardAddress || info.Network != networkID {
		return false
	}

	return isCanonicalShelleyAddress(addr, cardAddr.GetBytes(), info)
}

// isCanonicalShelleyAddress reports whether raw is exactly the encoding of its own decoded parts
// and addr is the bech32 string of raw. This rejects base addresses with trailing bytes (the
// parser only checks a minimum length), pointer addresses with non-minimal varints, and strings
// whose prefix does not match the network or that are not lower case.
func isCanonicalShelleyAddress(addr string, raw []byte, info wallet.CardanoAddressInfo) bool {
	if len(info.Extra) > 0 {
		return false
	}

	rebuilt, err := info.ToCardanoAddress()
	if err != nil || !bytes.Equal(rebuilt.GetBytes(), raw) {
		return false
	}

	// rebuilt has no cached string, so String encodes it again from the bytes
	return rebuilt.String() == addr
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
	knownTokens := make([]wallet.Token, 0)

	for _, tokenConfig := range cardanoConfig.Tokens {
		if tokenConfig.ChainSpecific == wallet.AdaTokenName {
			continue
		}

		token, err := wallet.NewTokenWithFullNameTry(tokenConfig.ChainSpecific)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to retrieve native token with name %s from config: %w", tokenConfig.ChainSpecific, err)
		}

		knownTokens = append(knownTokens, token)
	}

	return knownTokens, nil
}

func GetWrappedTokens(cardanoConfig *CardanoChainConfig) ([]wallet.Token, error) {
	wrappedTokens := make([]wallet.Token, 0)

	// Sort keys to have deterministic order
	keys := make([]uint16, 0, len(cardanoConfig.Tokens))
	for k := range cardanoConfig.Tokens {
		keys = append(keys, k)
	}

	slices.Sort(keys)

	for _, k := range keys {
		tokenConfig := cardanoConfig.Tokens[k]

		if !tokenConfig.IsWrappedCurrency {
			continue
		}

		token, err := wallet.NewTokenWithFullNameTry(tokenConfig.ChainSpecific)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to retrieve wrapped token with name %s from config: %w", tokenConfig.ChainSpecific, err)
		}

		wrappedTokens = append(wrappedTokens, token)
	}

	return wrappedTokens, nil
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

	minUtxo, err := txBuilder.SetProtocolParameters(protocolParams).CalculateMinUtxo(wallet.TxOutputWithRefScript{
		TxOutput: wallet.TxOutput{
			Addr:   addr,
			Amount: sumMap[wallet.AdaTokenName],
			Tokens: tokens,
		},
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
