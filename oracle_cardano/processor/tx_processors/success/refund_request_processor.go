package successtxprocessors

import (
	"fmt"
	"math/big"
	"slices"
	"strings"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/utils"
	cChain "github.com/Ethernal-Tech/apex-bridge/oracle_common/chain"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
)

const (
	unknownNativeTokensUtxoCntMax = 3
)

var _ core.CardanoTxSuccessRefundProcessor = (*RefundRequestProcessorImpl)(nil)

type RefundRequestProcessorImpl struct {
	logger     hclog.Logger
	chainInfos map[string]*cChain.CardanoChainInfo
}

func NewRefundRequestProcessor(
	logger hclog.Logger, chainInfos map[string]*cChain.CardanoChainInfo,
) *RefundRequestProcessorImpl {
	return &RefundRequestProcessorImpl{
		logger:     logger.Named("refund_request_processor"),
		chainInfos: chainInfos,
	}
}

func (*RefundRequestProcessorImpl) GetType() common.BridgingTxType {
	return common.TxTypeRefundRequest
}

func (*RefundRequestProcessorImpl) PreValidate(tx *core.CardanoTx, appConfig *cCore.AppConfig) error {
	return nil
}

func (*RefundRequestProcessorImpl) HandleBridgingProcessorPreValidate(
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

func (p *RefundRequestProcessorImpl) HandleBridgingProcessorError(
	claims *cCore.BridgeClaims, tx *core.CardanoTx, appConfig *cCore.AppConfig,
	err error, errContext string,
) error {
	p.logger.Warn(fmt.Sprintf("%s. handing over to refund processor", errContext),
		"tx", tx, "err", err)

	return p.ValidateAndAddClaim(claims, tx, appConfig)
}

func (p *RefundRequestProcessorImpl) ValidateAndAddClaim(
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

func (p *RefundRequestProcessorImpl) addRefundRequestClaim(
	claims *cCore.BridgeClaims, tx *core.CardanoTx,
	metadata *common.RefundBridgingRequestMetadata, appConfig *cCore.AppConfig,
) error {
	chainIDConverter := appConfig.ChainIDConverter
	chainConfig := appConfig.CardanoChains[tx.OriginChainID]
	senderAddr, _ := p.getSenderAddr(chainConfig, metadata)
	amount := big.NewInt(0)
	unknownTokenOutputIndexes := make([]common.TxOutputIndex, 0, unknownNativeTokensUtxoCntMax)

	for idx, out := range tx.Outputs {
		if !utils.IsBridgingAddrForChain(appConfig, chainConfig.ChainID, out.Address) {
			continue
		}

		// since this is a reactor only processor, we decline all tokens
		if len(out.Tokens) > 0 {
			unknownTokenOutputIndexes = append(unknownTokenOutputIndexes, common.TxOutputIndex(idx)) //nolint:gosec

			continue
		}

		amount.Add(amount, new(big.Int).SetUint64(out.Amount))
	}

	// tx contains unknown tokens
	// amounts are not used on batcher when unknown tokens are present
	if len(unknownTokenOutputIndexes) > 0 {
		amount = big.NewInt(0)
	}

	claim := cCore.RefundRequestClaim{
		OriginChainId:            chainIDConverter.ToChainIDNum(tx.OriginChainID),
		DestinationChainId:       chainIDConverter.ToChainIDNum(metadata.DestinationChainID), // unused for RefundRequestClaim
		OriginTransactionHash:    tx.Hash,
		OriginSenderAddress:      senderAddr,
		OriginAmount:             amount,
		OriginWrappedAmount:      big.NewInt(0),
		OutputIndexes:            common.PackNumbersToBytes(unknownTokenOutputIndexes),
		ShouldDecrementHotWallet: tx.BatchTryCount > 0,
		RetryCounter:             uint64(tx.RefundTryCount),
		TokenAmounts: []cCore.RefundTokenAmount{
			{TokenId: 0, AmountCurrency: amount, AmountTokens: big.NewInt(0)},
		},
	}

	claims.RefundRequestClaims = append(claims.RefundRequestClaims, claim)

	p.logger.Info("Added RefundRequestClaim",
		"txHash", tx.Hash, "claim", cCore.RefundRequestClaimString(claim, chainIDConverter))

	return nil
}

func (p *RefundRequestProcessorImpl) validate(
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

	amountSum := big.NewInt(0)
	unknownNativeTokensUtxoCnt := uint(0)

	var hasTokens bool

	for _, out := range tx.Outputs {
		if !utils.IsBridgingAddrForChain(appConfig, chainConfig.ChainID, out.Address) {
			continue
		}

		amountSum.Add(amountSum, new(big.Int).SetUint64(out.Amount))

		// since this is a reactor only processor, we decline all tokens
		if len(out.Tokens) > 0 {
			hasTokens = true
			unknownNativeTokensUtxoCnt++
		}
	}

	if unknownNativeTokensUtxoCnt > unknownNativeTokensUtxoCntMax {
		return fmt.Errorf("more UTxOs with unknown tokens than allowed. max: %d", unknownNativeTokensUtxoCntMax)
	}

	calculatedMinUtxo, err := calculateMinUtxoForRefund(chainConfig, tx, senderAddr,
		appConfig.BridgingAddressesManager.GetAllPaymentAddresses(
			appConfig.ChainIDConverter.ToChainIDNum(chainConfig.ChainID),
		), p.chainInfos)
	if err != nil {
		return fmt.Errorf("failed to calculate min utxo. err: %w", err)
	}

	minBridgingFee := chainConfig.GetMinBridgingFee(hasTokens)

	if amountSum.Cmp(new(big.Int).SetUint64(minBridgingFee+calculatedMinUtxo)) == -1 {
		return fmt.Errorf(
			"sum of amounts to the bridging address: %v is less than the minimum required for refund: %v",
			amountSum, minBridgingFee+calculatedMinUtxo)
	}

	return nil
}

func calculateMinUtxoForRefund(
	config *cCore.CardanoChainConfig, tx *core.CardanoTx,
	receiverAddr string, bridgingAddresses []string,
	chainInfos map[string]*cChain.CardanoChainInfo,
) (uint64, error) {
	builder, err := cardanowallet.NewTxBuilder(cardanowallet.ResolveCardanoCliBinary(config.NetworkID))
	if err != nil {
		return 0, err
	}

	defer builder.Dispose()

	chainInfo, exists := chainInfos[config.ChainID]
	if !exists {
		return 0, fmt.Errorf("chain info for chainID: %s, not found", config.ChainID)
	}

	builder.SetProtocolParameters(chainInfo.ProtocolParams)

	tokenNameToAmount := make(map[string]uint64)

	for _, out := range tx.Outputs {
		if !slices.Contains(bridgingAddresses, out.Address) {
			continue
		}

		for _, tok := range out.Tokens {
			tokenNameToAmount[tok.TokenName()] += tok.Amount
		}
	}

	tokens := make([]cardanowallet.TokenAmount, 0, len(tokenNameToAmount))

	for name, amount := range tokenNameToAmount {
		tok, err := cardanowallet.NewTokenWithFullNameTry(name)
		if err != nil {
			return 0, fmt.Errorf("failed to create Token. err: %w", err)
		}

		tokens = append(
			tokens,
			cardanowallet.NewTokenAmount(tok, amount),
		)
	}

	potentialTokenCost, err := cardanowallet.GetMinUtxoForSumMap(
		builder,
		receiverAddr,
		cardanowallet.GetTokensSumMap(tokens...),
		nil,
	)
	if err != nil {
		return 0, err
	}

	return max(config.UtxoMinAmount, potentialTokenCost), nil
}

func (p *RefundRequestProcessorImpl) getSenderAddr(
	config *cCore.CardanoChainConfig, metadata *common.RefundBridgingRequestMetadata,
) (string, error) {
	senderAddr := strings.Join(metadata.SenderAddr, "")

	if valid := cardanotx.IsValidOutputAddress(senderAddr, config.NetworkID); !valid {
		return "", fmt.Errorf("invalid sender addr: %s", senderAddr)
	}

	return senderAddr, nil
}
