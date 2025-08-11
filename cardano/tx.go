package cardanotx

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
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
	addrAndAmountToDeduct []common.AddressAndAmount,
	logger hclog.Logger,
) ([]byte, string, error) {
	// ensure there is at least one input for both the multisig and fee multisig.
	// in case that there are no certificates for the tx
	multisigLn := 0
	for _, multisig := range txInputInfos.MultiSig {
		multisigLn += len(multisig.Inputs)
	}

	feeLn := len(txInputInfos.MultiSigFee.Inputs)
	if certificatesData == nil && (multisigLn == 0 || feeLn == 0) {
		return nil, "", fmt.Errorf("no inputs found for multisig (%d) or fee multisig (%d)", multisigLn, feeLn)
	}

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

		stakeKeyRegistrationFee = certificatesData.RegistrationFee
	}

	outputsAmount := cardanowallet.GetOutputsSum(outputs)
	feeOutput, feeIndex := getOutputForAddress(outputs, txInputInfos.MultiSigFee.Address)

	// add multisigFee output
	if feeIndex == -1 {
		feeIndex = len(outputs)

		builder.AddOutputs(cardanowallet.TxOutput{
			Addr: txInputInfos.MultiSigFee.Address,
		})
	}

	builder.AddInputsWithScript(txInputInfos.MultiSigFee.PolicyScript, txInputInfos.MultiSigFee.Inputs...)

	carryOverChange := uint64(0)
	for _, multisig := range txInputInfos.MultiSig {
		multisigOutput, multiSigIndex := getOutputForAddress(outputs, multisig.Address)

		logger.Debug("CREATE TX addr", multisig.Address)
		logger.Debug("CREATE TX baseTxOutput amount", multisigOutput.Amount)
		logger.Debug("CREATE TX totalSum amount", multisig.Sum[cardanowallet.AdaTokenName])
		logger.Debug("CREATE TX outputsSum amount", outputsAmount[cardanowallet.AdaTokenName])

		if addrAndAmountToDeduct != nil {
			outputsAmountNew := GetOutputsSumForAddress(multisig.Address, addrAndAmountToDeduct)
			logger.Debug("CREATE TX addrAndAmountToDeduct", addrAndAmountToDeduct)
			logger.Debug("CREATE TX outputsSumNew amount", outputsAmountNew[cardanowallet.AdaTokenName])
		} else {
			logger.Debug("CREATE TX addrAndAmountToDeduct is nil")
		}

		multisigChangeTxOutput, err := cardanowallet.CreateTxOutputChange(
			multisigOutput, multisig.Sum, outputsAmount)
		if err != nil {
			return nil, "", err
		}

		multisigChangeTxOutput.Amount += carryOverChange

		// add multisig output if change is not zero
		if multisigChangeTxOutput.Amount > 0 || len(multisigChangeTxOutput.Tokens) > 0 {
			if multisigChangeTxOutput.Amount >= common.MinUtxoAmountDefault {
				if multiSigIndex == -1 {
					builder.AddOutputs(multisigChangeTxOutput)
				} else {
					builder.ReplaceOutput(multiSigIndex, multisigChangeTxOutput)
				}
			} else {
				carryOverChange += multisigChangeTxOutput.Amount
			}
		} else if multiSigIndex >= 0 {
			// we need to decrement feeIndex if it was after multisig in outputs
			if feeIndex > multiSigIndex {
				feeIndex--
			}

			builder.RemoveOutput(multiSigIndex)
		}

		for tokenName, amount := range multisig.Sum {
			outputsAmount[tokenName] = safeSubstract(outputsAmount[tokenName], amount)
		}

		builder.AddInputsWithScript(multisig.PolicyScript, multisig.Inputs...)
	}

	calcFee, err := builder.CalculateFee(0)
	if err != nil {
		return nil, "", err
	}

	builder.SetFee(calcFee)

	feeChangeTxOutput, err := cardanowallet.CreateTxOutputChange(
		feeOutput, txInputInfos.MultiSigFee.Sum, map[string]uint64{
			cardanowallet.AdaTokenName: calcFee + stakeKeyRegistrationFee,
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

func GetOutputsSumForAddress(addr string, addrAndAmountToDeduct []common.AddressAndAmount) map[string]uint64 {
	result := map[string]uint64{}

	for _, addrAndAmount := range addrAndAmountToDeduct {
		if addrAndAmount.Address == addr {
			for name, token := range addrAndAmount.TokensAmounts {
				result[name] += token
			}
		}
	}

	return result
}

func safeSubstract(a, b uint64) uint64 {
	if a >= b {
		return a - b
	}

	return 0
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
