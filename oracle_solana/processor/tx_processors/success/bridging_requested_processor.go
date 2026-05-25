package successtxprocessors

import (
	"fmt"
	"math/big"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	oChain "github.com/Ethernal-Tech/apex-bridge/oracle_common/chain"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	oUtils "github.com/Ethernal-Tech/apex-bridge/oracle_common/utils"
	"github.com/Ethernal-Tech/apex-bridge/oracle_solana/core"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	solanawallet "github.com/Ethernal-Tech/solana-infrastructure/wallet"
	goEthCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-hclog"
)

var _ core.SolanaTxSuccessProcessor = (*BridgingRequestedProcessorImpl)(nil)

type BridgingRequestedProcessorImpl struct {
	refundRequestProcessor core.SolanaTxSuccessRefundProcessor
	logger                 hclog.Logger

	cardanoChainInfos map[string]*oChain.CardanoChainInfo
}

type receiverValidationCtxSolanaSrc struct {
	oCore.ReceiverValidationContext
	solanaSrcConfig *oCore.SolanaChainConfig
	metadata        *core.BridgingRequestSolMetadata
	feeSum          *big.Int
}

func NewSolanaBridgingRequestedProcessor(
	refundRequestProcessor core.SolanaTxSuccessRefundProcessor,
	logger hclog.Logger,
	cardanoChainInfos map[string]*oChain.CardanoChainInfo,
) *BridgingRequestedProcessorImpl {
	return &BridgingRequestedProcessorImpl{
		refundRequestProcessor: refundRequestProcessor,
		logger:                 logger.Named("solana_bridging_requested_processor"),
		cardanoChainInfos:      cardanoChainInfos,
	}
}

func (*BridgingRequestedProcessorImpl) GetType() common.BridgingTxType {
	return common.BridgingTxTypeBridgingRequest
}

func (*BridgingRequestedProcessorImpl) PreValidate(tx *core.SolanaTx, appConfig *oCore.AppConfig) error {
	return nil
}

func (p *BridgingRequestedProcessorImpl) ValidateAndAddClaim(
	claims *oCore.BridgeClaims, tx *core.SolanaTx, appConfig *oCore.AppConfig,
) error {
	metadata, err := core.UnmarshalSolMetadata[core.BridgingRequestSolMetadata](tx.Metadata)
	if err != nil {
		return p.refundRequestProcessor.HandleBridgingProcessorError(
			claims, tx, appConfig, err, "failed to unmarshal sol metadata")
	}

	if metadata.BridgingTxType != p.GetType() {
		return p.refundRequestProcessor.HandleBridgingProcessorError(
			claims, tx, appConfig, nil, "ValidateAndAddClaim called for irrelevant tx")
	}

	p.logger.Debug("Validating relevant tx", "txHash", tx.TxSignature, "metadata", metadata)

	err = p.validate(tx, metadata, appConfig)
	if err == nil {
		return p.addBridgingRequestClaim(claims, tx, metadata, appConfig)
	}

	return p.refundRequestProcessor.HandleBridgingProcessorError(
		claims, tx, appConfig, err, "validation failed for tx")
}

func (p *BridgingRequestedProcessorImpl) validate(
	tx *core.SolanaTx, metadata *core.BridgingRequestSolMetadata, appConfig *oCore.AppConfig,
) error {
	if err := p.refundRequestProcessor.HandleBridgingProcessorPreValidate(tx, appConfig); err != nil {
		return err
	}

	srcChain := oUtils.GetChainConfigResult(appConfig, tx.OriginChainID)
	if srcChain.IsNone() || srcChain.Solana == nil {
		return fmt.Errorf("unsupported chain id found in tx. chain id: %v", tx.OriginChainID)
	}

	destChain := oUtils.GetChainConfigResult(appConfig, metadata.DestinationChainID)
	if destChain.IsNone() {
		return fmt.Errorf("unsupported chain id found in tx. chain id: %v", metadata.DestinationChainID)
	}

	destChainInfo, err := oUtils.GetDestChainInfoResult(metadata.DestinationChainID, destChain, appConfig)
	if err != nil {
		return err
	}

	if err := p.validateOperationAndReceiverLimits(metadata, srcChain.Solana, appConfig); err != nil {
		return err
	}

	srcCurrencyID, err := srcChain.Solana.GetCurrencyID()
	if err != nil {
		return err
	}

	receiverCtx := &receiverValidationCtxSolanaSrc{
		solanaSrcConfig: srcChain.Solana,
		metadata:        metadata,
		feeSum:          big.NewInt(0),
		ReceiverValidationContext: oCore.ReceiverValidationContext{
			CardanoDestConfig: destChain.Cardano,
			EthDestConfig:     destChain.Eth,
			SolanaDestConfig:  destChain.Solana,
			DestFeeAddress:    destChainInfo.FeeAddress,
			BridgingSettings:  &appConfig.BridgingSettings,
			MinColCoinsAllowedToBridge: oUtils.MaxBigInt(
				common.LamportToWei(new(big.Int).SetUint64(srcChain.Solana.MinColCoinsAllowedToBridge)),
				destChainInfo.MinColCoinsAllowedToBridge),
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

	if err := p.validateTokenAmounts(tx.Value, receiverCtx); err != nil {
		return err
	}

	return nil
}

func (p *BridgingRequestedProcessorImpl) validateReceiver(
	receiver *core.BridgingRequestSolMetadataTransaction,
	ctx *receiverValidationCtxSolanaSrc,
) error {
	if oUtils.NormalizeAddr(receiver.Address) == oUtils.NormalizeAddr(ctx.DestFeeAddress) {
		if ctx.CurrencySrcID != receiver.TokenID {
			return fmt.Errorf("fee receiver metadata invalid. metadata: %v, receiver: %v", ctx.metadata, receiver)
		}

		ctx.feeSum.Add(ctx.feeSum, receiver.Amount)

		return nil
	}

	tokenPair, err := oUtils.GetTokenPair(
		ctx.solanaSrcConfig.DestinationChains,
		ctx.solanaSrcConfig.ChainID,
		ctx.metadata.DestinationChainID,
		receiver.TokenID,
	)
	if err != nil {
		return fmt.Errorf("invalid receiver. metadata: %v, receiver: %v, err: %w", ctx.metadata, receiver, err)
	}

	switch {
	case ctx.CardanoDestConfig != nil:
		return p.validateReceiverCardano(ctx, receiver, tokenPair)
	case ctx.EthDestConfig != nil:
		return p.validateReceiverEth(ctx, receiver, tokenPair)
	case ctx.SolanaDestConfig != nil:
		return p.validateReceiverSolana(ctx, receiver, tokenPair)
	default:
		return fmt.Errorf("invalid destination chain config")
	}
}

func (p *BridgingRequestedProcessorImpl) validateReceiverCardano(
	ctx *receiverValidationCtxSolanaSrc,
	receiver *core.BridgingRequestSolMetadataTransaction,
	tokenPair *common.TokenPair,
) error {
	if !cardanotx.IsValidOutputAddress(receiver.Address, ctx.CardanoDestConfig.NetworkID) {
		return fmt.Errorf(
			"found an invalid receiver addr in metadata. metadata: %v, receiver: %v", ctx.metadata, receiver)
	}

	if tokenPair.DestinationTokenID == ctx.CurrencyDestID {
		utxoMinWeiDest := common.DfmToWei(new(big.Int).SetUint64(ctx.CardanoDestConfig.UtxoMinAmount))
		if receiver.Amount.Cmp(utxoMinWeiDest) < 0 {
			return fmt.Errorf("found an utxo value below minimum value in metadata receivers: %v", ctx.metadata)
		}
	} else if receiver.Amount.Cmp(ctx.MinColCoinsAllowedToBridge) < 0 {
		return fmt.Errorf("token amount below minimum allowed in metadata receivers: %v", ctx.metadata)
	}

	if tokensSum, ok := ctx.AmountsSums[tokenPair.SourceTokenID]; ok {
		tokensSum.Add(tokensSum, receiver.Amount)
	} else {
		ctx.AmountsSums[tokenPair.SourceTokenID] = new(big.Int).Set(receiver.Amount)
	}

	return nil
}

func (p *BridgingRequestedProcessorImpl) validateReceiverEth(
	ctx *receiverValidationCtxSolanaSrc,
	receiver *core.BridgingRequestSolMetadataTransaction,
	tokenPair *common.TokenPair,
) error {
	if !goEthCommon.IsHexAddress(receiver.Address) {
		return fmt.Errorf(
			"found an invalid eth receiver addr in metadata. metadata: %v, receiver: %v", ctx.metadata, receiver)
	}

	return p.validateReceiverEthOrSol(ctx, receiver, tokenPair)
}

func (p *BridgingRequestedProcessorImpl) validateReceiverSolana(
	ctx *receiverValidationCtxSolanaSrc,
	receiver *core.BridgingRequestSolMetadataTransaction,
	tokenPair *common.TokenPair,
) error {
	if !solanawallet.IsSolanaAddress(receiver.Address) {
		return fmt.Errorf("found an invalid receiver addr in metadata. metadata: %v, receiver: %v", ctx.metadata, receiver)
	}

	return p.validateReceiverEthOrSol(ctx, receiver, tokenPair)
}

func (p *BridgingRequestedProcessorImpl) validateReceiverEthOrSol(
	ctx *receiverValidationCtxSolanaSrc,
	receiver *core.BridgingRequestSolMetadataTransaction,
	tokenPair *common.TokenPair,
) error {
	if tokensSum, ok := ctx.AmountsSums[tokenPair.SourceTokenID]; ok {
		tokensSum.Add(tokensSum, receiver.Amount)
	} else {
		ctx.AmountsSums[tokenPair.SourceTokenID] = new(big.Int).Set(receiver.Amount)
	}

	if receiver.Amount.Cmp(ctx.MinColCoinsAllowedToBridge) < 0 {
		return fmt.Errorf("token amount below minimum allowed in metadata receivers: %v", ctx.metadata)
	}

	return nil
}

func (p *BridgingRequestedProcessorImpl) validateTokenAmounts(
	txValue *big.Int,
	receiverCtx *receiverValidationCtxSolanaSrc,
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

	if metadata.BridgingFee.Cmp(receiverCtx.solanaSrcConfig.MinFeeForBridging) < 0 {
		return fmt.Errorf("bridging fee in metadata is less than minimum: %v", metadata)
	}

	if txValue == nil || txValue.Cmp(nativeCurrencySum) != 0 {
		return fmt.Errorf("tx value is not equal to sum of receiver amounts + fee: expected %v but got %v",
			nativeCurrencySum, txValue)
	}

	return nil
}

func (p *BridgingRequestedProcessorImpl) validateOperationAndReceiverLimits(
	metadata *core.BridgingRequestSolMetadata,
	solanaSrcConfig *oCore.SolanaChainConfig,
	appConfig *oCore.AppConfig,
) error {
	if metadata.OperationFee.Cmp(common.LamportToWei(new(big.Int).SetUint64(solanaSrcConfig.MinOperationFee))) == -1 {
		return fmt.Errorf("operation fee in metadata is less than minimum: %v", metadata)
	}

	if len(metadata.Transactions) > appConfig.BridgingSettings.MaxReceiversPerBridgingRequest {
		return fmt.Errorf("number of receivers in metadata greater than maximum allowed - no: %v, max: %v, metadata: %v",
			len(metadata.Transactions), appConfig.BridgingSettings.MaxReceiversPerBridgingRequest, metadata)
	}

	return nil
}

func (p *BridgingRequestedProcessorImpl) addBridgingRequestClaim(
	claims *oCore.BridgeClaims,
	tx *core.SolanaTx,
	metadata *core.BridgingRequestSolMetadata,
	appConfig *oCore.AppConfig,
) error {
	srcChain := oUtils.GetChainConfigResult(appConfig, tx.OriginChainID)
	if srcChain.IsNone() || srcChain.Solana == nil {
		return fmt.Errorf("unsupported chain id found in tx. chain id: %v", tx.OriginChainID)
	}

	destChain := oUtils.GetChainConfigResult(appConfig, metadata.DestinationChainID)
	if destChain.IsNone() {
		return fmt.Errorf("unsupported chain id found in tx. chain id: %v", metadata.DestinationChainID)
	}

	chainIDConverter := appConfig.ChainIDConverter

	destChainInfo, err := oUtils.GetDestChainInfoResult(metadata.DestinationChainID, destChain, appConfig)
	if err != nil {
		return err
	}

	srcCurrencyID, err := srcChain.Solana.GetCurrencyID()
	if err != nil {
		return err
	}

	var (
		receivers         = make([]oCore.BridgingRequestReceiver, 0, len(metadata.Transactions))
		totalTokensAmount = oCore.NewTotalTokensAmount()
	)

	processReceiver := func(
		receiver *core.BridgingRequestSolMetadataTransaction,
	) (*oCore.BridgingRequestReceiver, error) {
		switch destChain.GetChainType() {
		case common.ChainTypeCardanoStr:
			return p.processReceiverCardano(
				srcChain.Solana,
				destChain.Cardano,
				receiver,
				srcCurrencyID,
				destChainInfo.CurrencyTokenID,
				totalTokensAmount,
			)
		case common.ChainTypeEVMStr:
			return p.processReceiverEthOrSol(
				srcChain.Solana,
				metadata.DestinationChainID,
				receiver,
				srcCurrencyID,
				destChainInfo.CurrencyTokenID,
				destChain.Eth.AlwaysTrackCurrencyAndWrappedCurrency,
				destChain.Eth.Tokens,
				totalTokensAmount,
			)
		case common.ChainTypeSolanaStr:
			return p.processReceiverEthOrSol(
				srcChain.Solana,
				metadata.DestinationChainID,
				receiver,
				srcCurrencyID,
				destChainInfo.CurrencyTokenID,
				destChain.Solana.AlwaysTrackCurrencyAndWrappedCurrency,
				destChain.Solana.Tokens,
				totalTokensAmount,
			)
		default:
			return nil, fmt.Errorf("unknown destination chain type: %s", destChain.GetChainType())
		}
	}

	for _, receiver := range metadata.Transactions {
		if oUtils.NormalizeAddr(receiver.Address) == oUtils.NormalizeAddr(destChainInfo.FeeAddress) {
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

	totalTokensAmount.TotalAmountCurrencyDst = new(big.Int).Add(
		totalTokensAmount.TotalAmountCurrencyDst, destChainInfo.FeeAddrBridgingWei)

	receivers = append(receivers, oCore.BridgingRequestReceiver{
		DestinationAddress: destChainInfo.FeeAddress,
		Amount:             destChainInfo.FeeAddrBridgingWei,
		AmountWrapped:      big.NewInt(0),
		TokenId:            0,
	})

	claim := oCore.BridgingRequestClaim{
		ObservedTransactionHash:         tx.TxSignature[:],
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
		"txHash", tx.TxSignature, "metadata", metadata, "claim", oCore.BridgingRequestClaimString(claim, chainIDConverter))

	return nil
}

func (p *BridgingRequestedProcessorImpl) processReceiverCardano(
	solanaSrcConfig *oCore.SolanaChainConfig,
	cardanoDestConfig *oCore.CardanoChainConfig,
	receiver *core.BridgingRequestSolMetadataTransaction,
	currencySrcID, currencyDestID uint16,
	totalTokensAmount *oCore.TotalTokensAmount,
) (*oCore.BridgingRequestReceiver, error) {
	var (
		receiverCurrency *big.Int
		receiverTokens   = big.NewInt(0)
	)

	// validation has already checked that there is no error
	tokenPair, _ := oUtils.GetTokenPair(
		solanaSrcConfig.DestinationChains, solanaSrcConfig.ChainID,
		cardanoDestConfig.ChainID, receiver.TokenID)

	receiverAmountWei := receiver.Amount

	if tokenPair.DestinationTokenID == currencyDestID {
		// currency on destination
		receiverCurrency = receiverAmountWei

		if cardanoDestConfig.AlwaysTrackCurrencyAndWrappedCurrency || tokenPair.TrackDestinationToken {
			totalTokensAmount.TrackDestTokenAmount(receiverAmountWei, big.NewInt(0))
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

		receiverCurrency = common.DfmToWei(new(big.Int).SetUint64(dstMinUtxo))
		totalTokensAmount.TrackDestTokenAmount(receiverCurrency, big.NewInt(0))

		receiverTokens = receiverAmountWei

		// wrapped token on destination
		if (cardanoDestConfig.AlwaysTrackCurrencyAndWrappedCurrency || tokenPair.TrackDestinationToken) &&
			cardanoDestConfig.Tokens[tokenPair.DestinationTokenID].IsWrappedCurrency {
			totalTokensAmount.TrackDestTokenAmount(big.NewInt(0), receiverAmountWei)
		}
	}

	if solanaSrcConfig.AlwaysTrackCurrencyAndWrappedCurrency || tokenPair.TrackSourceToken {
		totalTokensAmount.TrackSourceTokenAmount(
			tokenPair.SourceTokenID,
			currencySrcID,
			receiverAmountWei,
			solanaSrcConfig.Tokens,
		)
	}

	return &oCore.BridgingRequestReceiver{
		DestinationAddress: receiver.Address,
		Amount:             receiverCurrency,
		AmountWrapped:      receiverTokens,
		TokenId:            tokenPair.DestinationTokenID,
	}, nil
}

func (p *BridgingRequestedProcessorImpl) processReceiverEthOrSol(
	solanaSrcConfig *oCore.SolanaChainConfig,
	destinationChainID string,
	receiver *core.BridgingRequestSolMetadataTransaction,
	currencySrcID, currencyDestID uint16,
	destAlwaysTrackCurrencyAndWrappedCurrency bool,
	destTokens map[uint16]common.Token,
	totalTokensAmount *oCore.TotalTokensAmount,
) (*oCore.BridgingRequestReceiver, error) {
	tokenPair, err := oUtils.GetTokenPair(
		solanaSrcConfig.DestinationChains, solanaSrcConfig.ChainID,
		destinationChainID, receiver.TokenID)
	if err != nil {
		return nil, err
	}

	amount := big.NewInt(0)
	amountWrapped := big.NewInt(0)
	receiverAmountWei := receiver.Amount

	// currency on destination
	if tokenPair.DestinationTokenID == currencyDestID {
		amount = receiverAmountWei

		if destAlwaysTrackCurrencyAndWrappedCurrency || tokenPair.TrackDestinationToken {
			totalTokensAmount.TrackDestTokenAmount(
				receiverAmountWei, big.NewInt(0),
			)
		}
	} else {
		amountWrapped = receiverAmountWei

		// wrapped token on destination
		if (destAlwaysTrackCurrencyAndWrappedCurrency || tokenPair.TrackDestinationToken) &&
			destTokens[tokenPair.DestinationTokenID].IsWrappedCurrency {
			totalTokensAmount.TrackDestTokenAmount(
				big.NewInt(0),
				receiverAmountWei,
			)
		}
	}

	if solanaSrcConfig.AlwaysTrackCurrencyAndWrappedCurrency || tokenPair.TrackSourceToken {
		totalTokensAmount.TrackSourceTokenAmount(
			tokenPair.SourceTokenID,
			currencySrcID,
			receiverAmountWei,
			solanaSrcConfig.Tokens,
		)
	}

	return &oCore.BridgingRequestReceiver{
		DestinationAddress: receiver.Address,
		Amount:             amount,
		AmountWrapped:      amountWrapped,
		TokenId:            tokenPair.DestinationTokenID,
	}, nil
}

func (p *BridgingRequestedProcessorImpl) calculateMinUtxo(
	config *oCore.CardanoChainConfig, receiverAddr string, nativeTokensSum map[uint16]*big.Int,
) (uint64, error) {
	builder, err := cardanowallet.NewTxBuilder(cardanowallet.ResolveCardanoCliBinary(config.CardanoCliBinaryName))
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
