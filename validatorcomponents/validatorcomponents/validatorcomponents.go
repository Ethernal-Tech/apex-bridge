package validatorcomponents

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	batchermanager "github.com/Ethernal-Tech/apex-bridge/batcher/batcher_manager"
	batcherCore "github.com/Ethernal-Tech/apex-bridge/batcher/core"
	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	ethtxhelper "github.com/Ethernal-Tech/apex-bridge/eth/txhelper"
	eth_bridge "github.com/Ethernal-Tech/apex-bridge/eth_oracle/bridge"
	ethOracleCore "github.com/Ethernal-Tech/apex-bridge/eth_oracle/core"
	ethOracle "github.com/Ethernal-Tech/apex-bridge/eth_oracle/oracle"
	"github.com/Ethernal-Tech/apex-bridge/oracle/bridge"
	oracleCore "github.com/Ethernal-Tech/apex-bridge/oracle/core"

	"github.com/Ethernal-Tech/apex-bridge/oracle/oracle"
	"github.com/Ethernal-Tech/apex-bridge/telemetry"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/api"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/api/controllers"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	databaseaccess "github.com/Ethernal-Tech/apex-bridge/validatorcomponents/database_access"
	eventTrackerStore "github.com/Ethernal-Tech/blockchain-event-tracker/store"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	indexerDb "github.com/Ethernal-Tech/cardano-infrastructure/indexer/db"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
)

const (
	MainComponentName            = "validatorcomponents"
	RelayerImitatorComponentName = "relayerimitator"
)

type ValidatorComponentsImpl struct {
	ctx               context.Context
	shouldRunAPI      bool
	db                core.Database
	cardanoIndexerDbs map[string]indexer.Database
	oracle            oracleCore.Oracle
	ethOracle         ethOracleCore.Oracle
	batcherManager    batcherCore.BatcherManager
	relayerImitator   core.RelayerImitator
	api               core.API
	telemetry         *telemetry.Telemetry
	logger            hclog.Logger
	errorCh           *common.SafeCh[error]
}

var _ core.ValidatorComponents = (*ValidatorComponentsImpl)(nil)

func NewValidatorComponents(
	ctx context.Context,
	appConfig *core.AppConfig,
	shouldRunAPI bool,
	logger hclog.Logger,
) (*ValidatorComponentsImpl, error) {
	telemetry, err := telemetry.NewTelemetry(appConfig.Telemetry, logger.Named("telemetry"))
	if err != nil {
		return nil, fmt.Errorf("failed to create telemetry. err: %w", err)
	}

	db, err := databaseaccess.NewDatabase(filepath.Join(appConfig.Settings.DbsPath, MainComponentName+".db"))
	if err != nil {
		return nil, fmt.Errorf("failed to open validator components database: %w", err)
	}

	bridgingRequestStateManager, err := NewBridgingRequestStateManager(db, logger.Named("bridging_request_state_manager"))
	if err != nil {
		return nil, fmt.Errorf("failed to create BridgingRequestStateManager. err: %w", err)
	}

	oracleConfig, batcherConfig := appConfig.SeparateConfigs()

	err = fixChainsAndAddresses(
		ctx, oracleConfig, batcherConfig,
		eth.NewBridgeSmartContract(
			oracleConfig.Bridge.NodeURL, oracleConfig.Bridge.SmartContractAddress,
			oracleConfig.Bridge.DynamicTx, logger.Named("bridge_smart_contract")),
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to populate utxos and addresses. err: %w", err)
	}

	cardanoIndexerDbs := make(map[string]indexer.Database, len(oracleConfig.CardanoChains))

	for _, cardanoChainConfig := range oracleConfig.CardanoChains {
		indexerDB, err := indexerDb.NewDatabaseInit("",
			filepath.Join(appConfig.Settings.DbsPath, cardanoChainConfig.ChainID+".db"))
		if err != nil {
			return nil, fmt.Errorf("failed to open oracle indexer db for `%s`: %w", cardanoChainConfig.ChainID, err)
		}

		cardanoIndexerDbs[cardanoChainConfig.ChainID] = indexerDB
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

	secretsManager, err := common.GetSecretsManager(
		appConfig.ValidatorDataDir, appConfig.ValidatorConfigPath, true)
	if err != nil {
		return nil, fmt.Errorf("failed to create secrets manager: %w", err)
	}

	wallet, err := ethtxhelper.NewEthTxWalletFromSecretManager(secretsManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create blade wallet for oracle: %w", err)
	}

	oracleBridgeSC := eth.NewOracleBridgeSmartContract(
		appConfig.Bridge.NodeURL, appConfig.Bridge.SmartContractAddress,
		appConfig.Bridge.DynamicTx, logger.Named("oracle_bridge_smart_contract"))

	oracleBridgeSCWithWallet, err := eth.NewOracleBridgeSmartContractWithWallet(
		appConfig.Bridge.NodeURL, appConfig.Bridge.SmartContractAddress,
		wallet, appConfig.Bridge.DynamicTx, logger.Named("oracle_bridge_smart_contract"))
	if err != nil {
		return nil, fmt.Errorf("failed to create oracle bridge smart contract: %w", err)
	}

	bridgeSubmitter := bridge.NewBridgeSubmitter(ctx, oracleBridgeSCWithWallet, logger.Named("bridge_submitter"))

	oracle, err := oracle.NewOracle(
		ctx, oracleConfig, oracleBridgeSC, bridgeSubmitter, cardanoIndexerDbs,
		bridgingRequestStateManager, logger.Named("oracle"))
	if err != nil {
		return nil, fmt.Errorf("failed to create oracle. err %w", err)
	}

	ethBridgeSubmitter := eth_bridge.NewBridgeSubmitter(ctx, oracleBridgeSCWithWallet, logger.Named("bridge_submitter"))

	ethOracle, err := ethOracle.NewEthOracle(
		ctx, oracleConfig, oracleBridgeSC, ethBridgeSubmitter, ethIndexerDbs,
		bridgingRequestStateManager, logger.Named("eth_oracle"))
	if err != nil {
		return nil, fmt.Errorf("failed to create eth_oracle. err %w", err)
	}

	batcherManager, err := batchermanager.NewBatcherManager(
		ctx, batcherConfig, cardanoIndexerDbs, ethIndexerDbs, bridgingRequestStateManager, logger.Named("batcher"))
	if err != nil {
		return nil, fmt.Errorf("failed to create batcher manager: %w", err)
	}

	relayerBridgeSmartContract := eth.NewBridgeSmartContract(
		appConfig.Bridge.NodeURL, appConfig.Bridge.SmartContractAddress,
		appConfig.Bridge.DynamicTx, logger.Named("bridge_smart_contract"))

	relayerImitator, err := NewRelayerImitator(
		ctx, appConfig, bridgingRequestStateManager, relayerBridgeSmartContract, db,
		logger.Named("relayer_imitator"))
	if err != nil {
		return nil, fmt.Errorf("failed to create RelayerImitator. err: %w", err)
	}

	var apiObj *api.APIImpl

	if shouldRunAPI {
		apiControllers := []core.APIController{
			controllers.NewBridgingRequestStateController(
				bridgingRequestStateManager, logger.Named("bridging_request_state_controller")),
			controllers.NewCardanoTxController(
				oracleConfig, batcherConfig, logger.Named("cardano_tx_controller")),
			controllers.NewOracleStateController(
				cardanoIndexerDbs, getAddressesMap(oracleConfig.CardanoChains), logger.Named("oracle_state")),
		}

		apiObj, err = api.NewAPI(ctx, appConfig.APIConfig, apiControllers, logger.Named("api"))
		if err != nil {
			return nil, fmt.Errorf("failed to create api: %w", err)
		}
	}

	return &ValidatorComponentsImpl{
		ctx:               ctx,
		shouldRunAPI:      shouldRunAPI,
		db:                db,
		cardanoIndexerDbs: cardanoIndexerDbs,
		oracle:            oracle,
		ethOracle:         ethOracle,
		batcherManager:    batcherManager,
		relayerImitator:   relayerImitator,
		api:               apiObj,
		telemetry:         telemetry,
		logger:            logger,
	}, nil
}

func (v *ValidatorComponentsImpl) Start() error {
	v.logger.Debug("Starting ValidatorComponents")

	err := v.telemetry.Start()
	if err != nil {
		return err
	}

	err = v.oracle.Start()
	if err != nil {
		return fmt.Errorf("failed to start oracle. error: %w", err)
	}

	err = v.ethOracle.Start()
	if err != nil {
		return fmt.Errorf("failed to start eth_oracle. error: %w", err)
	}

	v.batcherManager.Start()

	if v.shouldRunAPI {
		go v.api.Start()
	}

	go v.relayerImitator.Start()

	v.errorCh = common.MakeSafeCh[error](1)

	go v.errorHandler()

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

	if err := v.oracle.Dispose(); err != nil {
		v.logger.Error("error while disposing oracle", "err", err)
		errs = append(errs, fmt.Errorf("error while disposing oracle. err: %w", err))
	}

	if err := v.ethOracle.Dispose(); err != nil {
		v.logger.Error("error while disposing eth_oracle", "err", err)
		errs = append(errs, fmt.Errorf("error while disposing eth_oracle. err: %w", err))
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

	if err := v.telemetry.Close(context.Background()); err != nil {
		v.logger.Error("Failed to close telemetry", "err", err)
		errs = append(errs, fmt.Errorf("failed to close telemetry. err: %w", err))
	}

	if err := v.errorCh.Close(); err != nil {
		v.logger.Error("Failed to close error channel", "err", err)
		errs = append(errs, fmt.Errorf("failed to close error channel. err: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors while disposing validatorcomponents. errors: %w", errors.Join(errs...))
	}

	v.logger.Info("ValidatorComponents disposed")

	return nil
}

func (v *ValidatorComponentsImpl) ErrorCh() (<-chan error, error) {
	return v.errorCh.ReadCh()
}

func (v *ValidatorComponentsImpl) errorHandler() {
outsideloop:
	for {
		select {
		case err := <-v.oracle.ErrorCh():
			wcerr := v.errorCh.Write(err)
			v.logger.Error("Failed to write to error channel", "err", wcerr)
		case <-v.ctx.Done():
			break outsideloop
		}
	}

	v.logger.Debug("Exiting validatorcomponents error handler")
}

func fixChainsAndAddresses(
	ctx context.Context,
	config *oracleCore.AppConfig,
	batcherConfig *batcherCore.BatcherManagerConfiguration,
	smartContract eth.IBridgeSmartContract,
	logger hclog.Logger,
) error {
	var (
		allRegisteredChains []eth.Chain
		validatorsData      []eth.ValidatorChainData
	)

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

	cardanoChains := make(map[string]*oracleCore.CardanoChainConfig)
	ethChains := make(map[string]*oracleCore.EthChainConfig)

	// handle config for oracles
	for _, regChain := range allRegisteredChains {
		chainID := common.ToStrChainID(regChain.Id)

		logger.Debug("Registered chain received", "chainID", chainID, "type", regChain.ChainType,
			"addr", regChain.AddressMultisig, "fee", regChain.AddressFeePayer)

		switch regChain.ChainType {
		case common.ChainTypeCardano:
			chainConfig, exists := config.CardanoChains[chainID]
			if !exists {
				return fmt.Errorf("no configuration for chain: %s", chainID)
			}

			chainConfig.BridgingAddresses = oracleCore.BridgingAddresses{
				BridgingAddress: regChain.AddressMultisig,
				FeeAddress:      regChain.AddressFeePayer,
			}

			err := common.RetryForever(ctx, 2*time.Second, func(ctxInner context.Context) (err error) {
				validatorsData, err = smartContract.GetValidatorsChainData(ctxInner, chainID)
				if err != nil {
					logger.Error("Failed to GetAllRegisteredChains while creating ValidatorComponents. Retrying...", "err", err)
				}

				return err
			})
			if err != nil {
				return fmt.Errorf("error while RetryForever of GetValidatorsChainData. err: %w", err)
			}

			multisigPolicyScript, multisigFeePolicyScript, err := cardanotx.GetPolicyScripts(validatorsData, logger)
			if err != nil {
				return fmt.Errorf("error while executing GetPolicyScripts. err: %w", err)
			}

			multisigAddr, feeAddr, err := cardanotx.GetMultisigAddresses(
				wallet.ResolveCardanoCliBinary(chainConfig.NetworkID), uint(chainConfig.NetworkMagic),
				multisigPolicyScript, multisigFeePolicyScript)
			if err != nil {
				return fmt.Errorf("error while executing GetMultisigAddresses. err: %w", err)
			}

			if multisigAddr != chainConfig.BridgingAddresses.BridgingAddress ||
				feeAddr != chainConfig.BridgingAddresses.FeeAddress {

				return fmt.Errorf("addresses do not match: (%s, %s) != (%s, %s)", multisigAddr, feeAddr,
					chainConfig.BridgingAddresses.BridgingAddress, chainConfig.BridgingAddresses.FeeAddress)
			} else {
				logger.Debug("Addresses are matching", "multisig", multisigAddr, "fee", feeAddr)
			}

			cardanoChains[chainID] = chainConfig
		case common.ChainTypeEVM:
			ethChainConfig, exists := config.EthChains[chainID]
			if !exists {
				return fmt.Errorf("no configuration for evm chain: %s", chainID)
			}

			ethChainConfig.BridgingAddresses = oracleCore.BridgingAddresses{
				BridgingAddress: regChain.AddressMultisig,
				FeeAddress:      regChain.AddressFeePayer,
			}

			ethChains[chainID] = ethChainConfig
		default:
			logger.Debug("Do not know how to handle chain type", "chainID", chainID, "type", regChain.ChainType)
		}
	}

	config.CardanoChains = cardanoChains
	config.EthChains = ethChains

	batcherChainConfigs := make([]batcherCore.ChainConfig, 0, len(batcherConfig.Chains))
	// handle config for batchers
	for _, regChain := range allRegisteredChains {
		chainID := common.ToStrChainID(regChain.Id)

		for _, chain := range batcherConfig.Chains {
			if chain.ChainID == chainID {
				batcherChainConfigs = append(batcherChainConfigs, chain)

				break
			}
		}
	}

	batcherConfig.Chains = batcherChainConfigs

	return nil
}

func getAddressesMap(cardanoChainConfig map[string]*oracleCore.CardanoChainConfig) map[string][]string {
	result := make(map[string][]string, len(cardanoChainConfig))

	for key, config := range cardanoChainConfig {
		result[key] = []string{config.BridgingAddresses.BridgingAddress, config.BridgingAddresses.FeeAddress}
	}

	return result
}
