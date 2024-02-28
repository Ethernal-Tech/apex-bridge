package bridge

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/hashicorp/go-hclog"
)

type ClaimsSubmitterImpl struct {
	logger hclog.Logger
}

var _ core.ClaimsSubmitter = (*ClaimsSubmitterImpl)(nil)

func NewClaimsSubmitter(logger hclog.Logger) *ClaimsSubmitterImpl {
	return &ClaimsSubmitterImpl{
		logger: logger,
	}
}

func (cs *ClaimsSubmitterImpl) SubmitClaims(claims *core.BridgeClaims) error {
	// TODO: implement sending a list of claims to bridge contract
	return fmt.Errorf("not implemented")
}
