package successtxprocessors

import (
	"fmt"
	"math/big"
	"strings"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/chain"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/utils"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	cUtils "github.com/Ethernal-Tech/apex-bridge/oracle_common/utils"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
)

var _ core.CardanoTxSuccessProcessor = (*BridgingRequestedProcessorSkylineImpl)(nil)

type BridgingRequestedProcessorSkylineImpl struct {
	logger hclog.Logger

	chainInfos map[string]*chain.CardanoChainInfo
}

func NewSkylineBridgingRequestedProcessor(
	logger hclog.Logger, chainInfos map[string]*chain.CardanoChainInfo,
) *BridgingRequestedProcessorSkylineImpl {
	return &BridgingRequestedProcessorSkylineImpl{
		logger:     logger.Named("bridging_requested_processor_skyline"),
		chainInfos: chainInfos,
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
		return p.addBridgingRequestClaim(claims, tx, metadata, appConfig)
	} else {
		//nolint:godox
		// TODO: Refund
		// p.addRefundRequestClaim(claims, tx, metadata)
		return fmt.Errorf("validation failed for tx: %s, err: %w", tx.Hash, err)
	}
}

func (p *BridgingRequestedProcessorSkylineImpl) addBridgingRequestClaim(
	claims *cCore.BridgeClaims, tx *core.CardanoTx,
	metadata *common.BridgingRequestMetadata, appConfig *cCore.AppConfig,
) error {
	var (
		totalAmountCurrencySrc = uint64(0)
		totalAmountCurrencyDst = uint64(0)
		totalAmountWrappedSrc  = uint64(0)
		totalAmountWrappedDst  = uint64(0)

		cardanoDestConfig = appConfig.CardanoChains[metadata.DestinationChainID]
	)

	receivers := make([]cCore.BridgingRequestReceiver, 0, len(metadata.Transactions))

	for _, receiver := range metadata.Transactions {
		receiverAddr := strings.Join(receiver.Address, "")

		if receiverAddr == cardanoDestConfig.BridgingAddresses.FeeAddress {
			// fee address will be added at the end
			continue
		}

		var (
			receiverAmountCurrencyDst uint64
			receiverAmountWrappedDst  uint64
		)

		if receiver.IsNativeTokenOnSource() {
			// receiverAmount represents the amount of native currency that is bridged to the receiver.
			// receiver.Amount of native tokens on the source will be converted to the same amount of native currency on
			// the destination.
			// totalAmountCurrencySrc stays the same
			totalAmountWrappedSrc += receiver.Amount

			// receiverAmountWrappedDst stays the same
			receiverAmountCurrencyDst = receiver.Amount
		} else {
			totalAmountCurrencySrc += receiver.Amount
			// totalAmountWrappedSrc stays the same

			dstMinUtxo, err := p.calculateMinUtxo(
				cardanoDestConfig, tx.OriginChainID, receiverAddr, receiver.Amount)
			if err != nil {
				return fmt.Errorf("failed to calculate destination minUtxo for chainID: %s. err: %w",
					cardanoDestConfig.ChainID, err)
			}

			receiverAmountCurrencyDst = dstMinUtxo
			receiverAmountWrappedDst = receiver.Amount
		}

		totalAmountCurrencyDst += receiverAmountCurrencyDst
		totalAmountWrappedDst += receiverAmountWrappedDst

		receivers = append(receivers, cCore.BridgingRequestReceiver{
			DestinationAddress: receiverAddr,
			Amount:             new(big.Int).SetUint64(receiverAmountCurrencyDst),
			AmountWrapped:      new(big.Int).SetUint64(receiverAmountWrappedDst),
		})
	}

	feeCurrencyDst := cardanoDestConfig.UtxoMinAmount
	totalAmountCurrencyDst += feeCurrencyDst
	totalAmountCurrencySrc += metadata.BridgingFee + metadata.OperationFee

	receivers = append(receivers, cCore.BridgingRequestReceiver{
		DestinationAddress: cardanoDestConfig.BridgingAddresses.FeeAddress,
		Amount:             new(big.Int).SetUint64(feeCurrencyDst),
		AmountWrapped:      new(big.Int).SetUint64(0),
	})

	claim := cCore.BridgingRequestClaim{
		ObservedTransactionHash:         tx.Hash,
		SourceChainId:                   common.ToNumChainID(tx.OriginChainID),
		DestinationChainId:              common.ToNumChainID(metadata.DestinationChainID),
		Receivers:                       receivers,
		NativeCurrencyAmountDestination: new(big.Int).SetUint64(totalAmountCurrencyDst),
		NativeCurrencyAmountSource:      new(big.Int).SetUint64(totalAmountCurrencySrc),
		WrappedTokenAmountSource:        new(big.Int).SetUint64(totalAmountWrappedSrc),
		WrappedTokenAmountDestination:   new(big.Int).SetUint64(totalAmountWrappedDst),
		RetryCounter:                    big.NewInt(int64(tx.BatchTryCount)),
	}

	claims.BridgingRequestClaims = append(claims.BridgingRequestClaims, claim)

	p.logger.Info("Added BridgingRequestClaim",
		"txHash", tx.Hash, "metadata", metadata, "claim", cCore.BridgingRequestClaimString(claim))

	return nil
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

	if err := utils.ValidateOutputsHaveUnknownTokens(tx, appConfig); err != nil {
		return err
	}

	multisigUtxo, err := utils.ValidateTxOutputs(tx, appConfig, false)
	if err != nil {
		return err
	}

	if metadata.OperationFee < cardanoSrcConfig.MinOperationFee {
		return fmt.Errorf("operation fee in metadata receivers is less than minimum: fee %d, minFee %d, metadata %v",
			metadata.OperationFee, cardanoSrcConfig.MinOperationFee, metadata)
	}

	if len(metadata.Transactions) > appConfig.BridgingSettings.MaxReceiversPerBridgingRequest {
		return fmt.Errorf("number of receivers in metadata greater than maximum allowed - no: %v, max: %v, metadata: %v",
			len(metadata.Transactions), appConfig.BridgingSettings.MaxReceiversPerBridgingRequest, metadata)
	}

	nativeCurrencyAmountSum := new(big.Int).SetUint64(metadata.OperationFee)
	wrappedTokenAmountSum := big.NewInt(0)

	feeSum := uint64(0)
	foundAUtxoValueBelowMinimumValue := false
	foundAnInvalidReceiverAddr := false

	for _, receiver := range metadata.Transactions {
		receiverAddr := strings.Join(receiver.Address, "")

		if !cardanotx.IsValidOutputAddress(receiverAddr, cardanoDestConfig.NetworkID) {
			foundAnInvalidReceiverAddr = true

			break
		}

		if receiverAddr == cardanoDestConfig.BridgingAddresses.FeeAddress {
			if receiver.IsNativeTokenOnSource() {
				return fmt.Errorf("fee receiver metadata invalid: %v", metadata)
			}

			feeSum += receiver.Amount

			continue
		}

		if receiver.IsNativeTokenOnSource() {
			// amount_to_bridge must be >= minUtxoAmount on destination
			if receiver.Amount < cardanoDestConfig.UtxoMinAmount {
				foundAUtxoValueBelowMinimumValue = true

				break
			}

			wrappedTokenAmountSum.Add(wrappedTokenAmountSum, new(big.Int).SetUint64(receiver.Amount))
		} else {
			// amount_to_bridge must be >= minUtxoAmount on source
			if receiver.Amount < cardanoSrcConfig.UtxoMinAmount {
				foundAUtxoValueBelowMinimumValue = true

				break
			}

			nativeCurrencyAmountSum.Add(nativeCurrencyAmountSum, new(big.Int).SetUint64(receiver.Amount))
		}
	}

	if foundAUtxoValueBelowMinimumValue {
		return fmt.Errorf("found a utxo value below minimum value in metadata receivers: %v", metadata)
	}

	if foundAnInvalidReceiverAddr {
		return fmt.Errorf("found an invalid receiver addr in metadata: %v", metadata)
	}

	// update fee amount if needed with sum of fee address receivers
	metadata.BridgingFee += feeSum
	nativeCurrencyAmountSum.Add(nativeCurrencyAmountSum, new(big.Int).SetUint64(metadata.BridgingFee))

	if metadata.BridgingFee < cardanoSrcConfig.MinFeeForBridging {
		return fmt.Errorf("bridging fee in metadata receivers is less than minimum: fee %d, minFee %d, metadata %v",
			metadata.BridgingFee, cardanoSrcConfig.MinFeeForBridging, metadata)
	}

	nativeToken, err := cardanoSrcConfig.GetNativeToken(metadata.DestinationChainID)
	if err != nil {
		return err
	}

	multisigWrappedTokenAmount := cardanotx.GetTokenAmount(multisigUtxo, nativeToken.String())

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

	if nativeCurrencyAmountSum.Cmp(new(big.Int).SetUint64(multisigUtxo.Amount)) != 0 {
		return fmt.Errorf("multisig amount is not equal to sum of receiver amounts + fee: expected %v but got %v",
			multisigUtxo.Amount, nativeCurrencyAmountSum)
	}

	return nil
}

func (p *BridgingRequestedProcessorSkylineImpl) calculateMinUtxo(
	config *cCore.CardanoChainConfig, destinationChainID, receiverAddr string, wrappedAmount uint64,
) (uint64, error) {
	builder, err := cardanowallet.NewTxBuilder(cardanowallet.ResolveCardanoCliBinary(config.NetworkID))
	if err != nil {
		return 0, err
	}

	defer builder.Dispose()

	chainInfo, exists := p.chainInfos[config.ChainID]
	if !exists {
		return 0, fmt.Errorf("chain info for chainID: %s, not found", config.ChainID)
	}

	builder.SetProtocolParameters(chainInfo.ProtocolParams)

	nativeToken, err := config.GetNativeToken(destinationChainID)
	if err != nil {
		return 0, err
	}

	potentialTokenCost, err := cardanowallet.GetMinUtxoForSumMap(
		builder,
		receiverAddr,
		cardanowallet.GetTokensSumMap(cardanowallet.NewTokenAmount(nativeToken, wrappedAmount)),
	)
	if err != nil {
		return 0, err
	}

	return max(config.UtxoMinAmount, potentialTokenCost), nil
}
