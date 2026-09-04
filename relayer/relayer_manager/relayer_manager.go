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
	"github.com/Ethernal-Tech/apex-bridge/telemetry"
	"github.com/hashicorp/go-hclog"
)

const telemetryCloseTimeout = 5 * time.Second

type RelayerManagerImpl struct {
	config          *core.RelayerManagerConfiguration
	cardanoRelayers []core.Relayer
	telemetry       *telemetry.Telemetry
	telemetryWorker *TelemetryWorker
	cancelCtx       context.CancelFunc
	logger          hclog.Logger
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

	var operations map[string]core.ChainOperations

	relayers, operations, config.Chains, err = getRelayersAndConfigurations(
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
		telemetry:       telemetry.NewTelemetry(config.Telemetry, logger.Named("telemetry")),
		telemetryWorker: NewTelemetryWorker(
			operations, config.Chains, config.Telemetry.PullTime, logger.Named("telemetry_worker")),
		logger: logger,
	}, nil
}

func (rm *RelayerManagerImpl) Start() error {
	ctx, cancelCtx := context.WithCancel(context.Background())
	rm.cancelCtx = cancelCtx

	for _, r := range rm.cardanoRelayers {
		go r.Start(ctx)
	}

	if rm.telemetry.IsEnabled() {
		if err := rm.telemetry.Start(); err != nil {
			return fmt.Errorf("failed to start telemetry. err: %w", err)
		}

		go rm.telemetryWorker.Start(ctx)
	}

	return nil
}

func (rm *RelayerManagerImpl) Stop() error {
	rm.cancelCtx()

	if rm.telemetry.IsEnabled() {
		ctx, cancelCtx := context.WithTimeout(context.Background(), telemetryCloseTimeout)
		defer cancelCtx()

		if err := rm.telemetry.Close(ctx); err != nil {
			rm.logger.Error("failed to close telemetry", "err", err)
		}
	}

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
) ([]core.Relayer, map[string]core.ChainOperations, map[string]core.ChainConfig, error) {
	logger.Debug("done GetAllRegisteredChains", "allRegisteredChains", allRegisteredChains)

	relayers := make([]core.Relayer, 0, len(allRegisteredChains))
	allOperations := make(map[string]core.ChainOperations, len(allRegisteredChains))
	newChainsConfigs := make(map[string]core.ChainConfig, len(allRegisteredChains))

	for _, chainData := range allRegisteredChains {
		chainID := config.ChainIDConverter.ToChainIDStr(chainData.Id)

		chainConfig, exists := config.Chains[chainID]
		if !exists {
			logger.Warn("No configuration for registered chain: %s. Chain type = %d", chainID, chainData.ChainType)

			continue
		}

		chainConfig.ChainID = chainID
		newChainsConfigs[chainID] = chainConfig

		operations, err := relayer.GetChainSpecificOperations(chainConfig, chainData, config.RunMode, logger)
		if err != nil {
			return nil, nil, nil, err
		}

		allOperations[chainID] = operations

		db, err := databaseaccess.NewDatabase(
			filepath.Join(chainConfig.DbsPath, chainConfig.ChainID+".db"))
		if err != nil {
			return nil, nil, nil, err
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

	return relayers, allOperations, newChainsConfigs, nil
}
