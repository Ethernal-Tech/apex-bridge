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

func (*BatchExecutionFailedProcessorImpl) IsTxRelevant(tx *core.BridgeExpectedCardanoTx, appConfig *core.AppConfig) (bool, error) {
	metadata, err := common.UnmarshalBaseMetadata(tx.Metadata)

	if err == nil && metadata != nil {
		return metadata.BridgingTxType == common.BridgingTxTypeBatchExecution, err
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

	metadata, err := common.UnmarshalBatchExecutedMetadata(tx.Metadata)
	if err != nil {
		return fmt.Errorf("failed to unmarshal metadata: tx: %v,\n err: %v", tx, err)
	}

	if err := p.validate(tx, metadata, appConfig); err != nil {
		return fmt.Errorf("validation failed for tx: %v, err: %v", tx, err)
	}

	p.addBatchExecutionFailedClaim(claims, tx, metadata)

	return nil
}

func (*BatchExecutionFailedProcessorImpl) addBatchExecutionFailedClaim(claims *core.BridgeClaims, tx *core.BridgeExpectedCardanoTx, metadata *common.BatchExecutedMetadata) {
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
