package validatorcomponents

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"time"

	api "github.com/Ethernal-Tech/apex-bridge/api"
	apiCore "github.com/Ethernal-Tech/apex-bridge/api/core"
	"github.com/Ethernal-Tech/apex-bridge/api/utils"
	batchermanager "github.com/Ethernal-Tech/apex-bridge/batcher/batcher_manager"
	batcherCore "github.com/Ethernal-Tech/apex-bridge/batcher/core"
	bac "github.com/Ethernal-Tech/apex-bridge/bridging_addresses_coordinator"
	bam "github.com/Ethernal-Tech/apex-bridge/bridging_addresses_manager"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	cardanoOracleCore "github.com/Ethernal-Tech/apex-bridge/oracle_cardano/core"
	cardanoOracle "github.com/Ethernal-Tech/apex-bridge/oracle_cardano/oracle"
	oracleCommonBridge "github.com/Ethernal-Tech/apex-bridge/oracle_common/bridge"
	oracleCommonChain "github.com/Ethernal-Tech/apex-bridge/oracle_common/chain"
	oracleCommonCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	oracleCommonDA "github.com/Ethernal-Tech/apex-bridge/oracle_common/database_access"
	ethOracleCore "github.com/Ethernal-Tech/apex-bridge/oracle_eth/core"
	ethOracle "github.com/Ethernal-Tech/apex-bridge/oracle_eth/oracle"
	solanaOracleCore "github.com/Ethernal-Tech/apex-bridge/oracle_solana/core"
	solanaOracle "github.com/Ethernal-Tech/apex-bridge/oracle_solana/oracle"
	"github.com/Ethernal-Tech/apex-bridge/telemetry"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/api/controllers"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	databaseaccess "github.com/Ethernal-Tech/apex-bridge/validatorcomponents/database_access"
	relayerDbAccess "github.com/Ethernal-Tech/apex-bridge/validatorcomponents/database_access/relayer_imitator"
	eventTrackerStore "github.com/Ethernal-Tech/blockchain-event-tracker/store"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	indexerDb "github.com/Ethernal-Tech/cardano-infrastructure/indexer/db"
	solanaStore "github.com/Ethernal-Tech/solana-infrastructure/tracker/store"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-hclog"
	"go.etcd.io/bbolt"
)

const (
	MainComponentName            = "validatorcomponents"
	RelayerImitatorComponentName = "relayerimitator"
)

type ValidatorComponentsImpl struct {
	ctx                          context.Context
	shouldRunAPI                 bool
	oracleDB                     *bbolt.DB
	db                           core.Database
	cardanoIndexerDbs            map[string]indexer.Database
	solanaIndexerDbs             map[string]solanaStore.StorageHandler
	oracle                       *cardanoOracle.OracleImpl
	ethOracle                    *ethOracle.OracleImpl
	solanaOracle                 *solanaOracle.OracleImpl
	batcherManager               batcherCore.BatcherManager
	relayerImitator              core.RelayerImitator
	api                          apiCore.API
	telemetry                    *telemetry.Telemetry
	telemetryWorker              *TelemetryWorker
	bridgingAddressesManager     common.BridgingAddressesManager
	bridgingAddressesCoordinator common.BridgingAddressesCoordinator
	logger                       hclog.Logger
}

var _ core.ValidatorComponents = (*ValidatorComponentsImpl)(nil)

func NewValidatorComponents(
	ctx context.Context,
	appConfig *core.AppConfig,
	shouldRunAPI bool,
	logger hclog.Logger,
) (*ValidatorComponentsImpl, error) {
	db, err := databaseaccess.NewDatabase(filepath.Join(appConfig.Settings.DbsPath, MainComponentName+".db"))
	if err != nil {
		return nil, fmt.Errorf("failed to open validator components database: %w", err)
	}

	relayerImitatorDB, err := relayerDbAccess.NewDatabase(
		filepath.Join(appConfig.Settings.DbsPath, RelayerImitatorComponentName+".db"))
	if err != nil {
		return nil, fmt.Errorf("failed to open relayer imitator database: %w", err)
	}

	secretsManager, err := common.GetSecretsManager(
		appConfig.ValidatorDataDir, appConfig.ValidatorConfigPath, true)
	if err != nil {
		return nil, fmt.Errorf("failed to create secrets manager: %w", err)
	}

	wallet, err := ethtxhelper.NewEthTxWalletFromSecretManager(secretsManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create blade wallet: %w", err)
	}

	bridgingRequestStateManager := NewBridgingRequestStateManager(db, logger.Named("bridging_request_state_manager"))

	ethHelper := eth.NewEthHelperWrapperWithWallet(
		wallet, logger.Named("tx_helper_wrapper"),
		ethtxhelper.WithNodeURL(appConfig.Bridge.NodeURL),
		ethtxhelper.WithInitClientAndChainIDFn(ctx),
		ethtxhelper.WithTxPoolCheck(true),
		ethtxhelper.WithDynamicTx(appConfig.Bridge.DynamicTx),
		ethtxhelper.WithLogger(logger.Named("tx_helper")),
	)

	oracleBridgeSmartContract := eth.NewOracleBridgeSmartContract(
		appConfig.Bridge.SmartContractAddress, ethHelper, appConfig.ChainIDConverter)

	bridgeSmartContract := eth.NewBridgeSmartContract(
		appConfig.Bridge.SmartContractAddress, ethHelper, appConfig.ChainIDConverter)

	err = fixChainsInConfig(ctx, appConfig, bridgeSmartContract, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to populate utxos and addresses. err: %w", err)
	}

	if err := appConfig.ValidateDirectionConfig(); err != nil {
		return nil, fmt.Errorf("directionConfig validation failed. err: %w", err)
	}

	oracleConfig, batcherConfig := appConfig.SeparateConfigs()

	cardanoIndexerDbs := make(map[string]indexer.Database, len(oracleConfig.CardanoChains))

	for _, cardanoChainConfig := range oracleConfig.CardanoChains {
		indexerDB, err := indexerDb.NewDatabaseInit("",
			filepath.Join(appConfig.Settings.DbsPath, cardanoChainConfig.ChainID+".db"))
		if err != nil {
			return nil, fmt.Errorf("failed to open oracle indexer db for `%s`: %w", cardanoChainConfig.ChainID, err)
		}

		cardanoIndexerDbs[cardanoChainConfig.ChainID] = indexerDB
	}

	solanaIndexerDbs := make(map[string]solanaStore.StorageHandler, len(oracleConfig.SolanaChains))

	for _, solanaChainConfig := range oracleConfig.SolanaChains {
		indexerDB, err := solanaStore.NewBoltStorageHandler(
			filepath.Join(appConfig.Settings.DbsPath, solanaChainConfig.ChainID+".db"),
			false,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to open oracle indexer db for `%s`: %w", solanaChainConfig.ChainID, err)
		}

		solanaIndexerDbs[solanaChainConfig.ChainID] = indexerDB
	}

	ethIndexerDbs := make(map[string]eventTrackerStore.EventTrackerStore, len(appConfig.EthChains))

	for _, ethChainConfig := range oracleConfig.EthChains {
		indexerDB, err := eventTrackerStore.NewBoltDBEventTrackerStore(filepath.Join(
			appConfig.Settings.DbsPath, ethChainConfig.ChainID+".db"))
		if err != nil {
			return nil, fmt.Errorf("failed to open oracle indexer db for `%s`: %w", ethChainConfig.ChainID, err)
		}

		ethIndexerDbs[ethChainConfig.ChainID] = indexerDB
	}

	oracleDB, err := oracleCommonDA.NewDatabase(
		filepath.Join(appConfig.Settings.DbsPath, "oracle.db"), oracleConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open oracle database: %w", err)
	}

	cardanoBridgeSubmitter := oracleCommonBridge.NewBridgeSubmitter(
		ctx, oracleBridgeSmartContract, logger.Named("bridge_submitter_cardano"))

	typeRegister := oracleCommonCore.NewTypeRegisterWithChains(
		oracleConfig, reflect.TypeOf(cardanoOracleCore.CardanoTx{}),
		reflect.TypeOf(ethOracleCore.EthTx{}), reflect.TypeOf(solanaOracleCore.SolanaTx{}))

	bridgingAddressesManager, err := bam.NewBridgingAdressesManager(
		ctx,
		appConfig.CardanoChains,
		appConfig.ChainIDConverter,
		bridgeSmartContract,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create bridging addresses manager: %w", err)
	}

	bridgingAddressesCoordinator := bac.NewBridgingAddressesCoordinator(
		bridgingAddressesManager, cardanoIndexerDbs, appConfig.CardanoChains, appConfig.ChainIDConverter, logger)

	oracleConfig.BridgingAddressesManager = bridgingAddressesManager

	cardanoChainInfos := make(map[string]*oracleCommonChain.CardanoChainInfo, len(appConfig.CardanoChains))

	for _, cc := range appConfig.CardanoChains {
		info, err := oracleCommonChain.NewCardanoChainInfo(ctx, cc, db, logger.Named("cardano_chain_info_"+cc.ChainID))
		if err != nil {
			return nil, fmt.Errorf("failed to create cardano chain info for chainID: %s, err: %w", cc.ChainID, err)
		}

		cardanoChainInfos[cc.ChainID] = info
	}

	cardanoOracleObj, err := cardanoOracle.NewCardanoOracle(
		ctx, oracleDB, typeRegister, oracleConfig, cardanoChainInfos,
		oracleBridgeSmartContract, cardanoBridgeSubmitter, cardanoIndexerDbs,
		bridgingRequestStateManager, logger.Named("oracle_cardano"))
	if err != nil {
		return nil, fmt.Errorf("failed to create oracle_cardano. err %w", err)
	}

	var ethOracleObj *ethOracle.OracleImpl

	if len(appConfig.EthChains) > 0 {
		ethBridgeSubmitter := oracleCommonBridge.NewBridgeSubmitter(
			ctx, oracleBridgeSmartContract, logger.Named("bridge_submitter_eth"))

		ethOracleObj, err = ethOracle.NewEthOracle(
			ctx, oracleDB, typeRegister, oracleConfig, cardanoChainInfos,
			oracleBridgeSmartContract, ethBridgeSubmitter, ethIndexerDbs,
			bridgingRequestStateManager, logger.Named("oracle_eth"))
		if err != nil {
			return nil, fmt.Errorf("failed to create oracle_eth. err %w", err)
		}
	}

	var solanaOracleObj *solanaOracle.OracleImpl

	if len(appConfig.SolanaChains) > 0 {
		solanaBridgeSubmitter := oracleCommonBridge.NewBridgeSubmitter(
			ctx, oracleBridgeSmartContract, logger.Named("bridge_submitter_solana"))

		solanaOracleObj, err = solanaOracle.NewSolanaOracle(
			ctx, oracleDB, typeRegister, oracleConfig, cardanoChainInfos, oracleBridgeSmartContract,
			solanaBridgeSubmitter, logger.Named("oracle_solana"), solanaIndexerDbs, bridgingRequestStateManager)
		if err != nil {
			return nil, fmt.Errorf("failed to create oracle_solana. err %w", err)
		}
	}

	logger.Info("Batcher configuration info", "address", wallet.GetAddress(), "bridge", appConfig.Bridge.NodeURL,
		"contract", appConfig.Bridge.SmartContractAddress, "dynamicTx", appConfig.Bridge.DynamicTx)

	batcherManager, err := batchermanager.NewBatcherManager(
		ctx, batcherConfig, secretsManager, bridgeSmartContract,
		cardanoIndexerDbs, ethIndexerDbs, solanaIndexerDbs, bridgingRequestStateManager, bridgingAddressesManager,
		bridgingAddressesCoordinator, logger.Named("batcher"))
	if err != nil {
		return nil, fmt.Errorf("failed to create batcher manager: %w", err)
	}

	relayerImitator, err := NewRelayerImitator(
		appConfig, bridgingRequestStateManager, bridgeSmartContract, relayerImitatorDB, logger.Named("relayer_imitator"))
	if err != nil {
		return nil, fmt.Errorf("failed to create RelayerImitator. err: %w", err)
	}

	var apiObj *api.APIImpl

	if shouldRunAPI {
		apiLogger, err := utils.NewAPILogger(appConfig)
		if err != nil {
			return nil, err
		}

		apiControllers := []apiCore.APIController{
			controllers.NewBridgingRequestStateController(
				bridgingRequestStateManager, apiLogger.Named("bridging_request_state_controller")),
			controllers.NewOracleStateController(
				appConfig, bridgingRequestStateManager, cardanoIndexerDbs, ethIndexerDbs, solanaIndexerDbs,
				getAddressesMap(oracleConfig), apiLogger.Named("oracle_state")),
			controllers.NewSettingsController(appConfig, apiLogger.Named("settings_controller")),
			controllers.NewBridgingAddressController(
				bridgingAddressesCoordinator, bridgingAddressesManager,
				*appConfig.ChainIDConverter, apiLogger.Named("bridging_address_controller")),
		}

		apiObj, err = api.NewAPI(ctx, appConfig.APIConfig, apiControllers, apiLogger.Named("api"))
		if err != nil {
			return nil, fmt.Errorf("failed to create api: %w", err)
		}
	}

	return &ValidatorComponentsImpl{
		ctx:               ctx,
		shouldRunAPI:      shouldRunAPI,
		oracleDB:          oracleDB,
		db:                db,
		cardanoIndexerDbs: cardanoIndexerDbs,
		solanaIndexerDbs:  solanaIndexerDbs,
		oracle:            cardanoOracleObj,
		ethOracle:         ethOracleObj,
		solanaOracle:      solanaOracleObj,
		batcherManager:    batcherManager,
		relayerImitator:   relayerImitator,
		api:               apiObj,
		telemetry:         telemetry.NewTelemetry(appConfig.Telemetry, logger.Named("telemetry")),
		telemetryWorker: NewTelemetryWorker(
			ethHelper, cardanoIndexerDbs, ethIndexerDbs, solanaIndexerDbs, oracleConfig,
			appConfig.Telemetry.PullTime, logger.Named("telemetry_worker")),
		logger:                       logger,
		bridgingAddressesManager:     bridgingAddressesManager,
		bridgingAddressesCoordinator: bridgingAddressesCoordinator,
	}, nil
}

func (v *ValidatorComponentsImpl) Start() error {
	v.logger.Debug("Starting ValidatorComponents")

	err := v.oracle.Start()
	if err != nil {
		return fmt.Errorf("failed to start oracle_cardano. error: %w", err)
	}

	if v.ethOracle != nil {
		err = v.ethOracle.Start()
		if err != nil {
			return fmt.Errorf("failed to start oracle_eth. error: %w", err)
		}
	}

	if v.solanaOracle != nil {
		err = v.solanaOracle.Start()
		if err != nil {
			return fmt.Errorf("failed to start oracle_solana. error: %w", err)
		}
	}

	v.batcherManager.Start()

	if v.shouldRunAPI {
		go v.api.Start()
	}

	go v.relayerImitator.Start(v.ctx)

	if v.telemetry.IsEnabled() {
		if err := v.telemetry.Start(); err != nil {
			return fmt.Errorf("failed to start telemetry. error: %w", err)
		}

		go v.telemetryWorker.Start(v.ctx)
	}

	v.logger.Debug("Started ValidatorComponents")

	return nil
}

func (v *ValidatorComponentsImpl) Dispose() error {
	v.logger.Info("Disposing ValidatorComponents")

	errs := make([]error, 0)

	for _, indexerDB := range v.cardanoIndexerDbs {
		err := indexerDB.Close()
		if err != nil {
			v.logger.Error("Failed to close cardano indexer db", "err", err)
			errs = append(errs, fmt.Errorf("failed to close cardano indexer db. err %w", err))
		}
	}

	for _, indexerDB := range v.solanaIndexerDbs {
		err := indexerDB.Close()
		if err != nil {
			v.logger.Error("Failed to close solana indexer db", "err", err)
			errs = append(errs, fmt.Errorf("failed to close solana indexer db. err %w", err))
		}
	}

	if err := v.oracle.Dispose(); err != nil {
		v.logger.Error("error while disposing oracle", "err", err)
		errs = append(errs, fmt.Errorf("error while disposing oracle. err: %w", err))
	}

	if v.ethOracle != nil {
		if err := v.ethOracle.Dispose(); err != nil {
			v.logger.Error("error while disposing oracle_eth", "err", err)
			errs = append(errs, fmt.Errorf("error while disposing oracle_eth. err: %w", err))
		}
	}

	if v.solanaOracle != nil {
		if err := v.solanaOracle.Dispose(); err != nil {
			v.logger.Error("error while disposing oracle_solana", "err", err)
			errs = append(errs, fmt.Errorf("error while disposing oracle_solana. err: %w", err))
		}
	}

	err := v.oracleDB.Close()
	if err != nil {
		v.logger.Error("Failed to close oracle db", "err", err)
		errs = append(errs, fmt.Errorf("failed to close oracle db. err %w", err))
	}

	if v.shouldRunAPI {
		if err := v.api.Dispose(); err != nil {
			v.logger.Error("error while disposing api", "err", err)
			errs = append(errs, fmt.Errorf("error while disposing api. err: %w", err))
		}
	}

	if err := v.db.Close(); err != nil {
		v.logger.Error("Failed to close validatorcomponents db", "err", err)
		errs = append(errs, fmt.Errorf("failed to close validatorcomponents db. err: %w", err))
	}

	if err := v.telemetry.Close(v.ctx); err != nil {
		v.logger.Error("Failed to close telemetry", "err", err)
		errs = append(errs, fmt.Errorf("failed to close telemetry. err: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors while disposing validatorcomponents. errors: %w", errors.Join(errs...))
	}

	v.logger.Info("ValidatorComponents disposed")

	return nil
}

func fixChainsInConfig(
	ctx context.Context,
	config *core.AppConfig,
	smartContract eth.IBridgeSmartContract,
	logger hclog.Logger,
) error {
	var allRegisteredChains []eth.Chain

	logger.Debug("Retrieving all registered chains...")

	err := common.RetryForever(ctx, 2*time.Second, func(ctxInner context.Context) (err error) {
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

	cardanoChains := make(map[string]*oracleCommonCore.CardanoChainConfig)
	ethChains := make(map[string]*oracleCommonCore.EthChainConfig)
	solanaChains := make(map[string]*oracleCommonCore.SolanaChainConfig)

	// handle config for oracles
	for _, regChain := range allRegisteredChains {
		chainID := config.ChainIDConverter.ToChainIDStr(regChain.Id)

		logger.Debug("Registered chain received", "chainID", chainID, "type", regChain.ChainType,
			"addr", regChain.AddressMultisig, "fee", regChain.AddressFeePayer)

		switch regChain.ChainType {
		case common.ChainTypeCardano:
			chainConfig, exists := config.CardanoChains[chainID]
			if !exists {
				return fmt.Errorf("no configuration for chain: %s", chainID)
			}

			chainConfig.ChainID = chainID
			cardanoChains[chainID] = chainConfig
		case common.ChainTypeEVM:
			ethChainConfig, exists := config.EthChains[chainID]
			if !exists {
				return fmt.Errorf("no configuration for evm chain: %s", chainID)
			}

			if !ethcommon.IsHexAddress(regChain.AddressMultisig) {
				return fmt.Errorf("invalid gateway address for chain %s: %s", chainID, regChain.AddressMultisig)
			}

			ethChainConfig.ChainID = chainID
			ethChainConfig.BridgingAddresses = oracleCommonCore.EthBridgingAddresses{
				BridgingAddress: regChain.AddressMultisig,
			}

			ethChains[chainID] = ethChainConfig
		case common.ChainTypeSolana:
			solanaChainConfig, exists := config.SolanaChains[chainID]
			if !exists {
				return fmt.Errorf("no configuration for solana chain: %s", chainID)
			}

			solanaChainConfig.ChainID = chainID
			solanaChains[chainID] = solanaChainConfig
		default:
			logger.Debug("Do not know how to handle chain type", "chainID", chainID, "type", regChain.ChainType)
		}
	}

	config.CardanoChains = cardanoChains
	config.EthChains = ethChains
	config.SolanaChains = solanaChains

	return nil
}

func getAddressesMap(appConfig *oracleCommonCore.AppConfig) map[string][]string {
	result := make(map[string][]string, len(appConfig.CardanoChains))

	for chainID := range appConfig.CardanoChains {
		result[chainID] = append(appConfig.GetBridgingMultisigAddresses(chainID), appConfig.GetFeeMultisigAddress(chainID))
	}

	return result
}
