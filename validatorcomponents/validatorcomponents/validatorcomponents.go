package validatorcomponents

import (
	"context"
	"errors"
	"fmt"
	"path"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/telemetry"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/api"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/api/controllers"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	databaseaccess "github.com/Ethernal-Tech/apex-bridge/validatorcomponents/database_access"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	indexerDb "github.com/Ethernal-Tech/cardano-infrastructure/indexer/db"
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

	db, err := databaseaccess.NewDatabase(path.Join(appConfig.Settings.DbsPath, MainComponentName+".db"))
	if err != nil {
		return nil, fmt.Errorf("failed to open validator components database: %w", err)
	}

	bridgingRequestStateManager, err := NewBridgingRequestStateManager(db, logger.Named("bridging_request_state_manager"))
	if err != nil {
		return nil, fmt.Errorf("failed to create BridgingRequestStateManager. err: %w", err)
	}

	oracleConfig, batcherConfig := appConfig.SeparateConfigs()

	err = populateUtxosAndAddresses(
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
			path.Join(appConfig.Settings.DbsPath, cardanoChainConfig.ChainID+".db"))
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
				oracleConfig, batcherConfig, logger.Named("bridging_request_state_controller")),
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

func populateUtxosAndAddresses(
	ctx context.Context,
	config *oracleCore.AppConfig,
	smartContract eth.IBridgeSmartContract,
	logger hclog.Logger,
) error {
	l := logger.Named("populateUtxosAndAddresses")

	l.Debug("trying to populate utxos and addresses")

	var allRegisteredChains []contractbinding.IBridgeStructsChain

	err := common.RetryForever(ctx, 2*time.Second, func(ctxInner context.Context) (err error) {
		allRegisteredChains, err = smartContract.GetAllRegisteredChains(ctxInner)
		if err != nil {
			l.Error("Failed to GetAllRegisteredChains while creating ValidatorComponents. Retrying...", "err", err)
		}

		return err
	})

	if err != nil {
		return fmt.Errorf("error while RetryForever of GetAllRegisteredChains. err: %w", err)
	}

	l.Debug("done GetAllRegisteredChains", "allRegisteredChains", allRegisteredChains)

	addUtxos := func(outputs *[]*indexer.TxInputOutput, address string, utxos []eth.UTXO) {
		for _, x := range utxos {
			*outputs = append(*outputs, &indexer.TxInputOutput{
				Input: indexer.TxInput{
					Hash:  x.TxHash,
					Index: uint32(x.TxIndex),
				},
				Output: indexer.TxOutput{
					Address: address,
					Amount:  x.Amount,
				},
			})
		}
	}

	resultChains := make(map[string]*oracleCore.CardanoChainConfig, len(allRegisteredChains))

	for _, regChain := range allRegisteredChains {
		chainID := common.ToStrChainID(regChain.Id)
		chainConfig, exists := config.CardanoChains[chainID]
		if !exists {
			return fmt.Errorf("no config for registered chain: %s", chainID)
		}

		var availableUtxos eth.UTXOs

		err := common.RetryForever(ctx, 2*time.Second, func(ctxInner context.Context) (err error) {
			availableUtxos, err = smartContract.GetAvailableUTXOs(ctxInner, chainID)
			if err != nil {
				l.Error(
					"Failed to GetAvailableUTXOs while creating ValidatorComponents. Retrying...",
					"chainId", chainID, "err", err)
			}

			return err
		})

		if err != nil {
			return fmt.Errorf("error while RetryForever of GetAvailableUTXOs. err: %w", err)
		}

		l.Debug("done GetAvailableUTXOs", "chainID", chainID, "availableUtxos", availableUtxos)

		chainConfig.BridgingAddresses = oracleCore.BridgingAddresses{
			BridgingAddress: regChain.AddressMultisig,
			FeeAddress:      regChain.AddressFeePayer,
		}

		chainConfig.InitialUtxos = make([]*indexer.TxInputOutput, 0,
			len(availableUtxos.MultisigOwnedUTXOs)+len(availableUtxos.FeePayerOwnedUTXOs))

		// InitialUtxos wont be needed, initially they should be included with GetAvailableUTXOs
		// addUtxos(&chainConfig.InitialUtxos, regChain.AddressMultisig, regChain.Utxos.MultisigOwnedUTXOs)
		// addUtxos(&chainConfig.InitialUtxos, regChain.AddressFeePayer, regChain.Utxos.FeePayerOwnedUTXOs)
		addUtxos(&chainConfig.InitialUtxos, regChain.AddressMultisig, availableUtxos.MultisigOwnedUTXOs)
		addUtxos(&chainConfig.InitialUtxos, regChain.AddressFeePayer, availableUtxos.FeePayerOwnedUTXOs)

		resultChains[chainID] = chainConfig

		l.Debug("updated chainConfig", "chainConfig", chainConfig)
	}

	config.CardanoChains = resultChains

	return nil
}
