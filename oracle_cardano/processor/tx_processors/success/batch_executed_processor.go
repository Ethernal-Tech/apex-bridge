package successtxprocessors

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/utils"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/hashicorp/go-hclog"
)

var _ core.CardanoTxSuccessProcessor = (*BatchExecutedProcessorImpl)(nil)

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

func (*BatchExecutedProcessorImpl) PreValidate(tx *core.CardanoTx, appConfig *cCore.AppConfig) error {
	return nil
}

func (p *BatchExecutedProcessorImpl) ValidateAndAddClaim(
	claims *cCore.BridgeClaims, tx *core.CardanoTx, appConfig *cCore.AppConfig,
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

	claims.BatchExecutedClaims = append(claims.BatchExecutedClaims, cCore.BatchExecutedClaim{
		ObservedTransactionHash: tx.Hash,
		ChainId:                 appConfig.ChainIDConverter.ToNumChainID(tx.OriginChainID),
		BatchNonceId:            metadata.BatchNonceID,
	})

	p.logger.Info("Added BatchExecutedClaim",
		"txHash", tx.Hash, "chain", tx.OriginChainID, "BatchNonceId", metadata.BatchNonceID)

	return nil
}

func (*BatchExecutedProcessorImpl) validate(
	tx *core.CardanoTx, _ *common.BatchExecutedMetadata, appConfig *cCore.AppConfig,
) error {
	// after BridgingTxType and inputs are validated, no further validation needed
	return utils.ValidateTxInputs(tx, appConfig)
}
