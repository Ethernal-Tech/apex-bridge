package batcher

import (
	"context"
	"fmt"
	"os"

	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
)

type BatcherManager interface {
	Start() error
}

type BatchManagerImpl struct {
	config          *BatcherManagerConfiguration
	cardanoBatchers map[string]*Batcher
	ctx             context.Context
}

func NewBatcherManager(config *BatcherManagerConfiguration, ctx context.Context) *BatchManagerImpl {
	var batchers = map[string]*Batcher{}
	for chain, cardanoChainConfig := range config.CardanoChains {
		logger, err := logger.NewLogger(config.Logger)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error while creating logger: %v\n", err)
			os.Exit(1)
		}

		batchers[chain] = NewBatcher(&BatcherConfiguration{
			Bridge:        config.Bridge,
			CardanoChain:  cardanoChainConfig,
			PullTimeMilis: config.PullTimeMilis,
		}, logger)
	}

	return &BatchManagerImpl{
		config:          config,
		cardanoBatchers: batchers,
		ctx:             ctx,
	}
}

func (bm *BatchManagerImpl) Start() error {
	for chain, b := range bm.cardanoBatchers {
		go b.Execute(bm.ctx)

		fmt.Fprintf(os.Stdin, "Started batcher for: %v chain\n", chain)
		b.logger.Debug(fmt.Sprintf("%s batcher started", chain))
	}

	return nil
}
