package successtxprocessors

import (
	"fmt"
	"math/big"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	oUtils "github.com/Ethernal-Tech/apex-bridge/oracle_common/utils"
	"github.com/Ethernal-Tech/apex-bridge/oracle_eth/core"
	"github.com/hashicorp/go-hclog"
)

var _ core.EthTxSuccessProcessor = (*BridgingRequestedProcessorImpl)(nil)

type BridgingRequestedProcessorImpl struct {
	refundRequestProcessor core.EthTxSuccessProcessor
	logger                 hclog.Logger
}

func NewEthBridgingRequestedProcessor(
	refundRequestProcessor core.EthTxSuccessProcessor, logger hclog.Logger,
) *BridgingRequestedProcessorImpl {
	return &BridgingRequestedProcessorImpl{
		refundRequestProcessor: refundRequestProcessor,
		logger:                 logger.Named("eth_bridging_requested_processor"),
	}
}

func (*BridgingRequestedProcessorImpl) GetType() common.BridgingTxType {
	return common.BridgingTxTypeBridgingRequest
}

func (*BridgingRequestedProcessorImpl) PreValidate(tx *core.EthTx, appConfig *oCore.AppConfig) error {
	return nil
}

func (p *BridgingRequestedProcessorImpl) ValidateAndAddClaim(
	claims *oCore.BridgeClaims, tx *core.EthTx, appConfig *oCore.AppConfig,
) error {
	metadata, err := core.UnmarshalEthMetadata[core.BridgingRequestEthMetadata](
		tx.Metadata)
	if err != nil {
		p.logger.Warn("failed to unmarshal metadata. handing over to refund processor",
			"tx", tx, "err", err)

		return p.refundRequestProcessor.ValidateAndAddClaim(claims, tx, appConfig)
	}

	if metadata.BridgingTxType != p.GetType() {
		p.logger.Warn("ValidateAndAddClaim called for irrelevant tx. handing over to refund processor",
			"tx", tx, "err", err)

		return p.refundRequestProcessor.ValidateAndAddClaim(claims, tx, appConfig)
	}

	p.logger.Debug("Validating relevant tx", "txHash", tx.Hash, "metadata", metadata)

	err = p.validate(tx, metadata, appConfig)
	if err == nil {
		p.addBridgingRequestClaim(claims, tx, metadata, appConfig)
	} else {
		p.logger.Warn("validation failed for tx. handing over to refund processor",
			"tx", tx, "err", err)

		return p.refundRequestProcessor.ValidateAndAddClaim(claims, tx, appConfig)
	}

	return nil
}

func (p *BridgingRequestedProcessorImpl) addBridgingRequestClaim(
	claims *oCore.BridgeClaims, tx *core.EthTx,
	metadata *core.BridgingRequestEthMetadata, appConfig *oCore.AppConfig,
) {
	totalAmount := big.NewInt(0)

	cardanoDestConfig, _ := oUtils.GetChainConfig(appConfig, metadata.DestinationChainID)

	receivers := make([]oCore.BridgingRequestReceiver, 0, len(metadata.Transactions))

	for _, receiver := range metadata.Transactions {
		if receiver.Address == cardanoDestConfig.BridgingAddresses.FeeAddress {
			// fee address will be added at the end
			continue
		}

		receiverAmountDfm := common.WeiToDfm(receiver.Amount)

		receivers = append(receivers, oCore.BridgingRequestReceiver{
			DestinationAddress: receiver.Address,
			Amount:             receiverAmountDfm,
		})

		totalAmount.Add(totalAmount, receiverAmountDfm)
	}

	feeAmountDfm := common.WeiToDfm(metadata.FeeAmount)
	totalAmount.Add(totalAmount, feeAmountDfm)

	receivers = append(receivers, oCore.BridgingRequestReceiver{
		DestinationAddress: cardanoDestConfig.BridgingAddresses.FeeAddress,
		Amount:             feeAmountDfm,
	})

	claim := oCore.BridgingRequestClaim{
		ObservedTransactionHash: tx.Hash,
		SourceChainId:           common.ToNumChainID(tx.OriginChainID),
		DestinationChainId:      common.ToNumChainID(metadata.DestinationChainID),
		Receivers:               receivers,
		TotalAmount:             totalAmount,
		RetryCounter:            big.NewInt(int64(tx.BatchTryCount)),
	}

	claims.BridgingRequestClaims = append(claims.BridgingRequestClaims, claim)

	p.logger.Info("Added BridgingRequestClaim",
		"txHash", tx.Hash, "metadata", metadata, "claim", oCore.BridgingRequestClaimString(claim))
}

func (p *BridgingRequestedProcessorImpl) validate(
	tx *core.EthTx, metadata *core.BridgingRequestEthMetadata, appConfig *oCore.AppConfig,
) error {
	if tx.BatchTryCount > appConfig.TryCountLimits.MaxBatchTryCount ||
		tx.SubmitTryCount > appConfig.TryCountLimits.MaxSubmitTryCount {
		return fmt.Errorf(
			"try count exceeded. BatchTryCount: (current, max)=(%d, %d), SubmitTryCount: (current, max)=(%d, %d)",
			tx.BatchTryCount, appConfig.TryCountLimits.MaxBatchTryCount,
			tx.SubmitTryCount, appConfig.TryCountLimits.MaxSubmitTryCount)
	}

	if err := common.IsTxDirectionAllowed(tx.OriginChainID, metadata.DestinationChainID); err != nil {
		return err
	}

	_, ethSrcConfig := oUtils.GetChainConfig(appConfig, tx.OriginChainID)
	if ethSrcConfig == nil {
		return fmt.Errorf("origin chain not registered: %v", tx.OriginChainID)
	}

	cardanoDestConfig, _ := oUtils.GetChainConfig(appConfig, metadata.DestinationChainID)
	if cardanoDestConfig == nil {
		return fmt.Errorf("destination chain not registered: %v", metadata.DestinationChainID)
	}

	if len(metadata.Transactions) > appConfig.BridgingSettings.MaxReceiversPerBridgingRequest {
		return fmt.Errorf("number of receivers in metadata greater than maximum allowed - no: %v, max: %v, metadata: %v",
			len(metadata.Transactions), appConfig.BridgingSettings.MaxReceiversPerBridgingRequest, metadata)
	}

	receiverAmountSum := big.NewInt(0)
	feeSum := big.NewInt(0)
	foundAUtxoValueBelowMinimumValue := false
	foundAnInvalidReceiverAddr := false

	for _, receiver := range metadata.Transactions {
		receiverAmountDfm := common.WeiToDfm(receiver.Amount)
		if receiverAmountDfm.Uint64() < cardanoDestConfig.UtxoMinAmount {
			foundAUtxoValueBelowMinimumValue = true

			break
		}

		if !cardanotx.IsValidOutputAddress(receiver.Address, cardanoDestConfig.NetworkID) {
			foundAnInvalidReceiverAddr = true

			break
		}

		if receiver.Address == cardanoDestConfig.BridgingAddresses.FeeAddress {
			feeSum.Add(feeSum, receiver.Amount)
		} else {
			receiverAmountSum.Add(receiverAmountSum, receiver.Amount)
		}
	}

	if foundAUtxoValueBelowMinimumValue {
		return fmt.Errorf("found a utxo value below minimum value in metadata receivers: %v", metadata)
	}

	if foundAnInvalidReceiverAddr {
		return fmt.Errorf("found an invalid receiver addr in metadata: %v", metadata)
	}

	// update fee amount if needed with sum of fee address receivers
	metadata.FeeAmount.Add(metadata.FeeAmount, feeSum)
	receiverAmountSum.Add(receiverAmountSum, metadata.FeeAmount)

	feeAmountDfm := common.WeiToDfm(metadata.FeeAmount)
	if feeAmountDfm.Uint64() < cardanoDestConfig.MinFeeForBridging {
		return fmt.Errorf("bridging fee in metadata receivers is less than minimum: %v", metadata)
	}

	if tx.Value == nil || tx.Value.Cmp(receiverAmountSum) != 0 {
		return fmt.Errorf("tx value is not equal to sum of receiver amounts + fee: expected %v but got %v",
			receiverAmountSum, tx.Value)
	}

	if appConfig.BridgingSettings.MaxAmountAllowedToBridge != nil &&
		appConfig.BridgingSettings.MaxAmountAllowedToBridge.Sign() > 0 &&
		receiverAmountSum.Cmp(common.DfmToWei(appConfig.BridgingSettings.MaxAmountAllowedToBridge)) == 1 {
		return fmt.Errorf("sum of receiver amounts + fee: %v greater than maximum allowed: %v",
			receiverAmountSum, common.DfmToWei(appConfig.BridgingSettings.MaxAmountAllowedToBridge))
	}

	return nil
}
