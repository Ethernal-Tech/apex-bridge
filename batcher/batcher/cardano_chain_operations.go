package batcher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	cardano "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/validatorobserver"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
)

var (
	_ core.ChainOperations = (*CardanoChainOperations)(nil)

	errUTXOsLimitReached   = errors.New("utxos limit reached, consolidation is required")
	errUTXOsCouldNotSelect = errors.New("couldn't select UTXOs")
	errTxSizeTooBig        = errors.New("batch tx size too big")
)

// Get real tx size from protocolParams/config
const (
	maxTxSize = 16000
)

type batchInitialData struct {
	BatchNonceID         uint64
	Metadata             []byte
	ProtocolParams       []byte
	MultisigPolicyScript *cardanowallet.PolicyScript
	FeePolicyScript      *cardanowallet.PolicyScript
	MultisigAddr         string
	FeeAddr              string
}

type CardanoChainOperations struct {
	config           *cardano.CardanoChainConfig
	wallet           *cardano.ApexCardanoWallet
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
	data, err := cco.createBatchInitialData(ctx, bridgeSmartContract, chainID, batchNonceID)
	if err != nil {
		return nil, err
	}

	txData, err := cco.generateBatchTransaction(data, confirmedTransactions)

	if cco.shouldConsolidate(err) {
		cco.logger.Warn("consolidation batch generation started", "err", err)

		txData, err = cco.generateConsolidationTransaction(data)
		if err != nil {
			err = fmt.Errorf("consolidation batch failed: %w", err)
		}
	}

	return txData, err
}

// SignBatchTransaction implements core.ChainOperations.
func (cco *CardanoChainOperations) SignBatchTransaction(
	generatedBatchData *core.GeneratedBatchTxData) ([]byte, []byte, error) {
	if generatedBatchData.BatchType == uint8(ValidatorSetFinal) {
		return []byte{}, []byte{}, nil
	}

	txBuilder, err := cardanowallet.NewTxBuilder(cco.cardanoCliBinary)
	if err != nil {
		return nil, nil, err
	}

	defer txBuilder.Dispose()

	witnessMultiSig, err := txBuilder.CreateTxWitness(generatedBatchData.TxRaw, cco.wallet.MultiSig)
	if err != nil {
		return nil, nil, err
	}

	witnessMultiSigFee, err := txBuilder.CreateTxWitness(generatedBatchData.TxRaw, cco.wallet.Fee)
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

func (cco *CardanoChainOperations) generateBatchTransaction(
	data *batchInitialData, confirmedTransactions []eth.ConfirmedTransaction,
) (*core.GeneratedBatchTxData, error) {
	txOutputs := getOutputs(confirmedTransactions, cco.config.NetworkID, cco.logger)

	multisigUtxos, feeUtxos, err := cco.getUTXOs(
		data.MultisigAddr, data.FeeAddr, txOutputs)
	if err != nil {
		return nil, err
	}

	slotNumber, err := cco.getSlotNumber()
	if err != nil {
		return nil, err
	}

	cco.logger.Info("Creating batch tx", "batchID", data.BatchNonceID,
		"magic", cco.config.NetworkMagic, "binary", cco.cardanoCliBinary,
		"slot", slotNumber, "multisig", len(multisigUtxos), "fee", len(feeUtxos), "outputs", len(txOutputs.Outputs))

	// Create Tx
	txRaw, txHash, err := cardano.CreateTx(
		cco.cardanoCliBinary,
		uint(cco.config.NetworkMagic),
		data.ProtocolParams,
		slotNumber+cco.config.TTLSlotNumberInc,
		data.Metadata,
		cardano.TxInputInfos{
			MultiSig: &cardano.TxInputInfo{
				PolicyScript: data.MultisigPolicyScript,
				Address:      data.MultisigAddr,
				TxInputs:     convertUTXOsToTxInputs(multisigUtxos),
			},
			MultiSigFee: &cardano.TxInputInfo{
				PolicyScript: data.FeePolicyScript,
				Address:      data.FeeAddr,
				TxInputs:     convertUTXOsToTxInputs(feeUtxos),
			},
		},
		txOutputs.Outputs,
	)
	if err != nil {
		return nil, err
	}

	if len(txRaw) > maxTxSize {
		return nil, fmt.Errorf("%w: (size, max) = (%d, %d)",
			errTxSizeTooBig, len(txRaw), maxTxSize)
	}

	return &core.GeneratedBatchTxData{
		TxRaw:  txRaw,
		TxHash: txHash,
	}, nil
}

func (cco *CardanoChainOperations) shouldConsolidate(err error) bool {
	return errors.Is(err, errUTXOsLimitReached) || errors.Is(err, errTxSizeTooBig)
}

func (cco *CardanoChainOperations) generateConsolidationTransaction(
	data *batchInitialData,
) (*core.GeneratedBatchTxData, error) {
	multisigUtxos, feeUtxos, err := cco.getUTXOsForConsolidation(data.MultisigAddr, data.FeeAddr)
	if err != nil {
		return nil, err
	}

	totalMultisigAmount := uint64(0)
	for _, utxo := range multisigUtxos {
		totalMultisigAmount += utxo.Output.Amount
	}

	slotNumber, err := cco.getSlotNumber()
	if err != nil {
		return nil, err
	}

	cco.logger.Info("Creating consolidation tx", "consolidationTxID", data.BatchNonceID,
		"magic", cco.config.NetworkMagic, "binary", cco.cardanoCliBinary,
		"slot", slotNumber, "multisig", len(multisigUtxos), "fee", len(feeUtxos))

	// Create Tx
	txRaw, txHash, err := cardano.CreateTx(
		cco.cardanoCliBinary,
		uint(cco.config.NetworkMagic),
		data.ProtocolParams,
		slotNumber+cco.config.TTLSlotNumberInc,
		data.Metadata,
		cardano.TxInputInfos{
			MultiSig: &cardano.TxInputInfo{
				PolicyScript: data.MultisigPolicyScript,
				Address:      data.MultisigAddr,
				TxInputs:     convertUTXOsToTxInputs(multisigUtxos),
			},
			MultiSigFee: &cardano.TxInputInfo{
				PolicyScript: data.FeePolicyScript,
				Address:      data.FeeAddr,
				TxInputs:     convertUTXOsToTxInputs(feeUtxos),
			},
		},
		[]cardanowallet.TxOutput{
			cardanowallet.NewTxOutput(data.MultisigAddr, totalMultisigAmount),
		},
	)
	if err != nil {
		return nil, err
	}

	if len(txRaw) > maxTxSize {
		return nil, fmt.Errorf("%w: (size, max) = (%d, %d)", errTxSizeTooBig, len(txRaw), maxTxSize)
	}

	return &core.GeneratedBatchTxData{
		BatchType: uint8(Consolidation),
		TxRaw:     txRaw,
		TxHash:    txHash,
	}, nil
}

func (cco *CardanoChainOperations) getUTXOsForConsolidation(
	multisigAddress, multisigFeeAddress string,
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

	// do not take more than maxFeeUtxoCount
	feeUtxos = feeUtxos[:min(int(cco.config.MaxFeeUtxoCount), len(feeUtxos))] //nolint:gosec
	// do not take more than maxUtxoCount - length of chosen fee utxos
	maxUtxosCnt := min(getMaxUtxoCount(cco.config, len(feeUtxos)), len(multisigUtxos))
	multisigUtxos = multisigUtxos[:maxUtxosCnt]

	cco.logger.Debug("UTXOs chosen", "multisig", multisigUtxos, "fee", feeUtxos)

	return
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

	feeUtxos = feeUtxos[:min(cco.config.MaxFeeUtxoCount, uint(len(feeUtxos)))] // do not take more than MaxFeeUtxoCount

	multisigUtxos, err = getNeededUtxos(
		multisigUtxos,
		txOutputs.Sum[cardanowallet.AdaTokenName],
		cco.config.UtxoMinAmount,
		getMaxUtxoCount(cco.config, len(feeUtxos)),
		int(cco.config.TakeAtLeastUtxoCount), //nolint:gosec
	)
	if err != nil {
		return
	}

	cco.logger.Debug("UTXOs chosen", "multisig", multisigUtxos, "fee", feeUtxos)

	return multisigUtxos, feeUtxos, nil
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

func (cco *CardanoChainOperations) getValidatorsChainData(
	ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract, chainID string,
) ([]eth.ValidatorChainData, error) {
	validatorsData, err := bridgeSmartContract.GetValidatorsChainData(ctx, chainID)
	if err != nil {
		return nil, err
	}

	for _, data := range validatorsData {
		if cardano.AreVerifyingKeysTheSame(cco.wallet, data) {
			return validatorsData, nil
		}
	}

	return nil, fmt.Errorf(
		"verifying keys of current batcher wasn't found in validators data queried from smart contract")
}
func (cco *CardanoChainOperations) createBatchInitialData(
	ctx context.Context,
	bridgeSmartContract eth.IBridgeSmartContract,
	chainID string,
	batchNonceID uint64,
) (*batchInitialData, error) {
	validatorsData, err := cco.getValidatorsChainData(ctx, bridgeSmartContract, chainID)
	if err != nil {
		return nil, err
	}

	metadata, err := common.MarshalMetadata(common.MetadataEncodingTypeJSON, common.BatchExecutedMetadata{
		BridgingTxType: common.BridgingTxTypeBatchExecution,
		BatchNonceID:   batchNonceID,
	})
	if err != nil {
		return nil, err
	}

	protocolParams, err := cco.txProvider.GetProtocolParameters(ctx)
	if err != nil {
		return nil, err
	}

	keyHashes, err := cardano.NewApexKeyHashes(validatorsData)
	if err != nil {
		return nil, err
	}

	policyScripts := cardano.NewApexPolicyScripts(keyHashes)

	addresses, err := cardano.NewApexAddresses(cco.cardanoCliBinary, uint(cco.config.NetworkMagic), policyScripts)
	if err != nil {
		return nil, err
	}

	return &batchInitialData{
		BatchNonceID:         batchNonceID,
		Metadata:             metadata,
		ProtocolParams:       protocolParams,
		MultisigPolicyScript: policyScripts.Multisig.Payment,
		FeePolicyScript:      policyScripts.Fee.Payment,
		MultisigAddr:         addresses.Multisig.Payment,
		FeeAddr:              addresses.Fee.Payment,
	}, nil
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
	maxUtxoCount int,
	takeAtLeastUtxoCount int,
) (chosenUTXOs []*indexer.TxInputOutput, err error) {
	// if there is a change then it must be greater than this amount
	txCostWithMinChange := minUtxoAmount + desiredAmount

	// algorithm that chooses multisig UTXOs
	chosenUTXOsSum := uint64(0)
	totalUTXOsSum := uint64(0)
	utxoCount := 0
	isUtxosOk := false

	for i, utxo := range inputUTXOs {
		chosenUTXOs = append(chosenUTXOs, utxo)
		utxoCount++

		chosenUTXOsSum += utxo.Output.Amount // in cardano we should not care about overflow
		totalUTXOsSum += utxo.Output.Amount

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
		if totalUTXOsSum >= txCostWithMinChange || totalUTXOsSum == desiredAmount {
			return nil, fmt.Errorf("%w: %d vs %d", errUTXOsLimitReached, totalUTXOsSum, txCostWithMinChange)
		}

		return nil, fmt.Errorf("%w: %d vs %d", errUTXOsCouldNotSelect, totalUTXOsSum, txCostWithMinChange)
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

func getMaxUtxoCount(config *cardano.CardanoChainConfig, prevUtxosCnt int) int {
	return max(int(config.MaxUtxoCount)-prevUtxosCnt, 0) //nolint:gosec
}

func (cco *CardanoChainOperations) generatePolicyAndMultisig(
	validatorsData []eth.ValidatorChainData) (*cardano.ApexPolicyScripts, *cardano.ApexAddresses, error) {
	keyHashes, err := cardano.NewApexKeyHashes(validatorsData)
	if err != nil {
		return nil, nil, err
	}

	policyScripts := cardano.NewApexPolicyScripts(keyHashes)

	addresses, err := cardano.NewApexAddresses(cco.cardanoCliBinary, uint(cco.config.NetworkMagic), policyScripts)
	if err != nil {
		return nil, nil, err
	}

	return &policyScripts, &addresses, nil
}

// CreateValidatorSetChangeTx implements core.ChainOperations.
func (cco *CardanoChainOperations) CreateValidatorSetChangeTx(ctx context.Context, chainID string, nextBatchID uint64,
	bridgeSmartContract eth.IBridgeSmartContract, validatorsKeys validatorobserver.ValidatorsPerChain,
	lastBatchID uint64, lastBatchType uint8,
) (bool, *core.GeneratedBatchTxData, error) {
	// get validators data
	validatorsData, ok := validatorsKeys[chainID]
	if !ok {
		return false, nil, fmt.Errorf("couldn't find keys for chain:%s", chainID)
	}

	// new validator set policy, multisig & fee address
	_, newAddresses, err :=
		cco.generatePolicyAndMultisig(validatorsData.Keys)
	if err != nil {
		return false, nil, err
	}

	cco.logger.Debug("NEW MULTISIG & FEE ADDRESS",
		"ms", newAddresses.Multisig.Payment,
		"fee", newAddresses.Fee.Payment,
		"ms.stake", newAddresses.Multisig.Stake,
		"fee.stake", newAddresses.Fee.Stake)

	// get active validator set from bridge smart contract
	activeValidatorsData, err := cco.getValidatorsChainData(ctx, bridgeSmartContract, chainID)
	if err != nil {
		return false, nil, err
	}

	// active validator set policy, multisig & fee address
	activePolicy, activeAddresses, err :=
		cco.generatePolicyAndMultisig(activeValidatorsData)
	if err != nil {
		return false, nil, err
	}

	protocolParams, err := cco.txProvider.GetProtocolParameters(ctx)
	if err != nil {
		return false, nil, err
	}

	// get filtered & limited utxos
	multisigUtxos, feeUtxos, isFeeOnly, err := cco.getUTXOsForValidatorChange(
		activeAddresses.Multisig.Payment, activeAddresses.Fee.Payment, validatorsData.SlotNumber)
	if err != nil {
		return false, nil, err
	}

	// if there are no transactions
	// or only one transaction with an amount less than twice the minimum UTXO amount
	// return the validator final
	if len(feeUtxos) == 0 || (len(feeUtxos) == 1 && feeUtxos[0].Output.Amount < common.MinUtxoAmountDefault*2) {
		return false, &core.GeneratedBatchTxData{
			BatchType: uint8(ValidatorSetFinal),
			TxRaw:     []byte{},
			TxHash:    "",
		}, nil
	}

	metadataStruct := common.BatchExecutedMetadata{
		BridgingTxType: common.BridgingTxTypeBatchExecution,
		BatchNonceID:   nextBatchID,
	}
	if isFeeOnly {
		metadataStruct.IsFeeOnlyTx = 1
	}

	metadata, err := common.MarshalMetadata(common.MetadataEncodingTypeJSON, metadataStruct)
	if err != nil {
		return false, nil, err
	}

	// last slot number from indexer
	slotNumber, err := cco.getSlotNumber()
	if err != nil {
		return false, nil, err
	}

	var (
		txRaw  []byte
		txHash string
	)

	if isFeeOnly {
		// last transaction sends funds from active fee key to new fee key
		output := cardanowallet.TxOutput{
			Addr:   newAddresses.Fee.Payment,
			Amount: 0,
			Tokens: []cardanowallet.TokenAmount{},
		}

		txRaw, txHash, err = cardano.CreateOnlyFeeTx(
			cco.cardanoCliBinary,
			uint(cco.config.NetworkMagic),
			protocolParams,
			slotNumber+cco.config.TTLSlotNumberInc,
			metadata,
			&cardano.TxInputInfo{
				PolicyScript: activePolicy.Fee.Payment,
				Address:      activeAddresses.Fee.Payment,
				TxInputs:     convertUTXOsToTxInputs(feeUtxos),
			},
			output,
		)
		if err != nil {
			return false, nil, err
		}
	} else {
		var sum uint64
		for _, u := range multisigUtxos {
			sum += u.Output.Amount
		}

		outputs := []cardanowallet.TxOutput{
			{
				Addr:   newAddresses.Multisig.Payment,
				Amount: sum,
				Tokens: []cardanowallet.TokenAmount{},
			},
		}

		txRaw, txHash, err = cardano.CreateTx(
			cco.cardanoCliBinary,
			uint(cco.config.NetworkMagic),
			protocolParams,
			slotNumber+cco.config.TTLSlotNumberInc,
			metadata,
			cardano.TxInputInfos{
				MultiSig: &cardano.TxInputInfo{
					PolicyScript: activePolicy.Multisig.Payment,
					Address:      activeAddresses.Multisig.Payment,
					TxInputs:     convertUTXOsToTxInputs(multisigUtxos),
				},
				MultiSigFee: &cardano.TxInputInfo{
					PolicyScript: activePolicy.Fee.Payment,
					Address:      activeAddresses.Fee.Payment,
					TxInputs:     convertUTXOsToTxInputs(feeUtxos),
				},
			},
			outputs,
		)
		if err != nil {
			return false, nil, err
		}
	}

	return true, &core.GeneratedBatchTxData{
		BatchType: uint8(ValidatorSet),
		TxRaw:     txRaw,
		TxHash:    txHash,
	}, nil
}

func (cco *CardanoChainOperations) getUTXOsForValidatorChange(
	multisigAddress, multisigFeeAddress string, slot uint64,
) (multisigUtxos []*indexer.TxInputOutput, feeUtxos []*indexer.TxInputOutput, isFeeOnly bool, err error) {
	// Fetch all UTXOs for the multisig address (including those after the given slot).
	allMultisigUtxos, err := cco.db.GetAllTxOutputs(multisigAddress, true)
	if err != nil {
		return
	}

	multisigUtxos = make([]*indexer.TxInputOutput, 0, len(allMultisigUtxos))

	// Filter out any UTXOs created after the given slot.
	for _, utxo := range allMultisigUtxos {
		if utxo.Output.Slot <= slot {
			multisigUtxos = append(multisigUtxos, utxo)
		}
	}

	feeUtxos, err = cco.db.GetAllTxOutputs(multisigFeeAddress, true)
	if err != nil {
		return
	}

	multisigUtxos = filterOutTokenUtxos(multisigUtxos)
	feeUtxos = filterOutTokenUtxos(feeUtxos)

	if len(feeUtxos) == 0 {
		return multisigUtxos, feeUtxos, false, nil
	}

	if len(multisigUtxos) == 0 {
		isFeeOnly = true
		feeUtxos = feeUtxos[:min(cco.config.MaxUtxoCount, uint(len(feeUtxos)))]

		return
	}

	feeUtxos = feeUtxos[:min(int(cco.config.MaxFeeUtxoCount), len(feeUtxos))] //nolint:gosec
	maxUtxosCnt := min(getMaxUtxoCount(cco.config, len(feeUtxos)), len(multisigUtxos))
	multisigUtxos = multisigUtxos[:maxUtxosCnt]

	return
}
