package batcher

import (
	"context"
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

func (cco *CardanoChainOperations) CreateBatchTx(inputUtxos *contractbinding.IBridgeContractStructsUTXOs, txCost *big.Int,
	metadata []byte, protocolParams []byte, txInfos *cardano.TxInputInfos, outputs []cardanowallet.TxOutput, slotNumber uint64) (
	[]byte, string, *eth.UTXOs, error) {

	var err error = nil

	// We are taking all available UTXOs as fee (for now)
	var multisigFeeInputs []cardanowallet.TxInput
	var multisigFeeInputsSum uint64 = 0
	for _, utxo := range inputUtxos.FeePayerOwnedUTXOs {
		multisigFeeInputs = append(multisigFeeInputs, cardanowallet.TxInput{
			Hash:  utxo.TxHash,
			Index: uint32(utxo.TxIndex.Uint64()),
		})
		multisigFeeInputsSum += utxo.Amount.Uint64()
	}
	txInfos.MultiSigFee.TxInputUTXOs = cardano.TxInputUTXOs{Inputs: multisigFeeInputs, InputsSum: multisigFeeInputsSum}

	// Create initial UTXO set
	chosenUTXOs := make([]contractbinding.IBridgeContractStructsUTXO, len(inputUtxos.MultisigOwnedUTXOs))
	var chosenUTXOsCount = 0
	var chosenUTXOsSum uint64 = 0
	for _, utxo := range inputUtxos.MultisigOwnedUTXOs {
		chosenUTXOs = append(chosenUTXOs, utxo)
		chosenUTXOsSum += utxo.Amount.Uint64()
		chosenUTXOsCount++

		if chosenUTXOsSum > txCost.Uint64()+minUtxoAmount || chosenUTXOsSum == txCost.Uint64() {
			break
		}
	}

	if chosenUTXOsSum < txCost.Uint64()+minUtxoAmount {
		// TODO:
		// reduce outputs
		// adjust txCost
		// continue
		err = fmt.Errorf("fatal error, not enough resources")
		return nil, "", nil, err
	}

	for {
		var multisigInputs []cardanowallet.TxInput
		var multisigInputsSum uint64 = 0
		for _, utxo := range chosenUTXOs {
			multisigInputs = append(multisigInputs, cardanowallet.TxInput{
				Hash:  utxo.TxHash,
				Index: uint32(utxo.TxIndex.Uint64()),
			})
			multisigInputsSum += utxo.Amount.Uint64()
		}

		inputUtxos.MultisigOwnedUTXOs = chosenUTXOs
		txInfos.MultiSig.TxInputUTXOs = cardano.TxInputUTXOs{Inputs: multisigInputs, InputsSum: multisigInputsSum}

		rawTx, txHash, err := cardano.CreateTx(cco.Config.TestNetMagic, protocolParams, slotNumber+cardano.TTLSlotNumberInc,
			metadata, txInfos, outputs)
		if err != nil {
			return nil, "", nil, err
		}

		// TODO: Get real tx size from protocolParams/config
		if len(rawTx) < 16000 {
			// Check if we used big UTXO replace it with biggest acceptable one
			if len(chosenUTXOs) > chosenUTXOsCount {
				offset := len(inputUtxos.MultisigOwnedUTXOs) - 1
				neededAmount := txCost.Uint64() - multisigInputsSum
				for {
					if neededAmount+inputUtxos.MultisigOwnedUTXOs[offset].Amount.Uint64() > 0 {
						offset--
						continue
					}
					chosenUTXOsCount++
					break
				}

				chosenUTXOs[chosenUTXOsCount] = inputUtxos.MultisigOwnedUTXOs[offset+1]
				continue
			}

			return rawTx, txHash, inputUtxos, err
		}

		// Reduce chosenUTXOs by one and add largest one
		if len(chosenUTXOs) == chosenUTXOsCount {
			chosenUTXOsCount--
		}
		chosenUTXOsCount--
		chosenUTXOs = chosenUTXOs[:chosenUTXOsCount]
		chosenUTXOs = append(chosenUTXOs, inputUtxos.MultisigOwnedUTXOs[len(inputUtxos.MultisigOwnedUTXOs)])
	}
}

func GetInputUtxos(ctx context.Context, bridgeSmartContract eth.IBridgeSmartContract, destinationChain string, txCost *big.Int) (
	*contractbinding.IBridgeContractStructsUTXOs, error) {

	inputUtxos, err := bridgeSmartContract.GetAvailableUTXOs(ctx, destinationChain)
	if err != nil {
		return nil, err
	}

	sort.Slice(inputUtxos.MultisigOwnedUTXOs, func(i, j int) bool {
		return inputUtxos.MultisigOwnedUTXOs[i].Amount.Cmp(inputUtxos.MultisigOwnedUTXOs[j].Amount) < 0
	})
	sort.Slice(inputUtxos.FeePayerOwnedUTXOs, func(i, j int) bool {
		return inputUtxos.FeePayerOwnedUTXOs[i].Amount.Cmp(inputUtxos.FeePayerOwnedUTXOs[j].Amount) < 0
	})

	return inputUtxos, err
}
