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

	chainInfos map[string]*chain.CardanoChainInfo
}

type receiverValidationContext struct {
	cardanoSrcConfig  *cCore.CardanoChainConfig
	cardanoDestConfig *cCore.CardanoChainConfig
	ethDestConfig     *cCore.EthChainConfig

	bridgingSettings *cCore.BridgingSettings
	destFeeAddress   string
	metadata         *common.BridgingRequestMetadata

	currencySrcID  uint16
	currencyDestID uint16

	nativeTokensSum map[uint16]*big.Int
	feeSum          uint64
}

type totalTokensAmount struct {
	totalAmountCurrencySrc uint64
	totalAmountWrappedSrc  uint64
	totalAmountCurrencyDst uint64
	totalAmountWrappedDst  uint64
}

func NewSkylineBridgingRequestedProcessor(
	refundRequestProcessor core.CardanoTxSuccessRefundProcessor,
	logger hclog.Logger, chainInfos map[string]*chain.CardanoChainInfo,
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
	metadata, err := common.UnmarshalMetadata[common.BridgingRequestMetadata](common.MetadataEncodingTypeCbor, tx.Metadata)
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
	cardanoSrcConfig, _ := cUtils.GetChainConfig(appConfig, tx.OriginChainID)
	cardanoDestConfig, ethDestConfig := cUtils.GetChainConfig(appConfig, metadata.DestinationChainID)

	destChainFeeAddress, feeCurrencyDst, err := p.getDestChainInfo(metadata, appConfig, cardanoDestConfig, ethDestConfig)
	if err != nil {
		return err
	}

	isCardanoDest := cardanoDestConfig != nil

	var (
		receivers         = make([]cCore.BridgingRequestReceiver, 0, len(metadata.Transactions))
		currencyDestID    = uint16(0)
		totalTokensAmount = totalTokensAmount{}
	)

	currencySrcID, err := cardanoSrcConfig.GetCurrencyID()
	if err != nil {
		return err
	}

	if isCardanoDest {
		currencyDestID, err = cardanoDestConfig.GetCurrencyID()
		if err != nil {
			return err
		}
	}

	processReceiver := func(
		receiver sendtx.BridgingRequestMetadataTransaction,
	) (*cCore.BridgingRequestReceiver, error) {
		if isCardanoDest {
			return p.processReceiverCardano(
				cardanoSrcConfig,
				cardanoDestConfig,
				receiver,
				currencySrcID,
				currencyDestID,
				&totalTokensAmount,
			)
		}

		return p.processReceiverEth(
			cardanoSrcConfig,
			ethDestConfig,
			metadata.DestinationChainID,
			receiver,
			currencySrcID,
			&totalTokensAmount,
		)
	}

	for _, receiver := range metadata.Transactions {
		receiverAddr := strings.Join(receiver.Address, "")

		if receiverAddr == destChainFeeAddress {
			// fee address will be added at the end
			continue
		}

		brReceiver, err := processReceiver(receiver)
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

	totalTokensAmount.totalAmountCurrencyDst += feeCurrencyDst
	totalTokensAmount.totalAmountCurrencySrc += metadata.BridgingFee + metadata.OperationFee

	receivers = append(receivers, cCore.BridgingRequestReceiver{
		DestinationAddress: destChainFeeAddress,
		Amount:             new(big.Int).SetUint64(feeCurrencyDst),
		AmountWrapped:      new(big.Int).SetUint64(0),
		TokenId:            0,
	})

	claim := cCore.BridgingRequestClaim{
		ObservedTransactionHash:         tx.Hash,
		SourceChainId:                   common.ToNumChainID(tx.OriginChainID),
		DestinationChainId:              common.ToNumChainID(metadata.DestinationChainID),
		Receivers:                       receivers,
		NativeCurrencyAmountDestination: new(big.Int).SetUint64(totalTokensAmount.totalAmountCurrencyDst),
		NativeCurrencyAmountSource:      new(big.Int).SetUint64(totalTokensAmount.totalAmountCurrencySrc),
		WrappedTokenAmountSource:        new(big.Int).SetUint64(totalTokensAmount.totalAmountWrappedSrc),
		WrappedTokenAmountDestination:   new(big.Int).SetUint64(totalTokensAmount.totalAmountWrappedDst),
		RetryCounter:                    big.NewInt(int64(tx.BatchTryCount)),
	}

	claims.BridgingRequestClaims = append(claims.BridgingRequestClaims, claim)

	p.logger.Info("Added BridgingRequestClaim",
		"txHash", tx.Hash, "metadata", metadata, "claim", cCore.BridgingRequestClaimString(claim))

	return nil
}

func (p *BridgingRequestedProcessorSkylineImpl) validate(
	tx *core.CardanoTx, metadata *common.BridgingRequestMetadata, appConfig *cCore.AppConfig,
) error {
	cardanoSrcConfig, _ := cUtils.GetChainConfig(appConfig, tx.OriginChainID)
	if cardanoSrcConfig == nil {
		return fmt.Errorf("unsupported chain id found in tx. chain id: %v", tx.OriginChainID)
	}

	if err := p.preValidate(tx, metadata, appConfig); err != nil {
		return err
	}

	cardanoDestConfig, ethDestConfig := cUtils.GetChainConfig(appConfig, metadata.DestinationChainID)

	destFeeAddress, _, err := p.getDestChainInfo(metadata, appConfig, cardanoDestConfig, ethDestConfig)
	if err != nil {
		return err
	}

	multisigUtxo, err := utils.ValidateTxOutputs(tx, appConfig, false)
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

	receiverCtx := &receiverValidationContext{
		cardanoSrcConfig:  cardanoSrcConfig,
		cardanoDestConfig: cardanoDestConfig,
		ethDestConfig:     ethDestConfig,
		destFeeAddress:    destFeeAddress,
		bridgingSettings:  &appConfig.BridgingSettings,
		metadata:          metadata,
		nativeTokensSum:   make(map[uint16]*big.Int),
		currencySrcID:     currencySrcID,
	}

	if cardanoDestConfig != nil {
		currencyDestID, err := cardanoDestConfig.GetCurrencyID()
		if err != nil {
			return err
		}

		receiverCtx.currencyDestID = currencyDestID
	}

	for _, receiver := range metadata.Transactions {
		if err := p.validateReceiver(&receiver, receiverCtx); err != nil {
			return err
		}
	}

	if err := p.validateTokenAmounts(
		multisigUtxo, receiverCtx,
	); err != nil {
		return err
	}

	return nil
}

func (p *BridgingRequestedProcessorSkylineImpl) preValidate(
	tx *core.CardanoTx,
	metadata *common.BridgingRequestMetadata,
	appConfig *cCore.AppConfig,
) error {
	if err := p.refundRequestProcessor.HandleBridgingProcessorPreValidate(tx, appConfig); err != nil {
		return err
	}

	if err := utils.ValidateOutputsHaveUnknownTokens(tx, appConfig); err != nil {
		return err
	}

	return nil
}

func (p *BridgingRequestedProcessorSkylineImpl) getDestChainInfo(
	metadata *common.BridgingRequestMetadata,
	appConfig *cCore.AppConfig,
	cardanoDestConfig *cCore.CardanoChainConfig,
	ethDestConfig *cCore.EthChainConfig,
) (string, uint64, error) {
	switch {
	case cardanoDestConfig != nil:
		return appConfig.GetFeeMultisigAddress(metadata.DestinationChainID), cardanoDestConfig.FeeAddrBridgingAmount, nil
	case ethDestConfig != nil:
		return common.EthZeroAddr, ethDestConfig.FeeAddrBridgingAmount, nil
	default:
		return "", 0, fmt.Errorf("destination chain not registered: %s", metadata.DestinationChainID)
	}
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
	ctx *receiverValidationContext,
) error {
	receiverAddr := strings.Join(receiver.Address, "")

	if receiverAddr == ctx.destFeeAddress {
		currencyID, err := ctx.cardanoSrcConfig.GetCurrencyID()
		if err != nil {
			return err
		}

		if currencyID != receiver.Token {
			return fmt.Errorf("fee receiver metadata invalid: %v", ctx.metadata)
		}

		ctx.feeSum += receiver.Amount

		return nil
	}

	tokenPair, err := cUtils.GetTokenPair(
		ctx.cardanoSrcConfig.DestinationChains,
		ctx.cardanoSrcConfig.ChainID,
		ctx.metadata.DestinationChainID,
		receiver.Token,
	)
	if err != nil {
		return err
	}

	if ctx.cardanoDestConfig != nil {
		return p.validateReceiverCardano(ctx, receiver, tokenPair)
	}

	return p.validateReceiverEth(ctx, receiver, tokenPair)
}

func (p *BridgingRequestedProcessorSkylineImpl) validateReceiverCardano(
	ctx *receiverValidationContext,
	receiver *sendtx.BridgingRequestMetadataTransaction,
	tokenPair *common.TokenPair,
) error {
	receiverAddr := strings.Join(receiver.Address, "")

	if !cardanotx.IsValidOutputAddress(receiverAddr, ctx.cardanoDestConfig.NetworkID) {
		return fmt.Errorf("found an invalid receiver addr in metadata: %v", ctx.metadata)
	}

	// check min utxo when on destination
	if tokenPair.DestinationTokenID == ctx.currencyDestID {
		if receiver.Amount < ctx.cardanoDestConfig.UtxoMinAmount {
			return fmt.Errorf("found an utxo value below minimum value in metadata receivers: %v", ctx.metadata)
		}
	} else if tokenPair.SourceTokenID == ctx.currencySrcID {
		// check min utxo when currency on source
		if receiver.Amount < ctx.cardanoSrcConfig.UtxoMinAmount {
			return fmt.Errorf("found an utxo value below minimum value in metadata receivers: %v", ctx.metadata)
		}
		// source token is not currency and is not wrapped token - it's colored coin on source
		// TODO - check this one more time. Also check whether we should have MinColCoinsAllowedToBridge
	} else if receiver.Amount < ctx.bridgingSettings.MinColCoinsAllowedToBridge {
		return fmt.Errorf(
			"colored coin receiver amount too low for Colored Coin %s: got %d, minimum allowed %d; metadata: %v",
			receiver.Token,
			receiver.Amount,
			ctx.bridgingSettings.MinColCoinsAllowedToBridge,
			ctx.metadata,
		)
	}

	if nativeTokensSum, ok := ctx.nativeTokensSum[tokenPair.SourceTokenID]; ok {
		nativeTokensSum.Add(nativeTokensSum, new(big.Int).SetUint64(receiver.Amount))
	} else {
		ctx.nativeTokensSum[tokenPair.SourceTokenID] = new(big.Int).SetUint64(receiver.Amount)
	}

	return nil
}

func (p *BridgingRequestedProcessorSkylineImpl) validateReceiverEth(
	ctx *receiverValidationContext,
	receiver *sendtx.BridgingRequestMetadataTransaction,
	tokenPair *common.TokenPair,
) error {
	receiverAddr := strings.Join(receiver.Address, "")

	if !goEthCommon.IsHexAddress(receiverAddr) {
		return fmt.Errorf("found an invalid eth receiver addr in metadata: %v", ctx.metadata)
	}

	if tokenPair.SourceTokenID == ctx.currencySrcID {
		// check min utxo when currency on source
		if receiver.Amount < ctx.cardanoSrcConfig.UtxoMinAmount {
			return fmt.Errorf("found an utxo value below minimum value in metadata receivers: %v", ctx.metadata)
		}
	} else if receiver.Amount < ctx.bridgingSettings.MinColCoinsAllowedToBridge {
		// check colored coin min amount
		return fmt.Errorf(
			"colored coin receiver amount too low for Colored Coin %s: got %d, minimum allowed %d; metadata: %v",
			receiver.Token,
			receiver.Amount,
			ctx.bridgingSettings.MinColCoinsAllowedToBridge,
			ctx.metadata,
		)
	}

	if nativeTokensSum, ok := ctx.nativeTokensSum[tokenPair.SourceTokenID]; ok {
		nativeTokensSum.Add(nativeTokensSum, new(big.Int).SetUint64(receiver.Amount))
	} else {
		ctx.nativeTokensSum[tokenPair.SourceTokenID] = new(big.Int).SetUint64(receiver.Amount)
	}

	return nil
}

func (p *BridgingRequestedProcessorSkylineImpl) validateTokenAmounts(
	multisigUtxo *indexer.TxOutput,
	receiverCtx *receiverValidationContext,
) error {
	cardanoSrcConfig := receiverCtx.cardanoSrcConfig
	metadata := receiverCtx.metadata

	nativeCurrencySum, ok := receiverCtx.nativeTokensSum[receiverCtx.currencySrcID]
	if !ok {
		nativeCurrencySum = new(big.Int).SetInt64(0)
	}

	// Remove currency entry from the map
	delete(receiverCtx.nativeTokensSum, receiverCtx.currencySrcID)

	if receiverCtx.bridgingSettings.MaxAmountAllowedToBridge != nil &&
		receiverCtx.bridgingSettings.MaxAmountAllowedToBridge.Sign() > 0 &&
		nativeCurrencySum.Cmp(receiverCtx.bridgingSettings.MaxAmountAllowedToBridge) == 1 {
		return fmt.Errorf("sum of receiver amounts: %v greater than maximum allowed: %v",
			nativeCurrencySum, receiverCtx.bridgingSettings.MaxAmountAllowedToBridge)
	}

	// update fee amount if needed with sum of fee address receivers
	metadata.BridgingFee += receiverCtx.feeSum
	nativeCurrencySum.Add(nativeCurrencySum, new(big.Int).SetUint64(metadata.BridgingFee))
	nativeCurrencySum.Add(nativeCurrencySum, new(big.Int).SetUint64(metadata.OperationFee))

	minBridgingFee := cardanoSrcConfig.GetMinBridgingFee(
		len(receiverCtx.nativeTokensSum) > 0 || len(multisigUtxo.Tokens) > 0,
	)

	if metadata.BridgingFee < minBridgingFee {
		return fmt.Errorf("bridging fee in metadata receivers is less than minimum: fee %d, minFee %d, metadata %v",
			metadata.BridgingFee, minBridgingFee, metadata)
	}

	// if there is at least one native token on source transfer or multi sig has tokens
	// -> native token on source should be defined
	for tokenID, tokenAmount := range receiverCtx.nativeTokensSum {
		tokenName := cardanoSrcConfig.Tokens[tokenID].ChainSpecific

		nativeToken, err := cardanotx.GetNativeTokenFromName(tokenName)
		if err != nil {
			return err
		}

		multisigWrappedTokenAmount := new(big.Int).SetUint64(cardanotx.GetTokenAmount(multisigUtxo, nativeToken.String()))

		if tokenAmount.Cmp(multisigWrappedTokenAmount) != 0 {
			return fmt.Errorf("multisig native token %s amount mismatch: expected %v but got %v",
				nativeToken.String(), tokenAmount, multisigWrappedTokenAmount)
		}
	}

	var err error

	srcMinUtxo := cardanoSrcConfig.UtxoMinAmount
	if len(receiverCtx.nativeTokensSum) > 0 {
		srcMinUtxo, err = p.calculateMinUtxo(
			cardanoSrcConfig, multisigUtxo.Address, receiverCtx.nativeTokensSum)
		if err != nil {
			return fmt.Errorf("failed to calculate src minUtxo for chainID: %s. err: %w",
				cardanoSrcConfig.ChainID, err)
		}
	}

	minCurrency := srcMinUtxo + minBridgingFee
	if new(big.Int).SetUint64(minCurrency).Cmp(nativeCurrencySum) == 1 {
		return fmt.Errorf("sum of receiver amounts + fee is under the minimum allowed: min %v but got %v",
			minCurrency, nativeCurrencySum)
	}

	maxTokenAmt := receiverCtx.bridgingSettings.MaxTokenAmountAllowedToBridge
	if maxTokenAmt != nil && maxTokenAmt.Sign() > 0 {
		for tokenName, tokenAmount := range receiverCtx.nativeTokensSum {
			if tokenAmount.Cmp(maxTokenAmt) == 1 {
				return fmt.Errorf("sum of native token %s: %v greater than maximum allowed: %v",
					tokenName, tokenAmount, maxTokenAmt)
			}
		}
	}

	if nativeCurrencySum.Cmp(new(big.Int).SetUint64(multisigUtxo.Amount)) != 0 {
		return fmt.Errorf("multisig amount is not equal to sum of receiver amounts + fee: expected %v but got %v",
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

		nativeToken, err := cardanotx.GetNativeTokenFromName(tokenName)
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
	receiver sendtx.BridgingRequestMetadataTransaction,
	currencySrcID, currencyDestID uint16,
	totalTokensAmount *totalTokensAmount,
) (*cCore.BridgingRequestReceiver, error) {
	var (
		receiverCurrency  uint64
		receiverTokens    uint64
		wrappedTokensDest uint64
	)

	receiverAddr := strings.Join(receiver.Address, "")

	tokenPair, err := cUtils.GetTokenPair(cardanoSrcConfig.DestinationChains, cardanoSrcConfig.ChainID, cardanoDestConfig.ChainID, receiver.Token)
	if err != nil {
		return nil, err
	}

	if tokenPair.DestinationTokenID == currencyDestID {
		// currency on destination
		receiverCurrency = receiver.Amount
	} else {
		nativeTokensSum := map[uint16]*big.Int{
			tokenPair.DestinationTokenID: new(big.Int).SetUint64(receiver.Amount),
		}

		dstMinUtxo, err := p.calculateMinUtxo(
			cardanoDestConfig, receiverAddr, nativeTokensSum)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate destination minUtxo for chainID: %s. err: %w",
				cardanoDestConfig.ChainID, err)
		}

		receiverCurrency = dstMinUtxo
		receiverTokens = receiver.Amount

		// wrapped token on destination
		if cardanoDestConfig.Tokens[tokenPair.DestinationTokenID].IsWrappedCurrency {
			wrappedTokensDest = receiver.Amount
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

	if tokenPair.TrackDestinationToken {
		trackDestTokenAmount(totalTokensAmount, receiverCurrency, wrappedTokensDest)
	}

	return &cCore.BridgingRequestReceiver{
		DestinationAddress: receiverAddr,
		Amount:             new(big.Int).SetUint64(receiverCurrency),
		AmountWrapped:      new(big.Int).SetUint64(receiverTokens),
		TokenId:            receiver.Token,
	}, nil
}

func (p *BridgingRequestedProcessorSkylineImpl) processReceiverEth(
	cardanoSrcConfig *cCore.CardanoChainConfig,
	ethDestConfig *cCore.EthChainConfig,
	destinationChainID string,
	receiver sendtx.BridgingRequestMetadataTransaction,
	currencySrcID uint16,
	totalTokensAmount *totalTokensAmount,
) (*cCore.BridgingRequestReceiver, error) {
	receiverAddr := strings.Join(receiver.Address, "")

	tokenPair, err := cUtils.GetTokenPair(cardanoSrcConfig.DestinationChains, cardanoSrcConfig.ChainID, destinationChainID, receiver.Token)
	if err != nil {
		return nil, err
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

	// wrapped token on destination
	if tokenPair.TrackDestinationToken && ethDestConfig.Tokens[tokenPair.DestinationTokenID].IsWrappedCurrency {
		trackDestTokenAmount(totalTokensAmount, 0, receiver.Amount)
	}

	return &cCore.BridgingRequestReceiver{
		DestinationAddress: receiverAddr,
		Amount:             new(big.Int).SetUint64(0),
		AmountWrapped:      new(big.Int).SetUint64(receiver.Amount),
		TokenId:            receiver.Token,
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
