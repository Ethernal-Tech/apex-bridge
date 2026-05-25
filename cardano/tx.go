package cardanotx

import (
	"context"
	"fmt"
	"math"

	"github.com/Ethernal-Tech/apex-bridge/common"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

type PlutusTxParams struct {
	CalcFee          uint64
	TotalCollateral  uint64
	CollateralOutput uint64
	CPU              uint64
	Memory           uint64
}

// CreateTx creates tx and returns cbor of raw transaction data, tx hash and error
func CreateTx(
	ctx context.Context,
	cardanoCliBinary string,
	testNetMagic uint,
	protocolParams []byte,
	timeToLive uint64,
	metadataBytes []byte,
	txInputInfos *TxInputInfos,
	refundTxMultisigInputs []*TxInputInfo,
	outputs []cardanowallet.TxOutput,
	certificatesData *CertificatesData,
	addrAndAmountToDeduct []common.AddressAndAmount,
	txPlutusMintData *PlutusMintData,
	additionalWitnessCount int,
) ([]byte, string, error) {
	// ensure there is at least one input for both the multisig and fee multisig.
	// in case that there are no certificates for the tx
	multisigLn := 0
	for _, multisig := range txInputInfos.MultiSig {
		multisigLn += len(multisig.Inputs)
	}

	feeLn := len(txInputInfos.MultiSigFee.Inputs)

	refundLn := 0

	for _, multisig := range refundTxMultisigInputs {
		refundLn += len(multisig.Inputs)
	}

	if certificatesData == nil && ((multisigLn == 0 && refundLn == 0) || feeLn == 0) {
		return nil, "", fmt.Errorf("no inputs found for either multisig (%d) and refund (%d) or fee multisig (%d)",
			multisigLn, refundLn, feeLn)
	}

	builder, err := cardanowallet.NewTxBuilder(cardanoCliBinary)
	if err != nil {
		return nil, "", err
	}

	defer builder.Dispose()

	builder.SetProtocolParameters(protocolParams).SetTimeToLive(timeToLive).
		SetMetaData(metadataBytes).SetTestNetMagic(testNetMagic).AddOutputs(outputs...)

	stakeKeyRegistrationFee := uint64(0)
	stakeKeyDeregistrationGain := uint64(0)

	if certificatesData != nil {
		for _, cert := range certificatesData.Certificates {
			builder.AddCertificates(cert.PolicyScript, cert.Certificates...)
		}

		stakeKeyRegistrationFee = certificatesData.RegistrationFee
		stakeKeyDeregistrationGain = certificatesData.DeregistrationFee
	}

	outputsAmount := cardanowallet.GetOutputsSum(outputs)
	feeOutput, feeIndex := getOutputForAddress(outputs, txInputInfos.MultiSigFee.Address)
	feeOutputPresent := feeIndex >= 0

	// add multisigFee output
	if feeIndex == -1 {
		feeIndex = len(outputs)
		feeOutputPresent = true

		builder.AddOutputs(cardanowallet.TxOutput{
			Addr: txInputInfos.MultiSigFee.Address,
		})
	}

	for _, multisig := range refundTxMultisigInputs {
		builder.AddInputsWithScript(multisig.PolicyScript, multisig.Inputs...)
	}

	builder.AddInputsWithScript(txInputInfos.MultiSigFee.PolicyScript, txInputInfos.MultiSigFee.Inputs...)

	if txInputInfos.Custodial != nil {
		builder.AddInputsWithScript(txInputInfos.Custodial.PolicyScript, txInputInfos.Custodial.Inputs...)
	}

	for _, multisig := range txInputInfos.MultiSig {
		multisigChangeTxOutput := cardanowallet.TxOutput{}

		if addrAndAmountToDeduct != nil {
			multisigOutput, multiSigIndex := getOutputForAddress(outputs, multisig.Address)
			outputsAmountNew := GetOutputsSumForAddress(multisig.Address, addrAndAmountToDeduct)

			multisigChangeTxOutput, err = cardanowallet.CreateTxOutputChange(
				multisigOutput, multisig.Sum, outputsAmountNew)
			if err != nil {
				return nil, "", err
			}

			// add multisig output if change is not zero
			if multisigChangeTxOutput.Amount > 0 || len(multisigChangeTxOutput.Tokens) > 0 {
				if multisigChangeTxOutput.Amount >= common.MinUtxoAmountDefaultDfm {
					if multiSigIndex == -1 {
						builder.AddOutputs(multisigChangeTxOutput)
					} else {
						builder.ReplaceOutput(multiSigIndex, multisigChangeTxOutput)
					}
				}
			} else if multiSigIndex >= 0 {
				// we need to decrement feeIndex if it was after multisig in outputs
				if feeOutputPresent && feeIndex > multiSigIndex {
					feeIndex--
				}

				builder.RemoveOutput(multiSigIndex)
			}

			for tokenName, amount := range multisig.Sum {
				outputsAmount[tokenName] = safeSubstract(outputsAmount[tokenName], amount)
			}
		}

		builder.AddInputsWithScript(multisig.PolicyScript, multisig.Inputs...)
	}

	if txPlutusMintData != nil && len(txPlutusMintData.Tokens) > 0 {
		builder.AddPlutusTokenMints(txPlutusMintData.Tokens, txPlutusMintData.TxInReference, txPlutusMintData.TokensPolicyID)
		builder.AddCollateralInputs(txPlutusMintData.Collateral.Inputs)
		builder.AddCollateralOutput(cardanowallet.NewTxOutput(txPlutusMintData.CollateralAddress, 0))
		builder.SetTotalCollateral(0)
	}

	calcFee, err := builder.SetAdditionalWitnessCount(additionalWitnessCount).CalculateFee(0)
	if err != nil {
		return nil, "", err
	}

	calcFee, err = handlePlutusTx(ctx, builder, txPlutusMintData, calcFee)
	if err != nil {
		return nil, "", err
	}

	// First pass sizes the draft for fee calculation. Use at most what fee inputs can cover;
	// otherwise CreateTxOutputChange fails when rough fee > final fee but inputs match final only.
	if affordable := maxAffordableFeeForFeePayer(
		feeOutput, txInputInfos.MultiSigFee.Sum, stakeKeyRegistrationFee); calcFee > affordable {
		calcFee = affordable
	}

	feeIndex, feeOutputPresent, _, err = applyFeePayerChangeOutput(
		builder, feeIndex, feeOutputPresent, feeOutput, txInputInfos.MultiSigFee.Sum,
		calcFee, stakeKeyRegistrationFee, stakeKeyDeregistrationGain,
	)
	if err != nil {
		return nil, "", err
	}

	finalFee, err := builder.SetAdditionalWitnessCount(additionalWitnessCount).CalculateFee(0)
	if err != nil {
		return nil, "", err
	}

	finalFee, err = handlePlutusTx(ctx, builder, txPlutusMintData, finalFee)
	if err != nil {
		return nil, "", err
	}

	_, _, _, err = applyFeePayerChangeOutput( //nolint:dogsled
		builder, feeIndex, feeOutputPresent, feeOutput, txInputInfos.MultiSigFee.Sum,
		finalFee, stakeKeyRegistrationFee, stakeKeyDeregistrationGain,
	)
	if err != nil {
		return nil, "", err
	}

	return builder.Build()
}

// maxAffordableFeeForFeePayer is the highest fee the fee-payer input can cover besides the stake deposit.
func maxAffordableFeeForFeePayer(
	feeOutput cardanowallet.TxOutput,
	feeInputSum map[string]uint64,
	stakeKeyRegistrationFee uint64,
) uint64 {
	total := feeOutput.Amount + feeInputSum[cardanowallet.AdaTokenName]
	if total <= stakeKeyRegistrationFee {
		return 0
	}

	return total - stakeKeyRegistrationFee
}

// applyFeePayerChangeOutput sets the tx fee and updates the fee-payer change output in the builder.
// feeOutputPresent is false when that output was removed; feeIndex is only valid when feeOutputPresent is true.
func applyFeePayerChangeOutput(
	builder *cardanowallet.TxBuilder,
	feeIndex int,
	feeOutputPresent bool,
	feeOutput cardanowallet.TxOutput,
	feeInputSum map[string]uint64,
	fee uint64,
	stakeKeyRegistrationFee uint64,
	stakeKeyDeregistrationGain uint64,
) (int, bool, uint64, error) {
	builder.SetFee(fee)

	feeChangeTxOutput, err := cardanowallet.CreateTxOutputChange(
		feeOutput, feeInputSum, map[string]uint64{
			cardanowallet.AdaTokenName: fee + stakeKeyRegistrationFee,
		})
	if err != nil {
		return feeIndex, feeOutputPresent, fee, err
	}

	// Include the key deregistration gain if exists
	feeChangeTxOutput.Amount += stakeKeyDeregistrationGain

	// update multisigFee amount if needed (feeAmountFinal > 0) or remove it from output
	if feeChangeTxOutput.Amount > 0 || len(feeChangeTxOutput.Tokens) > 0 {
		if feeOutputPresent {
			builder.ReplaceOutput(feeIndex, feeChangeTxOutput)

			return feeIndex, true, fee, nil
		}

		builder.AddOutputs(feeChangeTxOutput)

		return feeIndex, true, fee, nil
	}

	if feeOutputPresent {
		builder.RemoveOutput(feeIndex)

		return feeIndex, false, fee, nil
	}

	return feeIndex, false, fee, nil
}

func handlePlutusTx(
	ctx context.Context,
	builder *cardanowallet.TxBuilder,
	txPlutusMintData *PlutusMintData,
	initialFee uint64,
) (uint64, error) {
	if txPlutusMintData == nil {
		return initialFee, nil
	}

	if len(txPlutusMintData.Tokens) == 0 {
		return initialFee, nil
	}

	txRaw, _, err := builder.UncheckedBuild()
	if err != nil {
		return 0, err
	}

	result, err := calculatePlutusParams(ctx, txRaw, txPlutusMintData, initialFee)
	if err != nil {
		return 0, err
	}

	builder.SetExecutionUnitParams(result.CPU, result.Memory)
	builder.SetTotalCollateral(result.TotalCollateral)
	builder.UpdateCollateralOutputAmount(-1, result.CollateralOutput)

	return result.CalcFee, nil
}

func calculatePlutusParams(
	ctx context.Context,
	txRaw []byte,
	txPlutusMintData *PlutusMintData,
	initialFee uint64,
) (*PlutusTxParams, error) {
	data, err := txPlutusMintData.TxProvider.EvaluateTx(ctx, txRaw)
	if err != nil {
		return nil, err
	}

	if data.CPU > txPlutusMintData.ExecutionUnitData.MaxTxExecutionUnits.Steps {
		return nil, fmt.Errorf(
			"cpu exceeds max tx execution units: %d > %d",
			data.CPU, txPlutusMintData.ExecutionUnitData.MaxTxExecutionUnits.Steps,
		)
	}

	if data.Memory > txPlutusMintData.ExecutionUnitData.MaxTxExecutionUnits.Memory {
		return nil, fmt.Errorf(
			"memory exceeds max tx execution units: %d > %d",
			data.Memory, txPlutusMintData.ExecutionUnitData.MaxTxExecutionUnits.Memory,
		)
	}

	// Calculate Plutus execution cost
	plutusExecutionCost := uint64(math.Ceil(
		txPlutusMintData.ExecutionUnitData.ExecutionUnitPrices.PriceMemory*float64(data.Memory) +
			txPlutusMintData.ExecutionUnitData.ExecutionUnitPrices.PriceSteps*float64(data.CPU),
	))

	calcFee := initialFee + plutusExecutionCost

	// Calculate total collateral
	totalCollateral := uint64(math.Ceil(
		float64(calcFee) * float64(txPlutusMintData.ExecutionUnitData.CollateralPercentage) / 100,
	))

	if totalCollateral > txPlutusMintData.Collateral.Sum[cardanowallet.AdaTokenName] {
		return nil, fmt.Errorf(
			"total collateral is greater than collateral input amount: %d > %d",
			totalCollateral, txPlutusMintData.Collateral.Sum[cardanowallet.AdaTokenName],
		)
	}

	collateralOutput := txPlutusMintData.Collateral.Sum[cardanowallet.AdaTokenName] - totalCollateral
	if collateralOutput < common.MinUtxoAmountDefaultDfm {
		return nil, fmt.Errorf(
			"collateral output is less than min utxo amount: %d < %d",
			collateralOutput, common.MinUtxoAmountDefaultDfm,
		)
	}

	return &PlutusTxParams{
		CalcFee:          calcFee,
		TotalCollateral:  totalCollateral,
		CollateralOutput: collateralOutput,
		CPU:              data.CPU,
		Memory:           data.Memory,
	}, nil
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
