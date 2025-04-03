package successtxprocessors

import (
	"fmt"
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/common"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_eth/core"
	goEthCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-hclog"
)

var _ core.EthTxSuccessProcessor = (*BridgingRequestedProcessorImpl)(nil)

type RefundRequestProcessorImpl struct {
	logger hclog.Logger
}

func NewRefundRequestProcessor(logger hclog.Logger) *RefundRequestProcessorImpl {
	return &RefundRequestProcessorImpl{
		logger: logger.Named("refund_request_processor"),
	}
}

func (*RefundRequestProcessorImpl) GetType() common.BridgingTxType {
	return common.TxTypeRefundRequest
}

func (*RefundRequestProcessorImpl) PreValidate(tx *core.EthTx, appConfig *cCore.AppConfig) error {
	return nil
}

func (p *RefundRequestProcessorImpl) ValidateAndAddClaim(
	claims *cCore.BridgeClaims, tx *core.EthTx, appConfig *cCore.AppConfig,
) error {
	if !appConfig.RefundEnabled {
		return fmt.Errorf("refund is not enabled")
	}

	metadata, err := core.UnmarshalEthMetadata[core.RefundBridgingRequestEthMetadata](
		tx.Metadata)
	if err != nil {
		return fmt.Errorf("failed to unmarshal metadata: tx: %v, err: %w", tx, err)
	}

	if err := p.validate(tx, metadata, appConfig); err != nil {
		return fmt.Errorf("refund validation failed for tx: %v, err: %w", tx, err)
	}

	p.addRefundRequestClaim(claims, tx, metadata)

	return nil
}

func (p *RefundRequestProcessorImpl) addRefundRequestClaim(
	claims *cCore.BridgeClaims, tx *core.EthTx,
	metadata *core.RefundBridgingRequestEthMetadata,
) {
	claim := cCore.RefundRequestClaim{
		OriginChainId:            common.ToNumChainID(tx.OriginChainID),
		OriginTransactionHash:    tx.Hash,
		OriginSenderAddress:      metadata.SenderAddr,
		OriginAmount:             tx.Value,
		OutputIndexes:            []byte{},
		ShouldDecrementHotWallet: tx.BatchTryCount > 0,
		RetryCounter:             uint64(tx.RefundTryCount),
	}

	claims.RefundRequestClaims = append(claims.RefundRequestClaims, claim)

	p.logger.Info("Added RefundRequestClaim",
		"txHash", tx.Hash, "claim", cCore.RefundRequestClaimString(claim))
}

func (p *RefundRequestProcessorImpl) validate(
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

	if tx.Value.Cmp(new(big.Int).SetUint64(chainConfig.MinFeeForBridging)) != 1 {
		return fmt.Errorf(
			"tx.Value: %v is less than the minimum required for refund: %v",
			tx.Value, chainConfig.MinFeeForBridging+1)
	}

	return nil
}
