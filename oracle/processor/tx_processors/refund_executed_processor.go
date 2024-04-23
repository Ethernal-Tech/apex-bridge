package tx_processors

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
)

var _ core.CardanoTxProcessor = (*RefundExecutedProcessorImpl)(nil)

type RefundExecutedProcessorImpl struct {
}

func NewRefundExecutedProcessor() *RefundExecutedProcessorImpl {
	return &RefundExecutedProcessorImpl{}
}

func (*RefundExecutedProcessorImpl) GetType() core.TxProcessorType {
	return core.TxProcessorTypeRefundExecuted
}

func (*RefundExecutedProcessorImpl) IsTxRelevant(tx *core.CardanoTx) (bool, error) {
	metadata, err := common.UnmarshalMetadata[common.BaseMetadata](common.MetadataEncodingTypeCbor, tx.Metadata)

	if err == nil && metadata != nil {
		return metadata.BridgingTxType == common.BridgingTxTypeRefundExecution, err
	}

	return false, err
}

func (p *RefundExecutedProcessorImpl) ValidateAndAddClaim(claims *core.BridgeClaims, tx *core.CardanoTx, appConfig *core.AppConfig) error {
	relevant, err := p.IsTxRelevant(tx)
	if err != nil {
		return err
	}

	if !relevant {
		return fmt.Errorf("ValidateAndAddClaim called for irrelevant tx: %v", tx)
	}

	metadata, err := common.UnmarshalMetadata[common.RefundExecutedMetadata](common.MetadataEncodingTypeCbor, tx.Metadata)
	if err != nil {
		return fmt.Errorf("failed to unmarshal metadata: tx: %v, err: %w", tx, err)
	}

	if err := p.validate(tx, metadata, appConfig); err != nil {
		return fmt.Errorf("validation failed for tx: %v, err: %w", tx, err)
	}

	// TODO: Refund
	p.addRefundExecutedClaim(claims, tx, metadata)

	return nil
}

func (*RefundExecutedProcessorImpl) addRefundExecutedClaim(claims *core.BridgeClaims, tx *core.CardanoTx, metadata *common.RefundExecutedMetadata) {
	/*
		// implement logic for creating a claim for tx
		claim := core.RefundExecutedClaim{}

		claims.RefundExecuted = append(claims.RefundExecuted, claim)
	*/
}

func (*RefundExecutedProcessorImpl) validate(tx *core.CardanoTx, metadata *common.RefundExecutedMetadata, appConfig *core.AppConfig) error {
	// implement validating the tx for this specific claim if it is needed
	// once we figure out the structure of metadata and how the refund is applied
	return nil
}
