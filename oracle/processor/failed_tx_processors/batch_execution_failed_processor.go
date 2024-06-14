package failedtxprocessors

import (
	"fmt"

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

func (*BatchExecutionFailedProcessorImpl) GetType() common.BridgingTxType {
	return common.BridgingTxTypeBatchExecution
}

func (p *BatchExecutionFailedProcessorImpl) ValidateAndAddClaim(
	claims *core.BridgeClaims, tx *core.BridgeExpectedCardanoTx, appConfig *core.AppConfig,
) error {
	metadata, err := common.UnmarshalMetadata[common.BatchExecutedMetadata](common.MetadataEncodingTypeCbor, tx.Metadata)
	if err != nil {
		return fmt.Errorf("failed to unmarshal metadata: tx: %v, err: %w", tx, err)
	}

	if metadata.BridgingTxType != p.GetType() {
		return fmt.Errorf("ValidateAndAddClaim called for irrelevant tx: %v", tx)
	}

	p.logger.Debug("Validating relevant tx", "txHash", tx.Hash, "metadata", metadata)

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
		ObservedTransactionHash: common.MustHashToBytes32(tx.Hash),
		ChainId:                 common.ToNumChainID(tx.ChainID),
		BatchNonceId:            metadata.BatchNonceID,
	}

	claims.BatchExecutionFailedClaims = append(claims.BatchExecutionFailedClaims, claim)

	p.logger.Info("Added BatchExecutionFailedClaim",
		"txHash", tx.Hash, "metadata", metadata, "claim", core.BatchExecutionFailedClaimString(claim))
}

func (*BatchExecutionFailedProcessorImpl) validate(
	tx *core.BridgeExpectedCardanoTx, metadata *common.BatchExecutedMetadata, appConfig *core.AppConfig,
) error {
	// no validation needed
	return nil
}
