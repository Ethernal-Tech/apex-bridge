package txprocessors

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle/utils"
	"github.com/hashicorp/go-hclog"
)

var _ core.CardanoTxProcessor = (*BatchExecutedProcessorImpl)(nil)

type BatchExecutedProcessorImpl struct {
	logger hclog.Logger
}

func NewBatchExecutedProcessor(logger hclog.Logger) *BatchExecutedProcessorImpl {
	return &BatchExecutedProcessorImpl{
		logger: logger.Named("batch_executed_processor"),
	}
}

func (*BatchExecutedProcessorImpl) GetType() common.BridgingTxType {
	return common.BridgingTxTypeBatchExecution
}

func (p *BatchExecutedProcessorImpl) ValidateAndAddClaim(
	claims *core.BridgeClaims, tx *core.CardanoTx, appConfig *core.AppConfig,
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

	claims.BatchExecutedClaims = append(claims.BatchExecutedClaims, core.BatchExecutedClaim{
		ObservedTransactionHash: tx.Hash,
		ChainId:                 common.ToNumChainID(tx.OriginChainID),
		BatchNonceId:            metadata.BatchNonceID,
	})

	p.logger.Info("Added BatchExecutedClaim",
		"txHash", tx.Hash, "chain", tx.OriginChainID, "BatchNonceId", metadata.BatchNonceID)

	return nil
}

func (*BatchExecutedProcessorImpl) validate(
	tx *core.CardanoTx, metadata *common.BatchExecutedMetadata, appConfig *core.AppConfig,
) error {
	// after BridgingTxType and inputs are validated, no further validation needed
	return utils.ValidateTxInputs(tx, appConfig)
}
