package failedtxprocessors

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth_oracle/core"
	oracleCore "github.com/Ethernal-Tech/apex-bridge/oracle/core"
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
	claims *oracleCore.BridgeClaims, tx *core.BridgeExpectedEthTx, appConfig *oracleCore.AppConfig,
) error {
	metadata, err := common.UnmarshalMetadata[common.BatchExecutedMetadata](
		common.MetadataEncodingTypeJSON, tx.Metadata)
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
	claims *oracleCore.BridgeClaims, tx *core.BridgeExpectedEthTx, metadata *common.BatchExecutedMetadata,
) {
	claim := oracleCore.BatchExecutionFailedClaim{
		ObservedTransactionHash: tx.Hash,
		ChainId:                 common.ToNumChainID(tx.ChainID),
		BatchNonceId:            metadata.BatchNonceID,
	}

	claims.BatchExecutionFailedClaims = append(claims.BatchExecutionFailedClaims, claim)

	p.logger.Info("Added BatchExecutionFailedClaim",
		"txHash", tx.Hash, "metadata", metadata, "claim", oracleCore.BatchExecutionFailedClaimString(claim))
}

func (*BatchExecutionFailedProcessorImpl) validate(
	tx *core.BridgeExpectedEthTx, metadata *common.BatchExecutedMetadata, appConfig *oracleCore.AppConfig,
) error {
	// no validation needed
	return nil
}
