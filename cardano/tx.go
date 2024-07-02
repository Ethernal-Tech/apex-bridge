package cardanotx

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

// CreateTx creates tx and returns cbor of raw transaction data, tx hash and error
func CreateTx(
	cardanoCliBinary string,
	testNetMagic uint,
	protocolParams []byte,
	timeToLive uint64,
	metadataBytes []byte,
	txInputInfos TxInputInfos,
	outputs []cardanowallet.TxOutput,
) ([]byte, string, error) {
	outputsAmount := cardanowallet.GetOutputsSum(outputs)
	multiSigIndex, multisigAmount := isAddressInOutputs(outputs, txInputInfos.MultiSig.Address)
	feeIndex, feeAmount := isAddressInOutputs(outputs, txInputInfos.MultiSigFee.Address)
	changeAmount := common.SafeSubtract(txInputInfos.MultiSig.Sum+multisigAmount, outputsAmount, 0)

	builder, err := cardanowallet.NewTxBuilder(cardanoCliBinary)
	if err != nil {
		return nil, "", err
	}

	defer builder.Dispose()

	builder.SetProtocolParameters(protocolParams).SetTimeToLive(timeToLive).
		SetMetaData(metadataBytes).SetTestNetMagic(testNetMagic).AddOutputs(outputs...)

	// add multisigFee output
	if feeIndex == -1 {
		feeIndex = len(outputs)

		builder.AddOutputs(cardanowallet.TxOutput{
			Addr: txInputInfos.MultiSigFee.Address,
		})
	}

	// add multisig output if change is not zero
	if changeAmount > 0 {
		if multiSigIndex == -1 {
			builder.AddOutputs(cardanowallet.TxOutput{
				Addr:   txInputInfos.MultiSig.Address,
				Amount: changeAmount,
			})
		} else {
			builder.UpdateOutputAmount(multiSigIndex, changeAmount)
		}
	} else if multiSigIndex >= 0 {
		// we need to decrement feeIndex if it was after multisig in outputs
		if feeIndex > multiSigIndex {
			feeIndex--
		}

		builder.RemoveOutput(multiSigIndex)
	}

	builder.AddInputsWithScript(txInputInfos.MultiSig.PolicyScript, txInputInfos.MultiSig.Inputs...).
		AddInputsWithScript(txInputInfos.MultiSigFee.PolicyScript, txInputInfos.MultiSigFee.Inputs...)

	fee, err := builder.CalculateFee(0)
	if err != nil {
		return nil, "", err
	}

	builder.SetFee(fee)

	feeAmountFinal := common.SafeSubtract(txInputInfos.MultiSigFee.Sum+feeAmount, fee, 0)

	// update multisigFee amount if needed (feeAmountFinal > 0) or remove it from output
	if feeAmountFinal > 0 {
		builder.UpdateOutputAmount(feeIndex, feeAmountFinal)
	} else {
		builder.RemoveOutput(feeIndex)
	}

	return builder.Build()
}

// CreateTxWitness creates cbor of vkey+signature pair of tx hash
func CreateTxWitness(txHash string, key cardanowallet.ISigner) ([]byte, error) {
	return cardanowallet.CreateTxWitness(txHash, key)
}

// AssembleTxWitnesses assembles all witnesses in final cbor of signed tx
func AssembleTxWitnesses(cardanoCliBinary string, txRaw []byte, witnesses [][]byte) ([]byte, error) {
	builder, err := cardanowallet.NewTxBuilder(cardanoCliBinary)
	if err != nil {
		return nil, err
	}

	defer builder.Dispose()

	return builder.AssembleTxWitnesses(txRaw, witnesses)
}

func CreateBatchMetaData(v uint64) ([]byte, error) {
	return common.MarshalMetadata(common.MetadataEncodingTypeJSON, common.BatchExecutedMetadata{
		BridgingTxType: common.BridgingTxTypeBatchExecution,
		BatchNonceID:   v,
	})
}

func isAddressInOutputs(outputs []cardanowallet.TxOutput, addr string) (int, uint64) {
	for i, x := range outputs {
		if x.Addr == addr {
			return i, x.Amount
		}
	}

	return -1, 0
}

func GetAddressFromPolicyScript(
	cardanoCliBinary string, testNetMagic uint, ps *cardanowallet.PolicyScript,
) (string, error) {
	return cardanowallet.NewCliUtils(cardanoCliBinary).GetPolicyScriptAddress(testNetMagic, ps)
}
