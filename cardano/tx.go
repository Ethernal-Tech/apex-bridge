package cardanotx

import (
	"errors"
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
) ([]byte, string, error) {
	// ensure there is at least one input for both the multisig and fee multisig.
	if ln, feeLn := len(txInputInfos.MultiSig.Inputs), len(txInputInfos.MultiSigFee.Inputs); ln == 0 || feeLn == 0 {
		return nil, "", fmt.Errorf("no inputs found for multisig (%d) or fee multisig (%d)", ln, feeLn)
	}

	outputsAmount := cardanowallet.GetOutputsSum(outputs)
	lovelaceOutputsAmount := outputsAmount[cardanowallet.AdaTokenName]
	multiSigIndex, multisigAmount := isAddressInOutputs(outputs, txInputInfos.MultiSig.Address)
	feeIndex, feeAmount := isAddressInOutputs(outputs, txInputInfos.MultiSigFee.Address)
	multisigAmount += txInputInfos.MultiSig.Sum[cardanowallet.AdaTokenName]
	feeAmount += txInputInfos.MultiSigFee.Sum[cardanowallet.AdaTokenName]

	if multisigAmount < lovelaceOutputsAmount {
		return nil, "", fmt.Errorf("not enough funds on multisig: %d vs %vs", multisigAmount, lovelaceOutputsAmount)
	}

	changeAmount := multisigAmount - lovelaceOutputsAmount

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

	calcFee, err := builder.CalculateFee(0)
	if err != nil {
		return nil, "", err
	}

	builder.SetFee(calcFee)

	if feeAmount < calcFee {
		return nil, "", fmt.Errorf("not enough funds on fee multisig: %d vs %vs", feeAmount, calcFee)
	}

	feeAmountFinal := feeAmount - calcFee

	// update multisigFee amount if needed (feeAmountFinal > 0) or remove it from output
	if feeAmountFinal > 0 {
		builder.UpdateOutputAmount(feeIndex, feeAmountFinal)
	} else {
		builder.RemoveOutput(feeIndex)
	}

	return builder.Build()
}

func isAddressInOutputs(outputs []cardanowallet.TxOutput, addr string) (int, uint64) {
	for i, x := range outputs {
		if x.Addr == addr {
			return i, x.Amount
		}
	}

	return -1, 0
}

var ErrNotEnoughFee = errors.New("not enough fee for tx")

func CreateOnlyFeeTx(
	cardanoCliBinary string,
	testNetMagic uint,
	protocolParams []byte,
	timeToLive uint64,
	metadataBytes []byte,
	feeTxInput *TxInputInfo,
	output cardanowallet.TxOutput,
) ([]byte, string, error) {
	// ensure there is at least one input for both the multisig and fee multisig.
	if feeTxInput == nil || len(feeTxInput.Inputs) == 0 {
		return nil, "", ErrNotEnoughFee
	}

	feeAmount := feeTxInput.Sum[cardanowallet.AdaTokenName]

	builder, err := cardanowallet.NewTxBuilder(cardanoCliBinary)
	if err != nil {
		return nil, "", err
	}

	defer builder.Dispose()

	builder.SetProtocolParameters(protocolParams).SetTimeToLive(timeToLive).
		SetMetaData(metadataBytes).SetTestNetMagic(testNetMagic).AddOutputs(output)

	builder.AddInputsWithScript(feeTxInput.PolicyScript, feeTxInput.Inputs...)

	calcFee, err := builder.CalculateFee(0)
	if err != nil {
		return nil, "", err
	}

	builder.SetFee(calcFee)

	feeAmountFinal := feeAmount - calcFee

	builder.UpdateOutputAmount(0, feeAmountFinal)

	return builder.Build()
}
