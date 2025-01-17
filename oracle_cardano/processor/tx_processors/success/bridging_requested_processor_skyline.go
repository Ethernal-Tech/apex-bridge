package successtxprocessors

import (
	"fmt"
	"math/big"
	"strings"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/utils"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	cUtils "github.com/Ethernal-Tech/apex-bridge/oracle_common/utils"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/sendtx"
	"github.com/hashicorp/go-hclog"
)

var _ core.CardanoTxSuccessProcessor = (*BridgingRequestedProcessorSkylineImpl)(nil)

type BridgingRequestedProcessorSkylineImpl struct {
	logger hclog.Logger
}

func NewNativeBridgingRequestedProcessor(logger hclog.Logger) *BridgingRequestedProcessorSkylineImpl {
	return &BridgingRequestedProcessorSkylineImpl{
		logger: logger.Named("bridging_requested_processor_skyline"),
	}
}

func (*BridgingRequestedProcessorSkylineImpl) GetType() common.BridgingTxType {
	return common.BridgingTxTypeBridgingRequest
}

func (*BridgingRequestedProcessorSkylineImpl) PreValidate(tx *core.CardanoTx, appConfig *cCore.AppConfig) error {
	return nil
}

func (p *BridgingRequestedProcessorSkylineImpl) ValidateAndAddClaim(
	claims *cCore.BridgeClaims, tx *core.CardanoTx, appConfig *cCore.AppConfig,
) error {
	metadata, err := common.UnmarshalMetadata[common.BridgingRequestMetadata](common.MetadataEncodingTypeCbor, tx.Metadata)
	if err != nil {
		return fmt.Errorf("failed to unmarshal metadata: tx: %v, err: %w", tx, err)
	}

	if common.BridgingTxType(metadata.BridgingTxType) != p.GetType() {
		return fmt.Errorf("ValidateAndAddClaim called for irrelevant tx: %v", tx)
	}

	p.logger.Debug("Validating relevant tx", "txHash", tx.Hash, "metadata", metadata)

	err = p.validate(tx, metadata, appConfig)
	if err == nil {
		p.addBridgingRequestClaim(claims, tx, metadata, appConfig)
	} else {
		//nolint:godox
		// TODO: Refund
		// p.addRefundRequestClaim(claims, tx, metadata)
		return fmt.Errorf("validation failed for tx: %s, err: %w", tx.Hash, err)
	}

	return nil
}

func (p *BridgingRequestedProcessorSkylineImpl) addBridgingRequestClaim(
	claims *cCore.BridgeClaims, tx *core.CardanoTx,
	metadata *common.BridgingRequestMetadata, appConfig *cCore.AppConfig,
) {
	totalAmountSrc := uint64(0)
	totalAmountDest := uint64(0)
	totalAmountWrappedTokenSrc := uint64(0)
	totalAmountWrappedTokenDest := uint64(0)

	cardanoDestConfig := appConfig.CardanoChains[metadata.DestinationChainID]

	receivers := make([]cCore.BridgingRequestReceiver, 0, len(metadata.Transactions))

	for _, receiver := range metadata.Transactions {
		receiverAddr := strings.Join(receiver.Address, "")

		if receiver.Additional == nil {
			receiver.Additional = &sendtx.BridgingRequestMetadataCurrencyInfo{}
		}

		if receiverAddr == cardanoDestConfig.BridgingAddresses.FeeAddress {
			// fee address will be added at the end
			continue
		}

		var (
			receiverAmount        uint64
			receiverAmountWrapped uint64
		)

		if receiver.IsNativeTokenOnSource() {
			// receiverAmount represents the amount of native currency that is bridged to the receiver.
			// receiver.Amount of native tokens on the source will be converted to the same amount of native currency on
			// the destination.
			// Additionally, receiver.Additional.SrcAmount (native currency) will be transferred to the destination
			// as receiver.Additional.DestAmount (native currency).
			receiverAmount = receiver.Amount + receiver.Additional.DestAmount
			receiverAmountWrapped = uint64(0)

			totalAmountSrc += receiver.Additional.SrcAmount
			totalAmountWrappedTokenSrc += receiver.Amount
		} else {
			receiverAmount = receiver.Additional.DestAmount
			receiverAmountWrapped = receiver.Amount

			totalAmountSrc += receiver.Amount + receiver.Additional.SrcAmount
			totalAmountWrappedTokenSrc += uint64(0)
		}

		totalAmountDest += receiverAmount
		totalAmountWrappedTokenDest += receiverAmountWrapped

		receivers = append(receivers, cCore.BridgingRequestReceiver{
			DestinationAddress: receiverAddr,
			Amount:             new(big.Int).SetUint64(receiverAmount),
			AmountWrapped:      new(big.Int).SetUint64(receiverAmountWrapped),
		})
	}

	totalAmountDest += metadata.FeeAmount.DestAmount
	totalAmountSrc += metadata.FeeAmount.SrcAmount

	receivers = append(receivers, cCore.BridgingRequestReceiver{
		DestinationAddress: cardanoDestConfig.BridgingAddresses.FeeAddress,
		Amount:             new(big.Int).SetUint64(metadata.FeeAmount.DestAmount),
		AmountWrapped:      new(big.Int).SetUint64(0),
	})

	claim := cCore.BridgingRequestClaim{
		ObservedTransactionHash:         tx.Hash,
		SourceChainId:                   common.ToNumChainID(tx.OriginChainID),
		DestinationChainId:              common.ToNumChainID(metadata.DestinationChainID),
		Receivers:                       receivers,
		NativeCurrencyAmountDestination: new(big.Int).SetUint64(totalAmountDest),
		NativeCurrencyAmountSource:      new(big.Int).SetUint64(totalAmountSrc),
		WrappedTokenAmountSource:        new(big.Int).SetUint64(totalAmountWrappedTokenSrc),
		WrappedTokenAmountDestination:   new(big.Int).SetUint64(totalAmountWrappedTokenDest),
		RetryCounter:                    big.NewInt(int64(tx.BatchFailedCount)),
	}

	claims.BridgingRequestClaims = append(claims.BridgingRequestClaims, claim)

	p.logger.Info("Added BridgingRequestClaim",
		"txHash", tx.Hash, "metadata", metadata, "claim", cCore.BridgingRequestClaimString(claim))
}

/*
func (*BridgingRequestedProcessorSkylineImpl) addRefundRequestClaim(
	claims *core.BridgeClaims, tx *core.CardanoTx, metadata *common.BridgingRequestMetadata,
) {

		var outputUtxos []core.Utxo
		for _, output := range tx.Outputs {
			outputUtxos = append(outputUtxos, core.Utxo{
				Address: output.Address,
				Amount:  output.Amount,
			})
		}

		// what goes into UtxoTransaction
		claim := core.RefundRequestClaim{
			TxHash:             tx.Hash,
			RetryCounter:       0,
			RefundToAddress:    metadata.SenderAddr,
			DestinationChainId: metadata.DestinationChainId,
			OutputUtxos:        outputUtxos,
			UtxoTransaction:    core.UtxoTransaction{},
		}

		claims.RefundRequest = append(claims.RefundRequest, claim)

		p.logger.Info("Added RefundRequestClaim",
		"txHash", tx.Hash, "metadata", metadata, "claim", core.RefundRequestClaimString(claim))
}
*/

func (p *BridgingRequestedProcessorSkylineImpl) validate(
	tx *core.CardanoTx, metadata *common.BridgingRequestMetadata, appConfig *cCore.AppConfig,
) error {
	cardanoSrcConfig, _ := cUtils.GetChainConfig(appConfig, tx.OriginChainID)
	if cardanoSrcConfig == nil {
		return fmt.Errorf("unsupported chain id found in tx. chain id: %v", tx.OriginChainID)
	}

	cardanoDestConfig, _ := cUtils.GetChainConfig(appConfig, metadata.DestinationChainID)
	if cardanoDestConfig == nil {
		return fmt.Errorf("destination chain not registered: %v", metadata.DestinationChainID)
	}

	multisigUtxo, err := utils.ValidateTxOutputs(tx, appConfig, false)
	if err != nil {
		return err
	}

	if len(metadata.Transactions) > appConfig.BridgingSettings.MaxReceiversPerBridgingRequest {
		return fmt.Errorf("number of receivers in metadata greater than maximum allowed - no: %v, max: %v, metadata: %v",
			len(metadata.Transactions), appConfig.BridgingSettings.MaxReceiversPerBridgingRequest, metadata)
	}

	nativeCurrencyAmountSum := big.NewInt(0)
	wrappedTokenAmountSum := big.NewInt(0)

	feeSum := uint64(0)
	foundAUtxoValueBelowMinimumValue := false
	foundAnInvalidReceiverAddr := false

	exchangeMargin := float64(0.000001)

	exchangeRate, err := GetExchangeRate(tx.OriginChainID, metadata.DestinationChainID)
	if err != nil {
		return err
	}

	for _, receiver := range metadata.Transactions {
		receiverAddr := strings.Join(receiver.Address, "")

		if !cardanotx.IsValidOutputAddress(receiverAddr, cardanoDestConfig.NetworkID) {
			foundAnInvalidReceiverAddr = true

			break
		}

		var (
			totalNativeCurrencyAmountToBridge uint64
			totalWrappedTokenAmountToBridge   uint64
		)

		if receiverAddr == cardanoDestConfig.BridgingAddresses.FeeAddress {
			if receiver.IsNativeTokenOnSource() || receiver.Additional != nil {
				return fmt.Errorf("fee receiver metadata invalid: %v", metadata)
			}

			feeSum += receiver.Amount

			continue
		}

		if receiver.Additional == nil {
			receiver.Additional = &sendtx.BridgingRequestMetadataCurrencyInfo{}
		}

		if receiver.IsNativeTokenOnSource() {
			// amount_to_bridge must be >= minUtxoAmount on destination
			if receiver.Amount < cardanoDestConfig.UtxoMinAmount {
				foundAUtxoValueBelowMinimumValue = true

				break
			}

			totalNativeCurrencyAmountToBridge = receiver.Additional.SrcAmount
			totalWrappedTokenAmountToBridge = receiver.Amount
		} else {
			// amount_to_bridge must be >= minUtxoAmount on source
			if receiver.Amount < cardanoSrcConfig.UtxoMinAmount {
				foundAUtxoValueBelowMinimumValue = true

				break
			}

			if receiver.Additional.DestAmount < cardanoDestConfig.UtxoMinAmount {
				foundAUtxoValueBelowMinimumValue = true

				break
			}

			totalNativeCurrencyAmountToBridge = receiver.Amount + receiver.Additional.SrcAmount
		}

		if !common.IsEqual(float64(receiver.Additional.DestAmount)*exchangeRate,
			float64(receiver.Additional.SrcAmount), exchangeMargin) {
			return fmt.Errorf("found an exchange rate error in metadata: %v", metadata)
		}

		nativeCurrencyAmountSum.Add(nativeCurrencyAmountSum, new(big.Int).SetUint64(totalNativeCurrencyAmountToBridge))
		wrappedTokenAmountSum.Add(wrappedTokenAmountSum, new(big.Int).SetUint64(totalWrappedTokenAmountToBridge))
	}

	if foundAUtxoValueBelowMinimumValue {
		return fmt.Errorf("found a utxo value below minimum value in metadata receivers: %v", metadata)
	}

	if foundAnInvalidReceiverAddr {
		return fmt.Errorf("found an invalid receiver addr in metadata: %v", metadata)
	}

	if !common.IsEqual(float64(metadata.FeeAmount.DestAmount)*exchangeRate,
		float64(metadata.FeeAmount.SrcAmount), exchangeMargin) {
		return fmt.Errorf("found an exchage rate error if fee metadata: %v", metadata)
	}

	// update fee amount if needed with sum of fee address receivers
	if feeSum > 0 {
		metadata.FeeAmount.SrcAmount += feeSum
		metadata.FeeAmount.DestAmount = uint64(float64(metadata.FeeAmount.SrcAmount) / exchangeRate)
	}

	nativeCurrencyAmountSum.Add(nativeCurrencyAmountSum, new(big.Int).SetUint64(metadata.FeeAmount.SrcAmount))

	if metadata.FeeAmount.DestAmount < cardanoDestConfig.MinFeeForBridging {
		return fmt.Errorf("bridging fee in metadata receivers is less than minimum: fee %d, minFee %d, metadata %v",
			metadata.FeeAmount.DestAmount, cardanoDestConfig.MinFeeForBridging, metadata)
	}

	if nativeCurrencyAmountSum.Cmp(new(big.Int).SetUint64(multisigUtxo.Amount)) != 0 {
		return fmt.Errorf("multisig amount is not equal to sum of receiver amounts + fee: expected %v but got %v",
			multisigUtxo.Amount, nativeCurrencyAmountSum)
	}

	nativeToken, err := cardanoSrcConfig.GetNativeToken(metadata.DestinationChainID, true)
	if err != nil {
		return err
	}

	multisigWrappedTokenAmount := getNativeTokenAmount(multisigUtxo, nativeToken.String())

	if wrappedTokenAmountSum.Cmp(new(big.Int).SetUint64(multisigWrappedTokenAmount)) != 0 {
		return fmt.Errorf("multisig wrapped token is not equal to receiver wrapped token amount: expected %v but got %v",
			multisigWrappedTokenAmount, wrappedTokenAmountSum)
	}

	if appConfig.BridgingSettings.MaxAmountAllowedToBridge != nil &&
		appConfig.BridgingSettings.MaxAmountAllowedToBridge.Sign() > 0 &&
		nativeCurrencyAmountSum.Cmp(appConfig.BridgingSettings.MaxAmountAllowedToBridge) == 1 {
		return fmt.Errorf("sum of receiver amounts + fee: %v greater than maximum allowed: %v",
			nativeCurrencyAmountSum, appConfig.BridgingSettings.MaxAmountAllowedToBridge)
	}

	return nil
}

func GetExchangeRate(sourceChainID string, destinationChainID string) (float64, error) {
	if destinationChainID == "prime" {
		return 0.5, nil
	}

	return 2, nil
}

func getNativeTokenAmount(utxo *indexer.TxOutput, tokenName string) (amount uint64) {
	for _, token := range utxo.Tokens {
		if token.TokenName() == tokenName {
			amount += token.Amount
		}
	}

	return amount
}
