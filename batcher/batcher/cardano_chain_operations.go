package batcher

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sort"

	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	cardano "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
)

var (
	_ core.ChainOperations = (*CardanoChainOperations)(nil)

	errBatchProposerDataNotSet = errors.New("batch proposer data not set")
	errBatchKeyNotFound        = errors.New("key of current batcher wasn't found in validators data queried from smart contract") //nolint:lll
)

// nolintlint TODO: Get from protocol parameters, maybe add to core.CardanoChainConfig
// Get real tx size from protocolParams/config
const (
	minUtxoAmount        = uint64(1_000_000)
	maxFeeUtxoCount      = 60
	maxUtxoCount         = 300
	maxTxSize            = 16000
	minFeeUtxosSum       = 3_000_000
	takeAtLeastUtxoCount = 50

	proposerEpochBlockCount   = 20
	proposerSlotToleranceDiff = 1000
)

type txVariableDataInfo struct {
	slot          uint64
	multisigUtxos []cardanowallet.Utxo
	feeUtxos      []cardanowallet.Utxo
}

func (txvar txVariableDataInfo) toBatchProposal() eth.BatchProposerData {
	return eth.BatchProposerData{
		Slot:          txvar.slot,
		MultisigUTXOs: convertUTXOsToEthUTXOs(txvar.multisigUtxos),
		FeePayerUTXOs: convertUTXOsToEthUTXOs(txvar.feeUtxos),
	}
}

func (txvar *txVariableDataInfo) update(
	slotRoundingThreshold, txCost uint64,
) (err error) {
	txvar.slot = getRoundedSlot(txvar.slot, slotRoundingThreshold)
	txvar.feeUtxos = txvar.feeUtxos[:min(len(txvar.feeUtxos), maxFeeUtxoCount)]
	txvar.multisigUtxos, err = getNeededUtxos(
		txvar.multisigUtxos, txCost, minUtxoAmount, len(txvar.feeUtxos), maxUtxoCount, takeAtLeastUtxoCount)

	return err
}

func (txvar *txVariableDataInfo) validateAndUpdate(
	proposal eth.BatchProposerData, slotRoundingThreshold, txCost uint64,
) (err error) {
	txvar.slot = getRoundedSlot(txvar.slot, slotRoundingThreshold)

	if txvar.slot >= proposal.Slot && txvar.slot-proposal.Slot > proposerSlotToleranceDiff ||
		txvar.slot < proposal.Slot && proposal.Slot-txvar.slot > proposerSlotToleranceDiff {
		return fmt.Errorf("proposed slot is not good: %d. current: %d", proposal.Slot, txvar.slot)
	}

	if len(proposal.FeePayerUTXOs) > maxFeeUtxoCount {
		return fmt.Errorf("proposed fee utxos count is not good: %d. max: %d",
			len(proposal.FeePayerUTXOs), maxFeeUtxoCount)
	}

	if len(proposal.FeePayerUTXOs)+len(proposal.MultisigUTXOs) > maxUtxoCount {
		return fmt.Errorf("proposed fee utxos count is not good: %d. max: %d",
			len(proposal.FeePayerUTXOs)+len(proposal.MultisigUTXOs), maxUtxoCount)
	}

	txvar.multisigUtxos, err = validateAndRetrieveUtxos(
		txvar.multisigUtxos, proposal.MultisigUTXOs, txCost, minUtxoAmount)
	if err != nil {
		return err
	}

	txvar.feeUtxos, err = validateAndRetrieveUtxos(
		txvar.feeUtxos, proposal.FeePayerUTXOs, minFeeUtxosSum, minUtxoAmount)

	return err
}

type CardanoChainOperations struct {
	Config     *cardano.CardanoChainConfig
	Wallet     *cardano.CardanoWallet
	TxProvider cardanowallet.ITxProvider

	logger hclog.Logger
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
	multisigKeyHashes, multisigFeeKeyHashes, validatorIdx, err := cco.getCardanoData(
		ctx, bridgeSmartContract, destinationChain)
	if err != nil {
		return nil, err
	}

	blockNumber, err := bridgeSmartContract.GetBlockNumber(ctx)
	if err != nil {
		return nil, err
	}

	txOutputs := getOutputs(confirmedTransactions)
	// determine if current validator is proposer or not
	proposerIdx := getProposerIndex(blockNumber, proposerEpochBlockCount, len(multisigKeyHashes))
	isProposer := proposerIdx == validatorIdx

	// retrieve current proposal
	proposal, err := bridgeSmartContract.GetBatchProposerData(ctx, destinationChain)
	if err != nil {
		return nil, err
	} else if proposal.Slot == 0 && !isProposer { // if not proposer -> proposal must be set
		return nil, errBatchProposerDataNotSet
	}

	metadata, err := cardano.CreateBatchMetaData(batchNonceID)
	if err != nil {
		return nil, err
	}

	protocolParams, err := cco.TxProvider.GetProtocolParameters(ctx)
	if err != nil {
		return nil, err
	}

	txInfos, err := cco.createTxInfos(multisigKeyHashes, multisigFeeKeyHashes)
	if err != nil {
		return nil, err
	}

	txVariableData, err := cco.getTxVariableData(ctx, txInfos.MultiSig.Address, txInfos.MultiSigFee.Address)
	if err != nil {
		return nil, err
	}

	err = cco.validateAndUpdateTxVariableData(
		isProposer, &txVariableData, proposal, cco.Config.SlotRoundingThreshold, txOutputs.Sum.Uint64())
	if err != nil {
		return nil, err
	}

	txInfos.MultiSig.TxInputs = convertUTXOsToTxInputs(txVariableData.multisigUtxos)
	txInfos.MultiSigFee.TxInputs = convertUTXOsToTxInputs(txVariableData.feeUtxos)

	txRaw, txHash, err := cardano.CreateTx(
		uint(cco.Config.TestNetMagic),
		protocolParams,
		txVariableData.slot+cco.Config.TTLSlotNumberInc,
		metadata,
		txInfos,
		txOutputs.Outputs,
	)
	if err != nil {
		return nil, err
	}

	if len(txRaw) > maxTxSize {
		return nil, errors.New("fatal error, tx size too big")
	}

	return &core.GeneratedBatchTxData{
		TxRaw:        txRaw,
		TxHash:       txHash,
		Proposal:     txVariableData.toBatchProposal(),
		ProposerIdx:  proposerIdx,
		ValidatorIdx: validatorIdx,
		BlockNumber:  blockNumber,
	}, nil
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

func (cco *CardanoChainOperations) getCardanoData(
	ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract, chainID string,
) ([]string, []string, int, error) {
	validatorsData, err := bridgeSmartContract.GetValidatorsCardanoData(ctx, chainID)
	if err != nil {
		return nil, nil, -1, err
	}

	var (
		multisigKeyHashes     = make([]string, len(validatorsData))
		multisigFeeKeyHashes  = make([]string, len(validatorsData))
		verificationKeyIdx    = -1
		feeVerificationKeyIdx = -1
	)

	for i, validator := range validatorsData {
		multisigKeyHashes[i], err = cardanowallet.GetKeyHash(validator.VerifyingKey[:])
		if err != nil {
			return nil, nil, -1, err
		}

		if bytes.Equal(cco.Wallet.MultiSig.GetVerificationKey(), validator.VerifyingKey[:]) {
			verificationKeyIdx = i
		}

		multisigFeeKeyHashes[i], err = cardanowallet.GetKeyHash(validator.VerifyingKeyFee[:])
		if err != nil {
			return nil, nil, -1, err
		}

		if bytes.Equal(cco.Wallet.MultiSigFee.GetVerificationKey(), validator.VerifyingKeyFee[:]) {
			feeVerificationKeyIdx = i
		}
	}

	if verificationKeyIdx == -1 {
		return nil, nil, -1, fmt.Errorf("multisig: %w", errBatchKeyNotFound)
	}

	if feeVerificationKeyIdx == -1 {
		return nil, nil, -1, fmt.Errorf("fee: %w", errBatchKeyNotFound)
	}

	return multisigKeyHashes, multisigFeeKeyHashes, verificationKeyIdx, nil
}

func (cco *CardanoChainOperations) createTxInfos(
	multisigKeyHashes []string, multisigFeeKeyHashes []string,
) (cardano.TxInputInfos, error) {
	multisigPolicyScript, err := cardanowallet.NewPolicyScript(
		multisigKeyHashes, int(common.GetRequiredSignaturesForConsensus(uint64(len(multisigKeyHashes)))))
	if err != nil {
		return cardano.TxInputInfos{}, err
	}

	multisigFeePolicyScript, err := cardanowallet.NewPolicyScript(
		multisigFeeKeyHashes, int(common.GetRequiredSignaturesForConsensus(uint64(len(multisigFeeKeyHashes)))))
	if err != nil {
		return cardano.TxInputInfos{}, err
	}

	multisigAddress, err := multisigPolicyScript.CreateMultiSigAddress(uint(cco.Config.TestNetMagic))
	if err != nil {
		return cardano.TxInputInfos{}, err
	}

	multisigFeeAddress, err := multisigFeePolicyScript.CreateMultiSigAddress(uint(cco.Config.TestNetMagic))
	if err != nil {
		return cardano.TxInputInfos{}, err
	}

	return cardano.TxInputInfos{
		MultiSig: &cardano.TxInputInfo{
			PolicyScript: multisigPolicyScript,
			Address:      multisigAddress,
		},
		MultiSigFee: &cardano.TxInputInfo{
			PolicyScript: multisigFeePolicyScript,
			Address:      multisigFeeAddress,
		},
	}, nil
}

func (cco *CardanoChainOperations) getTxVariableData(
	ctx context.Context, multisigAddr, feeAddr string,
) (txVariableDataInfo, error) {
	tipData, err := cco.TxProvider.GetTip(ctx)
	if err != nil {
		return txVariableDataInfo{}, err
	}

	multisigUtxos, err := cco.TxProvider.GetUtxos(ctx, multisigAddr)
	if err != nil {
		return txVariableDataInfo{}, err
	}

	feeUtxos, err := cco.TxProvider.GetUtxos(ctx, feeAddr)
	if err != nil {
		return txVariableDataInfo{}, err
	}

	sort.Slice(multisigUtxos, func(i, j int) bool {
		return multisigUtxos[i].Amount > multisigUtxos[j].Amount
	})

	sort.Slice(feeUtxos, func(i, j int) bool {
		return feeUtxos[i].Amount > feeUtxos[j].Amount
	})

	return txVariableDataInfo{
		slot:          tipData.Slot,
		multisigUtxos: multisigUtxos,
		feeUtxos:      feeUtxos,
	}, nil
}

func (cco *CardanoChainOperations) validateAndUpdateTxVariableData(
	isProposer bool, txVariableData *txVariableDataInfo,
	proposal eth.BatchProposerData, slotRoundingThreshold, txCost uint64,
) error {
	if !isProposer {
		return txVariableData.validateAndUpdate(proposal, slotRoundingThreshold, txCost)
	}

	// first try to use previous proposal if exists. This way consensus will be reached sooner
	if proposal.Slot != 0 {
		oldTxVariableData := *txVariableData

		if err := txVariableData.validateAndUpdate(proposal, slotRoundingThreshold, txCost); err == nil {
			return nil
		}

		*txVariableData = oldTxVariableData
	}

	return txVariableData.update(slotRoundingThreshold, txCost)
}

func getRoundedSlot(slot, threshold uint64) uint64 {
	return ((slot + threshold - 1) / threshold) * threshold
}

func getProposerIndex(blockNumber, proposerEpochBlockCount uint64, validatorsCount int) int {
	return int((blockNumber / proposerEpochBlockCount) % uint64(validatorsCount))
}

func convertUTXOsToTxInputs(utxos []cardanowallet.Utxo) (result cardanowallet.TxInputs) {
	result.Inputs = make([]cardanowallet.TxInput, len(utxos))
	result.Sum = 0

	for i, utxo := range utxos {
		result.Inputs[i] = cardanowallet.TxInput{
			Hash:  utxo.Hash,
			Index: utxo.Index,
		}
		result.Sum += utxo.Amount
	}

	return result
}

func convertUTXOsToEthUTXOs(utxos []cardanowallet.Utxo) (result []eth.UTXO) {
	result = make([]eth.UTXO, len(utxos))

	for i, utxo := range utxos {
		result[i] = eth.UTXO{
			TxHash:  indexer.NewHashFromHexString(utxo.Hash),
			TxIndex: uint64(utxo.Index),
		}
	}

	return result
}

func validateAndRetrieveUtxos(
	inputUtxos []cardanowallet.Utxo, desiredUtxos []eth.UTXO, atLeastSum uint64, minUtxoAmount uint64,
) ([]cardanowallet.Utxo, error) {
	mp := map[indexer.Hash]map[uint32]cardanowallet.Utxo{}

	for _, inp := range inputUtxos {
		key := indexer.NewHashFromHexString(inp.Hash)
		subMap, exists := mp[key]

		if !exists {
			subMap = map[uint32]cardanowallet.Utxo{}
			mp[key] = subMap
		}

		subMap[inp.Index] = inp
	}

	sum := uint64(0)
	resultUtxos := make([]cardanowallet.Utxo, len(desiredUtxos))

	for i, utxo := range desiredUtxos {
		subMap, exists := mp[utxo.TxHash]
		if !exists {
			return nil, fmt.Errorf("proposed utxo does not exists: %s, %d", utxo.TxHash, utxo.TxIndex)
		}

		inp, exists := subMap[uint32(utxo.TxIndex)]
		if !exists {
			return nil, fmt.Errorf("proposed utxo does not exists: %s, %d", utxo.TxHash, utxo.TxIndex)
		}

		sum += inp.Amount
		resultUtxos[i] = inp
	}

	if sum != atLeastSum && sum < atLeastSum+minUtxoAmount {
		return nil, fmt.Errorf("proposed utxos sum is not good: %d vs %d", sum, atLeastSum)
	}

	return resultUtxos, nil
}

// getNeededUtxos returns only needed input utxos
func getNeededUtxos(
	inputUTXOs []cardanowallet.Utxo, desiredAmount uint64,
	minUtxoAmount uint64, utxoCount int, maxUtxoCount int, takeAtLeastUtxoCount int,
) (chosenUTXOs []cardanowallet.Utxo, err error) {
	// Create initial UTXO set
	txCostWithMinChange := desiredAmount + minUtxoAmount

	// algorithm that chooses multisig UTXOs
	chosenUTXOsSum := uint64(0)
	isUtxosOk := false

	for i, utxo := range inputUTXOs {
		chosenUTXOs = append(chosenUTXOs, utxo)
		utxoCount++

		chosenUTXOsSum += utxo.Amount // overflow is not considered

		if utxoCount > maxUtxoCount {
			minChosenUTXO, minChosenUTXOIdx := findMinUtxo(chosenUTXOs)

			chosenUTXOs[minChosenUTXOIdx] = utxo
			chosenUTXOsSum -= minChosenUTXO.Amount
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
		return nil, fmt.Errorf("could not select utxos for sum: %d. max sum is: %d", desiredAmount, chosenUTXOsSum)
	}

	return chosenUTXOs, nil
}

func findMinUtxo(utxos []cardanowallet.Utxo) (cardanowallet.Utxo, int) {
	min := utxos[0]
	idx := 0

	for i, utxo := range utxos[1:] {
		if utxo.Amount < min.Amount {
			min = utxo
			idx = i + 1
		}
	}

	return min, idx
}

func getOutputs(txs []eth.ConfirmedTransaction) cardano.TxOutputs {
	receiversMap := map[string]uint64{}

	for _, transaction := range txs {
		for _, receiver := range transaction.Receivers {
			receiversMap[receiver.DestinationAddress] += receiver.Amount
		}
	}

	result := cardano.TxOutputs{
		Outputs: make([]cardanowallet.TxOutput, 0, len(receiversMap)),
		Sum:     big.NewInt(0),
	}

	for addr, amount := range receiversMap {
		if amount <= 0 {
			// this should be logged once
			continue
		}

		result.Outputs = append(result.Outputs, cardanowallet.TxOutput{
			Addr:   addr,
			Amount: amount,
		})
		result.Sum.Add(result.Sum, new(big.Int).SetUint64(amount))
	}

	// sort outputs because all batchers should have same order of outputs
	sort.Slice(result.Outputs, func(i, j int) bool {
		return result.Outputs[i].Addr < result.Outputs[j].Addr
	})

	return result
}
