package successtxprocessors

import (
	"fmt"
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/chain"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
)

const (
	unknownNativeTokensUtxoCntMax = 3
)

var _ core.CardanoTxSuccessProcessor = (*BridgingRequestedProcessorImpl)(nil)

type RefundRequestProcessorImpl struct {
	logger     hclog.Logger
	chainInfos map[string]*chain.CardanoChainInfo
}

func NewRefundRequestProcessor(
	logger hclog.Logger, chainInfos map[string]*chain.CardanoChainInfo,
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

func (p *RefundRequestProcessorImpl) ValidateAndAddClaim(
	claims *cCore.BridgeClaims, tx *core.CardanoTx, appConfig *cCore.AppConfig,
) error {
	if err := p.validate(tx, appConfig); err != nil {
		return fmt.Errorf("refund validation failed for tx: %v, err: %w", tx, err)
	}

	p.addRefundRequestClaim(claims, tx, appConfig)

	return nil
}

func (p *RefundRequestProcessorImpl) addRefundRequestClaim(
	claims *cCore.BridgeClaims, tx *core.CardanoTx, appConfig *cCore.AppConfig,
) {
	chainConfig := appConfig.CardanoChains[tx.OriginChainID]
	receiverAddr := p.findReceiverAddr(chainConfig, tx)
	amount := big.NewInt(0)
	unknownTokenOutputIndexes := make([]int, 0)

	for idx, out := range tx.Outputs {
		if out.Address != chainConfig.BridgingAddresses.BridgingAddress {
			continue
		}

		if len(out.Tokens) > 0 {
			unknownTokenOutputIndexes = append(unknownTokenOutputIndexes, idx)
		}

		if len(unknownTokenOutputIndexes) == 0 {
			amount.Add(amount, new(big.Int).SetUint64(out.Amount))
		}
	}

	claim := cCore.RefundRequestClaim{
		OriginChainId:            common.ToNumChainID(tx.OriginChainID),
		OriginTransactionHash:    tx.Hash,
		OriginSenderAddress:      receiverAddr,
		OriginAmount:             amount,
		OutputIndexes:            common.PackNumbersToBytes(unknownTokenOutputIndexes),
		ShouldDecrementHotWallet: tx.BatchTryCount > 0,
		RetryCounter:             uint64(tx.RefundTryCount),
	}

	claims.RefundRequestClaims = append(claims.RefundRequestClaims, claim)

	p.logger.Info("Added RefundRequestClaim",
		"txHash", tx.Hash, "claim", cCore.RefundRequestClaimString(claim))
}

func (p *RefundRequestProcessorImpl) validate(
	tx *core.CardanoTx, appConfig *cCore.AppConfig,
) error {
	if tx.RefundTryCount > appConfig.TryCountLimits.MaxRefundTryCount {
		return fmt.Errorf("try count exceeded. RefundTryCount: (current, max)=(%d, %d)",
			tx.RefundTryCount, appConfig.TryCountLimits.MaxRefundTryCount)
	}

	chainConfig := appConfig.CardanoChains[tx.OriginChainID]
	if chainConfig == nil {
		return fmt.Errorf("unsupported chain id found in tx. chain id: %v", tx.OriginChainID)
	}

	amountSum := big.NewInt(0)
	unknownNativeTokensUtxoCnt := uint(0)

	for _, out := range tx.Outputs {
		if out.Address != chainConfig.BridgingAddresses.BridgingAddress {
			continue
		}

		amountSum.Add(amountSum, new(big.Int).SetUint64(out.Amount))

		if len(out.Tokens) > 0 {
			unknownNativeTokensUtxoCnt++

			if unknownNativeTokensUtxoCnt > unknownNativeTokensUtxoCntMax {
				return fmt.Errorf("more UTxOs with unknown tokens than allowed. max: %d", unknownNativeTokensUtxoCntMax)
			}
		}
	}

	calculatedMinUtxo, err := p.calculateMinUtxoForRefund(chainConfig, tx)
	if err != nil {
		return fmt.Errorf("failed to calculate min utxo. err: %w", err)
	}

	if amountSum.Cmp(new(big.Int).SetUint64(chainConfig.UtxoMinAmount+calculatedMinUtxo)) == -1 {
		return fmt.Errorf(
			"sum of amounts to the bridging address: %v is less than minimum required for refund: %v",
			amountSum, chainConfig.UtxoMinAmount+calculatedMinUtxo)
	}

	return nil
}

func (p *RefundRequestProcessorImpl) calculateMinUtxoForRefund(
	config *cCore.CardanoChainConfig, tx *core.CardanoTx,
) (uint64, error) {
	builder, err := cardanowallet.NewTxBuilder(cardanowallet.ResolveCardanoCliBinary(config.NetworkID))
	if err != nil {
		return 0, err
	}

	defer builder.Dispose()

	chainInfo, exists := p.chainInfos[config.ChainID]
	if !exists {
		return 0, fmt.Errorf("chain info for chainID: %s, not found", config.ChainID)
	}

	builder.SetProtocolParameters(chainInfo.ProtocolParams)

	receiverAddr := p.findReceiverAddr(config, tx)
	tokenNameToAmount := make(map[string]uint64)

	for _, out := range tx.Outputs {
		if out.Address != config.BridgingAddresses.BridgingAddress {
			continue
		}

		for _, tok := range out.Tokens {
			tokenNameToAmount[tok.TokenName()] += tok.Amount
		}
	}

	tokens := make([]cardanowallet.TokenAmount, 0, len(tokenNameToAmount))

	for name, amount := range tokenNameToAmount {
		tok, err := cardanowallet.NewTokenAmountWithFullName(name, amount, true)
		if err != nil {
			return 0, fmt.Errorf("failed to create TokenAmount. err: %w", err)
		}

		tokens = append(tokens, tok)
	}

	potentialTokenCost, err := cardanowallet.GetTokenCostSum(
		builder, receiverAddr, []cardanowallet.Utxo{
			{
				Amount: 0,
				Tokens: tokens,
			},
		},
	)
	if err != nil {
		return 0, err
	}

	return max(config.UtxoMinAmount, potentialTokenCost), nil
}

func (p *RefundRequestProcessorImpl) findReceiverAddr(
	config *cCore.CardanoChainConfig, tx *core.CardanoTx,
) string {
	return ""
}
