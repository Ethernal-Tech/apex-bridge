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
	"github.com/hashicorp/go-hclog"
)

var _ core.EthTxSuccessProcessor = (*BridgingRequestedProcessorImpl)(nil)

type BridgingRequestedProcessorSkylineImpl struct {
	refundRequestProcessor core.EthTxSuccessRefundProcessor
	logger                 hclog.Logger

	cardanoChainInfos map[string]*oChain.CardanoChainInfo
}

func NewEthBridgingRequestedProcessorSkyline(
	refundRequestProcessor core.EthTxSuccessRefundProcessor, logger hclog.Logger,
	cardanoChainInfos map[string]*oChain.CardanoChainInfo,
) *BridgingRequestedProcessorImpl {
	return &BridgingRequestedProcessorImpl{
		refundRequestProcessor: refundRequestProcessor,
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
	cardanoDestConfig, _ := oUtils.GetChainConfig(appConfig, metadata.DestinationChainID)
	cardanoDestChainFeeAddress := appConfig.GetFeeMultisigAddress(metadata.DestinationChainID)
	srcCurrencyID, _ := ethSrcConfig.GetCurrencyID()
	destCurrencyID, _ := cardanoDestConfig.GetCurrencyID()

	totalCurrencySrc := big.NewInt(0)
	totalWrappedSrc := big.NewInt(0)
	totalCurrencyDest := big.NewInt(0)
	totalWrappedDest := big.NewInt(0)
	receivers := make([]oCore.BridgingRequestReceiver, 0, len(metadata.Transactions))

	for _, receiver := range metadata.Transactions {
		if receiver.Address == cardanoDestChainFeeAddress {
			// fee address will be added at the end
			continue
		}

		// validation has already checked that there is no error
		tokenPair, _ := oUtils.GetTokenPair(
			ethSrcConfig.DestinationChain,
			ethSrcConfig.ChainID,
			cardanoDestConfig.ChainID,
			receiver.TokenID,
		)

		var amount *big.Int

		amountWrapped := big.NewInt(0)
		receiverAmountDfm := common.WeiToDfm(receiver.Amount)

		// currency on destination
		if tokenPair.DestinationTokenID == destCurrencyID {
			amount = receiverAmountDfm

			if tokenPair.TrackDestinationToken {
				totalCurrencyDest.Add(totalCurrencyDest, receiverAmountDfm)
			}
		} else {
			nativeTokensSum := map[uint16]*big.Int{
				tokenPair.DestinationTokenID: receiverAmountDfm,
			}

			dstMinUtxo, err := p.calculateMinUtxo(cardanoDestConfig, receiver.Address, nativeTokensSum)
			if err != nil {
				return fmt.Errorf("failed to calculate destination minUtxo for chainID: %s. err: %w",
					cardanoDestConfig.ChainID, err)
			}

			amount = new(big.Int).SetUint64(dstMinUtxo)
			totalCurrencyDest.Add(totalCurrencyDest, amount)

			amountWrapped = receiverAmountDfm

			// wrapped token on destination
			if tokenPair.TrackDestinationToken && cardanoDestConfig.Tokens[tokenPair.DestinationTokenID].IsWrappedCurrency {
				totalWrappedDest.Add(totalWrappedDest, receiverAmountDfm)
			}
		}

		if tokenPair.TrackSourceToken {
			if tokenPair.SourceTokenID == srcCurrencyID {
				// currency on source
				totalCurrencySrc.Add(totalCurrencySrc, receiverAmountDfm)
			} else if ethSrcConfig.Tokens[tokenPair.SourceTokenID].IsWrappedCurrency {
				// wrapped token on source
				totalWrappedSrc.Add(totalWrappedSrc, receiverAmountDfm)
			}
		}

		receivers = append(receivers, oCore.BridgingRequestReceiver{
			DestinationAddress: receiver.Address,
			Amount:             amount,
			AmountWrapped:      amountWrapped,
			TokenId:            tokenPair.DestinationTokenID,
		})
	}

	// wTODO: think about whether we should always track all currency amounts,
	// regardless of .TrackSource and .TrackDestination
	totalCurrencySrc = new(big.Int).Add(totalCurrencySrc, common.WeiToDfm(metadata.BridgingFee))
	totalCurrencySrc = new(big.Int).Add(totalCurrencySrc, common.WeiToDfm(metadata.OperationFee))

	feeCurrencyDfmDst := new(big.Int).SetUint64(cardanoDestConfig.FeeAddrBridgingAmount)
	totalCurrencyDest = new(big.Int).Add(totalCurrencyDest, feeCurrencyDfmDst)

	receivers = append(receivers, oCore.BridgingRequestReceiver{
		DestinationAddress: cardanoDestChainFeeAddress,
		Amount:             feeCurrencyDfmDst,
		AmountWrapped:      big.NewInt(0),
		TokenId:            0,
	})

	claim := oCore.BridgingRequestClaim{
		ObservedTransactionHash:         tx.Hash,
		SourceChainId:                   common.ToNumChainID(tx.OriginChainID),
		DestinationChainId:              common.ToNumChainID(metadata.DestinationChainID),
		Receivers:                       receivers,
		NativeCurrencyAmountSource:      totalCurrencySrc,
		NativeCurrencyAmountDestination: totalCurrencyDest,
		WrappedTokenAmountSource:        totalWrappedSrc,
		WrappedTokenAmountDestination:   totalWrappedDest,
		RetryCounter:                    big.NewInt(int64(tx.BatchTryCount)),
	}

	claims.BridgingRequestClaims = append(claims.BridgingRequestClaims, claim)

	p.logger.Info("Added BridgingRequestClaim",
		"txHash", tx.Hash, "metadata", metadata, "claim", oCore.BridgingRequestClaimString(claim))

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

	cardanoDestConfig, _ := oUtils.GetChainConfig(appConfig, metadata.DestinationChainID)
	if cardanoDestConfig == nil {
		return fmt.Errorf("destination chain not registered: %v", metadata.DestinationChainID)
	}

	cardanoDestChainFeeAddress := appConfig.GetFeeMultisigAddress(metadata.DestinationChainID)

	if len(metadata.Transactions) > appConfig.BridgingSettings.MaxReceiversPerBridgingRequest {
		return fmt.Errorf("number of receivers in metadata greater than maximum allowed - no: %v, max: %v, metadata: %v",
			len(metadata.Transactions), appConfig.BridgingSettings.MaxReceiversPerBridgingRequest, metadata)
	}

	feeSum := big.NewInt(0)
	receiverCurrencySum := big.NewInt(0)
	receiverTokensSum := make(map[uint16]*big.Int)

	srcCurrencyID, err := ethSrcConfig.GetCurrencyID()
	if err != nil {
		return err
	}

	destCurrencyID, err := cardanoDestConfig.GetCurrencyID()
	if err != nil {
		return err
	}

	for _, receiver := range metadata.Transactions {
		if !cardanotx.IsValidOutputAddress(receiver.Address, cardanoDestConfig.NetworkID) {
			return fmt.Errorf("found an invalid receiver addr in metadata: %v", metadata)
		}

		tokenPair, err := oUtils.GetTokenPair(
			ethSrcConfig.DestinationChain,
			ethSrcConfig.ChainID,
			cardanoDestConfig.ChainID,
			receiver.TokenID,
		)
		if err != nil {
			return err
		}

		// currency on source
		if tokenPair.SourceTokenID == srcCurrencyID {
			if receiver.Address == cardanoDestChainFeeAddress {
				feeSum.Add(feeSum, receiver.Amount)
			} else {
				receiverCurrencySum.Add(receiverCurrencySum, receiver.Amount)
			}
		} else {
			if receiver.Address == cardanoDestChainFeeAddress {
				return fmt.Errorf("fee receiver metadata invalid: %v", metadata)
			}
		}

		// currency on destination
		if tokenPair.DestinationTokenID == destCurrencyID {
			if common.WeiToDfm(receiver.Amount).Uint64() < cardanoDestConfig.UtxoMinAmount {
				return fmt.Errorf("found a utxo value below minimum value in metadata receivers: %v", metadata)
			}
		} else {
			if common.WeiToDfm(receiver.Amount).Uint64() < appConfig.BridgingSettings.MinColCoinsAllowedToBridge {
				return fmt.Errorf("found a colored coin amount below minimum allowed in metadata receivers: %v", metadata)
			}

			tokensSum, exists := receiverTokensSum[receiver.TokenID]
			if !exists {
				tokensSum = big.NewInt(0)
				receiverTokensSum[receiver.TokenID] = tokensSum
			}

			tokensSum.Add(tokensSum, receiver.Amount)
		}
	}

	if appConfig.BridgingSettings.MaxAmountAllowedToBridge != nil &&
		appConfig.BridgingSettings.MaxAmountAllowedToBridge.Sign() > 0 &&
		receiverCurrencySum.Cmp(common.DfmToWei(appConfig.BridgingSettings.MaxAmountAllowedToBridge)) == 1 {
		return fmt.Errorf("sum of receiver amounts: %v greater than maximum allowed: %v",
			receiverCurrencySum, common.DfmToWei(appConfig.BridgingSettings.MaxAmountAllowedToBridge))
	}

	for tokenID, tokenSum := range receiverTokensSum {
		if tokenSum.Cmp(common.DfmToWei(appConfig.BridgingSettings.MaxTokenAmountAllowedToBridge)) == 1 {
			return fmt.Errorf(
				"amount of tokens for receivers too high for token with ID %d: %v greater than maximum allowed: %v",
				tokenID, tokenSum, common.DfmToWei(appConfig.BridgingSettings.MaxTokenAmountAllowedToBridge))
		}
	}

	// update fee amount if needed with sum of fee address receivers
	metadata.BridgingFee.Add(metadata.BridgingFee, feeSum)
	receiverCurrencySum.Add(receiverCurrencySum, metadata.BridgingFee)
	receiverCurrencySum.Add(receiverCurrencySum, metadata.OperationFee)

	feeAmountDfm := common.WeiToDfm(metadata.BridgingFee)
	if feeAmountDfm.Uint64() < ethSrcConfig.MinFeeForBridging {
		return fmt.Errorf("bridging fee in metadata is less than minimum: %v", metadata)
	}

	if tx.Value == nil || tx.Value.Cmp(receiverCurrencySum) != 0 {
		return fmt.Errorf("tx value is not equal to sum of receiver amounts + fee: expected %v but got %v",
			receiverCurrencySum, tx.Value)
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
