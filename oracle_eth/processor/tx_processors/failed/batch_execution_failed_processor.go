package failedtxprocessors

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_eth/core"
	"github.com/hashicorp/go-hclog"
)

var _ core.EthTxFailedProcessor = (*BatchExecutionFailedProcessorImpl)(nil)

type BatchExecutionFailedProcessorImpl struct {
	logger hclog.Logger
}

func NewEthBatchExecutionFailedProcessor(logger hclog.Logger) *BatchExecutionFailedProcessorImpl {
	return &BatchExecutionFailedProcessorImpl{
		logger: logger.Named("eth_batch_execution_failed_processor"),
	}
}

func (*BatchExecutionFailedProcessorImpl) GetType() common.BridgingTxType {
	return common.BridgingTxTypeBatchExecution
}

func (p *BatchExecutionFailedProcessorImpl) ValidateAndAddClaim(
	claims *oCore.BridgeClaims, tx *core.BridgeExpectedEthTx, appConfig *oCore.AppConfig,
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

	p.addBatchExecutionFailedClaim(claims, tx, metadata)

	return nil
}

func (p *BatchExecutionFailedProcessorImpl) addBatchExecutionFailedClaim(
	claims *oCore.BridgeClaims, tx *core.BridgeExpectedEthTx, metadata *core.BatchExecutedEthMetadata,
) {
	claim := oCore.BatchExecutionFailedClaim{
		ObservedTransactionHash: tx.Hash,
		ChainId:                 common.ToNumChainID(tx.ChainID),
		BatchNonceId:            metadata.BatchNonceID,
	}

	claims.BatchExecutionFailedClaims = append(claims.BatchExecutionFailedClaims, claim)

	p.logger.Info("Added BatchExecutionFailedClaim",
		"txHash", tx.Hash, "metadata", metadata, "claim", oCore.BatchExecutionFailedClaimString(claim))
}

func (*BatchExecutionFailedProcessorImpl) validate(
	_ *core.BridgeExpectedEthTx, _ *core.BatchExecutedEthMetadata, _ *oCore.AppConfig,
) error {
	// no validation needed
	return nil
}
