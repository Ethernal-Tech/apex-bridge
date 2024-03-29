package relayer_manager

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/relayer/core"
	"github.com/Ethernal-Tech/apex-bridge/relayer/relayer"
	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
)

type RelayerManagerImpl struct {
	config          *core.RelayerManagerConfiguration
	cardanoRelayers map[string]core.Relayer
	cancelCtx       context.CancelFunc
}

var _ core.RelayerManager = (*RelayerManagerImpl)(nil)

func NewRelayerManager(config *core.RelayerManagerConfiguration, customBridgeSc ...eth.IBridgeSmartContract) *RelayerManagerImpl {
	var relayers = map[string]core.Relayer{}
	for chain, chainConfig := range config.Chains {
		logger, err := logger.NewLogger(config.Logger)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error while creating logger: %v\n", err)
			return nil
		}

		operations, err := relayer.GetChainSpecificOperations(chainConfig.ChainSpecific)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error while creating operations: %v\n", err)
			return nil
		}

		var bridgeSmartContract eth.IBridgeSmartContract
		if len(customBridgeSc) == 0 {
			bridgeSmartContract = eth.NewBridgeSmartContract(config.Bridge.NodeUrl, config.Bridge.SmartContractAddress)
		} else {
			bridgeSmartContract = customBridgeSc[0]
		}

		relayers[chain] = relayer.NewRelayer(&core.RelayerConfiguration{
			Bridge:        config.Bridge,
			Base:          chainConfig.Base,
			PullTimeMilis: config.PullTimeMilis,
		}, bridgeSmartContract, logger.Named(strings.ToUpper(chain)), operations)
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

		fmt.Fprintf(os.Stdin, "Started relayer for: %v chain\n", chain)
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
