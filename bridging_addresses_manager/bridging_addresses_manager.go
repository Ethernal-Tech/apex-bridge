package bridgingaddressmanager

import (
	"context"
	"fmt"
	"time"

	cardano "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	oracleCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
	"github.com/hashicorp/go-hclog"
)

type BridgingAddressesManagerImpl struct {
	bridgingPaymentAddresses     map[uint8][]string
	bridgingPaymentPolicyScripts map[uint8][]*cardanowallet.PolicyScript
	bridgingStakeAddresses       map[uint8][]string
	bridgingStakePolicyScripts   map[uint8][]*cardanowallet.PolicyScript
	feeMultisigAddresses         map[uint8]string
	feeMultisigPolicyScripts     map[uint8]*cardanowallet.PolicyScript

	custodialAddress       map[uint8]string
	custodialPolicyScripts map[uint8]*cardanowallet.PolicyScript

	cardanoChains       map[string]*oracleCore.CardanoChainConfig
	ctx                 context.Context
	bridgeSmartContract eth.IBridgeSmartContract
	logger              hclog.Logger
}

var _ common.BridgingAddressesManager = (*BridgingAddressesManagerImpl)(nil)

func NewBridgingAdressesManager(
	ctx context.Context,
	cardanoChains map[string]*oracleCore.CardanoChainConfig,
	bridgeSmartContract eth.IBridgeSmartContract,
	logger hclog.Logger,
) (common.BridgingAddressesManager, error) {
	registeredChains, err := getRegisteredChains(ctx, bridgeSmartContract, logger)
	if err != nil {
		return nil, fmt.Errorf("error while RetryForever of GetAllRegisteredChains in Bridging Address Manager. err: %w", err)
	}

	manager := &BridgingAddressesManagerImpl{
		bridgingPaymentAddresses:     make(map[uint8][]string),
		bridgingStakeAddresses:       make(map[uint8][]string),
		bridgingPaymentPolicyScripts: make(map[uint8][]*cardanowallet.PolicyScript),
		bridgingStakePolicyScripts:   make(map[uint8][]*cardanowallet.PolicyScript),
		feeMultisigAddresses:         make(map[uint8]string),
		feeMultisigPolicyScripts:     make(map[uint8]*cardanowallet.PolicyScript),
		custodialAddress:             make(map[uint8]string),
		custodialPolicyScripts:       make(map[uint8]*cardanowallet.PolicyScript),
		cardanoChains:                cardanoChains,
		ctx:                          ctx,
		bridgeSmartContract:          bridgeSmartContract,
		logger:                       logger,
	}

	for _, registeredChain := range registeredChains {
		registeredChainID := registeredChain.Id
		chainIDStr := common.ToStrChainID(registeredChainID)

		if !common.IsExistingChainID(chainIDStr) || registeredChain.ChainType != 0 {
			continue
		}

		validatorsData, err := getValidatorsChainData(ctx, bridgeSmartContract, chainIDStr, logger)
		if err != nil {
			return nil, fmt.Errorf("error while RetryForever of GetValidatorsChainData for %s. err: %w", chainIDStr, err)
		}

		keyHashes, err := cardano.NewApexKeyHashes(validatorsData)
		if err != nil {
			return nil, fmt.Errorf("error while executing NewApexKeyHashes for bridging addresses component. err: %w", err)
		}

		numberOfAddresses, err := getBridgingAddressesCount(ctx, bridgeSmartContract, chainIDStr, logger)
		if err != nil {
			return nil, fmt.Errorf("error while RetryForever of GetBridgingAddressesCount for %s. err: %w", chainIDStr, err)
		}

		chainConfig := cardanoChains[chainIDStr]

		for i := range uint64(numberOfAddresses) {
			if err := manager.buildBridgingAddress(registeredChainID, &keyHashes, chainConfig, i); err != nil {
				return nil, fmt.Errorf("failed to build bridging address %d for %s. err: %w", i, chainIDStr, err)
			}
		}

		if err := manager.buildCustodialAddress(registeredChainID, &keyHashes, chainConfig); err != nil {
			return nil, fmt.Errorf("failed to build custodial address for %s. err: %w", chainIDStr, err)
		}

		logger.Debug(
			fmt.Sprintf(
				"Bridging addresses manager initialized for %s chain with %d payment addresses: "+
					"%v, custodial address %s and fee address %s",
				chainIDStr, len(manager.bridgingPaymentAddresses[registeredChainID]),
				manager.bridgingPaymentAddresses[registeredChainID], manager.custodialAddress[registeredChainID],
				manager.feeMultisigAddresses[registeredChainID]))
	}

	return manager, nil
}

func (b *BridgingAddressesManagerImpl) GetAllPaymentAddresses(chainID uint8) []string {
	return b.bridgingPaymentAddresses[chainID]
}

func (b *BridgingAddressesManagerImpl) GetAllStakeAddresses(chainID uint8) []string {
	return b.bridgingStakeAddresses[chainID]
}

func (b *BridgingAddressesManagerImpl) GetPaymentPolicyScript(
	chainID uint8, index uint8,
) (*cardanowallet.PolicyScript, bool) {
	scripts, ok := b.bridgingPaymentPolicyScripts[chainID]
	if !ok || int(index) >= len(scripts) {
		return nil, false
	}

	return scripts[index], true
}

func (b *BridgingAddressesManagerImpl) GetStakePolicyScript(
	chainID uint8, index uint8,
) (*cardanowallet.PolicyScript, bool) {
	scripts, ok := b.bridgingStakePolicyScripts[chainID]
	if !ok || int(index) >= len(scripts) {
		return nil, false
	}

	return scripts[index], true
}

func (b *BridgingAddressesManagerImpl) GetPaymentAddressFromIndex(chainID uint8, index uint8) (string, bool) {
	addrs, ok := b.bridgingPaymentAddresses[chainID]
	if !ok || int(index) >= len(addrs) {
		return "", false
	}

	return addrs[index], true
}

func (b *BridgingAddressesManagerImpl) GetPaymentAddressIndex(chainID uint8, address string) (uint8, bool) {
	for i, addr := range b.bridgingPaymentAddresses[chainID] {
		if addr == address {
			return uint8(i), true //nolint:gosec
		}
	}

	return 0, false
}

func (b *BridgingAddressesManagerImpl) GetStakeAddressFromIndex(chainID uint8, index uint8) (string, bool) {
	addrs, ok := b.bridgingStakeAddresses[chainID]
	if !ok || int(index) >= len(addrs) {
		return "", false
	}

	return addrs[index], true
}

func (b *BridgingAddressesManagerImpl) GetStakeAddressIndex(chainID uint8, address string) (uint8, bool) {
	for i, addr := range b.bridgingStakeAddresses[chainID] {
		if addr == address {
			return uint8(i), true //nolint:gosec
		}
	}

	return 0, false
}

func (b *BridgingAddressesManagerImpl) GetFeeMultisigAddress(chainID uint8) string {
	return b.feeMultisigAddresses[chainID]
}

func (b *BridgingAddressesManagerImpl) GetFeeMultisigPolicyScript(chainID uint8) (*cardanowallet.PolicyScript, bool) {
	script, ok := b.feeMultisigPolicyScripts[chainID]

	return script, ok
}

func (b *BridgingAddressesManagerImpl) GetCustodialAddress(chainID uint8) (string, bool) {
	custodialAddr, ok := b.custodialAddress[chainID]

	return custodialAddr, ok
}

func (b *BridgingAddressesManagerImpl) buildBridgingAddress(
	chainID uint8,
	keyHashes *cardano.ApexKeyHashes,
	chainConfig *oracleCore.CardanoChainConfig,
	index uint64,
) error {
	policyScripts := cardano.NewApexPolicyScripts(*keyHashes, index)

	b.bridgingPaymentPolicyScripts[chainID] =
		append(b.bridgingPaymentPolicyScripts[chainID], policyScripts.Multisig.Payment)

	b.bridgingStakePolicyScripts[chainID] =
		append(b.bridgingStakePolicyScripts[chainID], policyScripts.Multisig.Stake)

	addrs, err := cardano.NewApexAddresses(
		cardanowallet.ResolveCardanoCliBinary(chainConfig.NetworkID), uint(chainConfig.NetworkMagic), policyScripts)
	if err != nil {
		return fmt.Errorf("error while executing NewApexAddresses for bridging addresses component. err: %w", err)
	}

	b.bridgingPaymentAddresses[chainID] =
		append(b.bridgingPaymentAddresses[chainID], addrs.Multisig.Payment)

	b.bridgingStakeAddresses[chainID] =
		append(b.bridgingStakeAddresses[chainID], addrs.Multisig.Stake)

	if index == 0 {
		b.feeMultisigAddresses[chainID] = addrs.Fee.Payment
		b.feeMultisigPolicyScripts[chainID] = policyScripts.Fee.Payment
	}

	return nil
}

func (b *BridgingAddressesManagerImpl) buildCustodialAddress(
	chainID uint8,
	keyHashes *cardano.ApexKeyHashes,
	chainConfig *oracleCore.CardanoChainConfig,
) error {
	custodialPolicyScript := cardano.NewCustodialPolicyScriptContainer(keyHashes.Multisig, 0)
	b.custodialPolicyScripts[chainID] = custodialPolicyScript.Payment

	addrs, err := cardano.NewAddressContainer(
		cardanowallet.ResolveCardanoCliBinary(chainConfig.NetworkID), uint(chainConfig.NetworkMagic), custodialPolicyScript)
	if err != nil {
		return fmt.Errorf("error while executing NewApexAddresses for custodial address. err: %w", err)
	}

	b.custodialAddress[chainID] = addrs.Payment

	return nil
}

func getRegisteredChains(
	ctx context.Context, bridge eth.IBridgeSmartContract, logger hclog.Logger,
) ([]eth.Chain, error) {
	var chains []eth.Chain

	err := common.RetryForever(ctx, 2*time.Second, func(inner context.Context) (err error) {
		chains, err = bridge.GetAllRegisteredChains(inner)
		if err != nil {
			logger.Error("Failed to GetAllRegisteredChains while creating Bridging Address Manager. Retrying...", "err", err)
		}

		return err
	})

	return chains, err
}

func getValidatorsChainData(
	ctx context.Context,
	bridge eth.IBridgeSmartContract,
	chainID string,
	logger hclog.Logger,
) ([]eth.ValidatorChainData, error) {
	var data []eth.ValidatorChainData

	err := common.RetryForever(ctx, 2*time.Second, func(ctxInner context.Context) (err error) {
		data, err = bridge.GetValidatorsChainData(ctxInner, chainID)
		if err != nil {
			logger.Error("Failed to GetValidatorsChainData while creating Bridging Address Manager. Retrying...", "chainID",
				chainID, "err", err)
		}

		return err
	})

	return data, err
}

func getBridgingAddressesCount(
	ctx context.Context,
	bridge eth.IBridgeSmartContract,
	chainID string,
	logger hclog.Logger,
) (uint8, error) {
	var count uint8

	err := common.RetryForever(ctx, 2*time.Second, func(ctxInner context.Context) (err error) {
		count, err = bridge.GetBridgingAddressesCount(ctxInner, chainID)
		if err != nil {
			logger.Error(
				"Failed to GetBridgingAddressesCount while creating Bridging Address Manager. Retrying...",
				"chainID", chainID, "err", err)
		}

		return err
	})

	return count, err
}
