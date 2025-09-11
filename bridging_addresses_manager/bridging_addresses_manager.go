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
	var (
		registeredChains  []eth.Chain
		validatorsData    []eth.ValidatorChainData
		numberOfAddresses uint8
	)

	err := common.RetryForever(ctx, 2*time.Second, func(ctxInner context.Context) (err error) {
		registeredChains, err = bridgeSmartContract.GetAllRegisteredChains(ctxInner)
		if err != nil {
			logger.Error("Failed to GetAllRegisteredChains while creating Bridging Address Manager. Retrying...", "err", err)
		}

		return err
	})
	if err != nil {
		return nil, fmt.Errorf("error while RetryForever of GetAllRegisteredChains in Bridging Address Manager. err: %w", err)
	}

	bridgingPaymentAddresses := make(map[uint8][]string)
	bridgingStakeAddresses := make(map[uint8][]string)
	bridgingPaymentPolicyScripts := make(map[uint8][]*cardanowallet.PolicyScript)
	bridgingStakePolicyScripts := make(map[uint8][]*cardanowallet.PolicyScript)
	feeMultisigAddresses := make(map[uint8]string)
	feeMultisigPolicyScripts := make(map[uint8]*cardanowallet.PolicyScript)

	for _, registeredChain := range registeredChains {
		chainIDStr := common.ToStrChainID(registeredChain.Id)
		if !common.IsExistingChainID(chainIDStr) || registeredChain.ChainType != 0 {
			continue
		}

		err := common.RetryForever(ctx, 2*time.Second, func(ctxInner context.Context) (err error) {
			validatorsData, err = bridgeSmartContract.GetValidatorsChainData(ctxInner, chainIDStr)
			if err != nil {
				logger.Error("Failed to GetValidatorsChainData while creating Bridging Address Manager. Retrying...",
					"chainID", chainIDStr, "err", err)
			}

			return err
		})
		if err != nil {
			return nil, fmt.Errorf("error while RetryForever of GetValidatorsChainData for %s. err: %w", chainIDStr, err)
		}

		keyHashes, err := cardano.NewApexKeyHashes(validatorsData)
		if err != nil {
			return nil, fmt.Errorf("error while executing NewApexKeyHashes for bridging addresses component. err: %w", err)
		}

		err = common.RetryForever(ctx, 2*time.Second, func(ctxInner context.Context) (err error) {
			numberOfAddresses, err = bridgeSmartContract.GetBridgingAddressesCount(ctxInner, chainIDStr)
			if err != nil {
				logger.Error("Failed to GetBridgingAddressesCount while creating Bridging Address Manager. Retrying...",
					"chainID", chainIDStr, "err", err)
			}

			return err
		})
		if err != nil {
			return nil, fmt.Errorf("error while RetryForever of GetBridgingAddressesCount for %s. err: %w", chainIDStr, err)
		}

		chainConfig := cardanoChains[chainIDStr]

		for i := range uint64(numberOfAddresses) {
			policyScripts := cardano.NewApexPolicyScripts(keyHashes, i)
			bridgingPaymentPolicyScripts[registeredChain.Id] =
				append(bridgingPaymentPolicyScripts[registeredChain.Id], policyScripts.Multisig.Payment)

			bridgingStakePolicyScripts[registeredChain.Id] =
				append(bridgingStakePolicyScripts[registeredChain.Id], policyScripts.Multisig.Stake)

			addrs, err := cardano.NewApexAddresses(
				cardanowallet.ResolveCardanoCliBinary(chainConfig.NetworkID), uint(chainConfig.NetworkMagic), policyScripts)
			if err != nil {
				return nil, fmt.Errorf("error while executing NewApexAddresses for bridging addresses component. err: %w", err)
			}

			bridgingPaymentAddresses[registeredChain.Id] =
				append(bridgingPaymentAddresses[registeredChain.Id], addrs.Multisig.Payment)

			bridgingStakeAddresses[registeredChain.Id] =
				append(bridgingStakeAddresses[registeredChain.Id], addrs.Multisig.Stake)

			if i == 0 {
				feeMultisigAddresses[registeredChain.Id] = addrs.Fee.Payment
				feeMultisigPolicyScripts[registeredChain.Id] = policyScripts.Fee.Payment
			}
		}

		logger.Debug(
			fmt.Sprintf("Bridging addresses manager initialized for %s chain with %d payment addresses: %v and fee address %s",
				chainIDStr, len(bridgingPaymentAddresses[registeredChain.Id]),
				bridgingPaymentAddresses[registeredChain.Id], feeMultisigAddresses[registeredChain.Id]))
	}

	return &BridgingAddressesManagerImpl{
		bridgingPaymentAddresses:     bridgingPaymentAddresses,
		bridgingStakeAddresses:       bridgingStakeAddresses,
		bridgingPaymentPolicyScripts: bridgingPaymentPolicyScripts,
		bridgingStakePolicyScripts:   bridgingStakePolicyScripts,
		feeMultisigAddresses:         feeMultisigAddresses,
		feeMultisigPolicyScripts:     feeMultisigPolicyScripts,
		cardanoChains:                cardanoChains,
		ctx:                          ctx,
		bridgeSmartContract:          bridgeSmartContract,
		logger:                       logger,
	}, nil
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
