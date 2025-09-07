package batcher

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"sort"

	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	cardano "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	txsend "github.com/Ethernal-Tech/cardano-infrastructure/sendtx"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
)

func getStakingCertificates(
	cardanoCliBinary string,
	data *batchInitialData,
	tx *eth.ConfirmedTransaction,
	policyScript *cardanowallet.PolicyScript,
	multisigStakeAddress string,
) (*cardano.CertificatesWithScript, uint64, error) {
	cliUtils := cardanowallet.NewCliUtils(cardanoCliBinary)

	certs := make([]cardanowallet.ICertificate, 0)

	// Generate certificates
	keyRegDepositAmount, err := extractStakeKeyDepositAmount(data.ProtocolParams)
	if err != nil {
		return nil, 0, err
	}

	if tx.TransactionSubType == uint8(common.StakeRegDelConfirmedTxSubType) {
		registrationCert, err := cliUtils.CreateRegistrationCertificate(multisigStakeAddress, keyRegDepositAmount)
		if err != nil {
			return nil, 0, errors.Join(errSkipConfirmedTx, err)
		}

		certs = append(certs, registrationCert)
	}

	if tx.TransactionSubType == uint8(common.StakeRegDelConfirmedTxSubType) ||
		tx.TransactionSubType == uint8(common.StakeDelConfirmedTxSubType) {
		delegationCert, err := cliUtils.CreateDelegationCertificate(multisigStakeAddress, tx.StakePoolId)
		if err != nil {
			return nil, 0, errors.Join(errSkipConfirmedTx, err)
		}

		certs = append(certs, delegationCert)
	}

	if tx.TransactionSubType == uint8(common.StakeDeregConfirmedTxSubType) {
		deregCert, err := cliUtils.CreateDeregistrationCertificate(multisigStakeAddress)
		if err != nil {
			return nil, 0, errors.Join(errSkipConfirmedTx, err)
		}

		certs = append(certs, deregCert)
	}

	return &cardano.CertificatesWithScript{
		PolicyScript: policyScript,
		Certificates: certs,
	}, keyRegDepositAmount, nil
}

func getOutputs(
	txs []eth.ConfirmedTransaction, cardanoConfig *cardano.CardanoChainConfig, logger hclog.Logger,
) (cardano.TxOutputs, bool, error) {
	receiversMap := map[string]cardanowallet.TxOutput{}
	isRedistribution := false

	for _, transaction := range txs {
		// stake delegation tx are not processed in this way
		if transaction.TransactionType == uint8(common.StakeConfirmedTxType) {
			continue
		}

		if transaction.TransactionType == uint8(common.RedistributionConfirmedTxType) {
			logger.Debug("Triggered token redistribution", "chain", transaction.DestinationChainId)

			isRedistribution = true

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
							return cardano.TxOutputs{}, false, err
						}
					} else {
						token, err = cardanoConfig.GetNativeToken(
							common.ToStrChainID(transaction.SourceChainId))
						if err != nil {
							return cardano.TxOutputs{}, false, err
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

	return result, isRedistribution, nil
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

func addRedistributionOutputs(
	outputs []cardanowallet.TxOutput,
	multisigAddresses []common.AddressAndAmount,
) ([]cardanowallet.TxOutput, error) {
	for _, addr := range multisigAddresses {
		output, err := getTxOutputFromSumMap(addr.Address, addr.TokensAmounts)
		if err != nil {
			return nil, err
		}

		outputs = append(outputs, output)
	}

	return outputs, nil
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

type addressConsolidation struct {
	Address      string
	AddressIndex uint8
	UtxoCount    int
	Share        float64
	Assigned     int
	Remainder    float64
	Utxos        []*indexer.TxInputOutput
}

type AddressConsolidationData struct {
	Address      string
	AddressIndex uint8
	UtxoCount    int
	Utxos        []*indexer.TxInputOutput
}

// Chose inputs for consolidation proportionally depending of how many there are for every address
// and the max number allowed.
func allocateInputsForConsolidation(
	inputs []AddressConsolidationData,
	maxUtxoCount int,
	totalNumberOfUtxos int,
	consolidationType core.ConsolidationType,
) ([]AddressConsolidationData, error) {
	if consolidationType == core.ConsolidationTypeToZeroAddress {
		return sequentialUtxoSelectionForConsolidation(inputs, maxUtxoCount), nil
	}

	n := len(inputs)
	alloc := make([]addressConsolidation, 0, n)

	if maxUtxoCount > totalNumberOfUtxos {
		maxUtxoCount = totalNumberOfUtxos
	}

	inputsWorkingSet := make([]AddressConsolidationData, 0, n)

	for _, input := range inputs {
		if input.UtxoCount > 1 {
			inputsWorkingSet = append(inputsWorkingSet, input)
		}
	}

	for {
		assigned := 0

		// First, assign the integer part of the proportional share
		for _, input := range inputsWorkingSet {
			share := (float64(input.UtxoCount) / float64(totalNumberOfUtxos)) * float64(maxUtxoCount)
			alloc = append(alloc, addressConsolidation{
				Address:      input.Address,
				AddressIndex: input.AddressIndex,
				UtxoCount:    input.UtxoCount,
				Share:        share,
				Assigned:     int(share),
				Remainder:    share - float64(int(share)),
				Utxos:        input.Utxos,
			})
			assigned += int(share)
		}

		if len(alloc) == 0 {
			return nil, fmt.Errorf("no elements found in addresses allocated for consolidation")
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

		indexesToRemove := make([]uint8, 0, len(inputsWorkingSet))

		for _, a := range alloc {
			if a.Assigned == 1 {
				indexesToRemove = append(indexesToRemove, a.AddressIndex)
			}
		}

		if len(indexesToRemove) == 0 {
			break
		}

		newInputsWorkingSet := make([]AddressConsolidationData, 0, len(inputsWorkingSet)-len(indexesToRemove))

		for _, input := range inputsWorkingSet {
			if !slices.Contains(indexesToRemove, input.AddressIndex) {
				newInputsWorkingSet = append(newInputsWorkingSet, input)
			}
		}

		inputsWorkingSet = newInputsWorkingSet
		alloc = make([]addressConsolidation, 0, n)
	}

	return generateAllocateInputsForConsolidationOutput(alloc), nil
}

// takes as many UTXOs as possible starting from the address with the most UTXOs
func sequentialUtxoSelectionForConsolidation(inputs []AddressConsolidationData, maxUtxoCount int,
) []AddressConsolidationData {
	sort.SliceStable(inputs, func(i, j int) bool {
		return inputs[i].UtxoCount > inputs[j].UtxoCount
	})

	alloc := make([]addressConsolidation, 0)

	var utxoCnt int

	for _, input := range inputs {
		if maxUtxoCount > 0 {
			if input.UtxoCount <= maxUtxoCount {
				utxoCnt = input.UtxoCount
				maxUtxoCount -= input.UtxoCount
			} else {
				// Take as many as we can from this address
				utxoCnt = maxUtxoCount
				maxUtxoCount = 0
			}

			alloc = append(alloc, addressConsolidation{
				Address:      input.Address,
				AddressIndex: input.AddressIndex,
				UtxoCount:    utxoCnt,
				Share:        0,
				Assigned:     utxoCnt,
				Remainder:    0,
				Utxos:        input.Utxos[:utxoCnt],
			})
		}
	}

	return generateAllocateInputsForConsolidationOutput(alloc)
}

func generateAllocateInputsForConsolidationOutput(alloc []addressConsolidation) []AddressConsolidationData {
	result := make([]AddressConsolidationData, 0)

	for _, input := range alloc {
		if input.Assigned > 0 {
			result = append(result, AddressConsolidationData{
				Address:      input.Address,
				AddressIndex: input.AddressIndex,
				UtxoCount:    input.Assigned,
				Utxos:        input.Utxos[:input.Assigned],
			})
		}
	}

	return result
}
