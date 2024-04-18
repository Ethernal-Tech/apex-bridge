package validatorcomponents

import (
	"context"
	"errors"
	"fmt"
	"path"

	"github.com/Ethernal-Tech/apex-bridge/batcher/batcher_manager"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/api"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/api/controllers"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/database_access"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/hashicorp/go-hclog"

	batcherCore "github.com/Ethernal-Tech/apex-bridge/batcher/core"
	oracleCore "github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle/oracle"
)

const (
	MainComponentName            = "validatorcomponents"
	RelayerImitatorComponentName = "relayerimitator"
)

type ValidatorComponentsImpl struct {
	shouldRunApi    bool
	db              core.Database
	oracle          oracleCore.Oracle
	batcherManager  batcherCore.BatcherManager
	relayerImitator core.RelayerImitator
	api             core.Api
	logger          hclog.Logger
}

var _ core.ValidatorComponents = (*ValidatorComponentsImpl)(nil)

func NewValidatorComponents(
	appConfig *core.AppConfig,
	shouldRunApi bool,
	logger hclog.Logger,
) (*ValidatorComponentsImpl, error) {
	if err := common.CreateDirectoryIfNotExists(appConfig.Settings.DbsPath, 0770); err != nil {
		return nil, fmt.Errorf("failed to create directory for validator components database: %w", err)
	}

	db, err := database_access.NewDatabase(path.Join(appConfig.Settings.DbsPath, MainComponentName+".db"))
	if err != nil {
		return nil, fmt.Errorf("failed to open validator components database: %w", err)
	}

	bridgingRequestStateManager, err := NewBridgingRequestStateManager(db, logger.Named("bridging_request_state_manager"))
	if err != nil {
		return nil, fmt.Errorf("failed to create BridgingRequestStateManager. err: %w", err)
	}

	oracleConfig, batcherConfig := appConfig.SeparateConfigs()

	err = populateUtxosAndAddresses(
		context.Background(), oracleConfig,
		eth.NewBridgeSmartContract(oracleConfig.Bridge.NodeUrl, oracleConfig.Bridge.SmartContractAddress),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to populate utxos and addresses. err: %w", err)
	}

	oracle, err := oracle.NewOracle(oracleConfig, bridgingRequestStateManager, logger.Named("oracle"))
	if err != nil {
		return nil, fmt.Errorf("failed to create oracle. err %w", err)
	}

	batcherManager, err := batcher_manager.NewBatcherManager(batcherConfig, bridgingRequestStateManager, logger.Named("batcher"))
	if err != nil {
		return nil, fmt.Errorf("failed to create batcher manager: %w", err)
	}

	relayerBridgeSmartContract := eth.NewBridgeSmartContract(appConfig.Bridge.NodeUrl, appConfig.Bridge.SmartContractAddress)

	relayerImitator, err := NewRelayerImitator(appConfig, bridgingRequestStateManager, relayerBridgeSmartContract, db, logger.Named("relayer_imitator"))
	if err != nil {
		return nil, fmt.Errorf("failed to create RelayerImitator. err: %w", err)
	}

	var apiObj *api.ApiImpl

	if shouldRunApi {
		bridgingRequestStateController, err := controllers.NewBridgingRequestStateController(bridgingRequestStateManager, logger.Named("bridging_request_state_controller"))
		if err != nil {
			return nil, fmt.Errorf("failed to create BridgingRequestStateController: %w", err)
		}

		apiControllers := []core.ApiController{bridgingRequestStateController}

		apiObj, err = api.NewApi(appConfig.ApiConfig, apiControllers, logger.Named("api"))
		if err != nil {
			return nil, fmt.Errorf("failed to create api: %w", err)
		}
	}

	return &ValidatorComponentsImpl{
		shouldRunApi:    shouldRunApi,
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

	err := v.oracle.Start()
	if err != nil {
		return fmt.Errorf("failed to start oracle. error: %v", err)
	}

	err = v.batcherManager.Start()
	if err != nil {
		return fmt.Errorf("failed to start batchers. error: %v", err)
	}

	go v.relayerImitator.Start()

	if v.shouldRunApi {
		go v.api.Start()
	}

	v.logger.Debug("Started ValidatorComponents")

	return nil
}

func (v *ValidatorComponentsImpl) Stop() error {
	v.logger.Debug("Stopping ValidatorComponents")

	errb := v.batcherManager.Stop()
	erro := v.oracle.Stop()
	errri := v.relayerImitator.Stop()

	var errapi error
	if v.shouldRunApi {
		errapi = v.api.Stop()
	}

	errdb := v.db.Close()

	v.logger.Debug("Stopped ValidatorComponents")

	return errors.Join(errb, erro, errri, errapi, errdb)
}

func (v *ValidatorComponentsImpl) ErrorCh() <-chan error {
	return v.oracle.ErrorCh()
}

func populateUtxosAndAddresses(
	ctx context.Context,
	config *oracleCore.AppConfig,
	smartContract eth.IBridgeSmartContract,
) error {
	allRegisteredChains, err := smartContract.GetAllRegisteredChains(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve registered chains: %w", err)
	}

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

		availableUtxos, err := smartContract.GetAvailableUTXOs(ctx, regChain.Id)
		if err != nil {
			return fmt.Errorf("failed to retrieve available utxos for %s: %w", regChain.Id, err)
		}

		chainConfig.BridgingAddresses = oracleCore.BridgingAddresses{
			BridgingAddress: regChain.AddressMultisig,
			FeeAddress:      regChain.AddressFeePayer,
		}

		chainConfig.InitialUtxos = make([]*indexer.TxInputOutput, 0,
			len(availableUtxos.MultisigOwnedUTXOs)+len(availableUtxos.FeePayerOwnedUTXOs))

		// InitialUtxos wont be needed, initially they should be included with GetAvailableUTXOs
		//addUtxos(&chainConfig.InitialUtxos, regChain.AddressMultisig, regChain.Utxos.MultisigOwnedUTXOs)
		//addUtxos(&chainConfig.InitialUtxos, regChain.AddressFeePayer, regChain.Utxos.FeePayerOwnedUTXOs)
		addUtxos(&chainConfig.InitialUtxos, regChain.AddressMultisig, availableUtxos.MultisigOwnedUTXOs)
		addUtxos(&chainConfig.InitialUtxos, regChain.AddressFeePayer, availableUtxos.FeePayerOwnedUTXOs)

		resultChains[regChain.Id] = chainConfig
	}

	config.CardanoChains = resultChains

	return nil
}
