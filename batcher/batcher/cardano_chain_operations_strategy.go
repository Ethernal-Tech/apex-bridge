package batcher

import (
	"fmt"
	"sort"

	cardano "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
)

type ICardanoChainOperationsStrategy interface {
	GetOutputs(
		txs []eth.ConfirmedTransaction, cardanoConfig *cardano.CardanoChainConfig,
		destChainID string, logger hclog.Logger,
	) (*cardano.TxOutputs, error)
	GetNeededUtxos(
		inputUTXOs []*indexer.TxInputOutput,
		desiredAmount map[string]uint64,
		minUtxoAmount uint64,
		utxoCount int,
		maxUtxoCount int,
		takeAtLeastUtxoCount int,
	) (chosenUTXOs []*indexer.TxInputOutput, err error)
	FindMinUtxo(utxos []*indexer.TxInputOutput) (*indexer.TxInputOutput, int)
}

type CardanoChainOperationReactorStrategy struct {
}

func (s *CardanoChainOperationReactorStrategy) GetOutputs(
	txs []eth.ConfirmedTransaction, cardanoConfig *cardano.CardanoChainConfig, _ string, logger hclog.Logger,
) (*cardano.TxOutputs, error) {
	receiversMap := map[string]uint64{}

	for _, transaction := range txs {
		for _, receiver := range transaction.Receivers {
			receiversMap[receiver.DestinationAddress] += receiver.Amount.Uint64()
		}
	}

	result := cardano.TxOutputs{
		Outputs: make([]cardanowallet.TxOutput, 0, len(receiversMap)),
		Sum:     map[string]uint64{},
	}

	for addr, amount := range receiversMap {
		if amount == 0 {
			logger.Warn("skipped output with zero amount", "addr", addr)

			continue
		} else if !cardano.IsValidOutputAddress(addr, cardanoConfig.NetworkID) {
			// apex-361 fix
			logger.Warn("skipped output because it is invalid", "addr", addr)

			continue
		}

		result.Outputs = append(result.Outputs, cardanowallet.TxOutput{
			Addr:   addr,
			Amount: amount,
		})
		result.Sum[cardanowallet.AdaTokenName] += amount
	}

	// sort outputs because all batchers should have same order of outputs
	sort.Slice(result.Outputs, func(i, j int) bool {
		return result.Outputs[i].Addr < result.Outputs[j].Addr
	})

	return &result, nil
}

// getNeededUtxos returns only needed input utxos
// It is expected that UTXOs are sorted by their Block Slot number (for example: returned sorted by db.GetAllTxOutput)
// and taken from first to last until desiredAmount has been met or maxUtxoCount reached
// if desiredAmount has been met, tx is created regularly
// if maxUtxoCount has been reached, we replace smallest UTXO with first next bigger one until we reach desiredAmount
func (s *CardanoChainOperationReactorStrategy) GetNeededUtxos(
	inputUTXOs []*indexer.TxInputOutput,
	desiredAmount map[string]uint64,
	minUtxoAmount uint64,
	utxoCount int,
	maxUtxoCount int,
	takeAtLeastUtxoCount int,
) (chosenUTXOs []*indexer.TxInputOutput, err error) {
	inputUTXOs = filterOutTokenUtxos(inputUTXOs)
	lovelaceDesiredAmount := desiredAmount[cardanowallet.AdaTokenName]
	// if we have change then it must be greater than this amount
	txCostWithMinChange := minUtxoAmount + lovelaceDesiredAmount

	// algorithm that chooses multisig UTXOs
	chosenUTXOsSum := uint64(0)
	isUtxosOk := false

	for i, utxo := range inputUTXOs {
		chosenUTXOs = append(chosenUTXOs, utxo)
		utxoCount++

		chosenUTXOsSum += utxo.Output.Amount // in cardano we should not care about overflow

		if utxoCount > maxUtxoCount {
			minChosenUTXO, minChosenUTXOIdx := s.FindMinUtxo(chosenUTXOs)

			chosenUTXOs[minChosenUTXOIdx] = utxo
			chosenUTXOsSum -= minChosenUTXO.Output.Amount
			chosenUTXOs = chosenUTXOs[:len(chosenUTXOs)-1]
			utxoCount--
		}

		if chosenUTXOsSum >= txCostWithMinChange || chosenUTXOsSum == lovelaceDesiredAmount {
			isUtxosOk = true

			// try to add utxos until we reach tryAtLeastUtxoCount
			cnt := min(
				len(inputUTXOs)-i-1,                   // still available in inputUTXOs
				takeAtLeastUtxoCount-len(chosenUTXOs), // needed to fill tryAtLeastUtxoCount
				maxUtxoCount-utxoCount,                // maxUtxoCount limit must be preserved
			)
			if cnt > 0 {
				chosenUTXOs = append(chosenUTXOs, inputUTXOs[i+1:i+1+cnt]...)
			}

			break
		}
	}

	if !isUtxosOk {
		return nil, fmt.Errorf("fatal error, couldn't select UTXOs for sum: %v", desiredAmount)
	}

	return chosenUTXOs, nil
}

func (s *CardanoChainOperationReactorStrategy) FindMinUtxo(
	utxos []*indexer.TxInputOutput,
) (*indexer.TxInputOutput, int) {
	mininimal := utxos[0]
	idx := 0

	for i, utxo := range utxos[1:] {
		if utxo.Output.Amount < mininimal.Output.Amount {
			mininimal = utxo
			idx = i + 1
		}
	}

	return mininimal, idx
}

type CardanoChainOperationSkylineStrategy struct {
}

func (s *CardanoChainOperationSkylineStrategy) GetOutputs(
	txs []eth.ConfirmedTransaction, cardanoConfig *cardano.CardanoChainConfig,
	destChainID string, logger hclog.Logger,
) (*cardano.TxOutputs, error) {
	receiversMap := map[string]cardanowallet.TxOutput{}

	for _, transaction := range txs {
		for _, receiver := range transaction.Receivers {
			data := receiversMap[receiver.DestinationAddress]
			data.Amount += receiver.Amount.Uint64()

			if receiver.AmountWrapped != nil {
				if len(data.Tokens) == 0 {
					tconf := getConfigTokenExchange(destChainID, true, cardanoConfig.Destinations)
					token, err := cardanowallet.NewTokenAmountWithFullName(tconf.DstTokenName, 0, true)

					if err != nil {
						return nil, fmt.Errorf("failed to create new token amount")
					}

					data.Tokens = []cardanowallet.TokenAmount{token}
				}

				data.Tokens[0].Amount += receiver.AmountWrapped.Uint64()
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

	return &result, nil
}

// getNeededSkylineUtxos returns only needed input utxos
// It is expected that UTXOs are sorted by their Block Slot number (for example: returned sorted by db.GetAllTxOutput)
// and taken from first to last until desiredAmount has been met or maxUtxoCount reached
// if desiredAmount has been met, tx is created regularly
// if maxUtxoCount has been reached, we replace smallest UTXO with first next bigger one until we reach desiredAmount
func (s *CardanoChainOperationSkylineStrategy) GetNeededUtxos(
	inputUTXOs []*indexer.TxInputOutput,
	desiredAmount map[string]uint64,
	minUtxoAmount uint64,
	utxoCount int,
	maxUtxoCount int,
	takeAtLeastUtxoCount int,
) (chosenUTXOs []*indexer.TxInputOutput, err error) {
	chosenUTXOsSum := map[string]uint64{}
	isUtxosOk := false
	txCostWithMinChange := map[string]uint64{}

	for chainName, desiredValue := range desiredAmount {
		if chainName == cardanowallet.AdaTokenName {
			// if we have change then it must be greater than this amount
			txCostWithMinChange[chainName] = desiredValue + minUtxoAmount
		} else {
			txCostWithMinChange[chainName] = desiredValue
		}
	}

	for i, utxo := range inputUTXOs {
		chosenUTXOs = append(chosenUTXOs, utxo)

		utxoCount++
		chosenUTXOsSum[cardanowallet.AdaTokenName] += utxo.Output.Amount // in cardano we should not care about overflow

		if len(utxo.Output.Tokens) > 0 {
			chosenUTXOsSum[utxo.Output.Tokens[0].Name] += utxo.Output.Tokens[0].Amount
		}

		var minChosenUTXO *indexer.TxInputOutput

		var minChosenUTXOIdx int

		if utxoCount > maxUtxoCount {
			// if len(utxo.Output.Tokens) == 0 {
			// 	minChosenUTXO, minChosenUTXOIdx = findMinUtxo(chosenUTXOs)
			// } else {
			// 	minChosenUTXO, minChosenUTXOIdx = findMinWrappedUtxo(chosenUTXOs)
			// }
			minChosenUTXO, minChosenUTXOIdx = s.FindMinUtxo(chosenUTXOs)

			chosenUTXOsSum[cardanowallet.AdaTokenName] -= minChosenUTXO.Output.Amount

			if len(minChosenUTXO.Output.Tokens) > 0 {
				chosenUTXOsSum[minChosenUTXO.Output.Tokens[0].Name] -= minChosenUTXO.Output.Tokens[0].Amount
			}

			chosenUTXOs[minChosenUTXOIdx] = utxo
			chosenUTXOs = chosenUTXOs[:len(chosenUTXOs)-1]
			utxoCount--
		}

		achievedDesiredAmountsCount := 0

		for tokenName, amount := range desiredAmount {
			if chosenUTXOsSum[tokenName] >= txCostWithMinChange[tokenName] ||
				chosenUTXOsSum[tokenName] == amount {
				achievedDesiredAmountsCount++
			}
		}

		if achievedDesiredAmountsCount == len(desiredAmount) {
			isUtxosOk = true

			// try to add utxos until we reach tryAtLeastUtxoCount
			cnt := min(
				len(inputUTXOs)-i-1,                   // still available in inputUTXOs
				takeAtLeastUtxoCount-len(chosenUTXOs), // needed to fill tryAtLeastUtxoCount
				maxUtxoCount-utxoCount,                // maxUtxoCount limit must be preserved
			)
			if cnt > 0 {
				chosenUTXOs = append(chosenUTXOs, inputUTXOs[i+1:i+1+cnt]...)
			}

			break
		}
	}

	if !isUtxosOk {
		return nil, fmt.Errorf("fatal error, couldn't select UTXOs for sum: %v", desiredAmount)
	}

	return chosenUTXOs, nil
}

func (s *CardanoChainOperationSkylineStrategy) FindMinUtxo(
	utxos []*indexer.TxInputOutput,
) (*indexer.TxInputOutput, int) {
	minimal := utxos[0]
	idx := 0

	for i, utxo := range utxos[1:] {
		if len(utxo.Output.Tokens) > 0 && len(minimal.Output.Tokens) > 0 {
			if utxo.Output.Tokens[0].Amount < minimal.Output.Tokens[0].Amount {
				minimal = utxo
				idx = i + 1
			}
		} else {
			if utxo.Output.Amount < minimal.Output.Amount {
				minimal = utxo
				idx = i + 1
			}
		}
	}

	return minimal, idx
}
