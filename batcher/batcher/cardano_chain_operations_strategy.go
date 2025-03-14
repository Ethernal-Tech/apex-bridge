package batcher

import (
	"fmt"
	"sort"

	cardano "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	txsend "github.com/Ethernal-Tech/cardano-infrastructure/sendtx"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
)

func filterUtxos(
	multisigUtxos, feeUtxos []*indexer.TxInputOutput, config *cardano.CardanoChainConfig,
) ([]*indexer.TxInputOutput, []*indexer.TxInputOutput, error) {
	knownTokens, err := cardano.GetKnownTokens(config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get known tokens: %w", err)
	}

	return filterOutUtxosWithUnknownTokens(multisigUtxos, knownTokens...),
		filterOutUtxosWithUnknownTokens(feeUtxos),
		nil
}

func getOutputs(
	txs []eth.ConfirmedTransaction, cardanoConfig *cardano.CardanoChainConfig, logger hclog.Logger,
) (cardano.TxOutputs, error) {
	receiversMap := map[string]cardanowallet.TxOutput{}

	for _, transaction := range txs {
		for _, receiver := range transaction.Receivers {
			data := receiversMap[receiver.DestinationAddress]
			data.Amount += receiver.Amount.Uint64()

			if receiver.AmountWrapped != nil && receiver.AmountWrapped.Sign() > 0 {
				if len(data.Tokens) == 0 {
					token, err := cardanoConfig.GetNativeToken(
						common.ToStrChainID(transaction.SourceChainId))
					if err != nil {
						return cardano.TxOutputs{}, err
					}

					data.Tokens = []cardanowallet.TokenAmount{
						cardanowallet.NewTokenAmount(token, receiver.AmountWrapped.Uint64()),
					}
				} else {
					data.Tokens[0].Amount += receiver.AmountWrapped.Uint64()
				}
			}

			receiversMap[receiver.DestinationAddress] = data
		}
	}

	result := cardano.TxOutputs{
		Outputs: make([]cardanowallet.TxOutput, 0, len(receiversMap)),
		Sum:     map[string]uint64{},
	}

	for addr, txOut := range receiversMap {
		if txOut.Amount == 0 {
			logger.Warn("skipped output with zero amount", "addr", addr)

			continue
		} else if !cardano.IsValidOutputAddress(addr, cardanoConfig.NetworkID) {
			logger.Warn("skipped output because it is invalid", "addr", addr)

			continue
		}

		txOut.Addr = addr

		result.Outputs = append(result.Outputs, txOut)

		result.Sum[cardanowallet.AdaTokenName] += txOut.Amount

		for _, token := range txOut.Tokens {
			result.Sum[token.TokenName()] += token.Amount
		}
	}

	// sort outputs because all batchers should have same order of outputs
	sort.Slice(result.Outputs, func(i, j int) bool {
		return result.Outputs[i].Addr < result.Outputs[j].Addr
	})

	return result, nil
}

func getUTXOsForAmounts(
	cardanoConfig *cardano.CardanoChainConfig,
	multisigFeeAddress string,
	multisigUtxos []*indexer.TxInputOutput,
	feeUtxos []*indexer.TxInputOutput,
	desiredAmounts map[string]uint64,
	minUtxoAmountLovelaceAmount uint64,
) ([]*indexer.TxInputOutput, []*indexer.TxInputOutput, error) {
	var err error

	if len(feeUtxos) == 0 {
		return nil, nil, fmt.Errorf("fee multisig does not have any utxo: %s", multisigFeeAddress)
	}

	feeUtxos = feeUtxos[:min(cardanoConfig.MaxFeeUtxoCount, len(feeUtxos))] // do not take more than maxFeeUtxoCount

	multisigUtxos, err = getNeededUtxos(
		multisigUtxos,
		desiredAmounts,
		minUtxoAmountLovelaceAmount,
		cardanoConfig.MaxUtxoCount-len(feeUtxos),
		cardanoConfig.TakeAtLeastUtxoCount,
	)
	if err != nil {
		return nil, nil, err
	}

	return multisigUtxos, feeUtxos, nil
}

func getNeededUtxos(
	txInputOutputs []*indexer.TxInputOutput,
	desiredAmounts map[string]uint64,
	minUtxoLovelaceAmount uint64,
	maxUtxoCount int,
	takeAtLeastUtxoCount int,
) ([]*indexer.TxInputOutput, error) {
	inputUtxos := make([]cardanowallet.Utxo, len(txInputOutputs))
	desiredAmounts[cardanowallet.AdaTokenName] += minUtxoLovelaceAmount

	for i, utxo := range txInputOutputs {
		inputUtxos[i] = cardanowallet.Utxo{
			Hash:   utxo.Input.Hash.String(),
			Index:  utxo.Input.Index,
			Amount: utxo.Output.Amount,
			Tokens: make([]cardanowallet.TokenAmount, len(utxo.Output.Tokens)),
		}
		for j, token := range utxo.Output.Tokens {
			inputUtxos[i].Tokens[j] = cardanowallet.NewTokenAmount(
				cardanowallet.NewToken(token.PolicyID, token.Name), token.Amount)
		}
	}

	outputUTXOs, err := txsend.GetUTXOsForAmounts(inputUtxos, desiredAmounts, maxUtxoCount, takeAtLeastUtxoCount)
	if err != nil {
		return nil, err
	}

	usedUtxoMap := map[string]bool{}
	for _, utxo := range outputUTXOs.Inputs {
		usedUtxoMap[utxo.String()] = true
	}

	chosenUTXOs := make([]*indexer.TxInputOutput, 0, len(outputUTXOs.Inputs))

	for _, utxo := range txInputOutputs {
		if usedUtxoMap[utxo.Input.String()] {
			chosenUTXOs = append(chosenUTXOs, utxo)
		}
	}

	return chosenUTXOs, nil
}

func filterOutUtxosWithUnknownTokens(
	utxos []*indexer.TxInputOutput, excludingTokens ...cardanowallet.Token,
) []*indexer.TxInputOutput {
	result := make([]*indexer.TxInputOutput, 0, len(utxos))

	for _, utxo := range utxos {
		if !cardano.UtxoContainsUnknownTokens(utxo.Output, excludingTokens...) {
			result = append(result, utxo)
		}
	}

	return result
}

func getSumMapFromTxInputOutput(utxos []*indexer.TxInputOutput) map[string]uint64 {
	totalSum := map[string]uint64{}

	for _, utxo := range utxos {
		totalSum[cardanowallet.AdaTokenName] += utxo.Output.Amount

		for _, token := range utxo.Output.Tokens {
			totalSum[token.TokenName()] += token.Amount
		}
	}

	return totalSum
}

func getTxOutputFromUtxos(utxos []*indexer.TxInputOutput, addr string) (cardanowallet.TxOutput, error) {
	totalSum := getSumMapFromTxInputOutput(utxos)
	tokens := make([]cardanowallet.TokenAmount, 0, len(totalSum)-1)

	for tokenName, amount := range totalSum {
		if tokenName != cardanowallet.AdaTokenName {
			newToken, err := cardanowallet.NewTokenWithFullName(tokenName, true)
			if err != nil {
				return cardanowallet.TxOutput{}, err
			}

			tokens = append(tokens, cardanowallet.NewTokenAmount(newToken, amount))
		}
	}

	return cardanowallet.NewTxOutput(addr, totalSum[cardanowallet.AdaTokenName], tokens...), nil
}

func subtractTxOutputsFromSumMap(
	sumMap map[string]uint64, txOutputs []cardanowallet.TxOutput,
) map[string]uint64 {
	for _, out := range txOutputs {
		if value, exists := sumMap[cardanowallet.AdaTokenName]; exists {
			if value > out.Amount {
				sumMap[cardanowallet.AdaTokenName] = value - out.Amount
			} else {
				delete(sumMap, cardanowallet.AdaTokenName)
			}
		}

		for _, token := range out.Tokens {
			tokenName := token.TokenName()
			if value, exists := sumMap[tokenName]; exists {
				if value > token.Amount {
					sumMap[tokenName] = value - token.Amount
				} else {
					delete(sumMap, tokenName)
				}
			}
		}
	}

	return sumMap
}

func calculateMinUtxoLovelaceAmount(
	cardanoCliBinary string,
	multisigAddr string, multisigUtxos []*indexer.TxInputOutput,
	protocolParams []byte, txOutputs []cardanowallet.TxOutput,
) (uint64, error) {
	sumMap := subtractTxOutputsFromSumMap(getSumMapFromTxInputOutput(multisigUtxos), txOutputs)

	tokens, err := cardanowallet.GetTokensFromSumMap(sumMap)
	if err != nil {
		return 0, err
	}

	txBuilder, err := cardanowallet.NewTxBuilder(cardanoCliBinary)
	if err != nil {
		return 0, err
	}

	defer txBuilder.Dispose()

	// calculate final multisig output change
	minUtxo, err := txBuilder.SetProtocolParameters(protocolParams).CalculateMinUtxo(cardanowallet.TxOutput{
		Addr:   multisigAddr,
		Amount: sumMap[cardanowallet.AdaTokenName],
		Tokens: tokens,
	})
	if err != nil {
		return 0, err
	}

	return minUtxo, nil
}
