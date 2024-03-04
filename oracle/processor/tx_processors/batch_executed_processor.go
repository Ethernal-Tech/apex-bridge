package tx_processors

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
)

var _ core.CardanoTxProcessor = (*BatchExecutedProcessorImpl)(nil)

type BatchExecutedProcessorImpl struct {
}

func NewBatchExecutedProcessor() *BatchExecutedProcessorImpl {
	return &BatchExecutedProcessorImpl{}
}

func (*BatchExecutedProcessorImpl) IsTxRelevant(tx *core.CardanoTx, appConfig *core.AppConfig) (bool, error) {
	metadata, err := core.UnmarshalBaseMetadata(tx.Metadata)

	if err == nil && metadata != nil {
		return metadata.BridgingTxType == core.BridgingTxTypeBatchExecution, err
	}

	return false, err
}

func (p *BatchExecutedProcessorImpl) ValidateAndAddClaim(claims *core.BridgeClaims, tx *core.CardanoTx, appConfig *core.AppConfig) error {
	relevant, err := p.IsTxRelevant(tx, appConfig)
	if err != nil {
		return err
	}

	if !relevant {
		return fmt.Errorf("ValidateAndAddClaim called for irrelevant tx: %v", tx)
	}

	metadata, err := core.UnmarshalBridgingRequestMetadata(tx.Metadata)
	if err != nil {
		return fmt.Errorf("failed to unmarshal metadata: tx: %v,\n err: %v", tx, err)
	}

	if err := p.validate(tx, metadata, appConfig); err != nil {
		return fmt.Errorf("validation failed for tx: %v", tx)
	}

	// TODO: implement logic for creating a claim for tx
	claim := core.BatchExecutedClaim{}

	claims.BatchExecuted = append(claims.BatchExecuted, claim)

	return nil
}

func (*BatchExecutedProcessorImpl) validate(tx *core.CardanoTx, metadata *core.BridgingRequestMetadata, appConfig *core.AppConfig) error {
	// TODO: implement validating the tx for this specific claim
	return nil
}
