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
	"github.com/Ethernal-Tech/cardano-infrastructure/sendtx"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
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
	cardanoDestConfig := appConfig.CardanoChains[metadata.DestinationChainID]
	cardanoDestChainFeeAddress := appConfig.GetFeeMultisigAddress(metadata.DestinationChainID)

	// Map colored coin ID to native tokens
	coloredCoinsIDToToken := make(map[uint16]*cardanowallet.Token, len(cardanoDestConfig.ColoredCoins))

	for _, cc := range cardanoDestConfig.ColoredCoins {
		ccToken, err := cardanotx.GetNativeTokenFromName(cc.TokenName)
		if err != nil {
			return fmt.Errorf("failed to resolve native token for colored coin %s: %w", cc.TokenName, err)
		}

		coloredCoinsIDToToken[cc.ColoredCoinID] = &ccToken
	}

	var (
		totalAmountCurrencySrc = uint64(0)
		totalAmountCurrencyDst = uint64(0)
		totalAmountWrappedSrc  = uint64(0)
		totalAmountWrappedDst  = uint64(0)

		receivers = make([]cCore.BridgingRequestReceiver, 0, len(metadata.Transactions))
	)

	for _, receiver := range metadata.Transactions {
		receiverAddr := strings.Join(receiver.Address, "")

		if receiverAddr == cardanoDestChainFeeAddress {
			// fee address will be added at the end
			continue
		}

		var (
			receiverAmountCurrencyDst    uint64
			receiverAmountWrappedDst     uint64
			receiverAmountNativeTokenDst uint64
			coloredCoinID                uint16
		)

		switch receiver.BridgingType {
		case sendtx.BridgingTypeWrappedTokenOnSource:
			// receiverAmount represents the amount of native currency that is bridged to the receiver.
			// receiver.Amount of native tokens on the source will be converted to the same amount of native currency on
			// the destination.
			// totalAmountCurrencySrc stays the same
			totalAmountWrappedSrc += receiver.Amount

			// receiverAmountWrappedDst stays the same
			receiverAmountCurrencyDst = receiver.Amount
			coloredCoinID = 0
		case sendtx.BridgingTypeColoredCoinOnSource:
			coloredCoinsAmountSum := map[uint16]*big.Int{
				receiver.ColoredCoinID: new(big.Int).SetUint64(receiver.Amount),
			}

			dstMinUtxo, err := p.calculateMinUtxo(
				cardanoDestConfig, tx.OriginChainID, receiverAddr, 0, coloredCoinsAmountSum, coloredCoinsIDToToken,
			)
			if err != nil {
				return fmt.Errorf("failed to calculate destination minUtxo for chainID: %s and colored coin: %d. err: %w",
					cardanoDestConfig.ChainID, receiver.ColoredCoinID, err)
			}

			receiverAmountCurrencyDst = dstMinUtxo
			receiverAmountNativeTokenDst = receiver.Amount
			coloredCoinID = receiver.ColoredCoinID

		default: // currency on source
			totalAmountCurrencySrc += receiver.Amount
			// totalAmountWrappedSrc stays the same

			dstMinUtxo, err := p.calculateMinUtxo(
				cardanoDestConfig, tx.OriginChainID, receiverAddr, receiver.Amount, nil, nil)
			if err != nil {
				return fmt.Errorf("failed to calculate destination minUtxo for chainID: %s. err: %w",
					cardanoDestConfig.ChainID, err)
			}

			receiverAmountCurrencyDst = dstMinUtxo
			receiverAmountWrappedDst = receiver.Amount
			receiverAmountNativeTokenDst = receiver.Amount
			coloredCoinID = 0
		}

		totalAmountCurrencyDst += receiverAmountCurrencyDst
		totalAmountWrappedDst += receiverAmountWrappedDst

		receivers = append(receivers, cCore.BridgingRequestReceiver{
			DestinationAddress: receiverAddr,
			Amount:             new(big.Int).SetUint64(receiverAmountCurrencyDst),
			AmountWrapped:      new(big.Int).SetUint64(receiverAmountNativeTokenDst),
			ColoredCoinId:      coloredCoinID,
		})
	}

	feeCurrencyDst := cardanoDestConfig.FeeAddrBridgingAmount
	totalAmountCurrencyDst += feeCurrencyDst
	totalAmountCurrencySrc += metadata.BridgingFee + metadata.OperationFee

	receivers = append(receivers, cCore.BridgingRequestReceiver{
		DestinationAddress: cardanoDestChainFeeAddress,
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

func (p *BridgingRequestedProcessorSkylineImpl) validate(
	tx *core.CardanoTx, metadata *common.BridgingRequestMetadata, appConfig *cCore.AppConfig,
) error {
	if err := p.refundRequestProcessor.HandleBridgingProcessorPreValidate(tx, appConfig); err != nil {
		return err
	}

	if err := utils.IsTxDirectionAllowed(appConfig, tx.OriginChainID, metadata); err != nil {
		return err
	}

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

	nativeCurrencyAmountSum := big.NewInt(0)
	wrappedTokenAmountSum := big.NewInt(0)
	coloredCoinsAmountSum := make(map[uint16]*big.Int)
	coloredCoinsIDToToken := make(map[uint16]*cardanowallet.Token)

	hasNativeTokenOnSource := false
	hasCurrencyOnSource := false
	feeSum := uint64(0)

	cardanoDestChainFeeAddress := appConfig.GetFeeMultisigAddress(metadata.DestinationChainID)

	for _, receiver := range metadata.Transactions {
		receiverAddr := strings.Join(receiver.Address, "")

		if !cardanotx.IsValidOutputAddress(receiverAddr, cardanoDestConfig.NetworkID) {
			return fmt.Errorf("found an invalid receiver addr in metadata: %v", metadata)
		}

		if receiverAddr == cardanoDestChainFeeAddress {
			if receiver.BridgingType == sendtx.BridgingTypeWrappedTokenOnSource ||
				receiver.BridgingType == sendtx.BridgingTypeColoredCoinOnSource {
				return fmt.Errorf("fee receiver metadata invalid: %v", metadata)
			}

			feeSum += receiver.Amount

			continue
		}

		switch receiver.BridgingType {
		case sendtx.BridgingTypeWrappedTokenOnSource:
			hasNativeTokenOnSource = true
			// amount_to_bridge must be >= minUtxoAmount on destination
			if receiver.Amount < cardanoDestConfig.UtxoMinAmount {
				return fmt.Errorf("found an utxo value below minimum value in metadata receivers: %v", metadata)
			}

			if receiver.ColoredCoinID != 0 {
				return fmt.Errorf(
					"wrapped-token receiver must have ColoredCoinId = 0 (got %d): %v", receiver.ColoredCoinID, metadata,
				)
			}

			wrappedTokenAmountSum.Add(wrappedTokenAmountSum, new(big.Int).SetUint64(receiver.Amount))
		case sendtx.BridgingTypeColoredCoinOnSource:
			hasNativeTokenOnSource = true

			if receiver.ColoredCoinID == 0 {
				return fmt.Errorf("colored coin receiver must have a non-zero ColoredCoinId: %v", metadata)
			}

			if receiver.Amount < appConfig.BridgingSettings.MinColCoinsAllowedToBridge {
				return fmt.Errorf(
					"colored coin receiver amount too low for ColoredCoinId %d: got %d, minimum allowed %d; metadata: %v",
					receiver.ColoredCoinID,
					receiver.Amount,
					appConfig.BridgingSettings.MinColCoinsAllowedToBridge,
					metadata,
				)
			}

			colCoinSum, ok := coloredCoinsAmountSum[receiver.ColoredCoinID]
			if ok {
				colCoinSum.Add(colCoinSum, new(big.Int).SetUint64(receiver.Amount))
			} else {
				coloredCoinsAmountSum[receiver.ColoredCoinID] = new(big.Int).SetUint64(receiver.Amount)
			}
		default: // currency on source
			hasCurrencyOnSource = true
			// amount_to_bridge must be >= minUtxoAmount on source
			if receiver.Amount < cardanoSrcConfig.UtxoMinAmount {
				return fmt.Errorf("found an utxo value below minimum value in metadata receivers: %v", metadata)
			}

			if receiver.ColoredCoinID != 0 {
				return fmt.Errorf(
					"currency-on-source receiver must have ColoredCoinId = 0 (got %d): %v", receiver.ColoredCoinID, metadata,
				)
			}

			nativeCurrencyAmountSum.Add(nativeCurrencyAmountSum, new(big.Int).SetUint64(receiver.Amount))
		}
	}

	if appConfig.BridgingSettings.MaxAmountAllowedToBridge != nil &&
		appConfig.BridgingSettings.MaxAmountAllowedToBridge.Sign() > 0 &&
		nativeCurrencyAmountSum.Cmp(appConfig.BridgingSettings.MaxAmountAllowedToBridge) == 1 {
		return fmt.Errorf("sum of receiver amounts + fee: %v greater than maximum allowed: %v",
			nativeCurrencyAmountSum, appConfig.BridgingSettings.MaxAmountAllowedToBridge)
	}

	// update fee amount if needed with sum of fee address receivers
	metadata.BridgingFee += feeSum
	nativeCurrencyAmountSum.Add(nativeCurrencyAmountSum, new(big.Int).SetUint64(metadata.BridgingFee))
	nativeCurrencyAmountSum.Add(nativeCurrencyAmountSum, new(big.Int).SetUint64(metadata.OperationFee))

	minBridgingFee := cardanoSrcConfig.GetMinBridgingFee(hasNativeTokenOnSource || len(multisigUtxo.Tokens) > 0)

	if metadata.BridgingFee < minBridgingFee {
		return fmt.Errorf("bridging fee in metadata receivers is less than minimum: fee %d, minFee %d, metadata %v",
			metadata.BridgingFee, minBridgingFee, metadata)
	}

	// if there is at least one native token on source transfer or multi sig has tokens
	// -> native token on source should be defined
	if hasNativeTokenOnSource || len(multisigUtxo.Tokens) > 0 {
		nativeToken, err := cardanoSrcConfig.GetNativeToken(metadata.DestinationChainID)
		if err != nil {
			return err
		}

		multisigWrappedTokenAmount := new(big.Int).SetUint64(cardanotx.GetTokenAmount(multisigUtxo, nativeToken.String()))

		if wrappedTokenAmountSum.Cmp(multisigWrappedTokenAmount) != 0 {
			return fmt.Errorf("multisig wrapped token is not equal to receiver wrapped token amount: expected %v but got %v",
				multisigWrappedTokenAmount, wrappedTokenAmountSum)
		}
	}

	for _, cc := range cardanoSrcConfig.ColoredCoins {
		ccToken, err := cardanotx.GetNativeTokenFromName(cc.TokenName)
		if err != nil {
			return fmt.Errorf("failed to resolve native token for colored coin %s: %w", cc.TokenName, err)
		}

		coloredCoinsIDToToken[cc.ColoredCoinID] = &ccToken
		multisigColCoinTokenAmount := new(big.Int).SetUint64(cardanotx.GetTokenAmount(multisigUtxo, ccToken.String()))

		ccAmountSum, exists := coloredCoinsAmountSum[cc.ColoredCoinID]

		// Case 1: receiver does NOT include this colored coin
		if !exists {
			if multisigColCoinTokenAmount.Sign() > 0 {
				return fmt.Errorf(
					"colored coin %d amount mismatch: expected %v but multisig has %v",
					cc.ColoredCoinID, 0, multisigColCoinTokenAmount,
				)
			}

			continue
		}

		// Case 2: receiver includes colored coin - amounts must match
		if ccAmountSum.Cmp(multisigColCoinTokenAmount) != 0 {
			return fmt.Errorf(
				"colored coin %d amount mismatch: expected %v but got %v",
				cc.ColoredCoinID, ccAmountSum, multisigColCoinTokenAmount,
			)
		}
	}

	// if there is at least one currency on source transfer -> native token on destination should be defined
	if hasCurrencyOnSource {
		if _, err := cardanoDestConfig.GetNativeToken(tx.OriginChainID); err != nil {
			return err
		}
	}

	srcMinUtxo := cardanoSrcConfig.UtxoMinAmount
	if wrappedTokenAmountSum.Sign() > 0 || len(coloredCoinsAmountSum) > 0 {
		srcMinUtxo, err = p.calculateMinUtxo(
			cardanoSrcConfig, metadata.DestinationChainID, multisigUtxo.Address, wrappedTokenAmountSum.Uint64(),
			coloredCoinsAmountSum, coloredCoinsIDToToken)
		if err != nil {
			return fmt.Errorf("failed to calculate src minUtxo for chainID: %s. err: %w",
				cardanoSrcConfig.ChainID, err)
		}
	}

	minCurrency := srcMinUtxo + minBridgingFee
	if new(big.Int).SetUint64(minCurrency).Cmp(nativeCurrencyAmountSum) == 1 {
		return fmt.Errorf("sum of receiver amounts + fee is under the minimum allowed: min %v but got %v",
			minCurrency, nativeCurrencyAmountSum)
	}

	maxTokenAmt := appConfig.BridgingSettings.MaxTokenAmountAllowedToBridge
	if maxTokenAmt != nil && maxTokenAmt.Sign() > 0 {
		if wrappedTokenAmountSum.Cmp(maxTokenAmt) == 1 {
			return fmt.Errorf("sum of wrapped token: %v greater than maximum allowed: %v", wrappedTokenAmountSum, maxTokenAmt)
		}

		for coloredCoinID, coloredCoinAmount := range coloredCoinsAmountSum {
			if coloredCoinAmount.Cmp(maxTokenAmt) == 1 {
				return fmt.Errorf("sum of colored token %d: %v greater than maximum allowed: %v",
					coloredCoinID, coloredCoinAmount, maxTokenAmt)
			}
		}
	}

	if nativeCurrencyAmountSum.Cmp(new(big.Int).SetUint64(multisigUtxo.Amount)) != 0 {
		return fmt.Errorf("multisig amount is not equal to sum of receiver amounts + fee: expected %v but got %v",
			multisigUtxo.Amount, nativeCurrencyAmountSum)
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
