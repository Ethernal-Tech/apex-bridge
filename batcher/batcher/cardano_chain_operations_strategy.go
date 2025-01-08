package batcher

import (
	"fmt"
	"sort"

	cardano "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/cardano-infrastructure/bridgingtx"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
)

type ICardanoChainOperationsStrategy interface {
	GetOutputs(
		txs []eth.ConfirmedTransaction,
		cardanoConfig *cardano.CardanoChainConfig,
		destChainID string,
		logger hclog.Logger,
	) (*cardano.TxOutputs, uint64, error)
	GetNeededUtxos(
		inputUTXOs []*indexer.TxInputOutput,
		desiredAmount map[string]uint64,
		minUtxoAmount uint64,
		utxoCount int,
		maxUtxoCount int,
		tokenHoldingOutputs uint64,
		takeAtLeastUtxoCount int,
	) (chosenUTXOs []*indexer.TxInputOutput, err error)
	GetUTXOs(
		multisigAddress,
		multisigFeeAddress string,
		txOutputs cardano.TxOutputs,
		tokenHoldingOutputs uint64,
		cardanoConfig *cardano.CardanoChainConfig,
		db indexer.Database,
		logger hclog.Logger,
	) (multisigUtxos []*indexer.TxInputOutput, feeUtxos []*indexer.TxInputOutput, err error)
}

type CardanoChainOperationReactorStrategy struct {
}

func (s *CardanoChainOperationReactorStrategy) GetOutputs(
	txs []eth.ConfirmedTransaction, cardanoConfig *cardano.CardanoChainConfig, _ string, logger hclog.Logger,
) (*cardano.TxOutputs, uint64, error) {
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

	return &result, 0, nil
}

func (s *CardanoChainOperationReactorStrategy) GetNeededUtxos(
	inputUTXOs []*indexer.TxInputOutput,
	desiredAmount map[string]uint64,
	minUtxoAmount uint64,
	utxoCount int,
	maxUtxoCount int,
	_ uint64,
	takeAtLeastUtxoCount int,
) (chosenUTXOs []*indexer.TxInputOutput, err error) {
	inputUTXOs = filterOutTokenUtxos(inputUTXOs)
	// if we have change then it must be greater than this amount
	desiredAmount[cardanowallet.AdaTokenName] += minUtxoAmount

	inUtxos := mapUtxos(inputUTXOs)

	outputUTXOs, err := bridgingtx.GetUTXOsForAmounts(inUtxos, desiredAmount, maxUtxoCount, takeAtLeastUtxoCount)
	if err != nil {
		return nil, err
	}

	usedUtxoMap := map[string]bool{}
	for _, utxo := range outputUTXOs.Inputs {
		usedUtxoMap[utxo.String()] = true
	}

	chosenUTXOs = make([]*indexer.TxInputOutput, 0, len(outputUTXOs.Inputs))

	for _, utxo := range inputUTXOs {
		if !usedUtxoMap[fmt.Sprintf("%s#%d", utxo.Input.Hash, utxo.Input.Index)] {
			continue
		}

		chosenUTXOs = append(chosenUTXOs, utxo)
	}

	return chosenUTXOs, nil
}

func (s *CardanoChainOperationReactorStrategy) GetUTXOs(
	multisigAddress, multisigFeeAddress string, txOutputs cardano.TxOutputs, _ uint64,
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

	feeUtxos = filterOutTokenUtxos(feeUtxos)

	if len(feeUtxos) == 0 {
		return nil, nil, fmt.Errorf("fee multisig does not have any utxo: %s", multisigFeeAddress)
	}

	logger.Debug("UTXOs retrieved",
		"multisig", multisigAddress, "utxos", multisigUtxos, "fee", multisigFeeAddress, "utxos", feeUtxos)

	feeUtxos = feeUtxos[:min(maxFeeUtxoCount, len(feeUtxos))] // do not take more than maxFeeUtxoCount

	multisigUtxos, err = s.GetNeededUtxos(
		multisigUtxos,
		txOutputs.Sum,
		cardanoConfig.UtxoMinAmount,
		len(feeUtxos)+len(txOutputs.Outputs),
		maxUtxoCount,
		0,
		cardanoConfig.TakeAtLeastUtxoCount,
	)
	if err != nil {
		return
	}

	logger.Debug("UTXOs chosen", "multisig", multisigUtxos, "fee", feeUtxos)

	return
}

func mapUtxos(inputUTXOs []*indexer.TxInputOutput) []cardanowallet.Utxo {
	output := make([]cardanowallet.Utxo, len(inputUTXOs))

	for i, utxo := range inputUTXOs {
		output[i] = cardanowallet.Utxo{
			Hash:   utxo.Input.Hash.String(),
			Index:  utxo.Input.Index,
			Amount: utxo.Output.Amount,
			Tokens: make([]cardanowallet.TokenAmount, len(utxo.Output.Tokens)),
		}
		for j, token := range utxo.Output.Tokens {
			output[i].Tokens[j] = cardanowallet.TokenAmount(token)
		}
	}

	return output
}
