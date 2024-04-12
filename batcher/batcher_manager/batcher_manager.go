package batcher_manager

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/batcher/batcher"
	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"

	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
)

type BatchManagerImpl struct {
	config          *core.BatcherManagerConfiguration
	cardanoBatchers map[string]core.Batcher
	cancelCtx       context.CancelFunc
}

var _ core.BatcherManager = (*BatchManagerImpl)(nil)

func NewBatcherManager(config *core.BatcherManagerConfiguration, customOperations map[string]core.ChainOperations, customBridgeSc ...eth.IBridgeSmartContract) *BatchManagerImpl {
	var batchers = map[string]core.Batcher{}
	for chain, chainConfig := range config.Chains {
		logger, err := logger.NewLogger(config.Logger)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error while creating logger: %v\n", err)
			return nil
		}

		var operations core.ChainOperations = customOperations[chain]
		if operations == nil {
			operations, err = batcher.GetChainSpecificOperations(chainConfig.ChainSpecific, chainConfig.Base.KeysDirPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error while creating operations: %v\n", err)
				logger.Error("error while creating operations", "err", err)
				return nil
			}
		}

		var bridgeSmartContract eth.IBridgeSmartContract
		if len(customBridgeSc) == 0 {
			wallet, err := ethtxhelper.NewEthTxWalletFromSecretManager(config.Bridge.SecretsManager)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error while creating wallet for bridge: %v\n", err)
				logger.Error("error while creating wallet for bridge", "err", err)
				return nil
			}

			bridgeSmartContract, err = eth.NewBridgeSmartContractWithWallet(
				config.Bridge.NodeUrl, config.Bridge.SmartContractAddress, wallet)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error while creating bridge smart contract instance: %v\n", err)
				logger.Error("error while creating bridge smart contract instance", "err", err)
				return nil
			}
		} else {
			bridgeSmartContract = customBridgeSc[0]
		}

		batchers[chain] = batcher.NewBatcher(&core.BatcherConfiguration{
			Bridge:        config.Bridge,
			Base:          chainConfig.Base,
			PullTimeMilis: config.PullTimeMilis,
		}, logger.Named(strings.ToUpper(chain)), operations, bridgeSmartContract)
	}

	return &BatchManagerImpl{
		config:          config,
		cardanoBatchers: batchers,
	}
}

func (bm *BatchManagerImpl) Start() error {
	ctx, cancelCtx := context.WithCancel(context.Background())
	bm.cancelCtx = cancelCtx

	for chain, b := range bm.cardanoBatchers {
		go b.Start(ctx)

		fmt.Fprintf(os.Stdin, "Started batcher for: %v chain\n", chain)
	}

	return nil
}

func (bm *BatchManagerImpl) Stop() error {
	bm.cancelCtx()

	return nil
}
