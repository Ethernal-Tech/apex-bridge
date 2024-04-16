package validatorcomponents

import (
	"context"
	"errors"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/hashicorp/go-hclog"

	"github.com/Ethernal-Tech/apex-bridge/batcher/batcher_manager"
	batcherCore "github.com/Ethernal-Tech/apex-bridge/batcher/core"
	oracleCore "github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle/oracle"
)

type ValidatorComponentsImpl struct {
	oracle         oracleCore.Oracle
	batcherManager batcherCore.BatcherManager
}

var _ core.ValidatorComponents = (*ValidatorComponentsImpl)(nil)

func NewValidatorComponents(
	appConfig *core.AppConfig,
	logger hclog.Logger,
) (*ValidatorComponentsImpl, error) {
	oracleConfig, batcherConfig := appConfig.SeparateConfigs()

	err := populateUtxosAndAddresses(
		context.Background(), oracleConfig,
		eth.NewBridgeSmartContract(oracleConfig.Bridge.NodeUrl, oracleConfig.Bridge.SmartContractAddress),
	)
	if err != nil {
		return nil, err
	}

	oracle, err := oracle.NewOracle(oracleConfig, logger.Named("oracle"))
	if err != nil {
		return nil, fmt.Errorf("failed to create oracle")
	}

	batcherManager, err := batcher_manager.NewBatcherManager(batcherConfig, logger.Named("batcher"))
	if err != nil {
		return nil, fmt.Errorf("failed to create batcher manager: %w", err)
	}

	return &ValidatorComponentsImpl{
		oracle:         oracle,
		batcherManager: batcherManager,
	}, nil
}

func (v *ValidatorComponentsImpl) Start() error {
	err := v.oracle.Start()
	if err != nil {
		return fmt.Errorf("failed to start oracle. error: %v", err)
	}

	err = v.batcherManager.Start()
	if err != nil {
		return fmt.Errorf("failed to start batchers. error: %v", err)
	}

	return nil
}

func (v *ValidatorComponentsImpl) Stop() error {
	errb := v.batcherManager.Stop()
	erro := v.oracle.Stop()

	return errors.Join(errb, erro)
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
