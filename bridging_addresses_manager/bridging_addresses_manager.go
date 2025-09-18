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
	bridgingPaymentAddresses           map[uint8][]string
	bridgingPaymentPolicyScripts       map[uint8][]*cardanowallet.PolicyScript
	bridgingStakeAddresses             map[uint8][]string
	bridgingStakePolicyScripts         map[uint8][]*cardanowallet.PolicyScript
	bridgingRewardPaymentAddresses     map[uint8][]string
	bridgingRewardPaymentPolicyScripts map[uint8][]*cardanowallet.PolicyScript
	bridgingRewardStakeAddresses       map[uint8][]string
	bridgingRewardStakePolicyScripts   map[uint8][]*cardanowallet.PolicyScript
	feeMultisigAddresses               map[uint8]string
	feeMultisigPolicyScripts           map[uint8]*cardanowallet.PolicyScript

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
	registeredChains, err := fetchRegisteredChains(ctx, bridgeSmartContract, logger)
	if err != nil {
		return nil, err
	}

	bridgingPaymentAddresses := make(map[uint8][]string)
	bridgingStakeAddresses := make(map[uint8][]string)
	bridgingPaymentPolicyScripts := make(map[uint8][]*cardanowallet.PolicyScript)
	bridgingStakePolicyScripts := make(map[uint8][]*cardanowallet.PolicyScript)
	bridgingRewardPaymentAddresses := make(map[uint8][]string)
	bridgingRewardStakeAddresses := make(map[uint8][]string)
	bridgingRewardPaymentPolicyScripts := make(map[uint8][]*cardanowallet.PolicyScript)
	bridgingRewardStakePolicyScripts := make(map[uint8][]*cardanowallet.PolicyScript)
	feeMultisigAddresses := make(map[uint8]string)
	feeMultisigPolicyScripts := make(map[uint8]*cardanowallet.PolicyScript)

	for _, registeredChain := range registeredChains {
		chainIDStr := common.ToStrChainID(registeredChain.Id)
		if !common.IsExistingChainID(chainIDStr) || registeredChain.ChainType != 0 {
			continue
		}

		validatorsData, err := fetchValidatorData(ctx, bridgeSmartContract, chainIDStr, logger)
		if err != nil {
			return nil, err
		}

		keyHashes, err := cardano.NewApexKeyHashes(validatorsData)
		if err != nil {
			return nil, fmt.Errorf("error while executing NewApexKeyHashes for bridging addresses component. err: %w", err)
		}

		numberOfAddresses, numberOfRewardAddresses, err := fetchAddressCounts(ctx, bridgeSmartContract, chainIDStr, logger)
		if err != nil {
			return nil, err
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

		for i := range uint64(numberOfRewardAddresses) {
			policyScripts := cardano.NewApexPolicyScripts(keyHashes, uint64(common.FirstRewardBridgingAddressIndex)+i)
			bridgingRewardPaymentPolicyScripts[registeredChain.Id] =
				append(bridgingRewardPaymentPolicyScripts[registeredChain.Id], policyScripts.Multisig.Payment)

			bridgingRewardStakePolicyScripts[registeredChain.Id] =
				append(bridgingRewardStakePolicyScripts[registeredChain.Id], policyScripts.Multisig.Stake)

			addrs, err := cardano.NewApexAddresses(
				cardanowallet.ResolveCardanoCliBinary(chainConfig.NetworkID), uint(chainConfig.NetworkMagic), policyScripts)
			if err != nil {
				return nil, fmt.Errorf("error while executing NewApexAddresses for bridging addresses component. err: %w", err)
			}

			bridgingRewardPaymentAddresses[registeredChain.Id] =
				append(bridgingRewardPaymentAddresses[registeredChain.Id], addrs.Multisig.Payment)

			bridgingRewardStakeAddresses[registeredChain.Id] =
				append(bridgingRewardStakeAddresses[registeredChain.Id], addrs.Multisig.Stake)
		}

		logger.Debug(
			fmt.Sprintf("Bridging addresses manager initialized for %s chain\n"+
				" - Payment addresses (%d): %v\n"+
				" - Reward addresses (%d): %v\n"+
				" - Fee address: %s",
				chainIDStr,
				len(bridgingPaymentAddresses[registeredChain.Id]),
				bridgingPaymentAddresses[registeredChain.Id],
				len(bridgingRewardPaymentAddresses[registeredChain.Id]),
				bridgingRewardPaymentAddresses[registeredChain.Id],
				feeMultisigAddresses[registeredChain.Id]))
	}

	return &BridgingAddressesManagerImpl{
		bridgingPaymentAddresses:           bridgingPaymentAddresses,
		bridgingStakeAddresses:             bridgingStakeAddresses,
		bridgingPaymentPolicyScripts:       bridgingPaymentPolicyScripts,
		bridgingStakePolicyScripts:         bridgingStakePolicyScripts,
		bridgingRewardPaymentAddresses:     bridgingRewardPaymentAddresses,
		bridgingRewardStakeAddresses:       bridgingRewardStakeAddresses,
		bridgingRewardPaymentPolicyScripts: bridgingRewardPaymentPolicyScripts,
		bridgingRewardStakePolicyScripts:   bridgingRewardStakePolicyScripts,
		feeMultisigAddresses:               feeMultisigAddresses,
		feeMultisigPolicyScripts:           feeMultisigPolicyScripts,
		cardanoChains:                      cardanoChains,
		ctx:                                ctx,
		bridgeSmartContract:                bridgeSmartContract,
		logger:                             logger,
	}, nil
}

func (b *BridgingAddressesManagerImpl) GetAllPaymentAddresses(chainID uint8, addressType common.AddressType) []string {
	switch addressType {
	case common.AddressTypeNormal:
		return b.bridgingPaymentAddresses[chainID]
	case common.AddressTypeReward:
		return b.bridgingRewardPaymentAddresses[chainID]
	default:
		return append(
			b.bridgingPaymentAddresses[chainID],
			b.bridgingRewardPaymentAddresses[chainID]...,
		)
	}
}

func (b *BridgingAddressesManagerImpl) GetAllStakeAddresses(chainID uint8, addressType common.AddressType) []string {
	switch addressType {
	case common.AddressTypeNormal:
		return b.bridgingStakeAddresses[chainID]
	case common.AddressTypeReward:
		return b.bridgingRewardStakeAddresses[chainID]
	default:
		return append(
			b.bridgingStakeAddresses[chainID],
			b.bridgingRewardStakeAddresses[chainID]...,
		)
	}
}

func (b *BridgingAddressesManagerImpl) GetPaymentPolicyScript(
	chainID uint8, index uint8,
) (*cardanowallet.PolicyScript, bool) {
	return getPolicyScriptFromIndex(
		index,
		b.bridgingPaymentPolicyScripts[chainID],
		b.bridgingRewardPaymentPolicyScripts[chainID],
	)
}

func (b *BridgingAddressesManagerImpl) GetStakePolicyScript(
	chainID uint8, index uint8,
) (*cardanowallet.PolicyScript, bool) {
	return getPolicyScriptFromIndex(
		index,
		b.bridgingStakePolicyScripts[chainID],
		b.bridgingRewardStakePolicyScripts[chainID],
	)
}

func (b *BridgingAddressesManagerImpl) GetPaymentAddressFromIndex(chainID uint8, index uint8) (string, bool) {
	return getAddressFromIndex(
		index,
		b.bridgingPaymentAddresses[chainID],
		b.bridgingRewardPaymentAddresses[chainID],
	)
}

func (b *BridgingAddressesManagerImpl) GetPaymentAddressIndex(chainID uint8, address string) (uint8, bool) {
	if index, found := findAddressIndex(b.bridgingPaymentAddresses[chainID], address, 0); found {
		return index, true
	}

	return findAddressIndex(
		b.bridgingRewardPaymentAddresses[chainID],
		address,
		common.FirstRewardBridgingAddressIndex,
	)
}

func (b *BridgingAddressesManagerImpl) GetStakeAddressFromIndex(chainID uint8, index uint8) (string, bool) {
	return getAddressFromIndex(
		index,
		b.bridgingStakeAddresses[chainID],
		b.bridgingRewardStakeAddresses[chainID],
	)
}

func (b *BridgingAddressesManagerImpl) GetStakeAddressIndex(chainID uint8, address string) (uint8, bool) {
	if index, found := findAddressIndex(b.bridgingStakeAddresses[chainID], address, 0); found {
		return index, true
	}

	return findAddressIndex(
		b.bridgingRewardStakeAddresses[chainID],
		address,
		common.FirstRewardBridgingAddressIndex,
	)
}

func (b *BridgingAddressesManagerImpl) GetFeeMultisigAddress(chainID uint8) string {
	return b.feeMultisigAddresses[chainID]
}

func (b *BridgingAddressesManagerImpl) GetFeeMultisigPolicyScript(chainID uint8) (*cardanowallet.PolicyScript, bool) {
	script, ok := b.feeMultisigPolicyScripts[chainID]

	return script, ok
}

func fetchRegisteredChains(
	ctx context.Context,
	bridgeSmartContract eth.IBridgeSmartContract,
	logger hclog.Logger,
) ([]eth.Chain, error) {
	var registeredChains []eth.Chain

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

	return registeredChains, nil
}

func fetchValidatorData(
	ctx context.Context,
	bridgeSmartContract eth.IBridgeSmartContract,
	chainID string,
	logger hclog.Logger,
) ([]eth.ValidatorChainData, error) {
	var validatorsData []eth.ValidatorChainData

	err := common.RetryForever(ctx, 2*time.Second, func(ctxInner context.Context) (err error) {
		validatorsData, err = bridgeSmartContract.GetValidatorsChainData(ctxInner, chainID)
		if err != nil {
			logger.Error("Failed to GetValidatorsChainData while creating Bridging Address Manager. Retrying...",
				"chainID", chainID, "err", err)
		}

		return err
	})
	if err != nil {
		return nil, fmt.Errorf("error while RetryForever of GetValidatorsChainData for %s. err: %w", chainID, err)
	}

	return validatorsData, nil
}

func fetchAddressCounts(
	ctx context.Context,
	bridgeSmartContract eth.IBridgeSmartContract,
	chainID string,
	logger hclog.Logger,
) (uint8, uint8, error) {
	var numberOfAddresses, numberOfRewardAddresses uint8

	err := common.RetryForever(ctx, 2*time.Second, func(ctxInner context.Context) (err error) {
		numberOfAddresses, err = bridgeSmartContract.GetBridgingAddressesCount(ctxInner, chainID)
		if err != nil {
			logger.Error("Failed to GetBridgingAddressesCount while creating Bridging Address Manager. Retrying...",
				"chainID", chainID, "err", err)
		}

		return err
	})
	if err != nil {
		return 0, 0, fmt.Errorf("error while RetryForever of GetBridgingAddressesCount for %s. err: %w", chainID, err)
	}

	err = common.RetryForever(ctx, 2*time.Second, func(ctxInner context.Context) (err error) {
		numberOfRewardAddresses, err = bridgeSmartContract.GetStakeBridgingAddressesCount(ctxInner, chainID)
		if err != nil {
			logger.Error("Failed to GetStakeBridgingAddressesCount while creating Bridging Address Manager. Retrying...",
				"chainID", chainID, "err", err)
		}

		return err
	})
	if err != nil {
		return 0, 0, fmt.Errorf("error while RetryForever of GetStakeBridgingAddressesCount for %s. err: %w", chainID, err)
	}

	return numberOfAddresses, numberOfRewardAddresses, nil
}

func findAddressIndex(
	addresses []string,
	target string,
	offset uint8,
) (uint8, bool) {
	for i, addr := range addresses {
		if addr == target {
			return uint8(i) + offset, true //nolint:gosec
		}
	}

	return 0, false
}

func getAddressFromIndex(
	index uint8,
	normal []string,
	reward []string,
) (string, bool) {
	var (
		addrs         []string
		adjustedIndex uint8
	)

	if index < common.FirstRewardBridgingAddressIndex {
		addrs = normal
		adjustedIndex = index
	} else {
		addrs = reward
		adjustedIndex = index - common.FirstRewardBridgingAddressIndex
	}

	if addrs == nil || int(adjustedIndex) >= len(addrs) {
		return "", false
	}

	return addrs[adjustedIndex], true
}

func getPolicyScriptFromIndex(
	index uint8,
	normal []*cardanowallet.PolicyScript,
	reward []*cardanowallet.PolicyScript,
) (*cardanowallet.PolicyScript, bool) {
	var (
		scripts       []*cardanowallet.PolicyScript
		adjustedIndex uint8
	)

	if index < common.FirstRewardBridgingAddressIndex {
		scripts = normal
		adjustedIndex = index
	} else {
		scripts = reward
		adjustedIndex = index - common.FirstRewardBridgingAddressIndex
	}

	if scripts == nil || int(adjustedIndex) >= len(scripts) {
		return nil, false
	}

	return scripts[adjustedIndex], true
}
