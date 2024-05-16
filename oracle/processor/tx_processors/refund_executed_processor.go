package txprocessors

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/hashicorp/go-hclog"
)

var _ core.CardanoTxProcessor = (*RefundExecutedProcessorImpl)(nil)

type RefundExecutedProcessorImpl struct {
	logger hclog.Logger
}

func NewRefundExecutedProcessor(logger hclog.Logger) *RefundExecutedProcessorImpl {
	return &RefundExecutedProcessorImpl{
		logger: logger.Named("refund_executed_processor"),
	}
}

func (*RefundExecutedProcessorImpl) GetType() core.TxProcessorType {
	return core.TxProcessorTypeRefundExecuted
}

func (p *RefundExecutedProcessorImpl) IsTxRelevant(tx *core.CardanoTx) (bool, error) {
	p.logger.Debug("Checking if tx is relevant", "tx", tx)

	metadata, err := common.UnmarshalMetadata[common.BaseMetadata](common.MetadataEncodingTypeCbor, tx.Metadata)

	p.logger.Debug("Unmarshaled metadata", "txHash", tx.Hash, "metadata", metadata, "err", err)

	if err == nil && metadata != nil {
		return metadata.BridgingTxType == common.BridgingTxTypeRefundExecution, err
	}

	return false, err
}

func (p *RefundExecutedProcessorImpl) ValidateAndAddClaim(
	claims *core.BridgeClaims, tx *core.CardanoTx, appConfig *core.AppConfig,
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
	p.addRefundExecutedClaim(claims, tx, metadata)

	return nil
}

func (*RefundExecutedProcessorImpl) addRefundExecutedClaim(
	claims *core.BridgeClaims, tx *core.CardanoTx, metadata *common.RefundExecutedMetadata,
) {
	/*
		// implement logic for creating a claim for tx
		claim := core.RefundExecutedClaim{}

		claims.RefundExecuted = append(claims.RefundExecuted, claim)

		p.logger.Info("Added RefundExecutedClaim",
		"txHash", tx.Hash, "metadata", metadata, "claim", core.RefundExecutedClaimString(claim))
	*/
}

func (*RefundExecutedProcessorImpl) validate(
	tx *core.CardanoTx, metadata *common.RefundExecutedMetadata, appConfig *core.AppConfig,
) error {
	// implement validating the tx for this specific claim if it is needed
	// once we figure out the structure of metadata and how the refund is applied
	return nil
}
