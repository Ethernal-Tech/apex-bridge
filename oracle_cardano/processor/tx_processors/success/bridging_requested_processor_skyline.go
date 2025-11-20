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
	var (
		feeCurrencyDst        uint64
		destChainFeeAddress   string
		coloredCoinsIDToToken map[uint16]*cardanowallet.Token
		err                   error

		totalAmountCurrencySrc = uint64(0)
		totalAmountCurrencyDst = uint64(0)
		totalAmountWrappedSrc  = uint64(0)
		totalAmountWrappedDst  = uint64(0)

		isCardanoDest = false
	)

	cardanoDestConfig, ethDestConfig := cUtils.GetChainConfig(appConfig, metadata.DestinationChainID)

	switch {
	case cardanoDestConfig != nil:
		isCardanoDest = true
		destChainFeeAddress = appConfig.GetFeeMultisigAddress(metadata.DestinationChainID)
		feeCurrencyDst = cardanoDestConfig.FeeAddrBridgingAmount

		coloredCoinsIDToToken, err = mapColoredCoinsToNativeTokens(cardanoDestConfig.ColoredCoins)
		if err != nil {
			return fmt.Errorf(
				"failed to map colored coins to native tokens for destination chain %s: %w",
				metadata.DestinationChainID,
				err,
			)
		}
	case ethDestConfig != nil:
		destChainFeeAddress = common.EthZeroAddr
		feeCurrencyDst = ethDestConfig.FeeAddrBridgingAmount
	default:
		return fmt.Errorf("added BridgingRequestClaim not supported chain %s", metadata.DestinationChainID)
	}

	var (
		receivers = make([]cCore.BridgingRequestReceiver, 0, len(metadata.Transactions))
	)

	for _, receiver := range metadata.Transactions {
		receiverAddr := strings.Join(receiver.Address, "")

		if receiverAddr == destChainFeeAddress {
			// fee address will be added at the end
			continue
		}

		var (
			brReceiver                *cCore.BridgingRequestReceiver
			receiverAmountCurrencyDst uint64
			receiverAmountWrappedDst  uint64
		)

		if isCardanoDest {
			brReceiver, receiverAmountCurrencyDst, receiverAmountWrappedDst, err = p.processReceiverCardano(
				cardanoDestConfig,
				tx.OriginChainID,
				receiver,
				coloredCoinsIDToToken,
				&totalAmountCurrencySrc,
				&totalAmountWrappedSrc,
			)
			if err != nil {
				return fmt.Errorf(
					"failed to process Cardano receiver (chain %s, receiver address: %v): %w",
					cardanoDestConfig.ChainID,
					receiver.Address,
					err,
				)
			}

			receivers = append(receivers, *brReceiver)

			totalAmountCurrencyDst += receiverAmountCurrencyDst
			totalAmountWrappedDst += receiverAmountWrappedDst
		} else {
			receivers = append(receivers, cCore.BridgingRequestReceiver{
				DestinationAddress: receiverAddr,
				Amount:             new(big.Int).SetUint64(0),
				AmountWrapped:      new(big.Int).SetUint64(receiver.Amount),
				ColoredCoinId:      receiver.ColoredCoinID,
			})
		}
	}

	if isCardanoDest {
		totalAmountCurrencyDst += feeCurrencyDst
		totalAmountCurrencySrc += metadata.BridgingFee + metadata.OperationFee
	}

	receivers = append(receivers, cCore.BridgingRequestReceiver{
		DestinationAddress: destChainFeeAddress,
		Amount:             new(big.Int).SetUint64(feeCurrencyDst),
		AmountWrapped:      new(big.Int).SetUint64(0),
		ColoredCoinId:      0,
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

type receiverValidationContext struct {
	CardanoSrcConfig  *cCore.CardanoChainConfig
	CardanoDestConfig *cCore.CardanoChainConfig
	EthDestConfig     *cCore.EthChainConfig
	BridgingSettings  *cCore.BridgingSettings
	DestFeeAddress    string
	Metadata          *common.BridgingRequestMetadata

	NativeCurrencySum *big.Int
	WrappedTokenSum   *big.Int
	ColoredCoinsSum   map[uint16]*big.Int
	ColoredCoinsToken map[uint16]*cardanowallet.Token

	HasNativeTokenOnSource bool
	HasCurrencyOnSource    bool
	FeeSum                 uint64
	WrappedColCoin         uint16
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

	destFeeAddress, isCardanoDest, err := p.getDestChainInfo(metadata, appConfig, cardanoDestConfig, ethDestConfig)
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

	receiverCtx := &receiverValidationContext{
		CardanoSrcConfig:  cardanoSrcConfig,
		CardanoDestConfig: cardanoDestConfig,
		EthDestConfig:     ethDestConfig,
		DestFeeAddress:    destFeeAddress,
		BridgingSettings:  &appConfig.BridgingSettings,
		Metadata:          metadata,
		NativeCurrencySum: big.NewInt(0),
		WrappedTokenSum:   big.NewInt(0),
		ColoredCoinsSum:   make(map[uint16]*big.Int),
		ColoredCoinsToken: make(map[uint16]*cardanowallet.Token),
	}

	for _, receiver := range metadata.Transactions {
		if err := p.validateReceiver(receiver, receiverCtx); err != nil {
			return err
		}
	}

	coloredCoinsIDToToken, err := mapColoredCoinsToNativeTokens(cardanoSrcConfig.ColoredCoins)
	if err != nil {
		return fmt.Errorf(
			"failed to map colored coins to native tokens for source chain %s: %w",
			cardanoSrcConfig.ChainID,
			err,
		)
	}

	if err := p.validateTokenAmounts(
		multisigUtxo, receiverCtx, isCardanoDest, tx.OriginChainID, coloredCoinsIDToToken,
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

	if err := utils.IsTxDirectionAllowed(appConfig, tx.OriginChainID, metadata); err != nil {
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
) (destFeeAddress string, isCardanoDest bool, err error) {
	switch {
	case cardanoDestConfig != nil:
		return appConfig.GetFeeMultisigAddress(metadata.DestinationChainID), true, nil
	case ethDestConfig != nil:
		return common.EthZeroAddr, false, nil
	default:
		return "", false, fmt.Errorf("destination chain not registered: %s", metadata.DestinationChainID)
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
	receiver sendtx.BridgingRequestMetadataTransaction,
	ctx *receiverValidationContext,
) error {
	receiverAddr := strings.Join(receiver.Address, "")

	if ctx.CardanoDestConfig != nil && !cardanotx.IsValidOutputAddress(receiverAddr, ctx.CardanoDestConfig.NetworkID) {
		return fmt.Errorf("found an invalid receiver addr in metadata: %v", ctx.Metadata)
	}

	if ctx.EthDestConfig != nil && !goEthCommon.IsHexAddress(receiverAddr) {
		return fmt.Errorf("found an invalid receiver addr in metadata: %v", ctx.Metadata)
	}

	if receiverAddr == ctx.DestFeeAddress {
		if receiver.BridgingType == sendtx.BridgingTypeWrappedTokenOnSource ||
			receiver.BridgingType == sendtx.BridgingTypeColoredCoinOnSource {
			return fmt.Errorf("fee receiver metadata invalid: %v", ctx.Metadata)
		}

		ctx.FeeSum += receiver.Amount

		return nil
	}

	if ctx.CardanoDestConfig != nil {
		return p.validateReceiverCardano(receiver, ctx)
	}

	return p.validateReceiverEth(receiver, ctx)
}

func (p *BridgingRequestedProcessorSkylineImpl) validateReceiverCardano(
	receiver sendtx.BridgingRequestMetadataTransaction,
	ctx *receiverValidationContext,
) error {
	switch receiver.BridgingType {
	case sendtx.BridgingTypeWrappedTokenOnSource:
		ctx.HasNativeTokenOnSource = true

		// amount_to_bridge must be >= minUtxoAmount on destination
		if receiver.Amount < ctx.CardanoDestConfig.UtxoMinAmount {
			return fmt.Errorf("found an utxo value below minimum value in metadata receivers: %v", ctx.Metadata)
		}

		if receiver.ColoredCoinID != 0 {
			return fmt.Errorf(
				"wrapped-token receiver on Cardano destination must have ColoredCoinId = 0 (got %d): %v",
				receiver.ColoredCoinID, ctx.Metadata,
			)
		}

		ctx.WrappedTokenSum.Add(ctx.WrappedTokenSum, new(big.Int).SetUint64(receiver.Amount))
	case sendtx.BridgingTypeColoredCoinOnSource:
		ctx.HasNativeTokenOnSource = true

		if receiver.ColoredCoinID == 0 {
			return fmt.Errorf("colored coin receiver must have a non-zero ColoredCoinId: %v", ctx.Metadata)
		}

		if receiver.Amount < ctx.BridgingSettings.MinColCoinsAllowedToBridge {
			return fmt.Errorf(
				"colored coin receiver amount too low for ColoredCoinId %d: got %d, minimum allowed %d; metadata: %v",
				receiver.ColoredCoinID,
				receiver.Amount,
				ctx.BridgingSettings.MinColCoinsAllowedToBridge,
				ctx.Metadata,
			)
		}

		if colCoinSum, ok := ctx.ColoredCoinsSum[receiver.ColoredCoinID]; ok {
			colCoinSum.Add(colCoinSum, new(big.Int).SetUint64(receiver.Amount))
		} else {
			ctx.ColoredCoinsSum[receiver.ColoredCoinID] = new(big.Int).SetUint64(receiver.Amount)
		}
	default: // currency on source
		ctx.HasCurrencyOnSource = true

		// amount_to_bridge must be >= minUtxoAmount on source
		if receiver.Amount < ctx.CardanoSrcConfig.UtxoMinAmount {
			return fmt.Errorf("found an utxo value below minimum value in metadata receivers: %v", ctx.Metadata)
		}

		ctx.NativeCurrencySum.Add(ctx.NativeCurrencySum, new(big.Int).SetUint64(receiver.Amount))
	}

	return nil
}

func (p *BridgingRequestedProcessorSkylineImpl) validateReceiverEth(
	receiver sendtx.BridgingRequestMetadataTransaction,
	ctx *receiverValidationContext,
) error {
	switch receiver.BridgingType {
	case sendtx.BridgingTypeWrappedTokenOnSource:
		ctx.HasNativeTokenOnSource = true

		if receiver.ColoredCoinID == 0 {
			return fmt.Errorf(
				"wrapped-token receiver on Eth destination must have a non-zero ColoredCoinId: %v", ctx.Metadata,
			)
		}

		if ctx.WrappedColCoin == 0 {
			ctx.WrappedColCoin = receiver.ColoredCoinID
		} else if ctx.WrappedColCoin != receiver.ColoredCoinID {
			return fmt.Errorf("inconsistent ColoredCoinID for wrapped tokens: expected %d, got %d; metadata: %v",
				ctx.WrappedColCoin, receiver.ColoredCoinID, ctx.Metadata)
		}

		ctx.WrappedTokenSum.Add(ctx.WrappedTokenSum, new(big.Int).SetUint64(receiver.Amount))
	case sendtx.BridgingTypeColoredCoinOnSource:
		ctx.HasNativeTokenOnSource = true

		if receiver.ColoredCoinID == 0 {
			return fmt.Errorf("colored coin receiver must have a non-zero ColoredCoinId: %v", ctx.Metadata)
		}

		if receiver.Amount < ctx.BridgingSettings.MinColCoinsAllowedToBridge {
			return fmt.Errorf(
				"colored coin receiver amount too low for ColoredCoinId %d: got %d, minimum allowed %d; metadata: %v",
				receiver.ColoredCoinID,
				receiver.Amount,
				ctx.BridgingSettings.MinColCoinsAllowedToBridge,
				ctx.Metadata,
			)
		}

		if colCoinSum, ok := ctx.ColoredCoinsSum[receiver.ColoredCoinID]; ok {
			colCoinSum.Add(colCoinSum, new(big.Int).SetUint64(receiver.Amount))
		} else {
			ctx.ColoredCoinsSum[receiver.ColoredCoinID] = new(big.Int).SetUint64(receiver.Amount)
		}
	default: // currency on source
		return fmt.Errorf(
			"currency bridging to Eth chain is not supported for destination chain %s (metadata: %v)",
			ctx.Metadata.DestinationChainID,
			ctx.Metadata,
		)
	}

	return nil
}

func (p *BridgingRequestedProcessorSkylineImpl) validateTokenAmounts(
	multisigUtxo *indexer.TxOutput,
	receiverCtx *receiverValidationContext,
	isCardanoDest bool,
	originChainID string,
	coloredCoinsIDToToken map[uint16]*cardanowallet.Token,
) error {
	cardanoSrcConfig := receiverCtx.CardanoSrcConfig
	metadata := receiverCtx.Metadata

	if receiverCtx.BridgingSettings.MaxAmountAllowedToBridge != nil &&
		receiverCtx.BridgingSettings.MaxAmountAllowedToBridge.Sign() > 0 &&
		receiverCtx.NativeCurrencySum.Cmp(receiverCtx.BridgingSettings.MaxAmountAllowedToBridge) == 1 {
		return fmt.Errorf("sum of receiver amounts + fee: %v greater than maximum allowed: %v",
			receiverCtx.NativeCurrencySum, receiverCtx.BridgingSettings.MaxAmountAllowedToBridge)
	}

	// update fee amount if needed with sum of fee address receivers
	metadata.BridgingFee += receiverCtx.FeeSum
	receiverCtx.NativeCurrencySum.Add(receiverCtx.NativeCurrencySum, new(big.Int).SetUint64(metadata.BridgingFee))
	receiverCtx.NativeCurrencySum.Add(receiverCtx.NativeCurrencySum, new(big.Int).SetUint64(metadata.OperationFee))

	minBridgingFee := cardanoSrcConfig.GetMinBridgingFee(
		receiverCtx.HasNativeTokenOnSource || len(multisigUtxo.Tokens) > 0,
	)

	if metadata.BridgingFee < minBridgingFee {
		return fmt.Errorf("bridging fee in metadata receivers is less than minimum: fee %d, minFee %d, metadata %v",
			metadata.BridgingFee, minBridgingFee, metadata)
	}

	// if there is at least one native token on source transfer or multi sig has tokens
	// -> native token on source should be defined
	if receiverCtx.HasNativeTokenOnSource {
		var (
			nativeToken cardanowallet.Token
			err         error
		)

		if isCardanoDest {
			nativeToken, err = cardanoSrcConfig.GetNativeToken(metadata.DestinationChainID)
			if err != nil {
				return err
			}
		} else {
			nativeToken, err = cardanotx.GetNativeTokenFromName(cardanoSrcConfig.ColoredCoins[receiverCtx.WrappedColCoin])
			if err != nil {
				return err
			}
		}

		multisigWrappedTokenAmount := new(big.Int).SetUint64(cardanotx.GetTokenAmount(multisigUtxo, nativeToken.String()))

		if receiverCtx.WrappedTokenSum.Cmp(multisigWrappedTokenAmount) != 0 {
			return fmt.Errorf("multisig wrapped token is not equal to receiver wrapped token amount: expected %v but got %v",
				multisigWrappedTokenAmount, receiverCtx.WrappedTokenSum)
		}
	}

	for ccID := range cardanoSrcConfig.ColoredCoins {
		ccToken := coloredCoinsIDToToken[ccID]
		multisigColCoinTokenAmount := new(big.Int).SetUint64(cardanotx.GetTokenAmount(multisigUtxo, ccToken.String()))

		ccAmountSum, exists := receiverCtx.ColoredCoinsSum[ccID]

		// Case 1: receiver does NOT include this colored coin
		if !exists {
			if multisigColCoinTokenAmount.Sign() > 0 {
				return fmt.Errorf(
					"colored coin %d amount mismatch: expected %v but multisig has %v",
					ccID, 0, multisigColCoinTokenAmount,
				)
			}

			continue
		}

		// Case 2: receiver includes colored coin - amounts must match
		if ccAmountSum.Cmp(multisigColCoinTokenAmount) != 0 {
			return fmt.Errorf(
				"colored coin %d amount mismatch: expected %v but got %v",
				ccID, ccAmountSum, multisigColCoinTokenAmount,
			)
		}
	}

	// if there is at least one currency on source transfer -> native token on destination should be defined
	if receiverCtx.HasCurrencyOnSource {
		if _, err := receiverCtx.CardanoDestConfig.GetNativeToken(originChainID); err != nil {
			return err
		}
	}

	var err error

	srcMinUtxo := cardanoSrcConfig.UtxoMinAmount
	if receiverCtx.WrappedTokenSum.Sign() > 0 || len(receiverCtx.ColoredCoinsSum) > 0 {
		srcMinUtxo, err = p.calculateMinUtxo(
			cardanoSrcConfig, metadata.DestinationChainID, multisigUtxo.Address, receiverCtx.WrappedTokenSum.Uint64(),
			receiverCtx.ColoredCoinsSum, coloredCoinsIDToToken)
		if err != nil {
			return fmt.Errorf("failed to calculate src minUtxo for chainID: %s. err: %w",
				cardanoSrcConfig.ChainID, err)
		}
	}

	minCurrency := srcMinUtxo + minBridgingFee
	if new(big.Int).SetUint64(minCurrency).Cmp(receiverCtx.NativeCurrencySum) == 1 {
		return fmt.Errorf("sum of receiver amounts + fee is under the minimum allowed: min %v but got %v",
			minCurrency, receiverCtx.NativeCurrencySum)
	}

	maxTokenAmt := receiverCtx.BridgingSettings.MaxTokenAmountAllowedToBridge
	if maxTokenAmt != nil && maxTokenAmt.Sign() > 0 {
		if receiverCtx.WrappedTokenSum.Cmp(maxTokenAmt) == 1 {
			return fmt.Errorf("sum of wrapped token: %v greater than maximum allowed: %v",
				receiverCtx.WrappedTokenSum, maxTokenAmt)
		}

		for coloredCoinID, coloredCoinAmount := range receiverCtx.ColoredCoinsSum {
			if coloredCoinAmount.Cmp(maxTokenAmt) == 1 {
				return fmt.Errorf("sum of colored token %d: %v greater than maximum allowed: %v",
					coloredCoinID, coloredCoinAmount, maxTokenAmt)
			}
		}
	}

	if receiverCtx.NativeCurrencySum.Cmp(new(big.Int).SetUint64(multisigUtxo.Amount)) != 0 {
		return fmt.Errorf("multisig amount is not equal to sum of receiver amounts + fee: expected %v but got %v",
			multisigUtxo.Amount, receiverCtx.NativeCurrencySum)
	}

	return nil
}

func (p *BridgingRequestedProcessorSkylineImpl) calculateMinUtxo(
	config *cCore.CardanoChainConfig, destinationChainID, receiverAddr string, wrappedAmount uint64,
	coloredCoinsAmountSum map[uint16]*big.Int, coloredCoinTokens map[uint16]*cardanowallet.Token,
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

	tokenAmounts := make([]cardanowallet.TokenAmount, 0, len(coloredCoinsAmountSum)+1)
	for coloredCoinID, coloredCoinAmount := range coloredCoinsAmountSum {
		tokenAmounts = append(
			tokenAmounts,
			cardanowallet.NewTokenAmount(*coloredCoinTokens[coloredCoinID], coloredCoinAmount.Uint64()),
		)
	}

	if wrappedAmount > 0 {
		nativeToken, err := config.GetNativeToken(destinationChainID)
		if err != nil {
			return 0, err
		}

		// Build full list: native wrapped token + colored coins
		tokenAmounts = append(tokenAmounts, cardanowallet.NewTokenAmount(nativeToken, wrappedAmount))
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

// mapColoredCoinsToNativeTokens returns a map from ColoredCoinID to the corresponding native token.
func mapColoredCoinsToNativeTokens(coloredCoins cardanotx.ColoredCoins) (map[uint16]*cardanowallet.Token, error) {
	coloredCoinsIDToToken := make(map[uint16]*cardanowallet.Token, len(coloredCoins))

	for ccID, ccName := range coloredCoins {
		ccToken, err := cardanotx.GetNativeTokenFromName(ccName)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve native token for colored coin %s: %w", ccName, err)
		}

		coloredCoinsIDToToken[ccID] = &ccToken
	}

	return coloredCoinsIDToToken, nil
}

func (p *BridgingRequestedProcessorSkylineImpl) processReceiverCardano(
	cardanoDestConfig *cCore.CardanoChainConfig,
	originChainID string,
	receiver sendtx.BridgingRequestMetadataTransaction,
	coloredCoinsIDToToken map[uint16]*cardanowallet.Token,
	totalAmountCurrencySrc, totalAmountWrappedSrc *uint64,
) (*cCore.BridgingRequestReceiver, uint64, uint64, error) {
	var (
		receiverAmountCurrencyDst    uint64
		receiverAmountWrappedDst     uint64
		receiverAmountNativeTokenDst uint64
		coloredCoinID                uint16
	)

	receiverAddr := strings.Join(receiver.Address, "")

	switch receiver.BridgingType {
	case sendtx.BridgingTypeWrappedTokenOnSource:
		// receiverAmount represents the amount of native currency that is bridged to the receiver.
		// receiver.Amount of native tokens on the source will be converted to the same amount of native currency on
		// the destination.
		// totalAmountCurrencySrc stays the same
		*totalAmountWrappedSrc += receiver.Amount

		// receiverAmountWrappedDst stays the same
		receiverAmountCurrencyDst = receiver.Amount
		coloredCoinID = 0
	case sendtx.BridgingTypeColoredCoinOnSource:
		coloredCoinsAmountSum := map[uint16]*big.Int{
			receiver.ColoredCoinID: new(big.Int).SetUint64(receiver.Amount),
		}

		dstMinUtxo, err := p.calculateMinUtxo(
			cardanoDestConfig, originChainID, receiverAddr, 0, coloredCoinsAmountSum, coloredCoinsIDToToken,
		)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("failed to calculate destination minUtxo for chainID: %s and colored coin: %d. err: %w",
				cardanoDestConfig.ChainID, receiver.ColoredCoinID, err)
		}

		receiverAmountCurrencyDst = dstMinUtxo
		receiverAmountNativeTokenDst = receiver.Amount
		coloredCoinID = receiver.ColoredCoinID

	default: // currency on source
		*totalAmountCurrencySrc += receiver.Amount
		// totalAmountWrappedSrc stays the same

		dstMinUtxo, err := p.calculateMinUtxo(
			cardanoDestConfig, originChainID, receiverAddr, receiver.Amount, nil, nil)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("failed to calculate destination minUtxo for chainID: %s. err: %w",
				cardanoDestConfig.ChainID, err)
		}

		receiverAmountCurrencyDst = dstMinUtxo
		receiverAmountWrappedDst = receiver.Amount
		receiverAmountNativeTokenDst = receiver.Amount
		coloredCoinID = 0
	}

	brReceiver := cCore.BridgingRequestReceiver{
		DestinationAddress: receiverAddr,
		Amount:             new(big.Int).SetUint64(receiverAmountCurrencyDst),
		AmountWrapped:      new(big.Int).SetUint64(receiverAmountNativeTokenDst),
		ColoredCoinId:      coloredCoinID,
	}

	return &brReceiver, receiverAmountCurrencyDst, receiverAmountWrappedDst, nil
}

/* func (p *BridgingRequestedProcessorSkylineImpl) processReceiverEth(
	receiver sendtx.BridgingRequestMetadataTransaction,
	totalAmountCurrencySrc, totalAmountWrappedSrc *uint64,
) (*cCore.BridgingRequestReceiver, error) {
	receiverAddr := strings.Join(receiver.Address, "")

	switch receiver.BridgingType {
	case sendtx.BridgingTypeWrappedTokenOnSource:
		*totalAmountWrappedSrc += receiver.Amount
	case sendtx.BridgingTypeCurrencyOnSource, sendtx.BridgingTypeNormal:
		*totalAmountCurrencySrc += receiver.Amount
	}

	brReceiver := cCore.BridgingRequestReceiver{
		DestinationAddress: receiverAddr,
		Amount:             new(big.Int).SetUint64(0),
		AmountWrapped:      new(big.Int).SetUint64(receiver.Amount),
		ColoredCoinId:      receiver.ColoredCoinID,
	}

	return &brReceiver, nil
} */
