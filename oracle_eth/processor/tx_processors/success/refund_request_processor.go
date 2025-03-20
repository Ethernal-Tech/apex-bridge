package successtxprocessors

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle_eth/core"
	"github.com/hashicorp/go-hclog"
)

var _ core.EthTxSuccessProcessor = (*BridgingRequestedProcessorImpl)(nil)

type RefundRequestProcessorImpl struct {
	logger hclog.Logger
}

func NewRefundRequestProcessor(logger hclog.Logger) *RefundRequestProcessorImpl {
	return &RefundRequestProcessorImpl{
		logger: logger.Named("refund_request_processor"),
	}
}

func (*RefundRequestProcessorImpl) GetType() common.BridgingTxType {
	return common.TxTypeRefundRequest
}

func (*RefundRequestProcessorImpl) PreValidate(tx *core.EthTx, appConfig *cCore.AppConfig) error {
	return nil
}

func (p *RefundRequestProcessorImpl) ValidateAndAddClaim(
	claims *cCore.BridgeClaims, tx *core.EthTx, appConfig *cCore.AppConfig,
) error {
	return nil
}

func (p *RefundRequestProcessorImpl) validate(
	tx *core.EthTx, appConfig *cCore.AppConfig,
) error {
	return nil
}
