package stakingcomponent

import (
	"context"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/staking/core"
	"github.com/hashicorp/go-hclog"
)

type StakingComponentImpl struct {
	config               *core.StakingConfiguration
	cardanoChainObserver core.CardanoChainObserver
	logger               hclog.Logger
}

var _ core.StakingComponent = (*StakingComponentImpl)(nil)

func NewStakingComponent(
	config *core.StakingConfiguration,
	cardanoChainObserver core.CardanoChainObserver,
	logger hclog.Logger,
) *StakingComponentImpl {
	return &StakingComponentImpl{
		config:               config,
		cardanoChainObserver: cardanoChainObserver,
		logger:               logger,
	}
}

func (sc *StakingComponentImpl) Start(ctx context.Context) {
	sc.logger.Debug("Starting Staking Component")

	sc.cardanoChainObserver.Start()

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
