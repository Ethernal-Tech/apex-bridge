package relayer_manager

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/relayer/core"
	"github.com/Ethernal-Tech/apex-bridge/relayer/relayer"
	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
)

type RelayerManagerImpl struct {
	config          *core.RelayerManagerConfiguration
	cardanoRelayers map[string]*relayer.RelayerImpl
	cancelCtx       context.CancelFunc
}

func NewRelayerManager(config *core.RelayerManagerConfiguration) *RelayerManagerImpl {
	var relayers = map[string]*relayer.RelayerImpl{}
	for chain, cardanoChainConfig := range config.CardanoChains {
		logger, err := logger.NewLogger(config.Logger)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error while creating logger: %v\n", err)
			os.Exit(1)
		}
		logger = logger.Named(strings.ToUpper(chain))

		relayers[chain] = relayer.NewRelayer(&core.RelayerConfiguration{
			Bridge:        config.Bridge,
			CardanoChain:  cardanoChainConfig,
			PullTimeMilis: config.PullTimeMilis,
		}, logger)
	}

	return &RelayerManagerImpl{
		config:          config,
		cardanoRelayers: relayers,
	}
}

func (rm *RelayerManagerImpl) Start() error {
	ctx, cancelCtx := context.WithCancel(context.Background())
	rm.cancelCtx = cancelCtx

	for chain, r := range rm.cardanoRelayers {
		go r.Start(ctx)

		fmt.Fprintf(os.Stdin, "Started batcher for: %v chain\n", chain)
	}

	return nil
}

func (rm *RelayerManagerImpl) Stop() error {
	rm.cancelCtx()

	return nil
}

func LoadConfig(path string) (*core.RelayerManagerConfiguration, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var appConfig core.RelayerManagerConfiguration
	decoder := json.NewDecoder(f)
	err = decoder.Decode(&appConfig)
	if err != nil {
		return nil, err
	}

	return &appConfig, nil
}
