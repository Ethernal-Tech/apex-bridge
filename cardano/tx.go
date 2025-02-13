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
	multisigOutput, multiSigIndex := getOutputForAddress(outputs, txInputInfos.MultiSig.Address)
	feeOutput, feeIndex := getOutputForAddress(outputs, txInputInfos.MultiSigFee.Address)

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

	multisigChangeTxOutput, err := cardanowallet.CreateTxOutputChange(
		multisigOutput, txInputInfos.MultiSig.Sum, outputsAmount)
	if err != nil {
		return nil, "", err
	}

	// add multisig output if change is not zero
	if multisigChangeTxOutput.Amount > 0 || len(multisigChangeTxOutput.Tokens) > 0 {
		if multiSigIndex == -1 {
			builder.AddOutputs(multisigChangeTxOutput)
		} else {
			builder.ReplaceOutput(multiSigIndex, multisigChangeTxOutput)
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

	feeChangeTxOutput, err := cardanowallet.CreateTxOutputChange(
		feeOutput, txInputInfos.MultiSigFee.Sum, map[string]uint64{
			cardanowallet.AdaTokenName: fee,
		})
	if err != nil {
		return nil, "", err
	}

	// update multisigFee amount if needed (feeAmountFinal > 0) or remove it from output
	if feeChangeTxOutput.Amount > 0 || len(feeChangeTxOutput.Tokens) > 0 {
		builder.ReplaceOutput(feeIndex, feeChangeTxOutput)
	} else {
		builder.RemoveOutput(feeIndex)
	}

	return builder.Build()
}

func CreateBatchMetaData(v uint64) ([]byte, error) {
	return common.MarshalMetadata(common.MetadataEncodingTypeJSON, common.BatchExecutedMetadata{
		BridgingTxType: common.BridgingTxTypeBatchExecution,
		BatchNonceID:   v,
	})
}

func getOutputForAddress(outputs []cardanowallet.TxOutput, addr string) (cardanowallet.TxOutput, int) {
	for i, x := range outputs {
		if x.Addr == addr {
			return x, i
		}
	}

	return cardanowallet.TxOutput{
		Addr: addr,
	}, -1
}

func GetAddressFromPolicyScript(
	cardanoCliBinary string, testNetMagic uint, ps *cardanowallet.PolicyScript,
) (string, error) {
	return cardanowallet.NewCliUtils(cardanoCliBinary).GetPolicyScriptAddress(testNetMagic, ps)
}
