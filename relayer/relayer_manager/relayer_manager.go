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
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
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
	ctx context.Context,
	config *core.RelayerManagerConfiguration,
	logger hclog.Logger,
) (*RelayerManagerImpl, error) {
	var (
		allRegisteredChains []eth.Chain
		relayers            []core.Relayer
		txHelper            = eth.NewEthHelperWrapper(
			logger.Named("bridge_smart_contract"),
			ethtxhelper.WithNodeURL(config.Bridge.NodeURL),
			ethtxhelper.WithInitClientAndChainIDFn(context.Background()),
			ethtxhelper.WithDynamicTx(config.Bridge.DynamicTx))
		bridgeSmartContract = eth.NewBridgeSmartContract(
			config.Bridge.SmartContractAddress, txHelper, config.ChainIDConverter)
	)

	err := common.RetryForever(ctx, 2*time.Second, func(ctxInner context.Context) (err error) {
		allRegisteredChains, err = bridgeSmartContract.GetAllRegisteredChains(ctxInner)
		if err != nil {
			logger.Error("Failed to GetAllRegisteredChains while creating Relayers. Retrying...", "err", err)
		}

		return err
	})
	if err != nil {
		return nil, fmt.Errorf("error while RetryForever of GetAllRegisteredChains. err: %w", err)
	}

	relayers, config.Chains, err = getRelayersAndConfigurations(
		bridgeSmartContract, allRegisteredChains, config, logger)
	if err != nil {
		return nil, err
	}

	if logger.IsDebug() {
		for chainID := range config.Chains {
			data, err := bridgeSmartContract.GetValidatorsChainData(ctx, chainID)

			logger.Debug("Validators data per chain", "chain", chainID,
				"data", eth.GetChainValidatorsDataInfoString(chainID, data, config.ChainIDConverter), "err", err)
		}
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

func getRelayersAndConfigurations(
	bridgeSmartContract eth.IBridgeSmartContract,
	allRegisteredChains []eth.Chain,
	config *core.RelayerManagerConfiguration,
	logger hclog.Logger,
) ([]core.Relayer, map[string]core.ChainConfig, error) {
	logger.Debug("done GetAllRegisteredChains", "allRegisteredChains", allRegisteredChains)

	relayers := make([]core.Relayer, 0, len(allRegisteredChains))
	newChainsConfigs := make(map[string]core.ChainConfig, len(allRegisteredChains))

	for _, chainData := range allRegisteredChains {
		chainID := config.ChainIDConverter.ToStrChainID(chainData.Id)

		chainConfig, exists := config.Chains[chainID]
		if !exists {
			logger.Warn("No configuration for registered chain: %s. Chain type = %d", chainID, chainData.ChainType)

			continue
		}

		chainConfig.ChainID = chainID
		newChainsConfigs[chainID] = chainConfig

		operations, err := relayer.GetChainSpecificOperations(chainConfig, chainData, config.RunMode, logger)
		if err != nil {
			return nil, nil, err
		}

		db, err := databaseaccess.NewDatabase(
			filepath.Join(chainConfig.DbsPath, chainConfig.ChainID+".db"))
		if err != nil {
			return nil, nil, err
		}

		relayers = append(relayers, relayer.NewRelayer(
			&core.RelayerConfiguration{
				Bridge:        config.Bridge,
				Chain:         chainConfig,
				PullTimeMilis: config.PullTimeMilis,
			},
			bridgeSmartContract,
			operations,
			db,
			logger.Named(strings.ToUpper(chainConfig.ChainID)),
		))
	}

	return relayers, newChainsConfigs, nil
}
