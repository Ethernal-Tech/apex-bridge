package stakingcomponent

import (
	"context"
	"fmt"
	"time"

	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	"github.com/Ethernal-Tech/apex-bridge/staking/core"
	"github.com/hashicorp/go-hclog"
)

type StakingComponentImpl struct {
	config               *core.StakingConfiguration
	cardanoChainObserver oCore.CardanoChainObserver
	logger               hclog.Logger
}

var _ core.StakingComponent = (*StakingComponentImpl)(nil)

func NewStakingComponent(
	config *core.StakingConfiguration,
	cardanoChainObserver oCore.CardanoChainObserver,
	logger hclog.Logger,
) *StakingComponentImpl {
	return &StakingComponentImpl{
		config:               config,
		cardanoChainObserver: cardanoChainObserver,
		logger:               logger,
	}
}

func (sc *StakingComponentImpl) Start(ctx context.Context) error {
	sc.logger.Debug("Starting Staking Component")

	err := sc.cardanoChainObserver.Start()
	if err != nil {
		return fmt.Errorf("failed to start observer for %s: %w", sc.cardanoChainObserver.GetConfig().GetChainID(), err)
	}

	waitTime := time.Millisecond * time.Duration(sc.config.PullTimeMilis)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(waitTime):
		}

		sc.logger.Debug("Staking Component execute...")
	}
}
