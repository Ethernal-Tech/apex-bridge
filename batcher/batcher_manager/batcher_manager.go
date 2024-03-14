package batcher_manager

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/batcher/batcher"
	"github.com/Ethernal-Tech/apex-bridge/batcher/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
)

type BatchManagerImpl struct {
	config          *core.BatcherManagerConfiguration
	cardanoBatchers map[string]core.Batcher
	cancelCtx       context.CancelFunc
}

var _ core.BatcherManager = (*BatchManagerImpl)(nil)

func NewBatcherManager(config *core.BatcherManagerConfiguration) *BatchManagerImpl {
	var batchers = map[string]core.Batcher{}
	for chain, cardanoChainConfig := range config.CardanoChains {
		logger, err := logger.NewLogger(config.Logger)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error while creating logger: %v\n", err)
			return nil
		}

		batchers[chain] = batcher.NewBatcher(&core.BatcherConfiguration{
			Bridge:        config.Bridge,
			CardanoChain:  cardanoChainConfig,
			PullTimeMilis: config.PullTimeMilis,
		}, logger.Named(strings.ToUpper(chain)))
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

func LoadConfig(path string) (*core.BatcherManagerConfiguration, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var appConfig core.BatcherManagerConfiguration
	decoder := json.NewDecoder(f)
	err = decoder.Decode(&appConfig)
	if err != nil {
		return nil, err
	}

	return &appConfig, nil
}
