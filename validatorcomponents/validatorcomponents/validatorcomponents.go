package validatorcomponents

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/telemetry"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/api"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/api/controllers"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	databaseaccess "github.com/Ethernal-Tech/apex-bridge/validatorcomponents/database_access"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	indexerDb "github.com/Ethernal-Tech/cardano-infrastructure/indexer/db"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"

	batchermanager "github.com/Ethernal-Tech/apex-bridge/batcher/batcher_manager"
	batcherCore "github.com/Ethernal-Tech/apex-bridge/batcher/core"
	oracleCore "github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle/oracle"
)

const (
	MainComponentName            = "validatorcomponents"
	RelayerImitatorComponentName = "relayerimitator"
)

type ValidatorComponentsImpl struct {
	ctx             context.Context
	shouldRunAPI    bool
	db              core.Database
	oracle          oracleCore.Oracle
	batcherManager  batcherCore.BatcherManager
	relayerImitator core.RelayerImitator
	api             core.API
	telemetry       *telemetry.Telemetry
	logger          hclog.Logger
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
		ctx, oracleConfig,
		eth.NewBridgeSmartContract(
			oracleConfig.Bridge.NodeURL, oracleConfig.Bridge.SmartContractAddress,
			oracleConfig.Bridge.DynamicTx, logger.Named("bridge_smart_contract")),
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to populate utxos and addresses. err: %w", err)
	}

	indexerDbs := make(map[string]indexer.Database, len(oracleConfig.CardanoChains))

	for _, cardanoChainConfig := range oracleConfig.CardanoChains {
		indexerDB, err := indexerDb.NewDatabaseInit("",
			filepath.Join(appConfig.Settings.DbsPath, cardanoChainConfig.ChainID+".db"))
		if err != nil {
			return nil, fmt.Errorf("failed to open oracle indexer db for `%s`: %w", cardanoChainConfig.ChainID, err)
		}

		indexerDbs[cardanoChainConfig.ChainID] = indexerDB
	}

	oracle, err := oracle.NewOracle(ctx, oracleConfig, indexerDbs, bridgingRequestStateManager, logger.Named("oracle"))
	if err != nil {
		return nil, fmt.Errorf("failed to create oracle. err %w", err)
	}

	batcherManager, err := batchermanager.NewBatcherManager(
		ctx, batcherConfig, indexerDbs, bridgingRequestStateManager, logger.Named("batcher"))
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
				indexerDbs, getAddressesMap(oracleConfig.CardanoChains), logger.Named("oracle_state")),
		}

		apiObj, err = api.NewAPI(ctx, appConfig.APIConfig, apiControllers, logger.Named("api"))
		if err != nil {
			return nil, fmt.Errorf("failed to create api: %w", err)
		}
	}

	return &ValidatorComponentsImpl{
		ctx:             ctx,
		shouldRunAPI:    shouldRunAPI,
		db:              db,
		oracle:          oracle,
		batcherManager:  batcherManager,
		relayerImitator: relayerImitator,
		api:             apiObj,
		telemetry:       telemetry,
		logger:          logger,
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

	v.batcherManager.Start()

	if v.shouldRunAPI {
		go v.api.Start()
	}

	go v.relayerImitator.Start()

	v.logger.Debug("Started ValidatorComponents")

	return nil
}

func (v *ValidatorComponentsImpl) Dispose() error {
	v.logger.Info("Disposing ValidatorComponents")

	errs := make([]error, 0)

	if err := v.oracle.Dispose(); err != nil {
		v.logger.Error("error while disposing oracle", "err", err)
		errs = append(errs, fmt.Errorf("error while disposing oracle. err: %w", err))
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

	if len(errs) > 0 {
		return fmt.Errorf("errors while disposing validatorcomponents. errors: %w", errors.Join(errs...))
	}

	v.logger.Info("ValidatorComponents disposed")

	return nil
}

func (v *ValidatorComponentsImpl) ErrorCh() <-chan error {
	return v.oracle.ErrorCh()
}

func fixChainsAndAddresses(
	ctx context.Context,
	config *oracleCore.AppConfig,
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

	resultChains := make(map[string]*oracleCore.CardanoChainConfig, len(allRegisteredChains))

	for _, regChain := range allRegisteredChains {
		chainID := common.ToStrChainID(regChain.Id)
		chainConfig, exists := config.CardanoChains[chainID]
		if !exists {
			return fmt.Errorf("no config for registered chain: %s", chainID)
		}

		logger.Debug("Registered chain received", "chainID", chainID, "type", regChain.ChainType,
			"addr", regChain.AddressMultisig, "fee", regChain.AddressFeePayer)

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

		// should handle evm too
		switch regChain.ChainType {
		case common.ChainTypeCardano:
			multisigAddr, feeAddr, err := getCardanoAddresses(
				wallet.ResolveCardanoCliBinary(chainConfig.NetworkID), chainConfig.NetworkMagic, validatorsData)
			if err != nil {
				return fmt.Errorf("error while RetryForever of GetValidatorsChainData. err: %w", err)
			}

			if multisigAddr != chainConfig.BridgingAddresses.BridgingAddress ||
				feeAddr != chainConfig.BridgingAddresses.FeeAddress {

				return fmt.Errorf("addresses do not match: (%s, %s) != (%s, %s)", multisigAddr, feeAddr,
					chainConfig.BridgingAddresses.BridgingAddress, chainConfig.BridgingAddresses.FeeAddress)
			} else {
				logger.Debug("Addresses are matching", "multisig", multisigAddr, "fee", feeAddr)
			}

			resultChains[chainID] = chainConfig
		default:
			logger.Debug("Do not know how to handle chain type", "chainID", chainID, "type", regChain.ChainType)
		}
	}

	config.CardanoChains = resultChains

	return nil
}

func getAddressesMap(cardanoChainConfig map[string]*oracleCore.CardanoChainConfig) map[string][]string {
	result := make(map[string][]string, len(cardanoChainConfig))

	for key, config := range cardanoChainConfig {
		result[key] = []string{config.BridgingAddresses.BridgingAddress, config.BridgingAddresses.FeeAddress}
	}

	return result
}

func getCardanoAddresses(
	cardanoCliBinary string, networkMagic uint32, validatorsData []eth.ValidatorChainData,
) (string, string, error) {
	multisigKeyHashes := make([]string, len(validatorsData))
	multisigFeeKeyHashes := make([]string, len(validatorsData))

	for i, x := range validatorsData {
		keyHash, err := wallet.GetKeyHash(x.VerifyingKey[:])
		if err != nil {
			return "", "", err
		}

		keyHashFee, err := wallet.GetKeyHash(x.VerifyingKeyFee[:])
		if err != nil {
			return "", "", err
		}

		multisigKeyHashes[i] = keyHash
		multisigFeeKeyHashes[i] = keyHashFee
	}

	multisigPolicyScript := wallet.NewPolicyScript(
		multisigKeyHashes, int(common.GetRequiredSignaturesForConsensus(uint64(len(validatorsData)))))
	multisigFeePolicyScript := wallet.NewPolicyScript(
		multisigFeeKeyHashes, int(common.GetRequiredSignaturesForConsensus(uint64(len(validatorsData)))))

	multisigAddress, err := cardanotx.GetAddressFromPolicyScript(
		cardanoCliBinary, uint(networkMagic), multisigPolicyScript)
	if err != nil {
		return "", "", err
	}

	multisigFeeAddress, err := cardanotx.GetAddressFromPolicyScript(
		cardanoCliBinary, uint(networkMagic), multisigFeePolicyScript)
	if err != nil {
		return "", "", err
	}

	return multisigAddress, multisigFeeAddress, nil
}
