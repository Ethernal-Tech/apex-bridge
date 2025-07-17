package stakingmanager

import (
	"context"
	"fmt"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/oracle_cardano/chain"
	"github.com/Ethernal-Tech/apex-bridge/staking/core"
	databaseaccess "github.com/Ethernal-Tech/apex-bridge/staking/database_access"
	cardanotxsprocessor "github.com/Ethernal-Tech/apex-bridge/staking/processor/txs_processor"
	stakingcomponent "github.com/Ethernal-Tech/apex-bridge/staking/staking"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/hashicorp/go-hclog"
	"go.etcd.io/bbolt"
)

type StakingManagerImpl struct {
	ctx               context.Context
	config            *core.StakingManagerConfiguration
	stakingComponents map[string]core.StakingComponent
	logger            hclog.Logger
}

var _ core.StakingManager = (*StakingManagerImpl)(nil)

func NewStakingManager(
	ctx context.Context,
	config *core.StakingManagerConfiguration,
	boltDB *bbolt.DB,
	indexerDbs map[string]indexer.Database,
	logger hclog.Logger,
) (*StakingManagerImpl, error) {
	stakingDB := databaseaccess.NewBBoltDatabase(stakingcomponent.DecodeStakingAddress)
	stakingDB.Init(boltDB, config)

	stakingComponents := make(map[string]core.StakingComponent, len(config.Chains))

	for _, chainConfig := range config.Chains {
		indexerDB := indexerDbs[chainConfig.ChainID]

		txsProcessorLogger := logger.Named("staking_cardano_txs_processor_")
		chainObserverLogger := logger.Named("staking_cardano_chain_observer_" + chainConfig.ChainID)

		cardanoTxsReceiver := cardanotxsprocessor.NewCardanoTxsReceiverImpl(config, stakingDB, txsProcessorLogger)

		cco, err := chain.NewCardanoChainObserver(
			ctx,
			chainConfig,
			cardanoTxsReceiver,
			stakingDB,
			indexerDB,
			chainObserverLogger,
			"staking",
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create staking Cardano chain observer for `%s`: %w", chainConfig.ChainID, err)
		}

		stakingComponent, err := stakingcomponent.NewStakingComponent(
			&core.StakingConfiguration{
				Chain:                  *chainConfig,
				UsersRewardsPercentage: config.UsersRewardsPercentage,
				PullTimeMilis:          config.PullTimeMilis,
			},
			cco,
			stakingDB,
			logger.Named(strings.ToUpper(chainConfig.ChainID)),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create staking component for chain `%s`: %w", chainConfig.ChainID, err)
		}

		stakingComponents[chainConfig.ChainID] = stakingComponent
	}

	return &StakingManagerImpl{
		ctx:               ctx,
		config:            config,
		stakingComponents: stakingComponents,
		logger:            logger.Named("staking_manager"),
	}, nil
}

func (sm *StakingManagerImpl) Start() {
	for _, sc := range sm.stakingComponents {
		go func() {
			if err := sc.Start(sm.ctx); err != nil {
				sm.logger.Error("Staking component exited with error", "err", err)
			}
		}()
	}
}

func (sm *StakingManagerImpl) GetStakingComponent(chainID string) (core.StakingComponent, error) {
	sc, ok := sm.stakingComponents[chainID]
	if !ok {
		return nil, fmt.Errorf("failed to get staking component for chainID: %s", chainID)
	}

	return sc, nil
}
