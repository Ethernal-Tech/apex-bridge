package successtxprocessors

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_eth/core"
)

var _ core.EthTxSuccessRefundProcessor = (*RefundDisabledProcessorImpl)(nil)

type RefundDisabledProcessorImpl struct {
}

func NewRefundDisabledProcessor() *RefundDisabledProcessorImpl {
	return &RefundDisabledProcessorImpl{}
}

func (*RefundDisabledProcessorImpl) GetType() common.BridgingTxType {
	return common.TxTypeRefundRequest
}

func (*RefundDisabledProcessorImpl) PreValidate(tx *core.EthTx, appConfig *cCore.AppConfig) error {
	return nil
}

func (*RefundDisabledProcessorImpl) HandleBridgingProcessorPreValidate(
	tx *core.EthTx, appConfig *cCore.AppConfig) error {
	return nil
}

func (p *RefundDisabledProcessorImpl) HandleBridgingProcessorError(
	claims *cCore.BridgeClaims, tx *core.EthTx, appConfig *cCore.AppConfig,
	err error, errContext string,
) error {
	return fmt.Errorf("%s. tx: %v, err: %w", errContext, tx, err)
}

func (p *RefundDisabledProcessorImpl) ValidateAndAddClaim(
	claims *cCore.BridgeClaims, tx *core.EthTx, appConfig *cCore.AppConfig,
) error {
	return fmt.Errorf("refund is not enabled")
}
