package batcher

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sort"

	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	cardano "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
)

var (
	errNonActiveBatchPeriod = errors.New("non active batch period")

	_ core.ChainOperations = (*CardanoChainOperations)(nil)
)

// nolintlint TODO: Get from protocol parameters, maybe add to core.CardanoChainConfig
// Get real tx size from protocolParams/config
const (
	minUtxoAmount = uint64(1000000)
	maxUtxoCount  = 410
	maxTxSize     = 16000

	noBatchPeriodPercent = 0.0625
)

type CardanoChainOperations struct {
	Config     *cardano.CardanoChainConfig
	Wallet     *cardano.CardanoWallet
	TxProvider cardanowallet.ITxDataRetriever
	logger     hclog.Logger
}

func NewCardanoChainOperations(
	jsonConfig json.RawMessage,
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

	cardanoWallet, err := cardanoConfig.LoadWallet()
	if err != nil {
		return nil, fmt.Errorf("error while loading wallet info: %w", err)
	}

	return &CardanoChainOperations{
		Wallet:     cardanoWallet,
		Config:     cardanoConfig,
		TxProvider: txProvider,
		logger:     logger,
	}, nil
}

// GenerateBatchTransaction implements core.ChainOperations.
func (cco *CardanoChainOperations) GenerateBatchTransaction(
	ctx context.Context,
	bridgeSmartContract eth.IBridgeSmartContract,
	destinationChain string,
	confirmedTransactions []eth.ConfirmedTransaction,
	batchNonceID uint64,
) (*core.GeneratedBatchTxData, error) {
	metadata, err := cardano.CreateBatchMetaData(batchNonceID)
	if err != nil {
		return nil, err
	}

	protocolParams, err := cco.TxProvider.GetProtocolParameters(ctx)
	if err != nil {
		return nil, err
	}

	slotNumber, err := cco.getSlotNumber(ctx, bridgeSmartContract, destinationChain, noBatchPeriodPercent)
	if err != nil {
		return nil, err
	}

	validatorsData, err := bridgeSmartContract.GetValidatorsCardanoData(ctx, destinationChain)
	if err != nil {
		return nil, err
	}

	var (
		multisigKeyHashes       = make([]string, len(validatorsData))
		multisigFeeKeyHashes    = make([]string, len(validatorsData))
		validatorKeyBytes       []byte
		foundVerificationKey    = false
		foundFeeVerificationKey = false
	)

	for i, validator := range validatorsData {
		validatorKeyBytes, err = hex.DecodeString(validator.VerifyingKey)
		if err != nil {
			return nil, err
		}

		multisigKeyHashes[i], err = cardanowallet.GetKeyHash(validatorKeyBytes)
		if err != nil {
			return nil, err
		}

		if bytes.Equal(cco.Wallet.MultiSig.GetVerificationKey(), validatorKeyBytes) {
			foundVerificationKey = true
		}

		validatorKeyBytes, err = hex.DecodeString(validator.VerifyingKeyFee)
		if err != nil {
			return nil, err
		}

		multisigFeeKeyHashes[i], err = cardanowallet.GetKeyHash(validatorKeyBytes)
		if err != nil {
			return nil, err
		}

		if bytes.Equal(cco.Wallet.MultiSigFee.GetVerificationKey(), validatorKeyBytes) {
			foundFeeVerificationKey = true
		}
	}

	if !foundVerificationKey {
		return nil, fmt.Errorf(
			"verifying key of current batcher wasn't found in validators data queried from smart contract")
	}

	if !foundFeeVerificationKey {
		return nil, fmt.Errorf(
			"verifying fee key of current batcher wasn't found in validators data queried from smart contract")
	}

	multisigPolicyScript, err := cardanowallet.NewPolicyScript(
		multisigKeyHashes, int(common.GetRequiredSignaturesForConsensus(uint64(len(multisigKeyHashes)))))
	if err != nil {
		return nil, err
	}

	multisigFeePolicyScript, err := cardanowallet.NewPolicyScript(
		multisigFeeKeyHashes, int(common.GetRequiredSignaturesForConsensus(uint64(len(multisigFeeKeyHashes)))))
	if err != nil {
		return nil, err
	}

	multisigAddress, err := multisigPolicyScript.CreateMultiSigAddress(uint(cco.Config.TestNetMagic))
	if err != nil {
		return nil, err
	}

	multisigFeeAddress, err := multisigFeePolicyScript.CreateMultiSigAddress(uint(cco.Config.TestNetMagic))
	if err != nil {
		return nil, err
	}

	txUtxos, err := getInputUtxos(ctx, bridgeSmartContract, destinationChain)
	if err != nil {
		return nil, err
	}

	txOutput := getOutputs(confirmedTransactions)

	txInfos := cardano.TxInputInfos{
		MultiSig: &cardano.TxInputInfo{
			PolicyScript: multisigPolicyScript,
			Address:      multisigAddress,
		},
		MultiSigFee: &cardano.TxInputInfo{
			PolicyScript: multisigFeePolicyScript,
			Address:      multisigFeeAddress,
		},
	}

	return cco.createBatchTx(
		txUtxos, metadata, protocolParams, txInfos, txOutput, slotNumber)
}

// SignBatchTransaction implements core.ChainOperations.
func (cco *CardanoChainOperations) SignBatchTransaction(txHash string) ([]byte, []byte, error) {
	witnessMultiSig, err := cardano.CreateTxWitness(txHash, cco.Wallet.MultiSig)
	if err != nil {
		return nil, nil, err
	}

	witnessMultiSigFee, err := cardano.CreateTxWitness(txHash, cco.Wallet.MultiSigFee)
	if err != nil {
		return nil, nil, err
	}

	return witnessMultiSig, witnessMultiSigFee, nil
}

/* UTXOs are sorted by Nonce and taken from first to last until txCost has been met or maxUtxoCount reached
 * if txCost has been met, tx is created regularly
 * if maxUtxoCount has been reached, we replace smallest UTXO with first next bigger one until we reach txCost
 */
func (cco *CardanoChainOperations) createBatchTx(
	inputUtxos eth.UTXOs,
	metadata []byte, protocolParams []byte,
	txInfos cardano.TxInputInfos,
	txOutputs cardano.TxOutputs, slotNumber uint64,
) (*core.GeneratedBatchTxData, error) {
	cco.logger.Info("creating batch tx",
		"slot", slotNumber, "ttl", cco.Config.TTLSlotNumberInc, "magic", cco.Config.TestNetMagic)

	feeTxInputs := convertUTXOsToTxInputs(inputUtxos.FeePayerOwnedUTXOs)

	multisigChosenUtxos, err := getNeededUtxos(
		inputUtxos.MultisigOwnedUTXOs, txOutputs.Sum, len(feeTxInputs.Inputs)+len(txOutputs.Outputs))
	if err != nil {
		return nil, err
	}

	txInfos.MultiSig.TxInputs = convertUTXOsToTxInputs(multisigChosenUtxos)
	txInfos.MultiSigFee.TxInputs = feeTxInputs

	// Create Tx
	txRaw, txHash, err := cardano.CreateTx(
		uint(cco.Config.TestNetMagic), protocolParams, slotNumber+cco.Config.TTLSlotNumberInc,
		metadata, txInfos, txOutputs.Outputs,
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
		Utxos: eth.UTXOs{
			MultisigOwnedUTXOs: multisigChosenUtxos,
			FeePayerOwnedUTXOs: inputUtxos.FeePayerOwnedUTXOs,
		},
		Slot: slotNumber,
	}, err
}

func (cco *CardanoChainOperations) getSlotNumber(
	ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract, chain string, noBatchPeriodPercent float64,
) (uint64, error) {
	if cco.Config.SlotRoundingThreshold == 0 {
		lastObservedBlock, err := bridgeSmartContract.GetLastObservedBlock(ctx, chain)
		if err != nil {
			return 0, err
		}

		return lastObservedBlock.BlockSlot, nil
	}

	data, err := cco.TxProvider.GetTip(ctx)
	if err != nil {
		return 0, err
	}

	newSlot, err := getSlotNumberWithRoundingThreshold(
		data.Slot, cco.Config.SlotRoundingThreshold, noBatchPeriodPercent)
	if err != nil {
		return 0, err
	}

	cco.logger.Debug("calculate slotNumber with rounding", "slot", data.Slot, "newSlot", newSlot)

	return newSlot, nil
}

func getSlotNumberWithRoundingThreshold(
	slotNumber, threshold uint64, noBatchPeriodPercent float64,
) (uint64, error) {
	if slotNumber == 0 {
		return 0, errors.New("slot number is zero")
	}

	newSlot := ((slotNumber + threshold - 1) / threshold) * threshold
	diffFromPrevious := slotNumber - (newSlot - threshold)

	if diffFromPrevious <= uint64(float64(threshold)*noBatchPeriodPercent) ||
		diffFromPrevious >= uint64(float64(threshold)*(1.0-noBatchPeriodPercent)) {
		return 0, fmt.Errorf("%w: (slot, rounded) = (%d, %d)", errNonActiveBatchPeriod, slotNumber, newSlot)
	}

	return newSlot, nil
}

func getInputUtxos(
	ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract, destinationChain string,
) (eth.UTXOs, error) {
	inputUtxos, err := bridgeSmartContract.GetAvailableUTXOs(ctx, destinationChain)
	if err != nil {
		return inputUtxos, err
	}

	sort.Slice(inputUtxos.MultisigOwnedUTXOs, func(i, j int) bool {
		return inputUtxos.MultisigOwnedUTXOs[i].Nonce < inputUtxos.MultisigOwnedUTXOs[j].Nonce
	})
	sort.Slice(inputUtxos.FeePayerOwnedUTXOs, func(i, j int) bool {
		return inputUtxos.FeePayerOwnedUTXOs[i].Nonce < inputUtxos.FeePayerOwnedUTXOs[j].Nonce
	})

	return inputUtxos, err
}

func convertUTXOsToTxInputs(utxos []eth.UTXO) (result cardanowallet.TxInputs) {
	// For now we are taking all available UTXOs as fee (should always be 1-2 of them)
	result.Inputs = make([]cardanowallet.TxInput, len(utxos))
	result.Sum = uint64(0)

	for i, utxo := range utxos {
		result.Inputs[i] = cardanowallet.TxInput{
			Hash:  utxo.TxHash,
			Index: uint32(utxo.TxIndex.Uint64()),
		}

		result.Sum += utxo.Amount.Uint64()
	}

	return result
}

// getNeededUtxos returns only needed input utxos
func getNeededUtxos(
	inputUTXOs []eth.UTXO, txCost *big.Int, utxoCount int,
) (chosenUTXOs []eth.UTXO, err error) {
	// Create initial UTXO set
	txCostWithMinChange := new(big.Int).Add(new(big.Int).SetUint64(minUtxoAmount), txCost)

	// algorithm that chooses multisig UTXOs
	chosenUTXOsSum := big.NewInt(0)
	isUtxosOk := false

	for _, utxo := range inputUTXOs {
		chosenUTXOs = append(chosenUTXOs, utxo)
		utxoCount++

		chosenUTXOsSum.Add(chosenUTXOsSum, utxo.Amount)

		if utxoCount > maxUtxoCount {
			minChosenUTXO, minChosenUTXOIdx := findMinUtxo(chosenUTXOs)

			chosenUTXOs[minChosenUTXOIdx] = utxo

			chosenUTXOsSum.Sub(chosenUTXOsSum, minChosenUTXO.Amount)

			chosenUTXOs = chosenUTXOs[:len(chosenUTXOs)-1]
			utxoCount--
		}

		if chosenUTXOsSum.Cmp(txCostWithMinChange) >= 0 || chosenUTXOsSum.Cmp(txCost) == 0 {
			isUtxosOk = true

			break
		}
	}

	if !isUtxosOk {
		return nil, errors.New("fatal error, couldn't select UTXOs")
	}

	return chosenUTXOs, nil
}

func findMinUtxo(utxos []eth.UTXO) (eth.UTXO, int) {
	min := utxos[0]
	idx := 0

	for i, utxo := range utxos[1:] {
		if utxo.Amount.Cmp(min.Amount) == -1 {
			min = utxo
			idx = i + 1
		}
	}

	return min, idx
}

func getOutputs(txs []eth.ConfirmedTransaction) cardano.TxOutputs {
	receiversMap := map[string]*big.Int{}

	for _, transaction := range txs {
		for _, receiver := range transaction.Receivers {
			if value, exists := receiversMap[receiver.DestinationAddress]; exists {
				value.Add(value, receiver.Amount)
			} else {
				receiversMap[receiver.DestinationAddress] = new(big.Int).Set(receiver.Amount)
			}
		}
	}

	result := cardano.TxOutputs{
		Outputs: make([]cardanowallet.TxOutput, 0, len(receiversMap)),
		Sum:     big.NewInt(0),
	}

	for addr, amount := range receiversMap {
		if amount.Cmp(big.NewInt(0)) <= 0 {
			// this should be logged once
			continue
		}

		result.Outputs = append(result.Outputs, cardanowallet.TxOutput{
			Addr:   addr,
			Amount: amount.Uint64(),
		})
		result.Sum.Add(result.Sum, amount)
	}

	// sort outputs because all batchers should have same order of outputs
	sort.Slice(result.Outputs, func(i, j int) bool {
		return result.Outputs[i].Addr < result.Outputs[j].Addr
	})

	return result
}
