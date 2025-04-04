package successtxprocessors

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
)

var _ core.CardanoTxSuccessRefundProcessor = (*RefundDisabledProcessorImpl)(nil)

type RefundDisabledProcessorImpl struct{}

func NewRefundDisabledProcessor() *RefundDisabledProcessorImpl {
	return &RefundDisabledProcessorImpl{}
}

func (*RefundDisabledProcessorImpl) GetType() common.BridgingTxType {
	return common.TxTypeRefundRequest
}

func (*RefundDisabledProcessorImpl) PreValidate(tx *core.CardanoTx, appConfig *cCore.AppConfig) error {
	return nil
}

func (p *RefundDisabledProcessorImpl) HandleBridgingProcessorError(
	claims *cCore.BridgeClaims, tx *core.CardanoTx, appConfig *cCore.AppConfig,
	err error, errContext string,
) error {
	return fmt.Errorf("%s. tx: %v, err: %w", errContext, tx, err)
}

func (p *RefundDisabledProcessorImpl) ValidateAndAddClaim(
	claims *cCore.BridgeClaims, tx *core.CardanoTx, appConfig *cCore.AppConfig,
) error {
	return fmt.Errorf("refund is not enabled")
}
