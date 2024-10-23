package successtxprocessors

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_eth/core"
	"github.com/hashicorp/go-hclog"
)

var _ core.EthTxSuccessProcessor = (*BatchExecutedProcessorImpl)(nil)

type BatchExecutedProcessorImpl struct {
	logger hclog.Logger
}

func NewEthBatchExecutedProcessor(logger hclog.Logger) *BatchExecutedProcessorImpl {
	return &BatchExecutedProcessorImpl{
		logger: logger.Named("eth_batch_executed_processor"),
	}
}

func (*BatchExecutedProcessorImpl) GetType() common.BridgingTxType {
	return common.BridgingTxTypeBatchExecution
}

func (*BatchExecutedProcessorImpl) PreValidate(tx *core.EthTx, appConfig *oCore.AppConfig) error {
	return nil
}

func (p *BatchExecutedProcessorImpl) ValidateAndAddClaim(
	claims *oCore.BridgeClaims, tx *core.EthTx, appConfig *oCore.AppConfig,
) error {
	metadata, err := core.UnmarshalEthMetadata[core.BatchExecutedEthMetadata](
		tx.Metadata)
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

	claims.BatchExecutedClaims = append(claims.BatchExecutedClaims, oCore.BatchExecutedClaim{
		ObservedTransactionHash: tx.InnerActionHash,
		ChainId:                 common.ToNumChainID(tx.OriginChainID),
		BatchNonceId:            metadata.BatchNonceID,
	})

	p.logger.Info("Added BatchExecutedClaim",
		"txHash", tx.Hash, "chain", tx.OriginChainID, "BatchNonceId", metadata.BatchNonceID)

	return nil
}

func (*BatchExecutedProcessorImpl) validate(
	_ *core.EthTx, _ *core.BatchExecutedEthMetadata, _ *oCore.AppConfig,
) error {
	return nil
}
