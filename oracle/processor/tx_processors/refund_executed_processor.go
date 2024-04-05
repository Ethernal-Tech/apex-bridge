package tx_processors

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
)

var _ core.CardanoTxProcessor = (*RefundExecutedProcessorImpl)(nil)

type RefundExecutedProcessorImpl struct {
}

func NewRefundExecutedProcessor() *RefundExecutedProcessorImpl {
	return &RefundExecutedProcessorImpl{}
}

func (*RefundExecutedProcessorImpl) IsTxRelevant(tx *core.CardanoTx, appConfig *core.AppConfig) (bool, error) {
	metadata, err := core.UnmarshalBaseMetadata(tx.Metadata)

	if err == nil && metadata != nil {
		return metadata.BridgingTxType == core.BridgingTxTypeRefundExecution, err
	}

	return false, err
}

func (p *RefundExecutedProcessorImpl) ValidateAndAddClaim(claims *core.BridgeClaims, tx *core.CardanoTx, appConfig *core.AppConfig) error {
	relevant, err := p.IsTxRelevant(tx, appConfig)
	if err != nil {
		return err
	}

	if !relevant {
		return fmt.Errorf("ValidateAndAddClaim called for irrelevant tx: %v", tx)
	}

	metadata, err := core.UnmarshalRefundExecutedMetadata(tx.Metadata)
	if err != nil {
		return fmt.Errorf("failed to unmarshal metadata: tx: %v,\n err: %v", tx, err)
	}

	if err := p.validate(tx, metadata, appConfig); err != nil {
		return fmt.Errorf("validation failed for tx: %v, err: %v", tx, err)
	}

	// TODO: Refund
	p.addRefundExecutedClaim(claims, tx, metadata)

	return nil
}

func (*RefundExecutedProcessorImpl) addRefundExecutedClaim(claims *core.BridgeClaims, tx *core.CardanoTx, metadata *core.RefundExecutedMetadata) {
	/*
		// implement logic for creating a claim for tx
		claim := core.RefundExecutedClaim{}

		claims.RefundExecuted = append(claims.RefundExecuted, claim)
	*/
}

func (*RefundExecutedProcessorImpl) validate(tx *core.CardanoTx, metadata *core.RefundExecutedMetadata, appConfig *core.AppConfig) error {
	// implement validating the tx for this specific claim if it is needed
	// once we figure out the structure of metadata and how the refund is applied
	return nil
}
