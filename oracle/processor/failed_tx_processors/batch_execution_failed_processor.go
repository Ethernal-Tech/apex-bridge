package failed_tx_processors

import (
	"fmt"
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
)

var _ core.CardanoTxFailedProcessor = (*BatchExecutionFailedProcessorImpl)(nil)

type BatchExecutionFailedProcessorImpl struct {
}

func NewBatchExecutionFailedProcessor() *BatchExecutionFailedProcessorImpl {
	return &BatchExecutionFailedProcessorImpl{}
}

func (*BatchExecutionFailedProcessorImpl) GetType() core.TxProcessorType {
	return core.TxProcessorTypeBatchExecuted
}

func (*BatchExecutionFailedProcessorImpl) IsTxRelevant(tx *core.BridgeExpectedCardanoTx) (bool, error) {
	metadata, err := common.UnmarshalMetadata[common.BaseMetadata](common.MetadataEncodingTypeCbor, tx.Metadata)

	if err == nil && metadata != nil {
		return metadata.BridgingTxType == common.BridgingTxTypeBatchExecution, err
	}

	return false, err
}

func (p *BatchExecutionFailedProcessorImpl) ValidateAndAddClaim(
	claims *core.BridgeClaims, tx *core.BridgeExpectedCardanoTx, appConfig *core.AppConfig,
) error {
	relevant, err := p.IsTxRelevant(tx)
	if err != nil {
		return err
	}

	if !relevant {
		return fmt.Errorf("ValidateAndAddClaim called for irrelevant tx: %v", tx)
	}

	metadata, err := common.UnmarshalMetadata[common.BatchExecutedMetadata](common.MetadataEncodingTypeCbor, tx.Metadata)
	if err != nil {
		return fmt.Errorf("failed to unmarshal metadata: tx: %v, err: %w", tx, err)
	}

	if err := p.validate(tx, metadata, appConfig); err != nil {
		return fmt.Errorf("validation failed for tx: %v, err: %w", tx, err)
	}

	p.addBatchExecutionFailedClaim(claims, tx, metadata)

	return nil
}

func (*BatchExecutionFailedProcessorImpl) addBatchExecutionFailedClaim(
	claims *core.BridgeClaims, tx *core.BridgeExpectedCardanoTx, metadata *common.BatchExecutedMetadata,
) {
	claims.BatchExecutionFailedClaims = append(claims.BatchExecutionFailedClaims, core.BatchExecutionFailedClaim{
		ObservedTransactionHash: tx.Hash,
		ChainID:                 tx.ChainId,
		BatchNonceID:            new(big.Int).SetUint64(metadata.BatchNonceId),
	})
}

func (*BatchExecutionFailedProcessorImpl) validate(tx *core.BridgeExpectedCardanoTx, metadata *common.BatchExecutedMetadata, appConfig *core.AppConfig) error {
	// no validation needed
	return nil
}
