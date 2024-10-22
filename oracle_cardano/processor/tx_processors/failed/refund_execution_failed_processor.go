package failedtxprocessors

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/hashicorp/go-hclog"
)

var _ core.CardanoTxFailedProcessor = (*RefundExecutionFailedProcessorImpl)(nil)

type RefundExecutionFailedProcessorImpl struct {
	logger hclog.Logger
}

func NewRefundExecutionFailedProcessor(logger hclog.Logger) *RefundExecutionFailedProcessorImpl {
	return &RefundExecutionFailedProcessorImpl{
		logger: logger.Named("refund_execution_failed_processor"),
	}
}

func (*RefundExecutionFailedProcessorImpl) GetType() common.BridgingTxType {
	return common.BridgingTxTypeRefundExecution
}

func (*RefundExecutionFailedProcessorImpl) PreValidate(
	tx *core.BridgeExpectedCardanoTx, appConfig *cCore.AppConfig,
) error {
	return nil
}

func (p *RefundExecutionFailedProcessorImpl) ValidateAndAddClaim(
	claims *cCore.BridgeClaims, tx *core.BridgeExpectedCardanoTx, appConfig *cCore.AppConfig,
) error {
	metadata, err := common.UnmarshalMetadata[common.RefundExecutedMetadata](common.MetadataEncodingTypeCbor, tx.Metadata)
	if err != nil {
		return fmt.Errorf("failed to unmarshal metadata: tx: %v, err: %w", tx, err)
	}

	if metadata.BridgingTxType != p.GetType() {
		return fmt.Errorf("ValidateAndAddClaim called for irrelevant tx: %v", tx)
	}

	p.logger.Debug("Validating relevant tx", "txHash", tx.Hash, "metadata", metadata)

	if err := p.validate(tx, metadata, appConfig); err != nil {
		return fmt.Errorf("validation failed for tx: %v, err: %w", tx, err)
	}

	//nolint:godox
	// TODO: Refund
	p.addRefundRequestClaim(claims, tx, metadata)

	return nil
}

func (*RefundExecutionFailedProcessorImpl) addRefundRequestClaim(
	claims *cCore.BridgeClaims, tx *core.BridgeExpectedCardanoTx, metadata *common.RefundExecutedMetadata,
) {
	/*
		// implement logic for creating a claim for tx
		claim := core.RefundRequestClaim{}

		claims.RefundRequest = append(claims.RefundRequest, claim)

		p.logger.Info("Added RefundRequestClaim",
		"txHash", tx.Hash, "metadata", metadata, "claim", core.RefundRequestClaimString(claim))
	*/
}

func (*RefundExecutionFailedProcessorImpl) validate(
	_ *core.BridgeExpectedCardanoTx, _ *common.RefundExecutedMetadata, _ *cCore.AppConfig,
) error {
	// implement validating the tx for this specific claim if it is needed
	// once we figure out the structure of metadata and how the refund is applied
	return nil
}
