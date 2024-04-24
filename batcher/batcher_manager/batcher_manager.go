package batcher_manager

import (
	"context"
	"fmt"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/batcher/batcher"
	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	"github.com/hashicorp/go-hclog"
)

type BatchManagerImpl struct {
	ctx             context.Context
	config          *core.BatcherManagerConfiguration
	cardanoBatchers []core.Batcher
}

var _ core.BatcherManager = (*BatchManagerImpl)(nil)

func NewBatcherManager(
	ctx context.Context,
	config *core.BatcherManagerConfiguration,
	bridgingRequestStateUpdater common.BridgingRequestStateUpdater,
	logger hclog.Logger,
) (*BatchManagerImpl, error) {
	var batchers = make([]core.Batcher, len(config.Chains))

	for _, chainConfig := range config.Chains {
		operations, err := batcher.GetChainSpecificOperations(chainConfig)
		if err != nil {
			return nil, err
		}

		wallet, err := ethtxhelper.NewEthTxWalletFromSecretManagerConfig(config.Bridge.SecretsManager)
		if err != nil {
			return nil, fmt.Errorf("failed to create blade wallet for batcher: %w", err)
		}

		bridgeSmartContract, err := eth.NewBridgeSmartContractWithWallet(
			config.Bridge.NodeUrl, config.Bridge.SmartContractAddress, wallet)
		if err != nil {
			return nil, err
		}

		batchers = append(batchers, batcher.NewBatcher(&core.BatcherConfiguration{
			Bridge:        config.Bridge,
			Chain:         chainConfig,
			PullTimeMilis: config.PullTimeMilis,
		}, logger.Named(strings.ToUpper(chainConfig.ChainId)),
			operations, bridgeSmartContract, bridgingRequestStateUpdater))
	}

	return &BatchManagerImpl{
		ctx:             ctx,
		config:          config,
		cardanoBatchers: batchers,
	}, nil
}

func (bm *BatchManagerImpl) Start() {
	for _, b := range bm.cardanoBatchers {
		go b.Start(bm.ctx)
	}
}
