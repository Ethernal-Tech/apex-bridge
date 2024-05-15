package failedtxprocessors

import (
	"fmt"
	"math/big"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/hashicorp/go-hclog"
)

var _ core.CardanoTxFailedProcessor = (*BatchExecutionFailedProcessorImpl)(nil)

type BatchExecutionFailedProcessorImpl struct {
	logger hclog.Logger
}

func NewBatchExecutionFailedProcessor(logger hclog.Logger) *BatchExecutionFailedProcessorImpl {
	return &BatchExecutionFailedProcessorImpl{
		logger: logger.Named("batch_execution_failed_processor"),
	}
}

func (*BatchExecutionFailedProcessorImpl) GetType() core.TxProcessorType {
	return core.TxProcessorTypeBatchExecuted
}

func (p *BatchExecutionFailedProcessorImpl) IsTxRelevant(tx *core.BridgeExpectedCardanoTx) (bool, error) {
	p.logger.Debug("Checking if tx is relevant", "tx", tx)

	metadata, err := common.UnmarshalMetadata[common.BaseMetadata](common.MetadataEncodingTypeCbor, tx.Metadata)

	p.logger.Debug("Unmarshaled metadata", "txHash", tx.Hash, "metadata", metadata, "err", err)

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

	p.logger.Debug("tx is relevant", "txHash", tx.Hash)

	metadata, err := common.UnmarshalMetadata[common.BatchExecutedMetadata](common.MetadataEncodingTypeCbor, tx.Metadata)
	if err != nil {
		return fmt.Errorf("failed to unmarshal metadata: tx: %v, err: %w", tx, err)
	}

	p.logger.Debug("Validating", "txHash", tx.Hash, "metadata", metadata)

	if err := p.validate(tx, metadata, appConfig); err != nil {
		return fmt.Errorf("validation failed for tx: %v, err: %w", tx, err)
	}

	p.addBatchExecutionFailedClaim(claims, tx, metadata)

	return nil
}

func (p *BatchExecutionFailedProcessorImpl) addBatchExecutionFailedClaim(
	claims *core.BridgeClaims, tx *core.BridgeExpectedCardanoTx, metadata *common.BatchExecutedMetadata,
) {
	claim := core.BatchExecutionFailedClaim{
		ObservedTransactionHash: tx.Hash,
		ChainID:                 tx.ChainID,
		BatchNonceID:            new(big.Int).SetUint64(metadata.BatchNonceID),
	}

	claims.BatchExecutionFailedClaims = append(claims.BatchExecutionFailedClaims, claim)

	p.logger.Info("Added BatchExecutionFailedClaim", "txHash", tx.Hash, "metadata", metadata, "claim", claim)
}

func (*BatchExecutionFailedProcessorImpl) validate(
	tx *core.BridgeExpectedCardanoTx, metadata *common.BatchExecutedMetadata, appConfig *core.AppConfig,
) error {
	// no validation needed
	return nil
}
