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
	) (cardano.TxOutputs, uint64, error)
	GetUTXOs(
		multisigAddress,
		multisigFeeAddress string,
		txOutputs cardano.TxOutputs,
		tokenHoldingOutputs uint64,
		destChainID string,
		cardanoConfig *cardano.CardanoChainConfig,
		db indexer.Database,
		logger hclog.Logger,
	) (multisigUtxos []*indexer.TxInputOutput, feeUtxos []*indexer.TxInputOutput, err error)
	FilterUTXOsForConsolidation(
		multisigUtxos, feeUtxos []*indexer.TxInputOutput,
		_ *cardano.CardanoChainConfig,
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
) (cardano.TxOutputs, uint64, error) {
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

	return result, 0, nil
}

func (s *CardanoChainOperationReactorStrategy) GetUTXOs(
	multisigAddress, multisigFeeAddress string, txOutputs cardano.TxOutputs, _ uint64, _ string,
	cardanoConfig *cardano.CardanoChainConfig, db indexer.Database, logger hclog.Logger,
) (multisigUtxos []*indexer.TxInputOutput, feeUtxos []*indexer.TxInputOutput, err error) {
	multisigUtxos, err = db.GetAllTxOutputs(multisigAddress, true)
	if err != nil {
		return
	}

	feeUtxos, err = db.GetAllTxOutputs(multisigFeeAddress, true)
	if err != nil {
		return
	}

	feeUtxos = filterOutUtxosWithUnknownTokens(feeUtxos)

	if len(feeUtxos) == 0 {
		return nil, nil, fmt.Errorf("fee multisig does not have any utxo: %s", multisigFeeAddress)
	}

	logger.Debug("UTXOs retrieved",
		"multisig", multisigAddress, "utxos", multisigUtxos, "fee", multisigFeeAddress, "utxos", feeUtxos)

	feeUtxos = feeUtxos[:min(cardanoConfig.MaxFeeUtxoCount, len(feeUtxos))] // do not take more than maxFeeUtxoCount

	multisigUtxos, err = s.getNeededUtxos(
		multisigUtxos,
		txOutputs.Sum,
		cardanoConfig.UtxoMinAmount,
		cardanoConfig.MaxUtxoCount-len(feeUtxos),
		0,
		cardanoConfig.TakeAtLeastUtxoCount,
	)
	if err != nil {
		return
	}

	logger.Debug("UTXOs chosen", "multisig", multisigUtxos, "fee", feeUtxos)

	return
}

func (s *CardanoChainOperationReactorStrategy) FilterUTXOsForConsolidation(
	multisigUtxos, feeUtxos []*indexer.TxInputOutput,
	_ *cardano.CardanoChainConfig,
) ([]*indexer.TxInputOutput, []*indexer.TxInputOutput, error) {
	return filterOutUtxosWithUnknownTokens(multisigUtxos), filterOutUtxosWithUnknownTokens(feeUtxos), nil
}

func (s *CardanoChainOperationReactorStrategy) getNeededUtxos(
	txInputsOutputs []*indexer.TxInputOutput,
	desiredAmount map[string]uint64,
	minUtxoAmount uint64,
	maxUtxoCount int,
	_ uint64,
	takeAtLeastUtxoCount int,
) (chosenUTXOs []*indexer.TxInputOutput, err error) {
	// if we have change then it must be greater than this amount
	desiredAmount[cardanowallet.AdaTokenName] += minUtxoAmount

	return getNeededUtxos(
		filterOutUtxosWithUnknownTokens(txInputsOutputs), desiredAmount, maxUtxoCount, takeAtLeastUtxoCount)
}

type CardanoChainOperationSkylineStrategy struct {
}

func (s *CardanoChainOperationSkylineStrategy) GetOutputs(
	txs []eth.ConfirmedTransaction, cardanoConfig *cardano.CardanoChainConfig, logger hclog.Logger,
) (cardano.TxOutputs, uint64, error) {
	receiversMap := map[string]cardanowallet.TxOutput{}
	tokenHoldingOutputs := uint64(0)

	for _, transaction := range txs {
		for _, receiver := range transaction.Receivers {
			data := receiversMap[receiver.DestinationAddress]
			data.Amount += receiver.Amount.Uint64()

			if receiver.AmountWrapped != nil && receiver.AmountWrapped.Sign() > 0 {
				if len(data.Tokens) == 0 {
					token, err := cardanoConfig.GetNativeToken(
						common.ToStrChainID(transaction.SourceChainId))
					if err != nil {
						return cardano.TxOutputs{}, 0, err
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

		if len(txOut.Tokens) > 0 {
			tokenHoldingOutputs++

			for _, token := range txOut.Tokens {
				result.Sum[token.TokenName()] += token.Amount
			}
		}
	}

	// sort outputs because all batchers should have same order of outputs
	sort.Slice(result.Outputs, func(i, j int) bool {
		return result.Outputs[i].Addr < result.Outputs[j].Addr
	})

	return result, tokenHoldingOutputs, nil
}

func (s CardanoChainOperationSkylineStrategy) GetUTXOs(
	multisigAddress, multisigFeeAddress string, txOutputs cardano.TxOutputs, tokenHoldingOutputs uint64,
	destChainID string, cardanoConfig *cardano.CardanoChainConfig, db indexer.Database, logger hclog.Logger,
) (multisigUtxos []*indexer.TxInputOutput, feeUtxos []*indexer.TxInputOutput, err error) {
	multisigUtxos, err = db.GetAllTxOutputs(multisigAddress, true)
	if err != nil {
		return
	}

	feeUtxos, err = db.GetAllTxOutputs(multisigFeeAddress, true)
	if err != nil {
		return
	}

	feeUtxos = filterOutUtxosWithUnknownTokens(feeUtxos)
	if len(feeUtxos) == 0 {
		return nil, nil, fmt.Errorf("fee multisig does not have any utxo: %s", multisigFeeAddress)
	}

	knownTokens, err := cardano.GetKnownTokens(cardanoConfig)
	if err != nil {
		return
	}

	multisigUtxos = filterOutUtxosWithUnknownTokens(multisigUtxos, knownTokens...)

	logger.Debug("UTXOs retrieved",
		"multisig", multisigAddress, "utxos", multisigUtxos, "fee", multisigFeeAddress, "utxos", feeUtxos)

	feeUtxos = feeUtxos[:min(cardanoConfig.MaxFeeUtxoCount, len(feeUtxos))] // do not take more than maxFeeUtxoCount

	multisigUtxos, err = s.getNeededUtxos(
		multisigUtxos,
		txOutputs.Sum,
		cardanoConfig.UtxoMinAmount,
		cardanoConfig.MaxUtxoCount-len(feeUtxos),
		tokenHoldingOutputs,
		cardanoConfig.TakeAtLeastUtxoCount,
	)
	if err != nil {
		return
	}

	logger.Debug("UTXOs chosen", "multisig", multisigUtxos, "fee", feeUtxos)

	return
}

func (s *CardanoChainOperationSkylineStrategy) FilterUTXOsForConsolidation(
	multisigUtxos, feeUtxos []*indexer.TxInputOutput,
	cardanoConfig *cardano.CardanoChainConfig,
) ([]*indexer.TxInputOutput, []*indexer.TxInputOutput, error) {
	knownTokens, err := cardano.GetKnownTokens(cardanoConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get known tokens: %w", err)
	}

	multisigUtxos = filterOutUtxosWithUnknownTokens(multisigUtxos, knownTokens...)
	feeUtxos = filterOutUtxosWithUnknownTokens(feeUtxos)

	return multisigUtxos, feeUtxos, nil
}

func (s *CardanoChainOperationSkylineStrategy) getNeededUtxos(
	txInputsOutputs []*indexer.TxInputOutput,
	desiredAmount map[string]uint64,
	minUtxoAmount uint64,
	maxUtxoCount int,
	tokenHoldingOutputs uint64,
	takeAtLeastUtxoCount int,
) ([]*indexer.TxInputOutput, error) {
	for tokenName, desiredValue := range desiredAmount {
		if tokenName == cardanowallet.AdaTokenName {
			// if we have change then it must be greater than this amount
			desiredAmount[tokenName] = desiredValue + minUtxoAmount*tokenHoldingOutputs
		}
	}

	return getNeededUtxos(txInputsOutputs, desiredAmount, maxUtxoCount, takeAtLeastUtxoCount)
}

func getNeededUtxos(
	txInputOutputs []*indexer.TxInputOutput, desiredAmount map[string]uint64,
	maxUtxoCount int, takeAtLeastUtxoCount int,
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

	outputUTXOs, err := txsend.GetUTXOsForAmounts(inputUtxos, desiredAmount, maxUtxoCount, takeAtLeastUtxoCount)
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
