package batcher_manager

import (
	"context"
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

	for chain, chainConfig := range config.Chains {
		operations, err := batcher.GetChainSpecificOperations(chainConfig.ChainSpecific, chainConfig.Base.KeysDirPath)
		if err != nil {
			return nil, err
		}

		wallet, err := ethtxhelper.NewEthTxWalletFromSecretManager(config.Bridge.SecretsManager)
		if err != nil {
			return nil, err
		}

		bridgeSmartContract, err := eth.NewBridgeSmartContractWithWallet(
			config.Bridge.NodeUrl, config.Bridge.SmartContractAddress, wallet)
		if err != nil {
			return nil, err
		}

		batchers = append(batchers, batcher.NewBatcher(&core.BatcherConfiguration{
			Bridge:        config.Bridge,
			Base:          chainConfig.Base,
			PullTimeMilis: config.PullTimeMilis,
		}, logger.Named(strings.ToUpper(chain)), operations, bridgeSmartContract, bridgingRequestStateUpdater))
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
