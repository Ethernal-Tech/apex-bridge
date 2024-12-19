package batcher

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	cardano "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
)

var (
	_ core.ChainOperations = (*CardanoChainOperations)(nil)
)

// Get real tx size from protocolParams/config
const (
	maxFeeUtxoCount = 4
	maxUtxoCount    = 50
	maxTxSize       = 16000
)

type CardanoChainOperations struct {
	config           *cardano.CardanoChainConfig
	wallet           *cardano.CardanoWallet
	txProvider       cardanowallet.ITxDataRetriever
	db               indexer.Database
	gasLimiter       eth.GasLimitHolder
	cardanoCliBinary string
	logger           hclog.Logger
}

func NewCardanoChainOperations(
	jsonConfig json.RawMessage,
	db indexer.Database,
	secretsManager secrets.SecretsManager,
	chainID string,
	logger hclog.Logger,
) (*CardanoChainOperations, error) {
	cardanoConfig, err := cardano.NewCardanoChainConfig(jsonConfig)
	if err != nil {
		return nil, err
	}

	txProvider, err := cardanoConfig.CreateTxProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to create tx provider: %w", err)
	}

	cardanoWallet, err := cardano.LoadWallet(secretsManager, chainID)
	if err != nil {
		return nil, err
	}

	return &CardanoChainOperations{
		wallet:           cardanoWallet,
		config:           cardanoConfig,
		txProvider:       txProvider,
		cardanoCliBinary: cardanowallet.ResolveCardanoCliBinary(cardanoConfig.NetworkID),
		gasLimiter:       eth.NewGasLimitHolder(submitBatchMinGasLimit, submitBatchMaxGasLimit, submitBatchStepsGasLimit),
		db:               db,
		logger:           logger,
	}, nil
}

// GenerateBatchTransaction implements core.ChainOperations.
func (cco *CardanoChainOperations) GenerateBatchTransaction(
	ctx context.Context,
	bridgeSmartContract eth.IBridgeSmartContract,
	chainID string,
	confirmedTransactions []eth.ConfirmedTransaction,
	batchNonceID uint64,
) (*core.GeneratedBatchTxData, error) {
	validatorsData, err := cco.getCardanoData(ctx, bridgeSmartContract, chainID)
	if err != nil {
		return nil, err
	}

	metadata, err := cardano.CreateBatchMetaData(batchNonceID)
	if err != nil {
		return nil, err
	}

	protocolParams, err := cco.txProvider.GetProtocolParameters(ctx)
	if err != nil {
		return nil, err
	}

	slotNumber, err := cco.getSlotNumber()
	if err != nil {
		return nil, err
	}

	multisigPolicyScript, multisigFeePolicyScript, err := cardano.GetPolicyScripts(validatorsData)
	if err != nil {
		return nil, err
	}

	multisigAddress, multisigFeeAddress, err := cardano.GetMultisigAddresses(
		cco.cardanoCliBinary, uint(cco.config.NetworkMagic), multisigPolicyScript, multisigFeePolicyScript)
	if err != nil {
		return nil, err
	}

	txOutputs := getOutputs(confirmedTransactions, cco.config.NetworkID, cco.logger)

	multisigUtxos, feeUtxos, err := cco.getUTXOs(
		multisigAddress, multisigFeeAddress, txOutputs)
	if err != nil {
		return nil, err
	}

	cco.logger.Info("Creating batch tx", "batchID", batchNonceID,
		"magic", cco.config.NetworkMagic, "binary", cco.cardanoCliBinary,
		"slot", slotNumber, "multisig", len(multisigUtxos), "fee", len(feeUtxos), "outputs", len(txOutputs.Outputs))

	// Create Tx
	txRaw, txHash, err := cardano.CreateTx(
		cco.cardanoCliBinary,
		uint(cco.config.NetworkMagic),
		protocolParams,
		slotNumber+cco.config.TTLSlotNumberInc,
		metadata,
		cardano.TxInputInfos{
			MultiSig: &cardano.TxInputInfo{
				PolicyScript: multisigPolicyScript,
				Address:      multisigAddress,
				TxInputs:     convertUTXOsToTxInputs(multisigUtxos),
			},
			MultiSigFee: &cardano.TxInputInfo{
				PolicyScript: multisigFeePolicyScript,
				Address:      multisigFeeAddress,
				TxInputs:     convertUTXOsToTxInputs(feeUtxos),
			},
		},
		txOutputs.Outputs,
	)
	if err != nil {
		return nil, err
	}

	if len(txRaw) > maxTxSize {
		return nil, errors.New("fatal error, tx size too big")
	}

	return &core.GeneratedBatchTxData{
		TxRaw:  txRaw,
		TxHash: txHash,
	}, nil
}

// SignBatchTransaction implements core.ChainOperations.
func (cco *CardanoChainOperations) SignBatchTransaction(txHash string) ([]byte, []byte, error) {
	witnessMultiSig, err := cardanowallet.CreateTxWitness(txHash, cco.wallet.MultiSig)
	if err != nil {
		return nil, nil, err
	}

	witnessMultiSigFee, err := cardanowallet.CreateTxWitness(txHash, cco.wallet.MultiSigFee)
	if err != nil {
		return nil, nil, err
	}

	return witnessMultiSig, witnessMultiSigFee, nil
}

// IsSynchronized implements core.IsSynchronized.
func (cco *CardanoChainOperations) IsSynchronized(
	ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract, chainID string,
) (bool, error) {
	lastObservedBlockBridge, err := bridgeSmartContract.GetLastObservedBlock(ctx, chainID)
	if err != nil {
		return false, err
	}

	lastOracleBlockPoint, err := cco.db.GetLatestBlockPoint()
	if err != nil {
		return false, err
	}

	return lastOracleBlockPoint != nil &&
		lastOracleBlockPoint.BlockSlot >= lastObservedBlockBridge.BlockSlot.Uint64(), nil
}

// Submit implements core.Submit.
func (cco *CardanoChainOperations) Submit(
	ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract, batch eth.SignedBatch,
) error {
	err := bridgeSmartContract.SubmitSignedBatch(ctx, batch, cco.gasLimiter.GetGasLimit())

	cco.gasLimiter.Update(err)

	return err
}

func (cco *CardanoChainOperations) getSlotNumber() (uint64, error) {
	data, err := cco.db.GetLatestBlockPoint()
	if err != nil {
		return 0, err
	}

	slot := uint64(0)
	if data != nil {
		slot = data.BlockSlot
	}

	newSlot, err := getNumberWithRoundingThreshold(
		slot, cco.config.SlotRoundingThreshold, cco.config.NoBatchPeriodPercent)
	if err != nil {
		return 0, err
	}

	cco.logger.Debug("calculate slotNumber with rounding", "slot", slot, "newSlot", newSlot)

	return newSlot, nil
}

func (cco *CardanoChainOperations) getCardanoData(
	ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract, chainID string,
) ([]eth.ValidatorChainData, error) {
	validatorsData, err := bridgeSmartContract.GetValidatorsChainData(ctx, chainID)
	if err != nil {
		return nil, err
	}

	hasVerificationKey, hasFeeVerificationKey := false, false

	for _, validator := range validatorsData {
		hasVerificationKey = hasVerificationKey || bytes.Equal(cco.wallet.MultiSig.VerificationKey,
			cardanowallet.PadKeyToSize(validator.Key[0].Bytes()))
		hasFeeVerificationKey = hasFeeVerificationKey || bytes.Equal(cco.wallet.MultiSigFee.VerificationKey,
			cardanowallet.PadKeyToSize(validator.Key[1].Bytes()))
	}

	if !hasVerificationKey {
		return nil, fmt.Errorf(
			"verifying key of current batcher wasn't found in validators data queried from smart contract")
	}

	if !hasFeeVerificationKey {
		return nil, fmt.Errorf(
			"verifying fee key of current batcher wasn't found in validators data queried from smart contract")
	}

	return validatorsData, nil
}

func (cco *CardanoChainOperations) getUTXOs(
	multisigAddress, multisigFeeAddress string, txOutputs cardano.TxOutputs,
) (multisigUtxos []*indexer.TxInputOutput, feeUtxos []*indexer.TxInputOutput, err error) {
	multisigUtxos, err = cco.db.GetAllTxOutputs(multisigAddress, true)
	if err != nil {
		return
	}

	feeUtxos, err = cco.db.GetAllTxOutputs(multisigFeeAddress, true)
	if err != nil {
		return
	}

	multisigUtxos = filterOutTokenUtxos(multisigUtxos)
	feeUtxos = filterOutTokenUtxos(feeUtxos)

	if len(feeUtxos) == 0 {
		return nil, nil, fmt.Errorf("fee multisig does not have any utxo: %s", multisigFeeAddress)
	}

	cco.logger.Debug("UTXOs retrieved",
		"multisig", multisigAddress, "utxos", multisigUtxos, "fee", multisigFeeAddress, "utxos", feeUtxos)

	feeUtxos = feeUtxos[:min(maxFeeUtxoCount, len(feeUtxos))] // do not take more than maxFeeUtxoCount

	multisigUtxos, err = getNeededUtxos(
		multisigUtxos,
		txOutputs.Sum[cardanowallet.AdaTokenName],
		cco.config.UtxoMinAmount,
		len(feeUtxos)+len(txOutputs.Outputs),
		maxUtxoCount,
		cco.config.TakeAtLeastUtxoCount,
	)
	if err != nil {
		return
	}

	cco.logger.Debug("UTXOs chosen", "multisig", multisigUtxos, "fee", feeUtxos)

	return
}

func filterOutTokenUtxos(utxos []*indexer.TxInputOutput) []*indexer.TxInputOutput {
	result := make([]*indexer.TxInputOutput, 0, len(utxos))

	for _, utxo := range utxos {
		if len(utxo.Output.Tokens) == 0 {
			result = append(result, utxo)
		}
	}

	return result
}

func convertUTXOsToTxInputs(utxos []*indexer.TxInputOutput) (result cardanowallet.TxInputs) {
	// For now we are taking all available UTXOs as fee (should always be 1-2 of them)
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

// getNeededUtxos returns only needed input utxos
// It is expected that UTXOs are sorted by their Block Slot number (for example: returned sorted by db.GetAllTxOutput)
// and taken from first to last until desiredAmount has been met or maxUtxoCount reached
// if desiredAmount has been met, tx is created regularly
// if maxUtxoCount has been reached, we replace smallest UTXO with first next bigger one until we reach desiredAmount
func getNeededUtxos(
	inputUTXOs []*indexer.TxInputOutput,
	desiredAmount uint64,
	minUtxoAmount uint64,
	utxoCount int,
	maxUtxoCount int,
	takeAtLeastUtxoCount int,
) (chosenUTXOs []*indexer.TxInputOutput, err error) {
	txCostWithMinChange := minUtxoAmount + desiredAmount // if we have change then it must be greater than this amount

	// algorithm that chooses multisig UTXOs
	chosenUTXOsSum := uint64(0)
	isUtxosOk := false

	for i, utxo := range inputUTXOs {
		chosenUTXOs = append(chosenUTXOs, utxo)
		utxoCount++

		chosenUTXOsSum += utxo.Output.Amount // in cardano we should not care about overflow

		if utxoCount > maxUtxoCount {
			minChosenUTXO, minChosenUTXOIdx := findMinUtxo(chosenUTXOs)

			chosenUTXOs[minChosenUTXOIdx] = utxo
			chosenUTXOsSum -= minChosenUTXO.Output.Amount
			chosenUTXOs = chosenUTXOs[:len(chosenUTXOs)-1]
			utxoCount--
		}

		if chosenUTXOsSum >= txCostWithMinChange || chosenUTXOsSum == desiredAmount {
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
		return nil, fmt.Errorf("fatal error, couldn't select UTXOs for sum: %d", desiredAmount)
	}

	return chosenUTXOs, nil
}

func findMinUtxo(utxos []*indexer.TxInputOutput) (*indexer.TxInputOutput, int) {
	min := utxos[0]
	idx := 0

	for i, utxo := range utxos[1:] {
		if utxo.Output.Amount < min.Output.Amount {
			min = utxo
			idx = i + 1
		}
	}

	return min, idx
}

func getOutputs(
	txs []eth.ConfirmedTransaction, networkID cardanowallet.CardanoNetworkType, logger hclog.Logger,
) cardano.TxOutputs {
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
		} else if !cardano.IsValidOutputAddress(addr, networkID) {
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

	return result
}
