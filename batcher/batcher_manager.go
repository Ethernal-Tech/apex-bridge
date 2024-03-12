package batcher

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
)

type BatcherManager interface {
	Start() error
	Stop() error
}

type BatchManagerImpl struct {
	config          *BatcherManagerConfiguration
	cardanoBatchers map[string]*Batcher
	cancelCtx       context.CancelFunc
}

func NewBatcherManager(config *BatcherManagerConfiguration) *BatchManagerImpl {
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
	}
}

func (bm *BatchManagerImpl) Start() error {
	ctx, cancelCtx := context.WithCancel(context.Background())
	bm.cancelCtx = cancelCtx

	for chain, b := range bm.cardanoBatchers {
		go b.Start(ctx)

		fmt.Fprintf(os.Stdin, "Started batcher for: %v chain\n", chain)
		b.logger.Debug(fmt.Sprintf("%s batcher started", chain))
	}

	return nil
}

func (bm *BatchManagerImpl) Stop() error {
	bm.cancelCtx()

	return nil
}

func LoadConfig() (*BatcherManagerConfiguration, error) {
	f, err := os.Open("config.json")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var appConfig BatcherManagerConfiguration
	decoder := json.NewDecoder(f)
	err = decoder.Decode(&appConfig)
	if err != nil {
		return nil, err
	}

	return &appConfig, nil
}
