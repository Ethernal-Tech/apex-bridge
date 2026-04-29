package failedtxprocessors

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_solana/core"
	"github.com/hashicorp/go-hclog"
)

var _ core.SolanaTxFailedProcessor = (*BatchExecutionFailedProcessorImpl)(nil)

type BatchExecutionFailedProcessorImpl struct {
	logger hclog.Logger
}

func NewSolanaBatchExecutionFailedProcessor(logger hclog.Logger) *BatchExecutionFailedProcessorImpl {
	return &BatchExecutionFailedProcessorImpl{
		logger: logger.Named("solana_batch_execution_failed_processor"),
	}
}

func (*BatchExecutionFailedProcessorImpl) GetType() common.BridgingTxType {
	return common.BridgingTxTypeBatchExecution
}

func (*BatchExecutionFailedProcessorImpl) PreValidate(
	tx *core.BridgeExpectedSolanaTx, appConfig *oCore.AppConfig,
) error {
	return nil
}

func (p *BatchExecutionFailedProcessorImpl) ValidateAndAddClaim(
	claims *oCore.BridgeClaims, tx *core.BridgeExpectedSolanaTx, appConfig *oCore.AppConfig,
) error {
	metadata, err := core.UnmarshalSolMetadata[core.BatchExecutedSolMetadata](tx.Metadata)
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

	p.addBatchExecutionFailedClaim(claims, tx, metadata, appConfig.ChainIDConverter)

	return nil
}

func (p *BatchExecutionFailedProcessorImpl) addBatchExecutionFailedClaim(
	claims *oCore.BridgeClaims, tx *core.BridgeExpectedSolanaTx,
	metadata *core.BatchExecutedSolMetadata, chainIDConverter *common.ChainIDConverter,
) {
	claim := oCore.BatchExecutionFailedClaim{
		ObservedTransactionHash: tx.Hash[:],
		ChainId:                 chainIDConverter.ToChainIDNum(tx.ChainID),
		BatchNonceId:            metadata.BatchNonceID,
	}

	claims.BatchExecutionFailedClaims = append(claims.BatchExecutionFailedClaims, claim)

	p.logger.Info("Added BatchExecutionFailedClaim",
		"txHash", tx.Hash, "metadata", metadata, "claim", oCore.BatchExecutionFailedClaimString(claim, chainIDConverter))
}

func (*BatchExecutionFailedProcessorImpl) validate(
	_ *core.BridgeExpectedSolanaTx, _ *core.BatchExecutedSolMetadata, _ *oCore.AppConfig,
) error {
	return nil
}
