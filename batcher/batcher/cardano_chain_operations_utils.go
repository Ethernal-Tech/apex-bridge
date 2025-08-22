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

func getOutputs(
	txs []eth.ConfirmedTransaction, cardanoConfig *cardano.CardanoChainConfig, logger hclog.Logger,
) (cardano.TxOutputs, error) {
	receiversMap := map[string]cardanowallet.TxOutput{}

	for _, transaction := range txs {
		// stake delegation tx are not processed in this way
		if transaction.TransactionType == uint8(common.StakeDelConfirmedTxType) {
			continue
		}

		for _, receiver := range transaction.Receivers {
			data := receiversMap[receiver.DestinationAddress]
			data.Amount += receiver.Amount.Uint64()

			if receiver.AmountWrapped != nil && receiver.AmountWrapped.Sign() > 0 {
				if len(data.Tokens) == 0 {
					var (
						err   error
						token cardanowallet.Token
					)

					if (transaction.TransactionType == uint8(common.DefundConfirmedTxType)) ||
						(transaction.TransactionType == uint8(common.RefundConfirmedTxType)) {
						token, err = cardano.GetNativeTokenFromConfig(cardanoConfig.NativeTokens[0])
						if err != nil {
							return cardano.TxOutputs{}, err
						}
					} else {
						token, err = cardanoConfig.GetNativeToken(
							common.ToStrChainID(transaction.SourceChainId))
						if err != nil {
							return cardano.TxOutputs{}, err
						}
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

// createUtxoSelectionAmounts creates a copy of desired amounts with adjusted ADA amount for UTXO selection
func createUtxoSelectionAmounts(desiredAmounts map[string]uint64, minUtxoLovelaceAmount uint64) map[string]uint64 {
	utxoSelectionAmounts := make(map[string]uint64, len(desiredAmounts))
	for k, v := range desiredAmounts {
		utxoSelectionAmounts[k] = v
	}
	utxoSelectionAmounts[cardanowallet.AdaTokenName] += minUtxoLovelaceAmount
	return utxoSelectionAmounts
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

	// Create a copy of desired amounts for UTXO selection so we don't modify the original map
	// We need to include minUtxoLovelace for change outputs (protocol rule)
	utxoSelectionAmounts := createUtxoSelectionAmounts(desiredAmounts, minUtxoLovelaceAmount)

	if maxUtxoCount == 0 {
		return nil, fmt.Errorf(
			"%w: maxUtxoCount equal to 0", cardanowallet.ErrUTXOsLimitReached)
	}

	outputUTXOs, err := txsend.GetUTXOsForAmounts(
		inputUtxos, utxoSelectionAmounts, maxUtxoCount, takeAtLeastUtxoCount)
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
	var params cardanowallet.ProtocolParameters

	if err := json.Unmarshal(protocolParams, &params); err != nil {
		return 0, err
	}

	return params.StakeAddressDeposit, nil
}

type AddressConsolidation struct {
	Address      string
	AddressIndex uint8
	UtxoCount    int
	Share        float64
	Assigned     int
	Remainder    float64
	IsFee        bool
}

type AddressConsolidationData struct {
	Address      string
	AddressIndex uint8
	UtxoCount    int
	IsFee        bool
	Utxos        []*indexer.TxInputOutput
}

// Chose inputs for consolidation proportionally depending of how many there are for every address
// and the max number allowed.
func allocateInputsForConsolidation(inputs []AddressConsolidationData, maxUtxoCount int) []AddressConsolidationData {
	total := 0
	for _, input := range inputs {
		total += input.UtxoCount
	}

	n := len(inputs)
	alloc := make([]AddressConsolidation, n)
	result := make([]AddressConsolidationData, n)

	if total <= maxUtxoCount {
		for i, input := range inputs {
			result[i] = AddressConsolidationData{
				Address:      input.Address,
				AddressIndex: input.AddressIndex,
				UtxoCount:    input.UtxoCount,
				IsFee:        input.IsFee,
			}
		}

		return result
	}

	assigned := 0

	// First, assign the integer part of the proportional share
	for i, input := range inputs {
		share := (float64(input.UtxoCount) / float64(total)) * float64(maxUtxoCount)
		alloc[i] = AddressConsolidation{
			Address:      input.Address,
			AddressIndex: input.AddressIndex,
			UtxoCount:    input.UtxoCount,
			Share:        share,
			Assigned:     int(share),
			Remainder:    share - float64(int(share)),
			IsFee:        input.IsFee,
		}
		assigned += alloc[i].Assigned
	}

	// Assign remaining utxos using the largest remainders
	remaining := maxUtxoCount - assigned
	if remaining > 0 {
		sort.SliceStable(alloc, func(i, j int) bool {
			return alloc[i].Remainder > alloc[j].Remainder
		})

		for i := 0; i < remaining; i++ {
			alloc[i%len(alloc)].Assigned += 1
		}
	}

	// Prepare the result
	maxIndex := 0
	maxAssigned := 0
	feeIndex := -1

	for i, input := range alloc {
		result[i] = AddressConsolidationData{
			Address:      input.Address,
			AddressIndex: input.AddressIndex,
			UtxoCount:    input.Assigned,
			IsFee:        input.IsFee,
		}

		if !input.IsFee && input.Assigned > maxAssigned {
			maxIndex = i
			maxAssigned = input.Assigned
		}

		if input.IsFee && input.Assigned == 0 {
			feeIndex = i
		}
	}

	if feeIndex != -1 {
		for {
			if result[feeIndex].UtxoCount < inputs[0].UtxoCount {
				result[maxIndex].UtxoCount -= 1
				result[feeIndex].UtxoCount += 1
			} else {
				break
			}

			// TODO: Not sure how to update this, it's not the best way to do it
			// Should we have the PotentialFeeDefault + MinUtxoAmountDefault?
			if calcualteUtxoSum(inputs[0].Utxos[:result[feeIndex].UtxoCount]) >= 2*common.MinUtxoAmountDefault {
				break
			}
		}
	}

	return result
}

func calcualteUtxoSum(inputs []*indexer.TxInputOutput) uint64 {
	sum := uint64(0)
	for _, input := range inputs {
		sum += input.Output.Amount
	}

	return sum
}
