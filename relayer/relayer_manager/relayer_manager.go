package relayer_manager

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/relayer/core"
	"github.com/Ethernal-Tech/apex-bridge/relayer/database_access"
	"github.com/Ethernal-Tech/apex-bridge/relayer/relayer"
	"github.com/hashicorp/go-hclog"
)

type RelayerManagerImpl struct {
	config          *core.RelayerManagerConfiguration
	cardanoRelayers []core.Relayer
	cancelCtx       context.CancelFunc
}

var _ core.RelayerManager = (*RelayerManagerImpl)(nil)

func NewRelayerManager(
	config *core.RelayerManagerConfiguration,
	logger hclog.Logger,
) (*RelayerManagerImpl, error) {
	relayers := make([]core.Relayer, 0, len(config.Chains))

	for chainId, chainConfig := range config.Chains {
		chainConfig.ChainId = chainId

		operations, err := relayer.GetChainSpecificOperations(chainConfig)
		if err != nil {
			return nil, err
		}

		db, err := database_access.NewDatabase(chainConfig.DbsPath + chainConfig.ChainId + ".db")
		if err != nil {
			return nil, err
		}

		relayers = append(relayers, relayer.NewRelayer(
			&core.RelayerConfiguration{
				Bridge:        config.Bridge,
				Chain:         chainConfig,
				PullTimeMilis: config.PullTimeMilis,
			},
			eth.NewBridgeSmartContract(config.Bridge.NodeUrl, config.Bridge.SmartContractAddress),
			logger.Named(strings.ToUpper(chainConfig.ChainId)),
			operations,
			db,
		))
	}

	return &RelayerManagerImpl{
		config:          config,
		cardanoRelayers: relayers,
	}, nil
}

func (rm *RelayerManagerImpl) Start() error {
	ctx, cancelCtx := context.WithCancel(context.Background())
	rm.cancelCtx = cancelCtx

	for _, r := range rm.cardanoRelayers {
		go r.Start(ctx)
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

	err = json.NewDecoder(f).Decode(&appConfig)
	if err != nil {
		return nil, err
	}

	return &appConfig, nil
}
