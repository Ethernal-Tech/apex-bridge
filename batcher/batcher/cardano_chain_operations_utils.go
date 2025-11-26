package batcher

import (
	"encoding/json"
	"errors"
	"fmt"
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

	certs := make([]cardanowallet.ICardanoArtifact, 0)

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
		deregCert, err := cliUtils.CreateDeregistrationCertificate(multisigStakeAddress, keyRegDepositAmount)
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

type getOutputsData struct {
	TxOutputs        cardano.TxOutputs
	IsRedistribution bool
	MintTokens       []cardanowallet.MintTokenAmount
}

func getOutputs(
	txs []eth.ConfirmedTransaction,
	cardanoConfig *cardano.CardanoChainConfig,
	feeAddress string,
	refundUtxosPerConfirmedTx [][]*indexer.TxInputOutput,
	logger hclog.Logger,
) (*getOutputsData, error) {
	receiversMap := map[string]cardanowallet.TxOutput{}
	mintTokens := make([]cardanowallet.MintTokenAmount, 0)
	isRedistribution := false

	for txIndex, transaction := range txs {
		// stake delegation tx are not processed in this way
		if transaction.TransactionType == uint8(common.StakeConfirmedTxType) {
			continue
		}

		if transaction.TransactionType == uint8(common.RedistributionConfirmedTxType) {
			logger.Debug("Triggered token redistribution", "chain", transaction.DestinationChainId)

			isRedistribution = true

			continue
		}

		// refund for unknown tokens will be handled later
		if transaction.TransactionType == uint8(common.RefundConfirmedTxType) && len(refundUtxosPerConfirmedTx[txIndex]) > 0 {
			continue
		}

		for _, receiver := range transaction.ReceiversWithToken {
			hasTokens := receiver.AmountWrapped != nil && receiver.AmountWrapped.Sign() > 0

			data := receiversMap[receiver.DestinationAddress]
			if transaction.TransactionType != uint8(common.RefundConfirmedTxType) {
				data.Amount += receiver.Amount.Uint64()
			} else {
				minBridgingFee := cardanoConfig.GetMinBridgingFee(hasTokens)

				data.Amount += receiver.Amount.Uint64() - minBridgingFee

				feeData := receiversMap[feeAddress]
				feeData.Amount += minBridgingFee
				receiversMap[feeAddress] = feeData
			}

			if hasTokens {
				if len(data.Tokens) == 0 {
					var (
						err        error
						token      cardanowallet.Token
						shouldMint bool
					)

					switch transaction.TransactionType {
					case uint8(common.DefundConfirmedTxType):
						token, err = cardanoConfig.GetTokenByID(receiver.TokenId)
						if err != nil {
							return nil, err
						}
					case uint8(common.RefundConfirmedTxType):
						//origDstChainID := common.ToStrChainID(transaction.DestinationChainId)

						//token, err = cardanoConfig.GetNativeToken(origDstChainID)
						token, err = cardanoConfig.GetTokenByID(receiver.TokenId)
						if err != nil {
							return nil, err
						}
					default:
						token, shouldMint, err = cardanoConfig.GetTokenData(
							receiver.TokenId)
						if err != nil {
							return nil, err
						}
					}

					if shouldMint {
						mintTokens = append(mintTokens, cardanowallet.MintTokenAmount{
							Token:  token,
							Amount: receiver.AmountWrapped.Uint64(),
						})
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
	sort.SliceStable(result.Outputs, func(i, j int) bool {
		return result.Outputs[i].Addr < result.Outputs[j].Addr
	})

	return &getOutputsData{
		TxOutputs:        result,
		IsRedistribution: isRedistribution,
		MintTokens:       mintTokens,
	}, nil
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

	sort.SliceStable(tokens, func(i, j int) bool {
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

func extractPlutusExecutionParams(protocolParams []byte) (cardano.ExecutionUnitData, error) {
	var params cardanowallet.ProtocolParameters

	if err := json.Unmarshal(protocolParams, &params); err != nil {
		return cardano.ExecutionUnitData{}, err
	}

	return cardano.ExecutionUnitData{
		CollateralPercentage: params.CollateralPercentage,
		ExecutionUnitPrices:  params.ExecutionUnitPrices,
		MaxTxExecutionUnits:  params.MaxTxExecutionUnits,
	}, nil
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

	if maxUtxoCount > totalNumberOfUtxos {
		maxUtxoCount = totalNumberOfUtxos
	}

	inputsWorkingSet := make([]AddressConsolidationData, 0, len(inputs))

	for _, input := range inputs {
		if input.UtxoCount > 1 {
			inputsWorkingSet = append(inputsWorkingSet, input)
		}
	}

	var alloc []addressConsolidation

	for {
		var remainingUtxosNum int
		alloc, remainingUtxosNum = calculateInitialAllocations(inputsWorkingSet, totalNumberOfUtxos, maxUtxoCount)

		if len(alloc) == 0 {
			return nil, fmt.Errorf("no elements found in addresses allocated for consolidation")
		}

		distributeRemainders(alloc, remainingUtxosNum)

		newInputsWorkingSet := filterWorkingSet(alloc)
		if len(newInputsWorkingSet) == len(alloc) {
			break
		}

		inputsWorkingSet = newInputsWorkingSet
	}

	return generateAllocateInputsForConsolidationOutput(alloc), nil
}

// takes as many UTXOs as possible starting from the address with the most UTXOs
func sequentialUtxoSelectionForConsolidation(inputs []AddressConsolidationData, maxUtxoCount int,
) []AddressConsolidationData {
	sort.SliceStable(inputs, func(i, j int) bool {
		return inputs[i].UtxoCount > inputs[j].UtxoCount
	})

	var (
		alloc   = make([]addressConsolidation, 0)
		utxoCnt int
	)

	for _, input := range inputs {
		if maxUtxoCount == 0 {
			break
		}

		utxoCnt = min(input.UtxoCount, maxUtxoCount)
		maxUtxoCount -= utxoCnt

		alloc = append(alloc, addressConsolidation{
			Address:      input.Address,
			AddressIndex: input.AddressIndex,
			UtxoCount:    utxoCnt,
			Assigned:     utxoCnt,
			Utxos:        input.Utxos[:utxoCnt],
		})
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

func calculateInitialAllocations(
	inputsWorkingSet []AddressConsolidationData,
	totalNumberOfUtxos, maxUtxoCount int,
) ([]addressConsolidation, int) {
	alloc := make([]addressConsolidation, 0, len(inputsWorkingSet))

	assigned := 0

	// First, assign the integer part of the proportional share
	for _, input := range inputsWorkingSet {
		share := float64(input.UtxoCount) / float64(totalNumberOfUtxos) * float64(maxUtxoCount)
		intShare := int(share)

		alloc = append(alloc, addressConsolidation{
			Address:      input.Address,
			AddressIndex: input.AddressIndex,
			UtxoCount:    input.UtxoCount,
			Share:        share,
			Assigned:     intShare,
			Remainder:    share - float64(intShare),
			Utxos:        input.Utxos,
		})
		assigned += intShare
	}

	return alloc, maxUtxoCount - assigned
}

func distributeRemainders(alloc []addressConsolidation, remainingUtxosNum int) {
	if remainingUtxosNum <= 0 {
		return
	}

	sort.SliceStable(alloc, func(i, j int) bool {
		return alloc[i].Remainder > alloc[j].Remainder
	})

	for i := range remainingUtxosNum {
		idx := i % len(alloc)
		if alloc[idx].Assigned < alloc[idx].UtxoCount {
			alloc[idx].Assigned++
		}
	}
}

func filterWorkingSet(alloc []addressConsolidation) []AddressConsolidationData {
	newInputsWorkingSet := make([]AddressConsolidationData, 0, len(alloc))

	for _, a := range alloc {
		if a.Assigned > 1 {
			newInputsWorkingSet = append(newInputsWorkingSet, AddressConsolidationData{
				Address:      a.Address,
				AddressIndex: a.AddressIndex,
				UtxoCount:    a.UtxoCount,
				Utxos:        a.Utxos,
			})
		}
	}

	return newInputsWorkingSet
}
