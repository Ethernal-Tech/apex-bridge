package successtxprocessors

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_solana/core"
)

var _ core.SolanaTxSuccessRefundProcessor = (*RefundDisabledProcessorImpl)(nil)

type RefundDisabledProcessorImpl struct{}

func NewRefundDisabledProcessor() *RefundDisabledProcessorImpl {
	return &RefundDisabledProcessorImpl{}
}

func (*RefundDisabledProcessorImpl) GetType() common.BridgingTxType {
	return common.TxTypeRefundRequest
}

func (*RefundDisabledProcessorImpl) PreValidate(tx *core.SolanaTx, appConfig *oCore.AppConfig) error {
	return nil
}

func (*RefundDisabledProcessorImpl) HandleBridgingProcessorPreValidate(
	tx *core.SolanaTx, appConfig *oCore.AppConfig) error {
	return nil
}

func (*RefundDisabledProcessorImpl) HandleBridgingProcessorError(
	claims *oCore.BridgeClaims, tx *core.SolanaTx, appConfig *oCore.AppConfig,
	err error, errContext string,
) error {
	return fmt.Errorf("%s. tx: %v, err: %w", errContext, tx, err)
}

func (*RefundDisabledProcessorImpl) ValidateAndAddClaim(
	claims *oCore.BridgeClaims, tx *core.SolanaTx, appConfig *oCore.AppConfig,
) error {
	return fmt.Errorf("refund is not enabled")
}
