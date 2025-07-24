package batcher

import (
	"encoding/json"
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

const StakeDepositFieldName = "stakeAddressDeposit"

func getReceiversMap(
	destChainID string,
	cardanoConfig *cardano.CardanoChainConfig,
	feeAddr string,
	txs []eth.ConfirmedTransaction,
	refundUtxosPerConfirmedTx [][]*indexer.TxInputOutput,
) (map[string]map[string]uint64, error) {
	receiversMap := map[string]map[string]uint64{}
	updateMap := func(addr string, tokenName string, value uint64) {
		subMap, exists := receiversMap[addr]
		if !exists {
			subMap = map[string]uint64{}
			receiversMap[addr] = subMap
		}

		subMap[tokenName] += value
	}

	for txIndx, tx := range txs {
		// stake delegation tx are not processed in this way
		if tx.TransactionType == uint8(common.StakeDelConfirmedTxType) {
			continue
		}

		srcChainID := common.ToStrChainID(tx.SourceChainId)

		for _, receiver := range tx.Receivers {
			amount := receiver.Amount.Uint64()

			switch tx.TransactionType {
			case uint8(common.RefundConfirmedTxType):
				// In case a transaction is of type refund, batcher should transfer minFeeForBridging
				// to fee payer address, and the rest is transferred to the user.
				updateMap(receiver.DestinationAddress, cardanowallet.AdaTokenName, amount-cardanoConfig.MinFeeForBridging)
				updateMap(feeAddr, cardanowallet.AdaTokenName, cardanoConfig.MinFeeForBridging)

				if receiver.AmountWrapped != nil && receiver.AmountWrapped.Sign() > 0 {
					// In case of refund, destChainID will be equal to srcChainID
					// to get the correct token name, original destination chain is needed.
					origDstChainID := common.ToStrChainID(tx.DestinationChainId)
					tokenName := cardanoConfig.GetNativeTokenName(origDstChainID)

					if tokenName == "" {
						return nil, fmt.Errorf("token is not defined for refund original destination chain: (%s -> %s)", srcChainID, origDstChainID) //nolint:lll
					}

					updateMap(receiver.DestinationAddress, tokenName, receiver.AmountWrapped.Uint64())
				}

				for _, utxo := range refundUtxosPerConfirmedTx[txIndx] {
					for _, token := range utxo.Output.Tokens {
						updateMap(receiver.DestinationAddress, token.TokenName(), token.Amount)
					}
				}

			case uint8(common.DefundConfirmedTxType):
				updateMap(receiver.DestinationAddress, cardanowallet.AdaTokenName, amount)

				if receiver.AmountWrapped != nil && receiver.AmountWrapped.Sign() > 0 {
					// defund tx should have correct destination chain id set.
					// this is hacky solution that will work for now
					token, err := cardano.GetNativeTokenFromConfig(cardanoConfig.NativeTokens[0])
					if err != nil {
						return nil, fmt.Errorf("token is not defined for defund: %s", srcChainID)
					}

					updateMap(receiver.DestinationAddress, token.String(), receiver.AmountWrapped.Uint64())
				}

			default:
				updateMap(receiver.DestinationAddress, cardanowallet.AdaTokenName, amount)

				if receiver.AmountWrapped != nil && receiver.AmountWrapped.Sign() > 0 {
					tokenName := cardanoConfig.GetNativeTokenName(srcChainID)
					if tokenName == "" {
						return nil, fmt.Errorf("token is not defined for destination chain: (%s -> %s)", srcChainID, destChainID) //nolint:lll
					}

					updateMap(receiver.DestinationAddress, tokenName, receiver.AmountWrapped.Uint64())
				}
			}
		}
	}

	return receiversMap, nil
}

func getOutputs(
	networkID cardanowallet.CardanoNetworkType, receiversMap map[string]map[string]uint64, logger hclog.Logger,
) cardano.TxOutputs {
	result := cardano.TxOutputs{
		Outputs: make([]cardanowallet.TxOutput, 0, len(receiversMap)),
		Sum:     map[string]uint64{},
	}

	for addr, amountMap := range receiversMap {
		if amountMap[cardanowallet.AdaTokenName] == 0 {
			logger.Warn("skipped output with zero amount", "addr", addr)

			continue
		} else if !cardano.IsValidOutputAddress(addr, networkID) {
			logger.Warn("skipped output because it is invalid", "addr", addr)

			continue
		}

		tokens, _ := cardanowallet.GetTokensFromSumMap(amountMap) // error can not happen here
		if len(tokens) == 0 {
			tokens = nil
		}

		result.Outputs = append(result.Outputs, cardanowallet.TxOutput{
			Addr:   addr,
			Amount: amountMap[cardanowallet.AdaTokenName],
			Tokens: tokens,
		})

		for tokenName, amount := range amountMap {
			result.Sum[tokenName] += amount
		}
	}

	// sort outputs because all batchers should have same order of outputs
	sort.Slice(result.Outputs, func(i, j int) bool {
		return result.Outputs[i].Addr < result.Outputs[j].Addr
	})

	return result
}

func getNeededUtxos(
	txInputOutputs []*indexer.TxInputOutput,
	desiredAmounts map[string]uint64,
	minUtxoLovelaceAmount uint64,
	maxUtxoCount int,
	takeAtLeastUtxoCount int,
) ([]*indexer.TxInputOutput, error) {
	inputUtxos := make([]cardanowallet.Utxo, len(txInputOutputs))

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

	// Change outputs require minUtxoLovelace (protocol rule)
	// Exact spends without change are rare (especially with tokens)
	desiredAmounts[cardanowallet.AdaTokenName] += minUtxoLovelaceAmount

	outputUTXOs, err := txsend.GetUTXOsForAmounts(
		inputUtxos, desiredAmounts, maxUtxoCount, takeAtLeastUtxoCount)
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

func getTxOutputFromSumMap(addr string, sumMap map[string]uint64) (cardanowallet.TxOutput, error) {
	if len(sumMap) == 0 {
		return cardanowallet.NewTxOutput(addr, 0), nil
	}

	tokens := make([]cardanowallet.TokenAmount, 0, len(sumMap)-1)

	for tokenName, amount := range sumMap {
		if tokenName != cardanowallet.AdaTokenName {
			newToken, err := cardanowallet.NewTokenWithFullNameTry(tokenName)
			if err != nil {
				return cardanowallet.TxOutput{}, err
			}

			tokens = append(tokens, cardanowallet.NewTokenAmount(newToken, amount))
		}
	}

	sort.Slice(tokens, func(i, j int) bool {
		return tokens[i].TokenName() < tokens[j].TokenName()
	})

	return cardanowallet.NewTxOutput(addr, sumMap[cardanowallet.AdaTokenName], tokens...), nil
}

func subtractTxOutputsFromSumMap(
	sumMap map[string]uint64, txOutputs []cardanowallet.TxOutput,
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
		updateTokenInMap(cardanowallet.AdaTokenName, out.Amount)

		for _, token := range out.Tokens {
			updateTokenInMap(token.TokenName(), token.Amount)
		}
	}

	return sumMap
}

func calculateMinUtxoLovelaceAmount(
	cardanoCliBinary string, protocolParams []byte,
	addr string, txInputOutputs []*indexer.TxInputOutput, txOutputs []cardanowallet.TxOutput,
) (uint64, error) {
	sumMap := subtractTxOutputsFromSumMap(getSumMapFromTxInputOutput(txInputOutputs), txOutputs)

	tokens, err := cardanowallet.GetTokensFromSumMap(sumMap)
	if err != nil {
		return 0, err
	}

	txBuilder, err := cardanowallet.NewTxBuilder(cardanoCliBinary)
	if err != nil {
		return 0, err
	}

	defer txBuilder.Dispose()

	minUtxo, err := txBuilder.SetProtocolParameters(protocolParams).CalculateMinUtxo(cardanowallet.TxOutput{
		Addr:   addr,
		Amount: sumMap[cardanowallet.AdaTokenName],
		Tokens: tokens,
	})
	if err != nil {
		return 0, err
	}

	return minUtxo, nil
}

func convertUTXOsToTxInputs(utxos []*indexer.TxInputOutput) (result cardanowallet.TxInputs) {
	result.Inputs = make([]cardanowallet.TxInput, len(utxos))
	result.Sum = make(map[string]uint64)

	for i, utxo := range utxos {
		result.Inputs[i] = cardanowallet.TxInput{
			Hash:  utxo.Input.Hash.String(),
			Index: utxo.Input.Index,
		}

		result.Sum[cardanowallet.AdaTokenName] += utxo.Output.Amount

		for _, token := range utxo.Output.Tokens {
			result.Sum[token.TokenName()] += token.Amount
		}
	}

	return result
}

func extractStakeKeyDepositAmount(protocolParams []byte) (uint64, error) {
	var params map[string]interface{}

	if err := json.Unmarshal(protocolParams, &params); err != nil {
		return 0, err
	}

	// Extract stakeAddressDeposit value
	if stakeDeposit, exists := params[StakeDepositFieldName]; exists {
		// Handle different number types that JSON might unmarshal to
		switch v := stakeDeposit.(type) {
		case float64:
			return uint64(v), nil
		case uint64:
			return v, nil
		case int:
			if v < 0 {
				return 0, fmt.Errorf("cannot convert negative int %d to uint64", v)
			}

			return uint64(v), nil
		case string:
			// If it's a string, try to parse it as a number
			var result uint64
			_, err := fmt.Sscanf(v, "%d", &result)

			return result, err
		default:
			return 0, fmt.Errorf("%s has unexpected type: %T", StakeDepositFieldName, stakeDeposit)
		}
	}

	return 0, fmt.Errorf("%s field not found in protocol parameters", StakeDepositFieldName)
}
