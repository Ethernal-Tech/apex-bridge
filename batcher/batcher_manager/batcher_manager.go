package batchermanager

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
	var batchers = make([]core.Batcher, 0, len(config.Chains))

	for _, chainConfig := range config.Chains {
		operations, err := batcher.GetChainSpecificOperations(chainConfig)
		if err != nil {
			return nil, err
		}

		secretsManager, err := common.GetSecretsManager(
			config.Bridge.ValidatorDataDir, config.Bridge.ValidatorConfigPath, true)
		if err != nil {
			return nil, fmt.Errorf("failed to create secrets manager: %w", err)
		}

		wallet, err := ethtxhelper.NewEthTxWalletFromSecretManager(secretsManager)
		if err != nil {
			return nil, fmt.Errorf("failed to create blade wallet for batcher: %w", err)
		}

		bridgeSmartContract, err := eth.NewBridgeSmartContractWithWallet(
			config.Bridge.NodeURL, config.Bridge.SmartContractAddress, wallet, config.Bridge.DynamicTx)
		if err != nil {
			return nil, err
		}

		batchers = append(batchers, batcher.NewBatcher(&core.BatcherConfiguration{
			Bridge:        config.Bridge,
			Chain:         chainConfig,
			PullTimeMilis: config.PullTimeMilis,
		}, logger.Named(strings.ToUpper(chainConfig.ChainID)),
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
