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
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/api"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/api/controllers"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	databaseaccess "github.com/Ethernal-Tech/apex-bridge/validatorcomponents/database_access"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
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
	logger          hclog.Logger
}

var _ core.ValidatorComponents = (*ValidatorComponentsImpl)(nil)

func NewValidatorComponents(
	ctx context.Context,
	appConfig *core.AppConfig,
	shouldRunAPI bool,
	logger hclog.Logger,
) (*ValidatorComponentsImpl, error) {
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

	oracle, err := oracle.NewOracle(ctx, oracleConfig, bridgingRequestStateManager, logger.Named("oracle"))
	if err != nil {
		return nil, fmt.Errorf("failed to create oracle. err %w", err)
	}

	batcherManager, err := batchermanager.NewBatcherManager(
		ctx, batcherConfig, bridgingRequestStateManager, logger.Named("batcher"))
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
		bridgingRequestStateController, err := controllers.NewBridgingRequestStateController(
			bridgingRequestStateManager, logger.Named("bridging_request_state_controller"))
		if err != nil {
			return nil, fmt.Errorf("failed to create BridgingRequestStateController: %w", err)
		}

		apiControllers := []core.APIController{bridgingRequestStateController}

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
		logger:          logger,
	}, nil
}

func (v *ValidatorComponentsImpl) Start() error {
	v.logger.Debug("Starting ValidatorComponents")

	if v.shouldRunAPI {
		go v.api.Start()
	}

	err := v.oracle.Start()
	if err != nil {
		return fmt.Errorf("failed to start oracle. error: %w", err)
	}

	v.batcherManager.Start()

	go v.relayerImitator.Start()

	v.logger.Debug("Started ValidatorComponents")

	return nil
}

func (v *ValidatorComponentsImpl) Dispose() error {
	errs := make([]error, 0)

	err := v.oracle.Dispose()
	if err != nil {
		v.logger.Error("error while disposing oracle", "err", err)
		errs = append(errs, fmt.Errorf("error while disposing oracle. err: %w", err))
	}

	if v.shouldRunAPI {
		err = v.api.Dispose()
		if err != nil {
			v.logger.Error("error while disposing api", "err", err)
			errs = append(errs, fmt.Errorf("error while disposing api. err: %w", err))
		}
	}

	err = v.db.Close()
	if err != nil {
		v.logger.Error("Failed to close validatorcomponents db", "err", err)
		errs = append(errs, fmt.Errorf("failed to close validatorcomponents db. err: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors while disposing validatorcomponents. errors: %w", errors.Join(errs...))
	}

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
					Index: uint32(x.TxIndex.Uint64()),
				},
				Output: indexer.TxOutput{
					Address: address,
					Amount:  x.Amount.Uint64(),
				},
			})
		}
	}

	resultChains := make(map[string]*oracleCore.CardanoChainConfig, len(allRegisteredChains))

	for _, regChain := range allRegisteredChains {
		chainConfig, exists := config.CardanoChains[regChain.Id]
		if !exists {
			return fmt.Errorf("no config for registered chain: %s", regChain.Id)
		}

		var availableUtxos eth.UTXOs

		err := common.RetryForever(ctx, 2*time.Second, func(ctxInner context.Context) (err error) {
			availableUtxos, err = smartContract.GetAvailableUTXOs(ctxInner, regChain.Id)
			if err != nil {
				l.Error(
					"Failed to GetAvailableUTXOs while creating ValidatorComponents. Retrying...",
					"chainId", regChain.Id, "err", err)
			}

			return err
		})

		if err != nil {
			return fmt.Errorf("error while RetryForever of GetAvailableUTXOs. err: %w", err)
		}

		l.Debug("done GetAvailableUTXOs", "chainID", regChain.Id, "availableUtxos", availableUtxos)

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

		resultChains[regChain.Id] = chainConfig

		l.Debug("updated chainConfig", "chainConfig", chainConfig)
	}

	config.CardanoChains = resultChains

	return nil
}
