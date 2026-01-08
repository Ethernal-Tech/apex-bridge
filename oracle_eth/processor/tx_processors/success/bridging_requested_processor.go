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
	refundRequestProcessor core.EthTxSuccessRefundProcessor
	logger                 hclog.Logger
}

func NewEthBridgingRequestedProcessor(
	refundRequestProcessor core.EthTxSuccessRefundProcessor, logger hclog.Logger,
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
		return p.refundRequestProcessor.HandleBridgingProcessorError(
			claims, tx, appConfig, err, "failed to unmarshal metadata")
	}

	if metadata.BridgingTxType != p.GetType() {
		return p.refundRequestProcessor.HandleBridgingProcessorError(
			claims, tx, appConfig, nil, "ValidateAndAddClaim called for irrelevant tx")
	}

	p.logger.Debug("Validating relevant tx", "txHash", tx.Hash, "metadata", metadata)

	err = p.validate(tx, metadata, appConfig)
	if err == nil {
		p.addBridgingRequestClaim(claims, tx, metadata, appConfig)
	} else {
		return p.refundRequestProcessor.HandleBridgingProcessorError(
			claims, tx, appConfig, err, "validation failed for tx")
	}

	return nil
}

func (p *BridgingRequestedProcessorImpl) addBridgingRequestClaim(
	claims *oCore.BridgeClaims, tx *core.EthTx,
	metadata *core.BridgingRequestEthMetadata, appConfig *oCore.AppConfig,
) {
	totalAmount := big.NewInt(0)

	cardanoDestConfig, _ := oUtils.GetChainConfig(appConfig, metadata.DestinationChainID)
	cardanoDestChainFeeAddress := appConfig.GetFeeMultisigAddress(metadata.DestinationChainID)
	chainIDConverter := appConfig.ChainIDConverter

	receivers := make([]oCore.BridgingRequestReceiver, 0, len(metadata.Transactions))

	for _, receiver := range metadata.Transactions {
		if receiver.Address == cardanoDestChainFeeAddress {
			// fee address will be added at the end
			continue
		}

		receiverAmountDfm := common.WeiToDfm(receiver.Amount)

		receivers = append(receivers, oCore.BridgingRequestReceiver{
			DestinationAddress: receiver.Address,
			Amount:             receiverAmountDfm,
			AmountWrapped:      big.NewInt(0),
		})

		totalAmount.Add(totalAmount, receiverAmountDfm)
	}

	feeCurrencyDfmDst := new(big.Int).SetUint64(cardanoDestConfig.FeeAddrBridgingAmount)
	totalAmountCurrencySrc := new(big.Int).Add(totalAmount, common.WeiToDfm(metadata.BridgingFee))
	totalAmountCurrencyDst := new(big.Int).Add(totalAmount, feeCurrencyDfmDst)

	receivers = append(receivers, oCore.BridgingRequestReceiver{
		DestinationAddress: cardanoDestChainFeeAddress,
		Amount:             feeCurrencyDfmDst,
		AmountWrapped:      big.NewInt(0),
	})

	claim := oCore.BridgingRequestClaim{
		ObservedTransactionHash:         tx.Hash,
		SourceChainId:                   chainIDConverter.ToChainIDNum(tx.OriginChainID),
		DestinationChainId:              chainIDConverter.ToChainIDNum(metadata.DestinationChainID),
		Receivers:                       receivers,
		NativeCurrencyAmountSource:      totalAmountCurrencySrc,
		NativeCurrencyAmountDestination: totalAmountCurrencyDst,
		WrappedTokenAmountSource:        big.NewInt(0),
		WrappedTokenAmountDestination:   big.NewInt(0),
		RetryCounter:                    big.NewInt(int64(tx.BatchTryCount)),
	}

	claims.BridgingRequestClaims = append(claims.BridgingRequestClaims, claim)

	p.logger.Info("Added BridgingRequestClaim",
		"txHash", tx.Hash, "metadata", metadata, "claim", oCore.BridgingRequestClaimString(claim, chainIDConverter))
}

func (p *BridgingRequestedProcessorImpl) validate(
	tx *core.EthTx, metadata *core.BridgingRequestEthMetadata, appConfig *oCore.AppConfig,
) error {
	if err := p.refundRequestProcessor.HandleBridgingProcessorPreValidate(tx, appConfig); err != nil {
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

	cardanoDestChainFeeAddress := appConfig.GetFeeMultisigAddress(metadata.DestinationChainID)

	if len(metadata.Transactions) > appConfig.BridgingSettings.MaxReceiversPerBridgingRequest {
		return fmt.Errorf("number of receivers in metadata greater than maximum allowed - no: %v, max: %v, metadata: %v",
			len(metadata.Transactions), appConfig.BridgingSettings.MaxReceiversPerBridgingRequest, metadata)
	}

	receiverAmountSum := big.NewInt(0)
	feeSum := big.NewInt(0)
	foundAUtxoValueBelowMinimumValue := false
	foundAnInvalidReceiverAddr := false

	currencySrcID, err := ethSrcConfig.GetCurrencyID()
	if err != nil {
		return err
	}

	_, err = oUtils.GetTokenPair(
		ethSrcConfig.DestinationChains,
		ethSrcConfig.ChainID,
		metadata.DestinationChainID,
		currencySrcID,
	)
	if err != nil {
		return fmt.Errorf("transaction direction not allowed. metadata: %v, err: %w", metadata, err)
	}

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

		if receiver.Address == cardanoDestChainFeeAddress {
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

	if appConfig.BridgingSettings.MaxAmountAllowedToBridge != nil &&
		appConfig.BridgingSettings.MaxAmountAllowedToBridge.Sign() > 0 &&
		receiverAmountSum.Cmp(common.DfmToWei(appConfig.BridgingSettings.MaxAmountAllowedToBridge)) == 1 {
		return fmt.Errorf("sum of receiver amounts + fee: %v greater than maximum allowed: %v",
			receiverAmountSum, common.DfmToWei(appConfig.BridgingSettings.MaxAmountAllowedToBridge))
	}

	// update fee amount if needed with sum of fee address receivers
	metadata.BridgingFee.Add(metadata.BridgingFee, feeSum)
	receiverAmountSum.Add(receiverAmountSum, metadata.BridgingFee)

	feeAmountDfm := common.WeiToDfm(metadata.BridgingFee)
	if feeAmountDfm.Uint64() < ethSrcConfig.MinFeeForBridging {
		return fmt.Errorf("bridging fee in metadata receivers is less than minimum: %v", metadata)
	}

	if tx.Value == nil || tx.Value.Cmp(receiverAmountSum) != 0 {
		return fmt.Errorf("tx value is not equal to sum of receiver amounts + fee: expected %v but got %v",
			receiverAmountSum, tx.Value)
	}

	return nil
}
