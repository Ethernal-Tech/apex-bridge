package failedtxprocessors

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
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

func (*RefundExecutionFailedProcessorImpl) GetType() core.TxProcessorType {
	return core.TxProcessorTypeRefundExecuted
}

func (p *RefundExecutionFailedProcessorImpl) IsTxRelevant(tx *core.BridgeExpectedCardanoTx) (bool, error) {
	p.logger.Debug("Checking if tx is relevant", "tx", tx)

	metadata, err := common.UnmarshalMetadata[common.BaseMetadata](common.MetadataEncodingTypeCbor, tx.Metadata)

	p.logger.Debug("Unmarshaled metadata", "txHash", tx.Hash, "metadata", metadata, "err", err)

	if err == nil && metadata != nil {
		return metadata.BridgingTxType == common.BridgingTxTypeRefundExecution, err
	}

	return false, err
}

func (p *RefundExecutionFailedProcessorImpl) ValidateAndAddClaim(
	claims *core.BridgeClaims, tx *core.BridgeExpectedCardanoTx, appConfig *core.AppConfig,
) error {
	relevant, err := p.IsTxRelevant(tx)
	if err != nil {
		return err
	}

	if !relevant {
		return fmt.Errorf("ValidateAndAddClaim called for irrelevant tx: %v", tx)
	}

	p.logger.Debug("tx is relevant", "txHash", tx.Hash)

	metadata, err := common.UnmarshalMetadata[common.RefundExecutedMetadata](common.MetadataEncodingTypeCbor, tx.Metadata)
	if err != nil {
		return fmt.Errorf("failed to unmarshal metadata: tx: %v, err: %w", tx, err)
	}

	p.logger.Debug("Validating", "txHash", tx.Hash, "metadata", metadata)

	if err := p.validate(tx, metadata, appConfig); err != nil {
		return fmt.Errorf("validation failed for tx: %v, err: %w", tx, err)
	}

	//nolint:godox
	// TODO: Refund
	p.addRefundRequestClaim(claims, tx, metadata)

	return nil
}

func (*RefundExecutionFailedProcessorImpl) addRefundRequestClaim(
	claims *core.BridgeClaims, tx *core.BridgeExpectedCardanoTx, metadata *common.RefundExecutedMetadata,
) {
	/*
		// implement logic for creating a claim for tx
		claim := core.RefundRequestClaim{}

		claims.RefundRequest = append(claims.RefundRequest, claim)

		p.logger.Info("Added RefundRequestClaim", "txHash", tx.Hash, "metadata", metadata, "claim", claim)
	*/
}

func (*RefundExecutionFailedProcessorImpl) validate(
	tx *core.BridgeExpectedCardanoTx, metadata *common.RefundExecutedMetadata, appConfig *core.AppConfig,
) error {
	// implement validating the tx for this specific claim if it is needed
	// once we figure out the structure of metadata and how the refund is applied
	return nil
}
