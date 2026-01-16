package successtxprocessors

import (
	"fmt"
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/common"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	oUtils "github.com/Ethernal-Tech/apex-bridge/oracle_common/utils"
	"github.com/Ethernal-Tech/apex-bridge/oracle_eth/core"
	goEthCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-hclog"
)

var _ core.EthTxSuccessRefundProcessor = (*RefundRequestProcessorSkylineImpl)(nil)

type RefundRequestProcessorSkylineImpl struct {
	logger hclog.Logger
}

func NewRefundRequestProcessorSkyline(logger hclog.Logger) *RefundRequestProcessorSkylineImpl {
	return &RefundRequestProcessorSkylineImpl{
		logger: logger.Named("refund_request_processor"),
	}
}

func (*RefundRequestProcessorSkylineImpl) GetType() common.BridgingTxType {
	return common.TxTypeRefundRequest
}

func (*RefundRequestProcessorSkylineImpl) PreValidate(tx *core.EthTx, appConfig *cCore.AppConfig) error {
	return nil
}

func (*RefundRequestProcessorSkylineImpl) HandleBridgingProcessorPreValidate(
	tx *core.EthTx, appConfig *cCore.AppConfig) error {
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
	claims *cCore.BridgeClaims, tx *core.EthTx, appConfig *cCore.AppConfig,
	err error, errContext string,
) error {
	p.logger.Warn(fmt.Sprintf("%s. handing over to refund processor", errContext),
		"tx", tx, "err", err)

	return p.ValidateAndAddClaim(claims, tx, appConfig)
}

func (p *RefundRequestProcessorSkylineImpl) ValidateAndAddClaim(
	claims *cCore.BridgeClaims, tx *core.EthTx, appConfig *cCore.AppConfig,
) error {
	metadata, err := core.UnmarshalEthMetadata[core.RefundBridgingRequestEthMetadata](
		tx.Metadata)
	if err != nil {
		return fmt.Errorf("failed to unmarshal metadata: tx: %v, err: %w", tx, err)
	}

	if err := p.validate(tx, metadata, appConfig); err != nil {
		return fmt.Errorf("refund validation failed for tx: %v, err: %w", tx, err)
	}

	p.addRefundRequestClaim(claims, tx, metadata, appConfig)

	return nil
}

func (p *RefundRequestProcessorSkylineImpl) addRefundRequestClaim(
	claims *cCore.BridgeClaims, tx *core.EthTx,
	metadata *core.RefundBridgingRequestEthMetadata,
	appConfig *cCore.AppConfig,
) {
	chainConfig := appConfig.EthChains[tx.OriginChainID]
	currencyID, _ := chainConfig.GetCurrencyID()
	chainIDConverter := appConfig.ChainIDConverter

	tokenAmounts, totalCurrency, totalWrapped :=
		buildRefundTokenAmounts(chainConfig, tx.Value, metadata, currencyID)

	claim := cCore.RefundRequestClaim{
		OriginChainId:            chainIDConverter.ToChainIDNum(tx.OriginChainID),
		DestinationChainId:       chainIDConverter.ToChainIDNum(metadata.DestinationChainID), // unused for RefundRequestClaim
		OriginTransactionHash:    tx.Hash,
		OriginSenderAddress:      metadata.SenderAddr,
		OriginAmount:             totalCurrency,
		OriginWrappedAmount:      totalWrapped,
		OutputIndexes:            []byte{},
		ShouldDecrementHotWallet: tx.BatchTryCount > 0,
		RetryCounter:             uint64(tx.RefundTryCount),
		TokenAmounts:             tokenAmounts,
	}

	claims.RefundRequestClaims = append(claims.RefundRequestClaims, claim)

	p.logger.Info("Added RefundRequestClaim",
		"txHash", tx.Hash, "claim", cCore.RefundRequestClaimString(claim, chainIDConverter))
}

func (p *RefundRequestProcessorSkylineImpl) validate(
	tx *core.EthTx, metadata *core.RefundBridgingRequestEthMetadata, appConfig *cCore.AppConfig,
) error {
	if tx.RefundTryCount > appConfig.TryCountLimits.MaxRefundTryCount {
		return fmt.Errorf("try count exceeded. RefundTryCount: (current, max)=(%d, %d)",
			tx.RefundTryCount, appConfig.TryCountLimits.MaxRefundTryCount)
	}

	chainConfig := appConfig.EthChains[tx.OriginChainID]
	if chainConfig == nil {
		return fmt.Errorf("unsupported chain id found in tx. chain id: %v", tx.OriginChainID)
	}

	if !goEthCommon.IsHexAddress(metadata.SenderAddr) {
		return fmt.Errorf("invalid sender addr: %s", metadata.SenderAddr)
	}

	minFeeForBridging := chainConfig.MinFeeForBridging.Int
	if tx.Value.Cmp(minFeeForBridging) != 1 {
		return fmt.Errorf(
			"tx.Value: %v is less than the minimum required for refund: %v",
			tx.Value, new(big.Int).Add(minFeeForBridging, big.NewInt(1)),
		)
	}

	for _, receiver := range metadata.Transactions {
		if _, exists := chainConfig.Tokens[receiver.TokenID]; !exists {
			return fmt.Errorf(
				"token with ID %d is not registered in chain %s",
				receiver.TokenID,
				chainConfig.ChainID,
			)
		}
	}

	return nil
}

func buildRefundTokenAmounts(
	chainConfig *cCore.EthChainConfig,
	txValue *big.Int,
	metadata *core.RefundBridgingRequestEthMetadata,
	currencyID uint16,
) (tokenAmounts []cCore.RefundTokenAmount, totalCurrency, totalWrapped *big.Int) {
	tokenAmounts = make([]cCore.RefundTokenAmount, 0)
	totalCurrency = big.NewInt(0)
	totalWrapped = big.NewInt(0)

	currencyAdded := false

	for _, receiver := range metadata.Transactions {
		tokenPair, _ := oUtils.GetTokenPair(
			chainConfig.DestinationChains,
			chainConfig.ChainID,
			metadata.DestinationChainID,
			receiver.TokenID,
		)

		// handle currency
		if receiver.TokenID == currencyID {
			if tokenPair != nil && tokenPair.TrackSourceToken {
				totalCurrency.Add(totalCurrency, receiver.Amount)
			}

			if !currencyAdded {
				tokenAmounts = append(tokenAmounts, cCore.RefundTokenAmount{
					TokenId:        receiver.TokenID,
					AmountCurrency: txValue,
					AmountTokens:   big.NewInt(0),
				})

				currencyAdded = true
			}

			continue
		}

		// handle wrapped token
		if chainConfig.Tokens[receiver.TokenID].IsWrappedCurrency {
			if tokenPair != nil && tokenPair.TrackSourceToken {
				totalWrapped.Add(totalWrapped, receiver.Amount)
			}
		}

		// build RefundTokenAmount entry
		currencyAmount := big.NewInt(0)
		if !currencyAdded {
			currencyAmount = txValue
			currencyAdded = true
		}

		tokenAmounts = append(tokenAmounts, cCore.RefundTokenAmount{
			TokenId:        receiver.TokenID,
			AmountCurrency: currencyAmount,
			AmountTokens:   receiver.Amount,
		})
	}

	return
}
