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
	"github.com/hashicorp/go-hclog"
)

var _ core.CardanoTxSuccessRefundProcessor = (*RefundRequestProcessorSkylineImpl)(nil)

type RefundRequestProcessorSkylineImpl struct {
	logger     hclog.Logger
	chainInfos map[string]*cChain.CardanoChainInfo
}

func NewRefundRequestProcessorSkyline(
	logger hclog.Logger, chainInfos map[string]*cChain.CardanoChainInfo,
) *RefundRequestProcessorSkylineImpl {
	return &RefundRequestProcessorSkylineImpl{
		logger:     logger.Named("refund_request_processor"),
		chainInfos: chainInfos,
	}
}

func (*RefundRequestProcessorSkylineImpl) GetType() common.BridgingTxType {
	return common.TxTypeRefundRequest
}

func (*RefundRequestProcessorSkylineImpl) PreValidate(tx *core.CardanoTx, appConfig *cCore.AppConfig) error {
	return nil
}

func (*RefundRequestProcessorSkylineImpl) HandleBridgingProcessorPreValidate(
	tx *core.CardanoTx, appConfig *cCore.AppConfig) error {
	if tx.BatchTryCount > appConfig.TryCountLimits.MaxBatchTryCount ||
		tx.SubmitTryCount > appConfig.TryCountLimits.MaxSubmitTryCount {
		return fmt.Errorf(
			"try count exceeded. BatchTryCount: (current, max)=(%d, %d), SubmitTryCount: (current, max)=(%d, %d)",
			tx.BatchTryCount, appConfig.TryCountLimits.MaxBatchTryCount,
			tx.SubmitTryCount, appConfig.TryCountLimits.MaxSubmitTryCount)
	}

	return nil
}

func (p *RefundRequestProcessorSkylineImpl) HandleBridgingProcessorError(
	claims *cCore.BridgeClaims, tx *core.CardanoTx, appConfig *cCore.AppConfig,
	err error, errContext string,
) error {
	p.logger.Warn(fmt.Sprintf("%s. handing over to refund processor", errContext),
		"tx", tx, "err", err)

	return p.ValidateAndAddClaim(claims, tx, appConfig)
}

func (p *RefundRequestProcessorSkylineImpl) ValidateAndAddClaim(
	claims *cCore.BridgeClaims, tx *core.CardanoTx, appConfig *cCore.AppConfig,
) error {
	metadata, err := common.UnmarshalMetadata[common.RefundBridgingRequestMetadata](
		common.MetadataEncodingTypeCbor, tx.Metadata)
	if err != nil {
		return fmt.Errorf("failed to unmarshal metadata: tx: %v, err: %w", tx, err)
	}

	if err := p.validate(tx, metadata, appConfig); err != nil {
		return fmt.Errorf("refund validation failed for tx: %v, err: %w", tx, err)
	}

	return p.addRefundRequestClaim(claims, tx, metadata, appConfig)
}

func (p *RefundRequestProcessorSkylineImpl) addRefundRequestClaim(
	claims *cCore.BridgeClaims, tx *core.CardanoTx,
	metadata *common.RefundBridgingRequestMetadata, appConfig *cCore.AppConfig,
) error {
	chainConfig := appConfig.CardanoChains[tx.OriginChainID]
	senderAddr, _ := p.getSenderAddr(chainConfig, metadata)
	currencyAmountSum := big.NewInt(0)
	tokenAmounts := make(map[uint16]*big.Int)
	unknownTokenOutputIndexes := make([]common.TxOutputIndex, 0, unknownNativeTokensUtxoCntMax)

	zeroAddress, ok := appConfig.BridgingAddressesManager.GetPaymentAddressFromIndex(
		common.ToNumChainID(tx.OriginChainID), 0)
	if !ok {
		return fmt.Errorf("failed to get zero address from bridging address manager")
	}

	tokenNamesAndIDs, err := chainConfig.GetFullTokenNamesAndIds()
	if err != nil {
		return fmt.Errorf("failed to get full token names and IDs from config. err: %w", err)
	}

	wrappedTokenAmount := big.NewInt(0)

	currencyID, err := chainConfig.GetCurrencyID()
	if err != nil {
		return err
	}

	currencyTokenPair, err := cUtils.GetTokenPair(
		chainConfig.DestinationChains, chainConfig.ChainID, metadata.DestinationChainID, currencyID)
	trackCurrency := err == nil && currencyTokenPair.TrackSourceToken

	wrappedTokenID, wrappedExists := chainConfig.GetWrappedTokenID()

	for idx, out := range tx.Outputs {
		if !utils.IsBridgingAddrForChain(appConfig, chainConfig.ChainID, out.Address) {
			continue
		}

		for _, token := range out.Tokens {
			tokenID, ok := tokenNamesAndIDs[token.TokenName()]

			if zeroAddress != out.Address || !ok {
				unknownTokenOutputIndexes = append(unknownTokenOutputIndexes, common.TxOutputIndex(idx)) //nolint:gosec

				break
			} else {
				// only wrapped tokens can be tracked on smart contracts
				if wrappedExists && wrappedTokenID == tokenID {
					tokenPair, err := cUtils.GetTokenPair(
						chainConfig.DestinationChains, chainConfig.ChainID, metadata.DestinationChainID, tokenID)

					if err == nil && tokenPair.TrackSourceToken {
						wrappedTokenAmount.Add(wrappedTokenAmount, new(big.Int).SetUint64(token.Amount))
					}
				}

				if tokenAmount, ok := tokenAmounts[tokenID]; ok {
					tokenAmount.Add(tokenAmount, new(big.Int).SetUint64(token.Amount))
				} else {
					tokenAmounts[tokenID] = new(big.Int).SetUint64(token.Amount)
				}
			}
		}

		if trackCurrency {
			currencyAmountSum.Add(currencyAmountSum, new(big.Int).SetUint64(out.Amount))
		}
	}

	refundTokensAmounts := buildRefundTokenAmounts(tokenAmounts, currencyAmountSum)

	// tx contains unknown tokens
	if len(unknownTokenOutputIndexes) > 0 {
		currencyAmountSum = big.NewInt(0)
		wrappedTokenAmount = big.NewInt(0)
	}

	claim := cCore.RefundRequestClaim{
		OriginChainId:            common.ToNumChainID(tx.OriginChainID),
		DestinationChainId:       common.ToNumChainID(metadata.DestinationChainID),
		OriginTransactionHash:    tx.Hash,
		OriginSenderAddress:      senderAddr,
		OriginAmount:             currencyAmountSum,
		OriginWrappedAmount:      wrappedTokenAmount,
		OutputIndexes:            common.PackNumbersToBytes(unknownTokenOutputIndexes),
		ShouldDecrementHotWallet: tx.BatchTryCount > 0,
		RetryCounter:             uint64(tx.RefundTryCount),
		TokenAmounts:             refundTokensAmounts,
	}

	claims.RefundRequestClaims = append(claims.RefundRequestClaims, claim)

	p.logger.Info("Added RefundRequestClaim",
		"txHash", tx.Hash, "claim", cCore.RefundRequestClaimString(claim))

	return nil
}

func (p *RefundRequestProcessorSkylineImpl) validate(
	tx *core.CardanoTx, metadata *common.RefundBridgingRequestMetadata, appConfig *cCore.AppConfig,
) error {
	if tx.RefundTryCount > appConfig.TryCountLimits.MaxRefundTryCount {
		return fmt.Errorf("try count exceeded. RefundTryCount: (current, max)=(%d, %d)",
			tx.RefundTryCount, appConfig.TryCountLimits.MaxRefundTryCount)
	}

	chainConfig := appConfig.CardanoChains[tx.OriginChainID]
	if chainConfig == nil {
		return fmt.Errorf("unsupported chain id found in tx. chain id: %v", tx.OriginChainID)
	}

	senderAddr, err := p.getSenderAddr(chainConfig, metadata)
	if err != nil {
		return err
	}

	zeroAddress, ok := appConfig.BridgingAddressesManager.GetPaymentAddressFromIndex(
		common.ToNumChainID(tx.OriginChainID), 0)
	if !ok {
		return fmt.Errorf("failed to get zero address from bridging address manager")
	}

	tokensNamesAndIds, err := chainConfig.GetFullTokenNamesAndIds()
	if err != nil {
		return fmt.Errorf("failed to get full token names and IDs from config. err: %w", err)
	}

	amountSum := big.NewInt(0)
	unknownNativeTokensUtxoCnt := uint(0)

	var hasTokens bool

	for _, out := range tx.Outputs {
		if !utils.IsBridgingAddrForChain(appConfig, chainConfig.ChainID, out.Address) {
			continue
		}

		amountSum.Add(amountSum, new(big.Int).SetUint64(out.Amount))

		if len(out.Tokens) > 0 {
			hasTokens = true

			if chainConfig.Tokens == nil || zeroAddress != out.Address {
				unknownNativeTokensUtxoCnt++
			} else {
				for _, token := range out.Tokens {
					if _, exists := tokensNamesAndIds[token.TokenName()]; !exists {
						unknownNativeTokensUtxoCnt++

						break
					}
				}
			}
		}
	}

	if unknownNativeTokensUtxoCnt > unknownNativeTokensUtxoCntMax {
		return fmt.Errorf("more UTxOs with unknown tokens than allowed. max: %d", unknownNativeTokensUtxoCntMax)
	}

	calculatedMinUtxo, err := calculateMinUtxoForRefund(chainConfig, tx, senderAddr,
		appConfig.BridgingAddressesManager.GetAllPaymentAddresses(common.ToNumChainID(chainConfig.ChainID)),
		p.chainInfos)
	if err != nil {
		return fmt.Errorf("failed to calculate min utxo. err: %w", err)
	}

	minBridgingFee := chainConfig.GetMinBridgingFee(hasTokens)

	if amountSum.Cmp(new(big.Int).SetUint64(minBridgingFee+calculatedMinUtxo)) == -1 {
		return fmt.Errorf(
			"sum of amounts to the bridging address: %v is less than the minimum required for refund: %v",
			amountSum, minBridgingFee+calculatedMinUtxo)
	}

	if appConfig.EthChains[metadata.DestinationChainID] == nil &&
		appConfig.CardanoChains[metadata.DestinationChainID] == nil {
		return fmt.Errorf("unsupported destination chain id found in metadata: %s", metadata.DestinationChainID)
	}

	return nil
}

func (p *RefundRequestProcessorSkylineImpl) getSenderAddr(
	config *cCore.CardanoChainConfig, metadata *common.RefundBridgingRequestMetadata,
) (string, error) {
	senderAddr := strings.Join(metadata.SenderAddr, "")

	if valid := cardanotx.IsValidOutputAddress(senderAddr, config.NetworkID); !valid {
		return "", fmt.Errorf("invalid sender addr: %s", senderAddr)
	}

	return senderAddr, nil
}

func buildRefundTokenAmounts(
	tokenAmounts map[uint16]*big.Int,
	currencyAmountSum *big.Int,
) []cCore.RefundTokenAmount {
	refundTokenAmounts := make([]cCore.RefundTokenAmount, 0, len(tokenAmounts))
	currencyAdded := false

	for tokenID, amount := range tokenAmounts {
		amountCurrency := big.NewInt(0)

		if !currencyAdded {
			// First token gets the full currency sum
			amountCurrency = currencyAmountSum
			currencyAdded = true
		}

		refundTokenAmounts = append(refundTokenAmounts, cCore.RefundTokenAmount{
			TokenId:        tokenID,
			AmountCurrency: amountCurrency,
			AmountTokens:   amount,
		})
	}

	return refundTokenAmounts
}
