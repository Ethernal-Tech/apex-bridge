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
	goEthCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-hclog"
)

var _ core.CardanoTxSuccessProcessor = (*BridgingRequestedProcessorImpl)(nil)

type BridgingRequestedProcessorImpl struct {
	refundRequestProcessor core.CardanoTxSuccessRefundProcessor
	logger                 hclog.Logger
}

func NewBridgingRequestedProcessor(
	refundRequestProcessor core.CardanoTxSuccessRefundProcessor,
	logger hclog.Logger,
) *BridgingRequestedProcessorImpl {
	return &BridgingRequestedProcessorImpl{
		refundRequestProcessor: refundRequestProcessor,
		logger:                 logger.Named("bridging_requested_processor"),
	}
}

func (*BridgingRequestedProcessorImpl) GetType() common.BridgingTxType {
	return common.BridgingTxTypeBridgingRequest
}

func (*BridgingRequestedProcessorImpl) PreValidate(tx *core.CardanoTx, appConfig *cCore.AppConfig) error {
	return nil
}

func (p *BridgingRequestedProcessorImpl) ValidateAndAddClaim(
	claims *cCore.BridgeClaims, tx *core.CardanoTx, appConfig *cCore.AppConfig,
) error {
	metadata, err := common.UnmarshalMetadata[common.BridgingRequestMetadata](
		common.MetadataEncodingTypeCbor, tx.Metadata)
	if err != nil {
		return p.refundRequestProcessor.HandleBridgingProcessorError(
			claims, tx, appConfig, err, "failed to unmarshal metadata")
	}

	if common.BridgingTxType(metadata.BridgingTxType) != p.GetType() {
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
	claims *cCore.BridgeClaims, tx *core.CardanoTx,
	metadata *common.BridgingRequestMetadata, appConfig *cCore.AppConfig,
) {
	var (
		totalAmount       = big.NewInt(0)
		feeCurrencyDstWei *big.Int
		feeAddress        string
		chainIDConverter  = appConfig.ChainIDConverter
	)

	cardanoDestConfig, ethDestConfig := cUtils.GetChainConfig(appConfig, metadata.DestinationChainID)

	switch {
	case cardanoDestConfig != nil:
		feeAddress = appConfig.GetFeeMultisigAddress(metadata.DestinationChainID)
		feeCurrencyDstWei = common.DfmToWei(new(big.Int).SetUint64(cardanoDestConfig.FeeAddrBridgingAmount))
	case ethDestConfig != nil:
		feeAddress = common.EthZeroAddr
		feeCurrencyDstWei = ethDestConfig.FeeAddrBridgingAmount
	default:
		p.logger.Warn("Added BridgingRequestClaim not supported chain", "chainId", metadata.DestinationChainID)

		return
	}

	receivers := make([]cCore.BridgingRequestReceiver, 0, len(metadata.Transactions))

	for _, receiver := range metadata.Transactions {
		receiverAddr := strings.Join(receiver.Address, "")

		if receiverAddr == feeAddress {
			// fee address will be added at the end
			continue
		}

		receiverAmountWei := common.DfmToWei(new(big.Int).SetUint64(receiver.Amount))
		receivers = append(receivers, cCore.BridgingRequestReceiver{
			DestinationAddress: receiverAddr,
			Amount:             receiverAmountWei,
			AmountWrapped:      big.NewInt(0),
		})

		totalAmount.Add(totalAmount, receiverAmountWei)
	}

	totalAmountCurrencySrc := new(big.Int).Add(totalAmount, common.DfmToWei(new(big.Int).SetUint64(metadata.BridgingFee)))
	totalAmountCurrencyDst := new(big.Int).Add(totalAmount, feeCurrencyDstWei)

	receivers = append(receivers, cCore.BridgingRequestReceiver{
		DestinationAddress: feeAddress,
		Amount:             feeCurrencyDstWei,
		AmountWrapped:      big.NewInt(0),
	})

	claim := cCore.BridgingRequestClaim{
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
		"txHash", tx.Hash, "metadata", metadata, "claim", cCore.BridgingRequestClaimString(claim, chainIDConverter))
}

func (p *BridgingRequestedProcessorImpl) validate(
	tx *core.CardanoTx, metadata *common.BridgingRequestMetadata, appConfig *cCore.AppConfig,
) error {
	if err := p.refundRequestProcessor.HandleBridgingProcessorPreValidate(tx, appConfig); err != nil {
		return err
	}

	chainConfig := appConfig.CardanoChains[tx.OriginChainID]
	if chainConfig == nil {
		return fmt.Errorf("unsupported chain id found in tx. chain id: %v", tx.OriginChainID)
	}

	if err := utils.ValidateOutputsHaveUnknownTokens(tx, appConfig, false); err != nil {
		return err
	}

	multisigUtxo, err := utils.ValidateTxOutputs(tx, appConfig, false)
	if err != nil {
		return err
	}

	cardanoSrcConfig, _ := cUtils.GetChainConfig(appConfig, tx.OriginChainID)
	if cardanoSrcConfig == nil {
		return fmt.Errorf("origin chain not registered: %v", tx.OriginChainID)
	}

	cardanoDestConfig, ethDestConfig := cUtils.GetChainConfig(appConfig, metadata.DestinationChainID)
	if cardanoDestConfig == nil && ethDestConfig == nil {
		return fmt.Errorf("destination chain not registered: %v", metadata.DestinationChainID)
	}

	if len(metadata.Transactions) > appConfig.BridgingSettings.MaxReceiversPerBridgingRequest {
		return fmt.Errorf("number of receivers in metadata greater than maximum allowed - no: %v, max: %v, metadata: %v",
			len(metadata.Transactions), appConfig.BridgingSettings.MaxReceiversPerBridgingRequest, metadata)
	}

	receiverAmountSum := big.NewInt(0)
	feeSum := uint64(0)
	foundAUtxoValueBelowMinimumValue := false
	foundAnInvalidReceiverAddr := false

	isCardanoDest := cardanoDestConfig != nil

	cardanoDestChainFeeAddress := ""
	if isCardanoDest {
		cardanoDestChainFeeAddress = appConfig.GetFeeMultisigAddress(metadata.DestinationChainID)
	}

	currencySrcID, err := cardanoSrcConfig.GetCurrencyID()
	if err != nil {
		return err
	}

	_, err = cUtils.GetTokenPair(
		cardanoSrcConfig.DestinationChains,
		cardanoSrcConfig.ChainID,
		metadata.DestinationChainID,
		currencySrcID,
	)
	if err != nil {
		return fmt.Errorf("transaction direction not allowed. metadata: %v, err: %w", metadata, err)
	}

	for _, receiver := range metadata.Transactions {
		receiverAddr := strings.Join(receiver.Address, "")

		if isCardanoDest {
			if receiver.Amount < cardanoDestConfig.UtxoMinAmount {
				foundAUtxoValueBelowMinimumValue = true

				break
			}

			if !cardanotx.IsValidOutputAddress(receiverAddr, cardanoDestConfig.NetworkID) {
				foundAnInvalidReceiverAddr = true

				break
			}

			if receiverAddr == cardanoDestChainFeeAddress {
				feeSum += receiver.Amount
			} else {
				receiverAmountSum.Add(receiverAmountSum, new(big.Int).SetUint64(receiver.Amount))
			}
		} else if ethDestConfig != nil {
			if !goEthCommon.IsHexAddress(receiverAddr) {
				foundAnInvalidReceiverAddr = true

				break
			}

			if receiverAddr == common.EthZeroAddr {
				feeSum += receiver.Amount
			} else {
				receiverAmountSum.Add(receiverAmountSum, new(big.Int).SetUint64(receiver.Amount))
			}
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
		receiverAmountSum.Cmp(common.WeiToDfm(appConfig.BridgingSettings.MaxAmountAllowedToBridge)) == 1 {
		return fmt.Errorf("sum of receiver amounts + fee: %v greater than maximum allowed: %v",
			receiverAmountSum, common.WeiToDfm(appConfig.BridgingSettings.MaxAmountAllowedToBridge))
	}

	// update fee amount if needed with sum of fee address receivers
	metadata.BridgingFee += feeSum
	receiverAmountSum.Add(receiverAmountSum, new(big.Int).SetUint64(metadata.BridgingFee))

	if metadata.BridgingFee < cardanoSrcConfig.DefaultMinFeeForBridging {
		return fmt.Errorf("bridging fee in metadata receivers is less than minimum: fee %d, minFee %d, metadata %v",
			metadata.BridgingFee, cardanoSrcConfig.DefaultMinFeeForBridging, metadata)
	}

	if receiverAmountSum.Cmp(new(big.Int).SetUint64(multisigUtxo.Amount)) != 0 {
		return fmt.Errorf("multisig amount is not equal to sum of receiver amounts + fee: expected %v but got %v",
			multisigUtxo.Amount, receiverAmountSum)
	}

	return nil
}
