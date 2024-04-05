package failed_tx_processors

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
)

var _ core.CardanoTxFailedProcessor = (*RefundExecutionFailedProcessorImpl)(nil)

type RefundExecutionFailedProcessorImpl struct {
}

func NewRefundExecutionFailedProcessor() *RefundExecutionFailedProcessorImpl {
	return &RefundExecutionFailedProcessorImpl{}
}

func (*RefundExecutionFailedProcessorImpl) IsTxRelevant(tx *core.BridgeExpectedCardanoTx, appConfig *core.AppConfig) (bool, error) {
	metadata, err := core.UnmarshalBaseMetadata(tx.Metadata)

	if err == nil && metadata != nil {
		return metadata.BridgingTxType == core.BridgingTxTypeRefundExecution, err
	}

	return false, err
}

func (p *RefundExecutionFailedProcessorImpl) ValidateAndAddClaim(claims *core.BridgeClaims, tx *core.BridgeExpectedCardanoTx, appConfig *core.AppConfig) error {
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
	p.addRefundRequestClaim(claims, tx, metadata)

	return nil
}

func (*RefundExecutionFailedProcessorImpl) addRefundRequestClaim(claims *core.BridgeClaims, tx *core.BridgeExpectedCardanoTx, metadata *core.RefundExecutedMetadata) {
	/*
		// implement logic for creating a claim for tx
		claim := core.RefundRequestClaim{}

		claims.RefundRequest = append(claims.RefundRequest, claim)
	*/
}

func (*RefundExecutionFailedProcessorImpl) validate(tx *core.BridgeExpectedCardanoTx, metadata *core.RefundExecutedMetadata, appConfig *core.AppConfig) error {
	// implement validating the tx for this specific claim if it is needed
	// once we figure out the structure of metadata and how the refund is applied
	return nil
}
