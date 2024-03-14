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
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

type RelayerManagerImpl struct {
	config          *core.RelayerManagerConfiguration
	cardanoRelayers map[string]core.Relayer
	cancelCtx       context.CancelFunc
}

var _ core.RelayerManager = (*RelayerManagerImpl)(nil)

func NewRelayerManager(config *core.RelayerManagerConfiguration) *RelayerManagerImpl {
	var relayers = map[string]core.Relayer{}
	for chain, cardanoChainConfig := range config.CardanoChains {
		logger, err := logger.NewLogger(config.Logger)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error while creating logger: %v\n", err)
			return nil
		}

		txProvider, err := cardanowallet.NewTxProviderBlockFrost(cardanoChainConfig.BlockfrostUrl, cardanoChainConfig.BlockfrostAPIKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error while creating tx provider: %v\n", err)
			return nil
		}

		relayers[chain] = relayer.NewRelayer(&core.RelayerConfiguration{
			Bridge:        config.Bridge,
			CardanoChain:  cardanoChainConfig,
			PullTimeMilis: config.PullTimeMilis,
		}, logger.Named(strings.ToUpper(chain)), relayer.NewCardanoChainOperations(txProvider))
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
