package stakingcomponent

import (
	"context"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/staking/core"
	"github.com/hashicorp/go-hclog"
)

type StakingComponentImpl struct {
	config *core.StakingConfiguration
	logger hclog.Logger
}

var _ core.StakingComponent = (*StakingComponentImpl)(nil)

func NewStakingComponent(
	config *core.StakingConfiguration,
	logger hclog.Logger,
) *StakingComponentImpl {
	return &StakingComponentImpl{
		config: config,
		logger: logger,
	}
}

func (sc *StakingComponentImpl) Start(ctx context.Context) {
	sc.logger.Debug("Staking Component started")

	waitTime := time.Millisecond * time.Duration(sc.config.PullTimeMilis)

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(waitTime):
		}

		sc.logger.Debug("Staking Component execute...")
	}
}
