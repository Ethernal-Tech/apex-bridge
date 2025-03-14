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

type ICardanoChainOperationsStrategy interface {
	GetOutputs(
		txs []eth.ConfirmedTransaction,
		cardanoConfig *cardano.CardanoChainConfig,
		logger hclog.Logger,
	) (cardano.TxOutputs, error)
	FilterUtxos(
		multisigUtxos, feeUtxos []*indexer.TxInputOutput, config *cardano.CardanoChainConfig,
	) ([]*indexer.TxInputOutput, []*indexer.TxInputOutput, error)
}

var (
	_ ICardanoChainOperationsStrategy = (*CardanoChainOperationReactorStrategy)(nil)
	_ ICardanoChainOperationsStrategy = (*CardanoChainOperationSkylineStrategy)(nil)
)

type CardanoChainOperationReactorStrategy struct {
}

func (s *CardanoChainOperationReactorStrategy) GetOutputs(
	txs []eth.ConfirmedTransaction, cardanoConfig *cardano.CardanoChainConfig, logger hclog.Logger,
) (cardano.TxOutputs, error) {
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

	return result, nil
}

func (s *CardanoChainOperationReactorStrategy) FilterUtxos(
	multisigUtxos, feeUtxos []*indexer.TxInputOutput, _ *cardano.CardanoChainConfig,
) ([]*indexer.TxInputOutput, []*indexer.TxInputOutput, error) {
	return filterOutUtxosWithUnknownTokens(multisigUtxos), filterOutUtxosWithUnknownTokens(feeUtxos), nil
}

type CardanoChainOperationSkylineStrategy struct {
}

func (s *CardanoChainOperationSkylineStrategy) GetOutputs(
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

func (s *CardanoChainOperationSkylineStrategy) FilterUtxos(
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
