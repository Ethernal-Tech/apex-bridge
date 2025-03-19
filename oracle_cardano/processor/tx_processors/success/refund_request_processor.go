package successtxprocessors

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/hashicorp/go-hclog"
)

var _ core.CardanoTxSuccessProcessor = (*BridgingRequestedProcessorImpl)(nil)

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

func (*RefundRequestProcessorImpl) PreValidate(tx *core.CardanoTx, appConfig *cCore.AppConfig) error {
	return nil
}

func (p *RefundRequestProcessorImpl) ValidateAndAddClaim(
	claims *cCore.BridgeClaims, tx *core.CardanoTx, appConfig *cCore.AppConfig,
) error {
	return nil
}

func (p *RefundRequestProcessorImpl) validate(
	tx *core.CardanoTx, metadata *common.BridgingRequestMetadata, appConfig *cCore.AppConfig,
) error {
	return nil
}
