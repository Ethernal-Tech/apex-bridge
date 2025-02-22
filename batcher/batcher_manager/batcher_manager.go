package batchermanager

import (
	"context"
	"fmt"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/batcher/batcher"
	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	eventTrackerStore "github.com/Ethernal-Tech/blockchain-event-tracker/store"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
	"github.com/hashicorp/go-hclog"
)

type BatchManagerImpl struct {
	ctx      context.Context
	config   *core.BatcherManagerConfiguration
	batchers []core.Batcher
}

var _ core.BatcherManager = (*BatchManagerImpl)(nil)

func NewBatcherManager(
	ctx context.Context,
	config *core.BatcherManagerConfiguration,
	secretsManager secrets.SecretsManager,
	bridgeSmartContract eth.IBridgeSmartContract,
	cardanoIndexerDbs map[string]indexer.Database,
	ethIndexerDbs map[string]eventTrackerStore.EventTrackerStore,
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater,
	logger hclog.Logger,
) (*BatchManagerImpl, error) {
	var (
		err      error
		batchers = make([]core.Batcher, 0, len(config.Chains))
	)

	for _, chainConfig := range config.Chains {
		chainLogger := logger.Named(strings.ToUpper(chainConfig.ChainID))

		var operations core.ChainOperations

		switch strings.ToLower(chainConfig.ChainType) {
		case common.ChainTypeCardanoStr:
			operations, err = getCardanoOperations(chainConfig, cardanoIndexerDbs, secretsManager, logger)
			if err != nil {
				return nil, err
			}
		case common.ChainTypeEVMStr:
			operations, err = getEthOperations(chainConfig, ethIndexerDbs, secretsManager, logger)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unknown chain type: %s", chainConfig.ChainType)
		}

		batcher := batcher.NewBatcher(
			&core.BatcherConfiguration{
				Chain:         chainConfig,
				PullTimeMilis: config.PullTimeMilis,
			},
			operations,
			bridgeSmartContract,
			bridgingRequestStateUpdater,
			chainLogger)

		batchers = append(batchers, batcher)
	}

	return &BatchManagerImpl{
		ctx:      ctx,
		config:   config,
		batchers: batchers,
	}, nil
}

func (bm *BatchManagerImpl) Start() {
	for _, b := range bm.batchers {
		go b.Start(bm.ctx)
	}
}

func getCardanoOperations(
	config core.ChainConfig, cardanoIndexerDbs map[string]indexer.Database,
	secretsManager secrets.SecretsManager, logger hclog.Logger,
) (core.ChainOperations, error) {
	db, exists := cardanoIndexerDbs[config.ChainID]
	if !exists {
		return nil, fmt.Errorf("database not exists for chain: %s", config.ChainID)
	}

	operations, err := batcher.NewCardanoChainOperations(
		config.ChainSpecific, db, secretsManager, config.ChainID, logger)
	if err != nil {
		return nil, err
	}

	return operations, nil
}

func getEthOperations(
	config core.ChainConfig, ethIndexerDbs map[string]eventTrackerStore.EventTrackerStore,
	secretsManager secrets.SecretsManager, logger hclog.Logger,
) (core.ChainOperations, error) {
	db, exists := ethIndexerDbs[config.ChainID]
	if !exists {
		return nil, fmt.Errorf("database not exists for chain: %s", config.ChainID)
	}

	operations, err := batcher.NewEVMChainOperations(
		config.ChainSpecific, secretsManager, db, config.ChainID, logger)
	if err != nil {
		return nil, err
	}

	return operations, nil
}
