package successtxprocessors

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_solana/core"
	"github.com/hashicorp/go-hclog"
)

var _ core.SolanaTxSuccessProcessor = (*BatchExecutedProcessorImpl)(nil)

type BatchExecutedProcessorImpl struct {
	logger hclog.Logger
}

func NewSolanaBatchExecutedProcessor(logger hclog.Logger) *BatchExecutedProcessorImpl {
	return &BatchExecutedProcessorImpl{
		logger: logger.Named("solana_batch_executed_processor"),
	}
}

func (*BatchExecutedProcessorImpl) GetType() common.BridgingTxType {
	return common.BridgingTxTypeBatchExecution
}

func (*BatchExecutedProcessorImpl) PreValidate(tx *core.SolanaTx, appConfig *oCore.AppConfig) error {
	return nil
}

func (p *BatchExecutedProcessorImpl) ValidateAndAddClaim(
	claims *oCore.BridgeClaims, tx *core.SolanaTx, appConfig *oCore.AppConfig,
) error {
	metadata, err := core.UnmarshalSolMetadata[core.BatchExecutedSolMetadata](tx.Metadata)
	if err != nil {
		return fmt.Errorf("failed to unmarshal metadata: tx: %v, err: %w", tx, err)
	}

	if metadata.BridgingTxType != p.GetType() {
		return fmt.Errorf("ValidateAndAddClaim called for irrelevant tx: %v", tx)
	}

	p.logger.Debug("Validating relevant tx", "txHash", tx.TxSignature, "metadata", metadata)

	if err := p.validate(tx, metadata, appConfig); err != nil {
		return fmt.Errorf("validation failed for tx: %v, err: %w", tx, err)
	}

	claims.BatchExecutedClaims = append(claims.BatchExecutedClaims, oCore.BatchExecutedClaim{
		ObservedTransactionHash: tx.TxSignature[:],
		ChainId:                 appConfig.ChainIDConverter.ToChainIDNum(tx.OriginChainID),
		BatchNonceId:            metadata.BatchNonceID,
	})

	p.logger.Info("Added BatchExecutedClaim",
		"txHash", tx.TxSignature, "chain", tx.OriginChainID, "BatchNonceId", metadata.BatchNonceID)

	return nil
}

func (*BatchExecutedProcessorImpl) validate(
	_ *core.SolanaTx, _ *core.BatchExecutedSolMetadata, _ *oCore.AppConfig,
) error {
	return nil
}
