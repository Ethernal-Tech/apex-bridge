package successtxprocessors

import (
	"fmt"
	"math/big"
	"strings"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/utils"
	cChain "github.com/Ethernal-Tech/apex-bridge/oracle_common/chain"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	cUtils "github.com/Ethernal-Tech/apex-bridge/oracle_common/utils"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/sendtx"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	goEthCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-hclog"
)

var _ core.CardanoTxSuccessProcessor = (*BridgingRequestedProcessorSkylineImpl)(nil)

type BridgingRequestedProcessorSkylineImpl struct {
	refundRequestProcessor core.CardanoTxSuccessRefundProcessor
	logger                 hclog.Logger

	chainInfos map[string]*cChain.CardanoChainInfo
}

type receiverValidationCtxCardanoSrc struct {
	cCore.ReceiverValidationContext
	cardanoSrcConfig *cCore.CardanoChainConfig
	metadata         *common.BridgingRequestMetadata
	feeSum           uint64
}

type totalTokensAmount struct {
	totalAmountCurrencySrc uint64
	totalAmountWrappedSrc  uint64
	totalAmountCurrencyDst uint64
	totalAmountWrappedDst  uint64
}

func NewSkylineBridgingRequestedProcessor(
	refundRequestProcessor core.CardanoTxSuccessRefundProcessor,
	logger hclog.Logger, chainInfos map[string]*cChain.CardanoChainInfo,
) *BridgingRequestedProcessorSkylineImpl {
	return &BridgingRequestedProcessorSkylineImpl{
		refundRequestProcessor: refundRequestProcessor,
		logger:                 logger.Named("bridging_requested_processor"),
		chainInfos:             chainInfos,
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
	chainConfig := appConfig.CardanoChains[tx.OriginChainID]
	if chainConfig == nil {
		return fmt.Errorf("unsupported chain id found in tx. chain id: %v", tx.OriginChainID)
	}

	metadata, err := unmarshalBridgingRequestMetadata(chainConfig, tx.Metadata)
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
		return p.addBridgingRequestClaim(claims, tx, metadata, appConfig)
	} else {
		return p.refundRequestProcessor.HandleBridgingProcessorError(
			claims, tx, appConfig, err, "validation failed for tx")
	}
}

func (p *BridgingRequestedProcessorSkylineImpl) addBridgingRequestClaim(
	claims *cCore.BridgeClaims, tx *core.CardanoTx,
	metadata *common.BridgingRequestMetadata, appConfig *cCore.AppConfig,
) error {
	chainIDConverter := appConfig.ChainIDConverter

	cardanoSrcConfig, _ := cUtils.GetChainConfig(appConfig, tx.OriginChainID)
	cardanoDestConfig, ethDestConfig := cUtils.GetChainConfig(appConfig, metadata.DestinationChainID)

	destChainInfo, err := cUtils.GetDestChainInfo(
		metadata.DestinationChainID, appConfig, cardanoDestConfig, ethDestConfig)
	if err != nil {
		return err
	}

	var (
		receivers         = make([]cCore.BridgingRequestReceiver, 0, len(metadata.Transactions))
		totalTokensAmount = totalTokensAmount{}
	)

	currencySrcID, err := cardanoSrcConfig.GetCurrencyID()
	if err != nil {
		return err
	}

	processReceiver := func(
		receiver *sendtx.BridgingRequestMetadataTransaction,
	) (*cCore.BridgingRequestReceiver, error) {
		if cardanoDestConfig != nil {
			return p.processReceiverCardano(
				cardanoSrcConfig,
				cardanoDestConfig,
				receiver,
				currencySrcID,
				destChainInfo.CurrencyTokenID,
				&totalTokensAmount,
			)
		}

		return p.processReceiverEth(
			cardanoSrcConfig,
			ethDestConfig,
			metadata.DestinationChainID,
			receiver,
			currencySrcID,
			destChainInfo.CurrencyTokenID,
			&totalTokensAmount,
		)
	}

	for _, receiver := range metadata.Transactions {
		receiverAddr := strings.Join(receiver.Address, "")

		if receiverAddr == destChainInfo.FeeAddress {
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
	totalTokensAmount.totalAmountCurrencyDst += destChainInfo.FeeAddrBridgingAmt
	totalTokensAmount.totalAmountCurrencySrc += metadata.BridgingFee

	receivers = append(receivers, cCore.BridgingRequestReceiver{
		DestinationAddress: destChainInfo.FeeAddress,
		Amount:             new(big.Int).SetUint64(destChainInfo.FeeAddrBridgingAmt),
		AmountWrapped:      new(big.Int).SetUint64(0),
		TokenId:            0,
	})

	claim := cCore.BridgingRequestClaim{
		ObservedTransactionHash:         tx.Hash,
		SourceChainId:                   chainIDConverter.ToChainIDNum(tx.OriginChainID),
		DestinationChainId:              chainIDConverter.ToChainIDNum(metadata.DestinationChainID),
		Receivers:                       receivers,
		NativeCurrencyAmountDestination: new(big.Int).SetUint64(totalTokensAmount.totalAmountCurrencyDst),
		NativeCurrencyAmountSource:      new(big.Int).SetUint64(totalTokensAmount.totalAmountCurrencySrc),
		WrappedTokenAmountSource:        new(big.Int).SetUint64(totalTokensAmount.totalAmountWrappedSrc),
		WrappedTokenAmountDestination:   new(big.Int).SetUint64(totalTokensAmount.totalAmountWrappedDst),
		RetryCounter:                    big.NewInt(int64(tx.BatchTryCount)),
	}

	claims.BridgingRequestClaims = append(claims.BridgingRequestClaims, claim)

	p.logger.Info("Added BridgingRequestClaim",
		"txHash", tx.Hash, "metadata", metadata, "claim", cCore.BridgingRequestClaimString(claim, chainIDConverter))

	return nil
}

func (p *BridgingRequestedProcessorSkylineImpl) validate(
	tx *core.CardanoTx, metadata *common.BridgingRequestMetadata, appConfig *cCore.AppConfig,
) error {
	cardanoSrcConfig, _ := cUtils.GetChainConfig(appConfig, tx.OriginChainID)
	if cardanoSrcConfig == nil {
		return fmt.Errorf("unsupported chain id found in tx. chain id: %v", tx.OriginChainID)
	}

	if err := p.preValidate(tx, appConfig); err != nil {
		return err
	}

	cardanoDestConfig, ethDestConfig := cUtils.GetChainConfig(appConfig, metadata.DestinationChainID)

	destChainInfo, err := cUtils.GetDestChainInfo(metadata.DestinationChainID, appConfig, cardanoDestConfig, ethDestConfig)
	if err != nil {
		return err
	}

	multisigUtxo, treasuryUtxoAmount, err := utils.ValidateTxOutputs(tx, appConfig, false, true)
	if err != nil {
		return err
	}

	if err := p.validateOperationAndReceiverLimits(metadata, cardanoSrcConfig, appConfig); err != nil {
		return err
	}

	currencySrcID, err := cardanoSrcConfig.GetCurrencyID()
	if err != nil {
		return err
	}

	receiverCtx := &receiverValidationCtxCardanoSrc{
		cardanoSrcConfig: cardanoSrcConfig,
		metadata:         metadata,
		ReceiverValidationContext: cCore.ReceiverValidationContext{
			CardanoDestConfig: cardanoDestConfig,
			EthDestConfig:     ethDestConfig,
			DestFeeAddress:    destChainInfo.FeeAddress,
			BridgingSettings:  &appConfig.BridgingSettings,
			AmountsSums:       make(map[uint16]*big.Int),
			CurrencySrcID:     currencySrcID,
			CurrencyDestID:    destChainInfo.CurrencyTokenID,
		},
	}

	for _, receiver := range metadata.Transactions {
		if err := p.validateReceiver(&receiver, receiverCtx); err != nil {
			return err
		}
	}

	if err := p.validateTokenAmounts(
		multisigUtxo, treasuryUtxoAmount, receiverCtx,
	); err != nil {
		return err
	}

	return nil
}

func (p *BridgingRequestedProcessorSkylineImpl) preValidate(
	tx *core.CardanoTx,
	appConfig *cCore.AppConfig,
) error {
	if err := p.refundRequestProcessor.HandleBridgingProcessorPreValidate(tx, appConfig); err != nil {
		return err
	}

	if err := utils.ValidateOutputsHaveUnknownTokens(tx, appConfig, false); err != nil {
		return err
	}

	return nil
}

func (p *BridgingRequestedProcessorSkylineImpl) validateOperationAndReceiverLimits(
	metadata *common.BridgingRequestMetadata,
	cardanoSrcConfig *cCore.CardanoChainConfig,
	appConfig *cCore.AppConfig,
) error {
	if metadata.OperationFee < cardanoSrcConfig.MinOperationFee {
		return fmt.Errorf("operation fee in metadata receivers is less than minimum: fee %d, minFee %d, metadata %v",
			metadata.OperationFee, cardanoSrcConfig.MinOperationFee, metadata)
	}

	if len(metadata.Transactions) > appConfig.BridgingSettings.MaxReceiversPerBridgingRequest {
		return fmt.Errorf("number of receivers in metadata greater than maximum allowed - no: %v, max: %v, metadata: %v",
			len(metadata.Transactions), appConfig.BridgingSettings.MaxReceiversPerBridgingRequest, metadata)
	}

	return nil
}

func (p *BridgingRequestedProcessorSkylineImpl) validateReceiver(
	receiver *sendtx.BridgingRequestMetadataTransaction,
	ctx *receiverValidationCtxCardanoSrc,
) error {
	receiverAddr := strings.Join(receiver.Address, "")

	if cUtils.NormalizeAddr(receiverAddr) == cUtils.NormalizeAddr(ctx.DestFeeAddress) {
		if ctx.CurrencySrcID != receiver.TokenID {
			return fmt.Errorf("fee receiver metadata invalid. metadata: %v, receiver: %v", ctx.metadata, receiver)
		}

		ctx.feeSum += receiver.Amount

		return nil
	}

	tokenPair, err := cUtils.GetTokenPair(
		ctx.cardanoSrcConfig.DestinationChains,
		ctx.cardanoSrcConfig.ChainID,
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
	ctx *receiverValidationCtxCardanoSrc,
	receiver *sendtx.BridgingRequestMetadataTransaction,
	tokenPair *common.TokenPair,
) error {
	receiverAddr := strings.Join(receiver.Address, "")

	if !cardanotx.IsValidOutputAddress(receiverAddr, ctx.CardanoDestConfig.NetworkID) {
		return fmt.Errorf(
			"found an invalid receiver addr in metadata. metadata: %v, receiver: %v", ctx.metadata, receiver)
	}

	// check min utxo when on destination
	if tokenPair.DestinationTokenID == ctx.CurrencyDestID {
		if receiver.Amount < ctx.CardanoDestConfig.UtxoMinAmount {
			return fmt.Errorf(
				"found an utxo value below minimum value in metadata receivers. metadata: %v, receiver: %v",
				ctx.metadata, receiver)
		}
	}

	if tokenPair.SourceTokenID == ctx.CurrencySrcID {
		// check min utxo when currency on source
		if receiver.Amount < ctx.cardanoSrcConfig.UtxoMinAmount {
			return fmt.Errorf(
				"found an utxo value below minimum value in metadata receivers. metadata: %v, receiver: %v",
				ctx.metadata, receiver)
		}
	}

	if tokenPair.SourceTokenID != ctx.CurrencySrcID &&
		tokenPair.DestinationTokenID != ctx.CurrencyDestID {
		// source token is not currency and is not wrapped token - it's colored coin on source
		if receiver.Amount < ctx.BridgingSettings.MinColCoinsAllowedToBridge {
			return fmt.Errorf(
				"receiver amount of token with ID %d too low: got %d, minimum allowed %d; metadata: %v, receiver: %v",
				receiver.TokenID,
				receiver.Amount,
				ctx.BridgingSettings.MinColCoinsAllowedToBridge,
				ctx.metadata,
				receiver,
			)
		}
	}

	if nativeTokensSum, ok := ctx.AmountsSums[tokenPair.SourceTokenID]; ok {
		nativeTokensSum.Add(nativeTokensSum, new(big.Int).SetUint64(receiver.Amount))
	} else {
		ctx.AmountsSums[tokenPair.SourceTokenID] = new(big.Int).SetUint64(receiver.Amount)
	}

	return nil
}

func (p *BridgingRequestedProcessorSkylineImpl) validateReceiverEth(
	ctx *receiverValidationCtxCardanoSrc,
	receiver *sendtx.BridgingRequestMetadataTransaction,
	tokenPair *common.TokenPair,
) error {
	receiverAddr := strings.Join(receiver.Address, "")

	if !goEthCommon.IsHexAddress(receiverAddr) {
		return fmt.Errorf(
			"found an invalid eth receiver addr in metadata. metadata: %v, receiver: %v",
			ctx.metadata, receiver)
	}

	if tokenPair.SourceTokenID == ctx.CurrencySrcID {
		// check min utxo when currency on source
		if receiver.Amount < ctx.cardanoSrcConfig.UtxoMinAmount {
			return fmt.Errorf(
				"found an utxo value below minimum value in metadata receivers. metadata: %v, receiver: %v",
				ctx.metadata, receiver)
		}
	} else if receiver.Amount < ctx.BridgingSettings.MinColCoinsAllowedToBridge {
		// check colored coin min amount
		return fmt.Errorf(
			"receiver amount of token with ID %d too low: got %d, minimum allowed %d; metadata: %v, receiver: %v",
			receiver.TokenID,
			receiver.Amount,
			ctx.BridgingSettings.MinColCoinsAllowedToBridge,
			ctx.metadata,
			receiver,
		)
	}

	if nativeTokensSum, ok := ctx.AmountsSums[tokenPair.SourceTokenID]; ok {
		nativeTokensSum.Add(nativeTokensSum, new(big.Int).SetUint64(receiver.Amount))
	} else {
		ctx.AmountsSums[tokenPair.SourceTokenID] = new(big.Int).SetUint64(receiver.Amount)
	}

	return nil
}

func (p *BridgingRequestedProcessorSkylineImpl) validateTokenAmounts(
	multisigUtxo *indexer.TxOutput, treasuryUtxoAmount uint64,
	receiverCtx *receiverValidationCtxCardanoSrc,
) error {
	cardanoSrcConfig := receiverCtx.cardanoSrcConfig
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

	// update fee amount if needed with sum of fee address receivers
	metadata.BridgingFee += receiverCtx.feeSum
	nativeCurrencySum.Add(nativeCurrencySum, new(big.Int).SetUint64(metadata.BridgingFee))

	minBridgingFee := cardanoSrcConfig.GetMinBridgingFee(
		len(receiverCtx.AmountsSums) > 0 || len(multisigUtxo.Tokens) > 0,
	)

	if metadata.BridgingFee < minBridgingFee {
		return fmt.Errorf("bridging fee in metadata receivers is less than minimum: fee %d, minFee %d, metadata %v",
			metadata.BridgingFee, minBridgingFee, metadata)
	}

	if treasuryUtxoAmount != metadata.OperationFee {
		return fmt.Errorf("treasury utxo amount %d does not match operation fee %d in metadata",
			treasuryUtxoAmount, metadata.OperationFee)
	}

	nativeTokensNamesInMetadata := make(map[string]struct{}, len(receiverCtx.AmountsSums))

	for tokenID, tokenAmount := range receiverCtx.AmountsSums {
		tokenName := cardanoSrcConfig.Tokens[tokenID].ChainSpecific

		nativeToken, err := cardanowallet.NewTokenWithFullNameTry(tokenName)
		if err != nil {
			return err
		}

		nativeTokensNamesInMetadata[nativeToken.String()] = struct{}{}

		multisigTokenAmount := new(big.Int).SetUint64(cardanotx.GetTokenAmount(multisigUtxo, nativeToken.String()))

		if tokenAmount.Cmp(multisigTokenAmount) != 0 {
			return fmt.Errorf("multisig native token with ID: %d amount mismatch: expected %v but got %v",
				tokenID, tokenAmount, multisigTokenAmount)
		}
	}

	for _, tokenAmount := range multisigUtxo.Tokens {
		if _, ok := nativeTokensNamesInMetadata[tokenAmount.TokenName()]; !ok {
			tokFullAmnt := cardanotx.GetTokenAmount(multisigUtxo, tokenAmount.TokenName())

			return fmt.Errorf("multisig native token: %s amount mismatch: expected 0 but got %v",
				tokenAmount.TokenName(), tokFullAmnt)
		}
	}

	var err error

	srcMinUtxo := cardanoSrcConfig.UtxoMinAmount
	if len(receiverCtx.AmountsSums) > 0 {
		srcMinUtxo, err = p.calculateMinUtxo(
			cardanoSrcConfig, multisigUtxo.Address, receiverCtx.AmountsSums)
		if err != nil {
			return fmt.Errorf("failed to calculate src minUtxo for chainID: %s. err: %w",
				cardanoSrcConfig.ChainID, err)
		}
	}

	minCurrency := srcMinUtxo + minBridgingFee
	if new(big.Int).SetUint64(minCurrency).Cmp(nativeCurrencySum) == 1 {
		return fmt.Errorf("sum of receiver amounts+fee is under the minimum allowed: min %v but got %v",
			minCurrency, nativeCurrencySum)
	}

	maxTokenAmt := receiverCtx.BridgingSettings.MaxTokenAmountAllowedToBridge
	if maxTokenAmt != nil && maxTokenAmt.Sign() > 0 {
		for tokenID, tokenAmount := range receiverCtx.AmountsSums {
			if tokenAmount.Cmp(maxTokenAmt) == 1 {
				return fmt.Errorf("sum of native token with ID %d: %v greater than maximum allowed: %v",
					tokenID, tokenAmount, maxTokenAmt)
			}
		}
	}

	if nativeCurrencySum.Cmp(new(big.Int).SetUint64(multisigUtxo.Amount)) != 0 {
		return fmt.Errorf(
			"multisig amount is not equal to sum of receiver amounts+fee: expected %v but got %v",
			multisigUtxo.Amount, nativeCurrencySum)
	}

	return nil
}

func (p *BridgingRequestedProcessorSkylineImpl) calculateMinUtxo(
	config *cCore.CardanoChainConfig, receiverAddr string, nativeTokensSum map[uint16]*big.Int,
) (uint64, error) {
	builder, err := cardanowallet.NewTxBuilder(cardanowallet.ResolveCardanoCliBinary(config.NetworkID))
	if err != nil {
		return 0, err
	}

	defer builder.Dispose()

	chainInfo, exists := p.chainInfos[config.ChainID]
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
	cardanoSrcConfig *cCore.CardanoChainConfig,
	cardanoDestConfig *cCore.CardanoChainConfig,
	receiver *sendtx.BridgingRequestMetadataTransaction,
	currencySrcID, currencyDestID uint16,
	totalTokensAmount *totalTokensAmount,
) (*cCore.BridgingRequestReceiver, error) {
	var (
		receiverCurrency  uint64
		receiverTokens    uint64
		wrappedTokensDest uint64
	)

	receiverAddr := strings.Join(receiver.Address, "")

	tokenPair, err := cUtils.GetTokenPair(
		cardanoSrcConfig.DestinationChains, cardanoSrcConfig.ChainID,
		cardanoDestConfig.ChainID, receiver.TokenID)
	if err != nil {
		return nil, err
	}

	if tokenPair.DestinationTokenID == currencyDestID {
		// currency on destination
		receiverCurrency = receiver.Amount

		if tokenPair.TrackDestinationToken {
			trackDestTokenAmount(totalTokensAmount, receiverCurrency, 0)
		}
	} else {
		nativeTokensSum := map[uint16]*big.Int{
			tokenPair.DestinationTokenID: new(big.Int).SetUint64(receiver.Amount),
		}

		dstMinUtxo, err := p.calculateMinUtxo(cardanoDestConfig, receiverAddr, nativeTokensSum)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate destination minUtxo for chainID: %s. err: %w",
				cardanoDestConfig.ChainID, err)
		}

		receiverCurrency = dstMinUtxo
		trackDestTokenAmount(totalTokensAmount, receiverCurrency, 0)

		receiverTokens = receiver.Amount

		// wrapped token on destination
		if cardanoDestConfig.Tokens[tokenPair.DestinationTokenID].IsWrappedCurrency {
			wrappedTokensDest = receiver.Amount
		}

		if tokenPair.TrackDestinationToken {
			trackDestTokenAmount(totalTokensAmount, 0, wrappedTokensDest)
		}
	}

	if tokenPair.TrackSourceToken {
		trackSourceTokenAmount(
			tokenPair.SourceTokenID,
			currencySrcID,
			receiver.Amount,
			cardanoSrcConfig,
			totalTokensAmount,
		)
	}

	return &cCore.BridgingRequestReceiver{
		DestinationAddress: receiverAddr,
		Amount:             new(big.Int).SetUint64(receiverCurrency),
		AmountWrapped:      new(big.Int).SetUint64(receiverTokens),
		TokenId:            tokenPair.DestinationTokenID,
	}, nil
}

func (p *BridgingRequestedProcessorSkylineImpl) processReceiverEth(
	cardanoSrcConfig *cCore.CardanoChainConfig,
	ethDestConfig *cCore.EthChainConfig,
	destinationChainID string,
	receiver *sendtx.BridgingRequestMetadataTransaction,
	currencySrcID, currencyDestID uint16,
	totalTokensAmount *totalTokensAmount,
) (*cCore.BridgingRequestReceiver, error) {
	receiverAddr := strings.Join(receiver.Address, "")

	tokenPair, err := cUtils.GetTokenPair(
		cardanoSrcConfig.DestinationChains, cardanoSrcConfig.ChainID,
		destinationChainID, receiver.TokenID)
	if err != nil {
		return nil, err
	}

	var amount, amountWrapped uint64

	// currency on destination
	if tokenPair.DestinationTokenID == currencyDestID {
		amount = receiver.Amount

		if tokenPair.TrackDestinationToken {
			trackDestTokenAmount(
				totalTokensAmount, receiver.Amount, 0,
			)
		}
	} else {
		amountWrapped = receiver.Amount

		// wrapped token on destination
		if tokenPair.TrackDestinationToken && ethDestConfig.Tokens[tokenPair.DestinationTokenID].IsWrappedCurrency {
			trackDestTokenAmount(
				totalTokensAmount, 0, receiver.Amount,
			)
		}
	}

	if tokenPair.TrackSourceToken {
		trackSourceTokenAmount(
			tokenPair.SourceTokenID,
			currencySrcID,
			receiver.Amount,
			cardanoSrcConfig,
			totalTokensAmount,
		)
	}

	return &cCore.BridgingRequestReceiver{
		DestinationAddress: receiverAddr,
		Amount:             new(big.Int).SetUint64(amount),
		AmountWrapped:      new(big.Int).SetUint64(amountWrapped),
		TokenId:            tokenPair.DestinationTokenID,
	}, nil
}

func trackDestTokenAmount(totalTokensAmount *totalTokensAmount, receiverCurrency uint64, receiverTokens uint64) {
	totalTokensAmount.totalAmountCurrencyDst += receiverCurrency
	totalTokensAmount.totalAmountWrappedDst += receiverTokens
}

func trackSourceTokenAmount(
	sourceTokenID uint16,
	currencySrcID uint16,
	receiverAmount uint64,
	cardanoSrcConfig *cCore.CardanoChainConfig,
	totalTokensAmount *totalTokensAmount,
) {
	if sourceTokenID == currencySrcID {
		// currency on source
		totalTokensAmount.totalAmountCurrencySrc += receiverAmount
	} else if cardanoSrcConfig.Tokens[sourceTokenID].IsWrappedCurrency {
		// source token is wrapped currency
		totalTokensAmount.totalAmountWrappedSrc += receiverAmount
	}
}

func unmarshalBridgingRequestMetadata(
	chainConfig *cCore.CardanoChainConfig, txMetadata []byte,
) (*common.BridgingRequestMetadata, error) {
	metadataBC, err := common.UnmarshalMetadata[common.BridgingRequestMetadataBC](
		common.MetadataEncodingTypeCbor, txMetadata)
	if err != nil {
		return nil, err
	}

	return mapBCMetadataToCurrent(chainConfig, metadataBC)
}

// backward compatible metadata version
func mapBCMetadataToCurrent(
	chainConfig *cCore.CardanoChainConfig, metadataBC *common.BridgingRequestMetadataBC,
) (*common.BridgingRequestMetadata, error) {
	txs := make([]sendtx.BridgingRequestMetadataTransaction, len(metadataBC.Transactions))

	for i, tx := range metadataBC.Transactions {
		var (
			err     error
			ok      bool
			tokenID = tx.TokenID
		)

		// Token should never be 0
		if tx.TokenID == 0 {
			if tx.IsNativeTokenOnSrc_Obsolete == 0 {
				tokenID, err = chainConfig.GetCurrencyID()
				if err != nil {
					return nil, err
				}
			} else {
				tokenID, ok = chainConfig.GetWrappedTokenID()
				if !ok {
					return nil, fmt.Errorf("wrapped currency not found in chain config")
				}
			}
		}

		txs[i] = sendtx.BridgingRequestMetadataTransaction{
			Address: tx.Address,
			Amount:  tx.Amount,
			TokenID: tokenID,
		}
	}

	return &common.BridgingRequestMetadata{
		BridgingTxType:     sendtx.BridgingRequestType(metadataBC.BridgingTxType),
		DestinationChainID: metadataBC.DestinationChainID,
		SenderAddr:         metadataBC.SenderAddr,
		Transactions:       txs,
		BridgingFee:        metadataBC.BridgingFee,
		OperationFee:       metadataBC.OperationFee,
	}, nil
}
