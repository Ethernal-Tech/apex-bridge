package batcher

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sort"

	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	cardano "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

var _ core.ChainOperations = (*CardanoChainOperations)(nil)

// TODO: Get from protocol parameters, maybe add to core.CardanoChainConfig
var minUtxoAmount = uint64(1000000)
var maxUtxoCount = 410

// TODO: Get real tx size from protocolParams/config
var maxTxSize = 16000

type CardanoChainOperations struct {
	Config        core.CardanoChainConfig
	CardanoWallet cardano.CardanoWallet
}

func NewCardanoChainOperations(config core.CardanoChainConfig, wallet cardano.CardanoWallet) *CardanoChainOperations {
	return &CardanoChainOperations{
		Config:        config,
		CardanoWallet: wallet,
	}
}

// GenerateBatchTransaction implements core.ChainOperations.
func (cco *CardanoChainOperations) GenerateBatchTransaction(
	ctx context.Context,
	bridgeSmartContract eth.IBridgeSmartContract,
	destinationChain string,
	confirmedTransactions []eth.ConfirmedTransaction,
	batchNonceId *big.Int) ([]byte, string, *eth.UTXOs, error) {

	var outputs []cardanowallet.TxOutput
	var txCost *big.Int = big.NewInt(0)
	for _, transaction := range confirmedTransactions {
		for _, receiver := range transaction.Receivers {
			outputs = append(outputs, cardanowallet.TxOutput{
				Addr:   receiver.DestinationAddress,
				Amount: receiver.Amount.Uint64(),
			})
			txCost.Add(txCost, receiver.Amount)
		}
	}

	metadata, err := cardano.CreateBatchMetaData(batchNonceId)
	if err != nil {
		return nil, "", nil, err
	}

	txProvider, err := cardanowallet.NewTxProviderBlockFrost(cco.Config.BlockfrostUrl, cco.Config.BlockfrostAPIKey)
	if err != nil {
		return nil, "", nil, err
	}

	protocolParams, err := txProvider.GetProtocolParameters(ctx)
	if err != nil {
		return nil, "", nil, err
	}

	lastObservedBlock, err := bridgeSmartContract.GetLastObservedBlock(ctx, destinationChain)
	if err != nil {
		return nil, "", nil, err
	}

	validatorsData, err := bridgeSmartContract.GetValidatorsCardanoData(ctx, destinationChain)
	if err != nil {
		return nil, "", nil, err
	}

	var multisigKeyHashes []string = make([]string, len(validatorsData))
	var multisigFeeKeyHashes []string = make([]string, len(validatorsData))
	foundVerificationKey := false
	foundFeeVerificationKey := false
	for i, validator := range validatorsData {
		multisigKeyHashes[i] = validator.KeyHash
		multisigFeeKeyHashes[i] = validator.KeyHashFee
		validatorKeyBytes, err := hex.DecodeString(validator.VerifyingKey)
		if err != nil {
			return nil, "", nil, err
		}

		if bytes.Equal(cco.CardanoWallet.MultiSig.GetVerificationKey(), validatorKeyBytes) {
			foundVerificationKey = true
		}

		validatorKeyBytes, err = hex.DecodeString(validator.VerifyingKeyFee)
		if err != nil {
			return nil, "", nil, err
		}

		if bytes.Equal(cco.CardanoWallet.MultiSigFee.GetVerificationKey(), validatorKeyBytes) {
			foundFeeVerificationKey = true
		}
	}

	if !foundVerificationKey {
		return nil, "", nil, fmt.Errorf("verifying key of current batcher wasn't found in validators data queried from smart contract")
	}

	if !foundFeeVerificationKey {
		return nil, "", nil, fmt.Errorf("verifying fee key of current batcher wasn't found in validators data queried from smart contract")
	}

	multisigPolicyScript, err := cardanowallet.NewPolicyScript(multisigKeyHashes, int(cco.Config.AtLeastValidators))
	if err != nil {
		return nil, "", nil, err
	}

	multisigFeePolicyScript, err := cardanowallet.NewPolicyScript(multisigFeeKeyHashes, int(cco.Config.AtLeastValidators))
	if err != nil {
		return nil, "", nil, err
	}

	multisigAddress, err := multisigPolicyScript.CreateMultiSigAddress(cco.Config.TestNetMagic)
	if err != nil {
		return nil, "", nil, err
	}

	multisigFeeAddress, err := multisigFeePolicyScript.CreateMultiSigAddress(cco.Config.TestNetMagic)
	if err != nil {
		return nil, "", nil, err
	}

	txUtxos, err := GetInputUtxos(ctx, bridgeSmartContract, destinationChain, txCost)
	if err != nil {
		return nil, "", nil, err
	}

	txInfos := &cardano.TxInputInfos{
		TestNetMagic: cco.Config.TestNetMagic,
		MultiSig: &cardano.TxInputInfo{
			PolicyScript: multisigPolicyScript,
			Address:      multisigAddress,
		},
		MultiSigFee: &cardano.TxInputInfo{
			PolicyScript: multisigFeePolicyScript,
			Address:      multisigFeeAddress,
		},
	}

	return cco.CreateBatchTx(txUtxos, txCost, metadata, protocolParams, txInfos, outputs, lastObservedBlock.BlockSlot)
}

// SignBatchTransaction implements core.ChainOperations.
func (cco *CardanoChainOperations) SignBatchTransaction(txHash string) ([]byte, []byte, error) {

	witnessMultiSig, err := cardano.CreateTxWitness(txHash, cco.CardanoWallet.MultiSig)
	if err != nil {
		return nil, nil, err
	}

	witnessMultiSigFee, err := cardano.CreateTxWitness(txHash, cco.CardanoWallet.MultiSigFee)
	if err != nil {
		return nil, nil, err
	}

	return witnessMultiSig, witnessMultiSigFee, nil
}

/* UTXOs are sorted by Nonce and taken from first to last until txCost has been met or maxUtxoCount reached
 * if txCost has been met, tx is created regularly
 * if maxUtxoCount has been reached, we replace smallest UTXO with first next bigger one until we reach txCost
 */
func (cco *CardanoChainOperations) CreateBatchTx(inputUtxos *contractbinding.IBridgeContractStructsUTXOs, txCost *big.Int,
	metadata []byte, protocolParams []byte, txInfos *cardano.TxInputInfos, outputs []cardanowallet.TxOutput, slotNumber uint64) (
	[]byte, string, *eth.UTXOs, error) {

	// For now we are taking all available UTXOs as fee (should always be 1-2 of them)
	multisigFeeInputs := make([]cardanowallet.TxInput, len(inputUtxos.FeePayerOwnedUTXOs))
	multisigFeeInputsSum := big.NewInt(0)
	for i, utxo := range inputUtxos.FeePayerOwnedUTXOs {
		multisigFeeInputs[i] = cardanowallet.TxInput{
			Hash:  utxo.TxHash,
			Index: uint32(utxo.TxIndex.Uint64()),
		}
		multisigFeeInputsSum.Add(multisigFeeInputsSum, utxo.Amount)
	}
	txInfos.MultiSigFee.TxInputUTXOs = cardano.TxInputUTXOs{Inputs: multisigFeeInputs, InputsSum: multisigFeeInputsSum.Uint64()}

	utxoCount := len(outputs) + len(multisigFeeInputs)

	// Create initial UTXO set
	txCostWithMinChange := new(big.Int).SetUint64(0)
	txCostWithMinChange.Add(txCost, big.NewInt(int64(minUtxoAmount)))

	chosenUTXOs := make([]contractbinding.IBridgeContractStructsUTXO, 0)
	chosenUTXOsSum := big.NewInt(0)
	isUtxosOk := false
	for _, utxo := range inputUtxos.MultisigOwnedUTXOs {
		chosenUTXOs = append(chosenUTXOs, utxo)
		utxoCount++
		chosenUTXOsSum.Add(chosenUTXOsSum, utxo.Amount)

		if utxoCount > maxUtxoCount {
			minChosenUTXO, minChosenUTXOIdx := FindMinUtxo(chosenUTXOs)

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
		return nil, "", nil, errors.New("fatal error, couldn't select UTXOs")
	}

	// Create inputs needed for tx from chosenUTXOs set
	var multisigInputs []cardanowallet.TxInput
	for _, utxo := range chosenUTXOs {
		multisigInputs = append(multisigInputs, cardanowallet.TxInput{
			Hash:  utxo.TxHash,
			Index: uint32(utxo.TxIndex.Uint64()),
		})
	}

	inputUtxos.MultisigOwnedUTXOs = chosenUTXOs
	txInfos.MultiSig.TxInputUTXOs = cardano.TxInputUTXOs{Inputs: multisigInputs, InputsSum: chosenUTXOsSum.Uint64()}

	// Create Tx
	rawTx, txHash, err := cardano.CreateTx(cco.Config.TestNetMagic, protocolParams, slotNumber+cardano.TTLSlotNumberInc,
		metadata, txInfos, outputs)
	if err != nil {
		return nil, "", nil, err
	}

	if len(rawTx) > maxTxSize {
		return nil, "", nil, errors.New("fatal error, tx size too big")
	}

	return rawTx, txHash, inputUtxos, err
}

func GetInputUtxos(ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract, destinationChain string, txCost *big.Int) (
	*contractbinding.IBridgeContractStructsUTXOs, error) {

	inputUtxos, err := bridgeSmartContract.GetAvailableUTXOs(ctx, destinationChain)
	if err != nil {
		return nil, err
	}

	sort.Slice(inputUtxos.MultisigOwnedUTXOs, func(i, j int) bool {
		return inputUtxos.MultisigOwnedUTXOs[i].Nonce < inputUtxos.MultisigOwnedUTXOs[j].Nonce
	})
	sort.Slice(inputUtxos.FeePayerOwnedUTXOs, func(i, j int) bool {
		return inputUtxos.FeePayerOwnedUTXOs[i].Nonce < inputUtxos.FeePayerOwnedUTXOs[j].Nonce
	})

	return inputUtxos, err
}

func FindMinUtxo(utxos []eth.UTXO) (eth.UTXO, int) {
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
