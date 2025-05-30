package stakingmanager

import (
	"context"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/staking/core"
	stakingcomponent "github.com/Ethernal-Tech/apex-bridge/staking/staking"
	"github.com/hashicorp/go-hclog"
)

type StakingManagerImpl struct {
	ctx               context.Context
	config            *core.StakingManagerConfiguration
	stakingComponents []core.StakingComponent
}

var _ core.StakingManager = (*StakingManagerImpl)(nil)

func NewStakingManager(
	ctx context.Context,
	config *core.StakingManagerConfiguration,
	logger hclog.Logger,
) (*StakingManagerImpl, error) {
	var stakingComponents = make([]core.StakingComponent, 0, len(config.Chains))

	for _, chainConfig := range config.Chains {
		stakingComponent := stakingcomponent.NewStakingComponent(
			&core.StakingConfiguration{
				Chain:         chainConfig,
				PullTimeMilis: config.PullTimeMilis,
			},
			logger.Named(strings.ToUpper(chainConfig.ChainID)),
		)
		stakingComponents = append(stakingComponents, stakingComponent)
	}

	return &StakingManagerImpl{
		ctx:               ctx,
		config:            config,
		stakingComponents: stakingComponents,
	}, nil
}

func (sm *StakingManagerImpl) Start() {
	for _, sc := range sm.stakingComponents {
		go sc.Start(sm.ctx)
	}
}
