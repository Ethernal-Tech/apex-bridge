package batcher

import (
	"context"
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

	var keyHashes []string = make([]string, len(validatorsData))
	foundVerificationKey := false
	for _, validator := range validatorsData {
		keyHashes = append(keyHashes, validator.KeyHash)
		if string(cco.CardanoWallet.MultiSig.GetVerificationKey()) == validator.VerifyingKey {
			foundVerificationKey = true
		}
	}

	if !foundVerificationKey {
		return nil, "", nil, fmt.Errorf("verifying key of current batcher wasn't found in validators data queried from smart contract")
	}

	multisigPolicyScript, err := cardanowallet.NewPolicyScript(keyHashes, int(cco.Config.AtLeastValidators))
	if err != nil {
		return nil, "", nil, err
	}

	keyHashes = make([]string, len(validatorsData))
	foundVerificationKey = false
	for _, validator := range validatorsData {
		keyHashes = append(keyHashes, validator.KeyHash)
		if string(cco.CardanoWallet.MultiSigFee.GetVerificationKey()) == validator.VerifyingKeyFee {
			foundVerificationKey = true
		}
	}

	if !foundVerificationKey {
		return nil, "", nil, fmt.Errorf("verifying fee key of current batcher wasn't found in validators data queried from smart contract")
	}

	multisigFeePolicyScript, err := cardanowallet.NewPolicyScript(keyHashes, int(cco.Config.AtLeastValidators))
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
 * if maxUtxoCount has been reached, we replace smallest UTXO with first biggest one that will cover the txCost
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
	minChosenUTXO := inputUtxos.MultisigOwnedUTXOs[0]
	minChosenUTXOIdx := 0
	isUtxosOk := false
	for idx, utxo := range inputUtxos.MultisigOwnedUTXOs {

		// If max UTXO count is reached we will replace last UTXO with bigger UTXO
		if utxoCount >= maxUtxoCount {
			// Check if curent UTXO is bigger than smallest UTXO
			if utxo.Amount.Cmp(minChosenUTXO.Amount) < 1 {
				continue
			}
			chosenUTXOsSum.Sub(chosenUTXOsSum, minChosenUTXO.Amount)
			chosenUTXOsSum.Add(chosenUTXOsSum, utxo.Amount)

			chosenUTXOs[minChosenUTXOIdx] = utxo
			minChosenUTXO = utxo
		} else {
			chosenUTXOs = append(chosenUTXOs, utxo)
			utxoCount++
			chosenUTXOsSum.Add(chosenUTXOsSum, utxo.Amount)

			if utxo.Amount.Cmp(minChosenUTXO.Amount) == -1 {
				minChosenUTXO = utxo
				minChosenUTXOIdx = idx
			}
		}

		// If required txCost was reached we don't need more UTXOs
		// chosenUTXOsSum >= txCostWithMinUtxo || chosenUTXOsSum == txCost
		if chosenUTXOsSum.Cmp(txCostWithMinChange) >= 1 || chosenUTXOsSum.Cmp(txCost) == 0 {
			isUtxosOk = true
			break
		}
	}

	if !isUtxosOk {
		return nil, "", nil, errors.New("fatal error, couldn't select UTXOs")
	}

	// Create inputs and sums needed for tx from chosenUTXOs set
	var multisigInputs []cardanowallet.TxInput
	multisigInputsSum := big.NewInt(0)
	for _, utxo := range chosenUTXOs {
		multisigInputs = append(multisigInputs, cardanowallet.TxInput{
			Hash:  utxo.TxHash,
			Index: uint32(utxo.TxIndex.Uint64()),
		})
		multisigInputsSum.Add(multisigInputsSum, utxo.Amount)
	}

	// inputUtxos.MultisigOwnedUTXOs = chosenUTXOs
	txInfos.MultiSig.TxInputUTXOs = cardano.TxInputUTXOs{Inputs: multisigInputs, InputsSum: multisigInputsSum.Uint64()}

	// Create Tx
	rawTx, txHash, err := cardano.CreateTx(cco.Config.TestNetMagic, protocolParams, slotNumber+cardano.TTLSlotNumberInc,
		metadata, txInfos, outputs)
	if err != nil {
		return nil, "", nil, err
	}

	if len(rawTx) > maxTxSize {
		return nil, "", nil, errors.New("fatal error, tx size too big")
	}

	inputUtxos.MultisigOwnedUTXOs = chosenUTXOs
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
