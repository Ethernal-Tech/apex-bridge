package failedtxprocessors

import (
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

// wTODO: Implement
func (*BatchExecutionFailedProcessorImpl) PreValidate(
	tx *core.BridgeExpectedSolanaTx, appConfig *oCore.AppConfig,
) error {
	return nil
}

func (p *BatchExecutionFailedProcessorImpl) ValidateAndAddClaim(
	claims *oCore.BridgeClaims, tx *core.BridgeExpectedSolanaTx, appConfig *oCore.AppConfig,
) error {
	return nil
}
