package successtxprocessors

import (
	"fmt"
	"math/big"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	oChain "github.com/Ethernal-Tech/apex-bridge/oracle_common/chain"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	oUtils "github.com/Ethernal-Tech/apex-bridge/oracle_common/utils"
	"github.com/Ethernal-Tech/apex-bridge/oracle_eth/core"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	goEthCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-hclog"
)

var _ core.EthTxSuccessProcessor = (*BridgingRequestedProcessorImpl)(nil)

type BridgingRequestedProcessorSkylineImpl struct {
	refundRequestProcessor core.EthTxSuccessRefundProcessor
	logger                 hclog.Logger

	cardanoChainInfos map[string]*oChain.CardanoChainInfo
}

type receiverValidationCtxEthSrc struct {
	oCore.ReceiverValidationContext
	ethSrcConfig *oCore.EthChainConfig
	metadata     *core.BridgingRequestEthMetadata
	feeSum       *big.Int
}

func NewEthBridgingRequestedProcessorSkyline(
	refundRequestProcessor core.EthTxSuccessRefundProcessor, logger hclog.Logger,
	cardanoChainInfos map[string]*oChain.CardanoChainInfo,
) *BridgingRequestedProcessorSkylineImpl {
	return &BridgingRequestedProcessorSkylineImpl{
		refundRequestProcessor: refundRequestProcessor,
		cardanoChainInfos:      cardanoChainInfos,
		logger:                 logger.Named("eth_bridging_requested_processor"),
	}
}

func (*BridgingRequestedProcessorSkylineImpl) GetType() common.BridgingTxType {
	return common.BridgingTxTypeBridgingRequest
}

func (*BridgingRequestedProcessorSkylineImpl) PreValidate(tx *core.EthTx, appConfig *oCore.AppConfig) error {
	return nil
}

func (p *BridgingRequestedProcessorSkylineImpl) ValidateAndAddClaim(
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
		return p.addBridgingRequestClaim(claims, tx, metadata, appConfig)
	} else {
		return p.refundRequestProcessor.HandleBridgingProcessorError(
			claims, tx, appConfig, err, "validation failed for tx")
	}
}

func (p *BridgingRequestedProcessorSkylineImpl) addBridgingRequestClaim(
	claims *oCore.BridgeClaims, tx *core.EthTx,
	metadata *core.BridgingRequestEthMetadata, appConfig *oCore.AppConfig,
) error {
	_, ethSrcConfig := oUtils.GetChainConfig(appConfig, tx.OriginChainID)
	cardanoDestConfig, ethDestConfig := oUtils.GetChainConfig(appConfig, metadata.DestinationChainID)
	chainIDConverter := appConfig.ChainIDConverter

	destChainInfo, err := oUtils.GetDestChainInfo(
		metadata.DestinationChainID, appConfig, cardanoDestConfig, ethDestConfig)
	if err != nil {
		return err
	}

	srcCurrencyID, err := ethSrcConfig.GetCurrencyID()
	if err != nil {
		return err
	}

	receivers := make([]oCore.BridgingRequestReceiver, 0, len(metadata.Transactions))
	totalTokensAmount := oCore.NewTotalTokensAmount()

	processReceiver := func(
		receiver *core.BridgingRequestEthMetadataTransaction,
	) (*oCore.BridgingRequestReceiver, error) {
		if cardanoDestConfig != nil {
			return p.processReceiverCardano(
				ethSrcConfig,
				cardanoDestConfig,
				receiver,
				srcCurrencyID,
				destChainInfo.CurrencyTokenID,
				totalTokensAmount,
			)
		}

		return p.processReceiverEth(
			ethSrcConfig,
			ethDestConfig,
			metadata.DestinationChainID,
			receiver,
			srcCurrencyID,
			destChainInfo.CurrencyTokenID,
			totalTokensAmount,
		)
	}

	for _, receiver := range metadata.Transactions {
		if receiver.Address == destChainInfo.FeeAddress {
			// fee address will be added at the end
			continue
		}

		brReceiver, err := processReceiver(&receiver)
		if err != nil {
			return fmt.Errorf(
				"failed to process receiver (chain %s, receiver address: %v): %w",
				metadata.DestinationChainID,
				receiver.Address,
				err,
			)
		}

		receivers = append(receivers, *brReceiver)
	}

	// wTODO: think about whether we should always track all currency amounts,
	// regardless of .TrackSource and .TrackDestination
	totalTokensAmount.TotalAmountCurrencySrc = new(big.Int).Add(
		totalTokensAmount.TotalAmountCurrencySrc, metadata.BridgingFee)

	totalTokensAmount.TotalAmountCurrencySrc = new(big.Int).Add(
		totalTokensAmount.TotalAmountCurrencySrc, metadata.OperationFee)

	totalTokensAmount.TotalAmountCurrencyDst = new(big.Int).Add(
		totalTokensAmount.TotalAmountCurrencyDst, destChainInfo.FeeAddrBridgingWei)

	receivers = append(receivers, oCore.BridgingRequestReceiver{
		DestinationAddress: destChainInfo.FeeAddress,
		Amount:             destChainInfo.FeeAddrBridgingWei,
		AmountWrapped:      big.NewInt(0),
		TokenId:            0,
	})

	claim := oCore.BridgingRequestClaim{
		ObservedTransactionHash:         tx.Hash,
		SourceChainId:                   chainIDConverter.ToChainIDNum(tx.OriginChainID),
		DestinationChainId:              chainIDConverter.ToChainIDNum(metadata.DestinationChainID),
		Receivers:                       receivers,
		NativeCurrencyAmountSource:      totalTokensAmount.TotalAmountCurrencySrc,
		NativeCurrencyAmountDestination: totalTokensAmount.TotalAmountCurrencyDst,
		WrappedTokenAmountSource:        totalTokensAmount.TotalAmountWrappedSrc,
		WrappedTokenAmountDestination:   totalTokensAmount.TotalAmountWrappedDst,
		RetryCounter:                    big.NewInt(int64(tx.BatchTryCount)),
	}

	claims.BridgingRequestClaims = append(claims.BridgingRequestClaims, claim)

	p.logger.Info("Added BridgingRequestClaim",
		"txHash", tx.Hash, "metadata", metadata, "claim", oCore.BridgingRequestClaimString(claim, chainIDConverter))

	return nil
}

func (p *BridgingRequestedProcessorSkylineImpl) validate(
	tx *core.EthTx, metadata *core.BridgingRequestEthMetadata, appConfig *oCore.AppConfig,
) error {
	if err := p.refundRequestProcessor.HandleBridgingProcessorPreValidate(tx, appConfig); err != nil {
		return err
	}

	_, ethSrcConfig := oUtils.GetChainConfig(appConfig, tx.OriginChainID)
	if ethSrcConfig == nil {
		return fmt.Errorf("origin chain not registered: %v", tx.OriginChainID)
	}

	cardanoDestConfig, ethDestConfig := oUtils.GetChainConfig(appConfig, metadata.DestinationChainID)

	destChainInfo, err := oUtils.GetDestChainInfo(metadata.DestinationChainID, appConfig, cardanoDestConfig, ethDestConfig)
	if err != nil {
		return err
	}

	if err := p.validateOperationAndReceiverLimits(metadata, ethSrcConfig, appConfig); err != nil {
		return err
	}

	srcCurrencyID, err := ethSrcConfig.GetCurrencyID()
	if err != nil {
		return err
	}

	receiverCtx := &receiverValidationCtxEthSrc{
		ethSrcConfig: ethSrcConfig,
		metadata:     metadata,
		feeSum:       big.NewInt(0),
		ReceiverValidationContext: oCore.ReceiverValidationContext{
			CardanoDestConfig: cardanoDestConfig,
			EthDestConfig:     ethDestConfig,
			DestFeeAddress:    destChainInfo.FeeAddress,
			BridgingSettings:  &appConfig.BridgingSettings,

			AmountsSums:    make(map[uint16]*big.Int),
			CurrencySrcID:  srcCurrencyID,
			CurrencyDestID: destChainInfo.CurrencyTokenID,
		},
	}

	for _, receiver := range metadata.Transactions {
		if err := p.validateReceiver(&receiver, receiverCtx); err != nil {
			return err
		}
	}

	if err := p.validateTokenAmounts(
		tx.Value, receiverCtx,
	); err != nil {
		return err
	}

	return nil
}

func (p *BridgingRequestedProcessorSkylineImpl) validateOperationAndReceiverLimits(
	metadata *core.BridgingRequestEthMetadata,
	ethSrcConfig *oCore.EthChainConfig,
	appConfig *oCore.AppConfig,
) error {
	if metadata.OperationFee.Cmp(ethSrcConfig.MinOperationFee) < 0 {
		return fmt.Errorf("operation fee in metadata is less than minimum: %v", metadata)
	}

	if len(metadata.Transactions) > appConfig.BridgingSettings.MaxReceiversPerBridgingRequest {
		return fmt.Errorf("number of receivers in metadata greater than maximum allowed - no: %v, max: %v, metadata: %v",
			len(metadata.Transactions), appConfig.BridgingSettings.MaxReceiversPerBridgingRequest, metadata)
	}

	return nil
}

func (p *BridgingRequestedProcessorSkylineImpl) validateReceiver(
	receiver *core.BridgingRequestEthMetadataTransaction,
	ctx *receiverValidationCtxEthSrc,
) error {
	if oUtils.NormalizeAddr(receiver.Address) == oUtils.NormalizeAddr(ctx.DestFeeAddress) {
		if ctx.CurrencySrcID != receiver.TokenID {
			return fmt.Errorf("fee receiver metadata invalid. metadata: %v, receiver: %v", ctx.metadata, receiver)
		}

		ctx.feeSum.Add(ctx.feeSum, receiver.Amount)

		return nil
	}

	tokenPair, err := oUtils.GetTokenPair(
		ctx.ethSrcConfig.DestinationChains,
		ctx.ethSrcConfig.ChainID,
		ctx.metadata.DestinationChainID,
		receiver.TokenID,
	)
	if err != nil {
		return fmt.Errorf("invalid receiver. metadata: %v, receiver: %v, err: %w", ctx.metadata, receiver, err)
	}

	if ctx.CardanoDestConfig != nil {
		return p.validateReceiverCardano(ctx, receiver, tokenPair)
	}

	return p.validateReceiverEth(ctx, receiver, tokenPair)
}

func (p *BridgingRequestedProcessorSkylineImpl) validateReceiverCardano(
	ctx *receiverValidationCtxEthSrc,
	receiver *core.BridgingRequestEthMetadataTransaction,
	tokenPair *common.TokenPair,
) error {
	if !cardanotx.IsValidOutputAddress(receiver.Address, ctx.CardanoDestConfig.NetworkID) {
		return fmt.Errorf(
			"found an invalid receiver addr in metadata. metadata: %v, receiver: %v", ctx.metadata, receiver)
	}

	// currency on destination
	if tokenPair.DestinationTokenID == ctx.CurrencyDestID {
		utxoMinWeiDest := common.DfmToWei(new(big.Int).SetUint64(ctx.CardanoDestConfig.UtxoMinAmount))
		if receiver.Amount.Cmp(utxoMinWeiDest) < 0 {
			return fmt.Errorf("found an utxo value below minimum value in metadata receivers: %v", ctx.metadata)
		}
	} else {
		if receiver.Amount.Cmp(ctx.BridgingSettings.MinColCoinsAllowedToBridge) < 0 {
			return fmt.Errorf("token amount below minimum allowed in metadata receivers: %v", ctx.metadata)
		}
	}

	if tokensSum, ok := ctx.AmountsSums[tokenPair.SourceTokenID]; ok {
		tokensSum.Add(tokensSum, receiver.Amount)
	} else {
		ctx.AmountsSums[tokenPair.SourceTokenID] = new(big.Int).Set(receiver.Amount)
	}

	return nil
}

func (p *BridgingRequestedProcessorSkylineImpl) validateReceiverEth(
	ctx *receiverValidationCtxEthSrc,
	receiver *core.BridgingRequestEthMetadataTransaction,
	tokenPair *common.TokenPair,
) error {
	if !goEthCommon.IsHexAddress(receiver.Address) {
		return fmt.Errorf(
			"found an invalid eth receiver addr in metadata. metadata: %v, receiver: %v",
			ctx.metadata, receiver)
	}

	if tokensSum, ok := ctx.AmountsSums[tokenPair.SourceTokenID]; ok {
		tokensSum.Add(tokensSum, receiver.Amount)
	} else {
		ctx.AmountsSums[tokenPair.SourceTokenID] = new(big.Int).Set(receiver.Amount)
	}

	if receiver.Amount.Cmp(ctx.BridgingSettings.MinColCoinsAllowedToBridge) < 0 {
		return fmt.Errorf("token amount below minimum allowed in metadata receivers: %v", ctx.metadata)
	}

	return nil
}

func (p *BridgingRequestedProcessorSkylineImpl) validateTokenAmounts(
	txValue *big.Int,
	receiverCtx *receiverValidationCtxEthSrc,
) error {
	metadata := receiverCtx.metadata

	nativeCurrencySum, ok := receiverCtx.AmountsSums[receiverCtx.CurrencySrcID]
	if !ok {
		nativeCurrencySum = new(big.Int).SetInt64(0)
	}

	// Remove currency entry from the map
	delete(receiverCtx.AmountsSums, receiverCtx.CurrencySrcID)

	maxCurrAmt := receiverCtx.BridgingSettings.MaxAmountAllowedToBridge
	if maxCurrAmt != nil && maxCurrAmt.Sign() > 0 && nativeCurrencySum.Cmp(maxCurrAmt) == 1 {
		return fmt.Errorf("sum of receiver amounts: %v greater than maximum allowed: %v",
			nativeCurrencySum, maxCurrAmt)
	}

	maxTokenAmt := receiverCtx.BridgingSettings.MaxTokenAmountAllowedToBridge
	if maxTokenAmt != nil && maxTokenAmt.Sign() > 0 {
		for tokenID, tokenSum := range receiverCtx.AmountsSums {
			if tokenSum.Cmp(maxTokenAmt) == 1 {
				return fmt.Errorf(
					"amount of tokens for receivers too high for token with ID %d: %v greater than maximum allowed: %v",
					tokenID, tokenSum, maxTokenAmt)
			}
		}
	}

	// update fee amount if needed with sum of fee address receivers
	metadata.BridgingFee.Add(metadata.BridgingFee, receiverCtx.feeSum)
	nativeCurrencySum.Add(nativeCurrencySum, metadata.BridgingFee)
	nativeCurrencySum.Add(nativeCurrencySum, metadata.OperationFee)

	if metadata.BridgingFee.Cmp(receiverCtx.ethSrcConfig.MinFeeForBridging) < 0 {
		return fmt.Errorf("bridging fee in metadata is less than minimum: %v", metadata)
	}

	if txValue == nil || txValue.Cmp(nativeCurrencySum) != 0 {
		return fmt.Errorf("tx value is not equal to sum of receiver amounts + fee: expected %v but got %v",
			nativeCurrencySum, txValue)
	}

	return nil
}

func (p *BridgingRequestedProcessorSkylineImpl) calculateMinUtxo(
	config *oCore.CardanoChainConfig, receiverAddr string, nativeTokensSum map[uint16]*big.Int,
) (uint64, error) {
	builder, err := cardanowallet.NewTxBuilder(cardanowallet.ResolveCardanoCliBinary(config.NetworkID))
	if err != nil {
		return 0, err
	}

	defer builder.Dispose()

	chainInfo, exists := p.cardanoChainInfos[config.ChainID]
	if !exists {
		return 0, fmt.Errorf("chain info not found for chainID: %s", config.ChainID)
	}

	builder.SetProtocolParameters(chainInfo.ProtocolParams)

	tokenAmounts := make([]cardanowallet.TokenAmount, 0, len(nativeTokensSum))

	for tokenID, tokenAmount := range nativeTokensSum {
		tokenName := config.Tokens[tokenID].ChainSpecific

		nativeToken, err := cardanowallet.NewTokenWithFullNameTry(tokenName)
		if err != nil {
			return 0, err
		}

		tokenAmounts = append(tokenAmounts, cardanowallet.NewTokenAmount(nativeToken, tokenAmount.Uint64()))
	}

	potentialTokenCost, err := cardanowallet.GetMinUtxoForSumMap(
		builder,
		receiverAddr,
		cardanowallet.GetTokensSumMap(tokenAmounts...),
		nil,
	)
	if err != nil {
		return 0, err
	}

	return max(config.UtxoMinAmount, potentialTokenCost), nil
}

func (p *BridgingRequestedProcessorSkylineImpl) processReceiverCardano(
	ethSrcConfig *oCore.EthChainConfig,
	cardanoDestConfig *oCore.CardanoChainConfig,
	receiver *core.BridgingRequestEthMetadataTransaction,
	currencySrcID, currencyDestID uint16,
	totalTokensAmount *oCore.TotalTokensAmount,
) (*oCore.BridgingRequestReceiver, error) {
	// validation has already checked that there is no error
	tokenPair, _ := oUtils.GetTokenPair(
		ethSrcConfig.DestinationChains,
		ethSrcConfig.ChainID,
		cardanoDestConfig.ChainID,
		receiver.TokenID,
	)

	var amount *big.Int

	amountWrapped := big.NewInt(0)

	// currency on destination
	if tokenPair.DestinationTokenID == currencyDestID {
		amount = receiver.Amount

		if tokenPair.TrackDestinationToken {
			totalTokensAmount.TrackDestTokenAmount(
				receiver.Amount,
				big.NewInt(0),
			)
		}
	} else {
		nativeTokensSum := map[uint16]*big.Int{
			tokenPair.DestinationTokenID: common.WeiToDfm(receiver.Amount),
		}

		dstMinUtxo, err := p.calculateMinUtxo(cardanoDestConfig, receiver.Address, nativeTokensSum)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate destination minUtxo for chainID: %s. err: %w",
				cardanoDestConfig.ChainID, err)
		}

		amount = common.DfmToWei(new(big.Int).SetUint64(dstMinUtxo))
		totalTokensAmount.TrackDestTokenAmount(amount, big.NewInt(0))

		amountWrapped = receiver.Amount

		// wrapped token on destination
		if tokenPair.TrackDestinationToken && cardanoDestConfig.Tokens[tokenPair.DestinationTokenID].IsWrappedCurrency {
			totalTokensAmount.TrackDestTokenAmount(
				big.NewInt(0),
				receiver.Amount,
			)
		}
	}

	if tokenPair.TrackSourceToken {
		totalTokensAmount.TrackSourceTokenAmount(
			tokenPair.SourceTokenID,
			currencySrcID,
			receiver.Amount,
			ethSrcConfig.Tokens,
		)
	}

	return &oCore.BridgingRequestReceiver{
		DestinationAddress: receiver.Address,
		Amount:             amount,
		AmountWrapped:      amountWrapped,
		TokenId:            tokenPair.DestinationTokenID,
	}, nil
}

func (p *BridgingRequestedProcessorSkylineImpl) processReceiverEth(
	ethSrcConfig *oCore.EthChainConfig,
	ethDestConfig *oCore.EthChainConfig,
	destinationChainID string,
	receiver *core.BridgingRequestEthMetadataTransaction,
	currencySrcID, currencyDestID uint16,
	totalTokensAmount *oCore.TotalTokensAmount,
) (*oCore.BridgingRequestReceiver, error) {
	tokenPair, err := oUtils.GetTokenPair(
		ethSrcConfig.DestinationChains, ethSrcConfig.ChainID,
		destinationChainID, receiver.TokenID)
	if err != nil {
		return nil, err
	}

	amount := big.NewInt(0)
	amountWrapped := big.NewInt(0)
	//receiverAmountDfm := common.WeiToDfm(receiver.Amount)

	// currency on destination
	if tokenPair.DestinationTokenID == currencyDestID {
		amount = receiver.Amount

		if tokenPair.TrackDestinationToken {
			totalTokensAmount.TrackDestTokenAmount(
				receiver.Amount, big.NewInt(0),
			)
		}
	} else {
		amountWrapped = receiver.Amount

		// wrapped token on destination
		if tokenPair.TrackDestinationToken && ethDestConfig.Tokens[tokenPair.DestinationTokenID].IsWrappedCurrency {
			totalTokensAmount.TrackDestTokenAmount(
				big.NewInt(0), receiver.Amount,
			)
		}
	}

	if tokenPair.TrackSourceToken {
		totalTokensAmount.TrackSourceTokenAmount(
			tokenPair.SourceTokenID, currencySrcID, receiver.Amount, ethSrcConfig.Tokens,
		)
	}

	return &oCore.BridgingRequestReceiver{
		DestinationAddress: receiver.Address,
		Amount:             amount,
		AmountWrapped:      amountWrapped,
		TokenId:            tokenPair.DestinationTokenID,
	}, nil
}
