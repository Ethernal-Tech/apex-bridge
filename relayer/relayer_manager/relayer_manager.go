package relayermanager

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/relayer/core"
	databaseaccess "github.com/Ethernal-Tech/apex-bridge/relayer/database_access"
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

	for chainID, chainConfig := range config.Chains {
		chainConfig.ChainID = chainID
		config.Chains[chainID] = chainConfig // update just to be sure that chainID is populated everywhere

		operations, err := relayer.GetChainSpecificOperations(chainConfig, logger)
		if err != nil {
			return nil, err
		}

		db, err := databaseaccess.NewDatabase(
			filepath.Join(chainConfig.DbsPath, chainConfig.ChainID+".db"))
		if err != nil {
			return nil, err
		}

		relayers = append(relayers, relayer.NewRelayer(
			&core.RelayerConfiguration{
				Bridge:        config.Bridge,
				Chain:         chainConfig,
				PullTimeMilis: config.PullTimeMilis,
			},
			eth.NewBridgeSmartContract(
				config.Bridge.NodeURL, config.Bridge.SmartContractAddress,
				config.Bridge.DynamicTx, logger.Named("bridge_smart_contract")),
			operations,
			db,
			logger.Named(strings.ToUpper(chainConfig.ChainID)),
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

func FixChains(config *core.RelayerManagerConfiguration, logger hclog.Logger) error {
	allRegisteredChains := []eth.Chain(nil)
	smartContract := eth.NewBridgeSmartContract(
		config.Bridge.NodeURL, config.Bridge.SmartContractAddress,
		config.Bridge.DynamicTx, logger)

	err := common.RetryForever(context.Background(), 2*time.Second, func(ctxInner context.Context) (err error) {
		allRegisteredChains, err = smartContract.GetAllRegisteredChains(ctxInner)
		if err != nil {
			logger.Error("Failed to GetAllRegisteredChains while creating ValidatorComponents. Retrying...", "err", err)
		}

		return err
	})
	if err != nil {
		return fmt.Errorf("error while RetryForever of GetAllRegisteredChains. err: %w", err)
	}

	logger.Debug("done GetAllRegisteredChains", "allRegisteredChains", allRegisteredChains)

	chainConfigs := make(map[string]core.ChainConfig, len(config.Chains))

	for _, regChain := range allRegisteredChains {
		chainID := common.ToStrChainID(regChain.Id)

		if cfg, exists := config.Chains[chainID]; exists {
			chainConfigs[chainID] = cfg
		}
	}

	config.Chains = chainConfigs

	return nil
}
