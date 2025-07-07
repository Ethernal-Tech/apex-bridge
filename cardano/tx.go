package cardanotx

import (
	"fmt"

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
	certificatesData *CertificatesData,
) ([]byte, string, error) {
	// ensure there is at least one input for both the multisig and fee multisig.
	if ln, feeLn := len(txInputInfos.MultiSig.Inputs), len(txInputInfos.MultiSigFee.Inputs); ln == 0 || feeLn == 0 {
		return nil, "", fmt.Errorf("no inputs found for multisig (%d) or fee multisig (%d)", ln, feeLn)
	}

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

	stakeKeyRegistrationFee := uint64(0)
	if certificatesData != nil {
		for _, cert := range certificatesData.Certificates {
			builder.AddCertificates(cert.PolicyScript, cert.Certificates...)
		}
		stakeKeyRegistrationFee += certificatesData.RegistrationFee
	}

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

	calcFee, err := builder.CalculateFee(0)
	if err != nil {
		return nil, "", err
	}

	builder.SetFee(calcFee)

	feeChangeTxOutput, err := cardanowallet.CreateTxOutputChange(
		feeOutput, txInputInfos.MultiSigFee.Sum, map[string]uint64{
			cardanowallet.AdaTokenName: calcFee,
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
