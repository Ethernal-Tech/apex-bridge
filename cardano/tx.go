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
	changeAmount := common.SafeSubtract(
		txInputInfos.MultiSig.Sum[cardanowallet.AdaTokenName]+multisigAmount, outputsAmount[cardanowallet.AdaTokenName], 0)

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

	tokenChange, tokenChangeVals, err := calculateTokenChange(txInputInfos.MultiSig.Sum, outputsAmount)
	if err != nil {
		return nil, "", err
	}

	// add multisig output if change is not zero
	if changeAmount > 0 {
		if multiSigIndex == -1 {
			builder.AddOutputs(cardanowallet.TxOutput{
				Addr:   txInputInfos.MultiSig.Address,
				Amount: changeAmount,
				Tokens: tokenChange,
			})
		} else {
			builder.UpdateOutputAmount(multiSigIndex, changeAmount, tokenChangeVals...)
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

	feeAmountFinal := common.SafeSubtract(txInputInfos.MultiSigFee.Sum[cardanowallet.AdaTokenName]+feeAmount, fee, 0)

	_, feeTokenChangeVals, err := calculateTokenChange(txInputInfos.MultiSigFee.Sum, outputsAmount)
	if err != nil {
		return nil, "", err
	}

	// update multisigFee amount if needed (feeAmountFinal > 0) or remove it from output
	if feeAmountFinal > 0 {
		builder.UpdateOutputAmount(feeIndex, feeAmountFinal, feeTokenChangeVals...)
	} else {
		builder.RemoveOutput(feeIndex)
	}

	return builder.Build()
}

func calculateTokenChange(tokenSum map[string]uint64, outputsAmount map[string]uint64,
) ([]cardanowallet.TokenAmount, []uint64, error) {
	tokenChange := []cardanowallet.TokenAmount(nil)
	tokenChangeVals := []uint64(nil)

	for token, amount := range tokenSum {
		if token == cardanowallet.AdaTokenName {
			continue
		}

		newToken, err := cardanowallet.NewTokenWithFullName(token, true)
		if err != nil {
			return nil, nil, err
		}

		tokenChangeAmount := amount - outputsAmount[token]
		if tokenChangeAmount > 0 {
			tokenChange = append(tokenChange, cardanowallet.NewTokenAmount(newToken, tokenChangeAmount))
			tokenChangeVals = append(tokenChangeVals, tokenChangeAmount)
		}
	}

	return tokenChange, tokenChangeVals, nil
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
