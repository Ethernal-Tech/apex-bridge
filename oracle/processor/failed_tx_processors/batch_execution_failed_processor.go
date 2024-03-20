package failed_tx_processors

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
)

var _ core.CardanoTxFailedProcessor = (*BatchExecutionFailedProcessorImpl)(nil)

type BatchExecutionFailedProcessorImpl struct {
}

func NewBatchExecutionFailedProcessor() *BatchExecutionFailedProcessorImpl {
	return &BatchExecutionFailedProcessorImpl{}
}

func (*BatchExecutionFailedProcessorImpl) IsTxRelevant(tx *core.BridgeExpectedCardanoTx, appConfig *core.AppConfig) (bool, error) {
	metadata, err := core.UnmarshalBaseMetadata(tx.Metadata)

	if err == nil && metadata != nil {
		return metadata.BridgingTxType == core.BridgingTxTypeBatchExecution, err
	}

	return false, err
}

func (p *BatchExecutionFailedProcessorImpl) ValidateAndAddClaim(claims *core.BridgeClaims, tx *core.BridgeExpectedCardanoTx, appConfig *core.AppConfig) error {
	relevant, err := p.IsTxRelevant(tx, appConfig)
	if err != nil {
		return err
	}

	if !relevant {
		return fmt.Errorf("ValidateAndAddClaim called for irrelevant tx: %v", tx)
	}

	metadata, err := core.UnmarshalBatchExecutedMetadata(tx.Metadata)
	if err != nil {
		return fmt.Errorf("failed to unmarshal metadata: tx: %v,\n err: %v", tx, err)
	}

	if err := p.validate(tx, metadata, appConfig); err != nil {
		return fmt.Errorf("validation failed for tx: %v, err: %v", tx, err)
	}

	p.addBatchExecutionFailedClaim(claims, tx, metadata)

	return nil
}

func (*BatchExecutionFailedProcessorImpl) addBatchExecutionFailedClaim(claims *core.BridgeClaims, tx *core.BridgeExpectedCardanoTx, metadata *core.BatchExecutedMetadata) {
	claim := core.BatchExecutionFailedClaim{
		TxHash:       tx.Hash,
		BatchNonceId: metadata.BatchNonceId,
	}

	claims.BatchExecutionFailed = append(claims.BatchExecutionFailed, claim)
}

func (*BatchExecutionFailedProcessorImpl) validate(tx *core.BridgeExpectedCardanoTx, metadata *core.BatchExecutedMetadata, appConfig *core.AppConfig) error {
	// TODO: implement validating the tx for this specific claim if it is needed
	// once we figure out the structure of metadata and how the batch is applied
	// to destination chain
	return nil
}