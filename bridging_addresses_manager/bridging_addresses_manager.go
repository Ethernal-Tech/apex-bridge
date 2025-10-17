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
	isReward            bool
	logger              hclog.Logger
}

var _ common.BridgingAddressesManager = (*BridgingAddressesManagerImpl)(nil)

func NewBridgingAdressesManager(
	ctx context.Context,
	cardanoChains map[string]*oracleCore.CardanoChainConfig,
	bridgeSmartContract eth.IBridgeSmartContract,
	isReward bool,
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
	feeMultisigAddresses := make(map[uint8]string)
	feeMultisigPolicyScripts := make(map[uint8]*cardanowallet.PolicyScript)

	logPrefix := ""
	firstIndex := uint8(0)

	if isReward {
		logPrefix = "Reward "
		firstIndex = common.FirstRewardBridgingAddressIndex
	}

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
			return nil, fmt.Errorf("error while executing NewApexKeyHashes for %sbridging addresses component. err: %w", logPrefix, err)
		}

		numberOfAddresses, err := fetchAddressCounts(ctx, bridgeSmartContract, chainIDStr, isReward, logger)
		if err != nil {
			return nil, err
		}

		chainConfig := cardanoChains[chainIDStr]

		for i := range uint64(numberOfAddresses) {
			policyScripts := cardano.NewApexPolicyScripts(keyHashes, i+uint64(firstIndex))
			bridgingPaymentPolicyScripts[registeredChain.Id] =
				append(bridgingPaymentPolicyScripts[registeredChain.Id], policyScripts.Multisig.Payment)

			bridgingStakePolicyScripts[registeredChain.Id] =
				append(bridgingStakePolicyScripts[registeredChain.Id], policyScripts.Multisig.Stake)

			addrs, err := cardano.NewApexAddresses(
				cardanowallet.ResolveCardanoCliBinary(chainConfig.NetworkID), uint(chainConfig.NetworkMagic), policyScripts)
			if err != nil {
				return nil, fmt.Errorf("error while executing NewApexAddresses for %sbridging addresses component. err: %w", logPrefix, err)
			}

			bridgingPaymentAddresses[registeredChain.Id] =
				append(bridgingPaymentAddresses[registeredChain.Id], addrs.Multisig.Payment)

			bridgingStakeAddresses[registeredChain.Id] =
				append(bridgingStakeAddresses[registeredChain.Id], addrs.Multisig.Stake)

			if i == 0 && !isReward {
				feeMultisigAddresses[registeredChain.Id] = addrs.Fee.Payment
				feeMultisigPolicyScripts[registeredChain.Id] = policyScripts.Fee.Payment
			}
		}

		msg := fmt.Sprintf(
			"%sBridging addresses manager initialized for %s chain\n"+
				" - Payment addresses (%d): %v",
			logPrefix,
			chainIDStr,
			len(bridgingPaymentAddresses[registeredChain.Id]),
			bridgingPaymentAddresses[registeredChain.Id])

		if !isReward {
			msg += fmt.Sprintf("\n - Fee address: %s", feeMultisigAddresses[registeredChain.Id])
		}

		logger.Debug(msg)
	}

	return &BridgingAddressesManagerImpl{
		bridgingPaymentAddresses:     bridgingPaymentAddresses,
		bridgingStakeAddresses:       bridgingStakeAddresses,
		bridgingPaymentPolicyScripts: bridgingPaymentPolicyScripts,
		bridgingStakePolicyScripts:   bridgingStakePolicyScripts,
		feeMultisigAddresses:         feeMultisigAddresses,
		feeMultisigPolicyScripts:     feeMultisigPolicyScripts,
		isReward:                     isReward,
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
	if validIdx := b.validateIndex(index); !validIdx {
		return nil, false
	}

	return b.getPolicyScriptFromIndex(
		index,
		b.bridgingPaymentPolicyScripts[chainID],
	)
}

func (b *BridgingAddressesManagerImpl) GetStakePolicyScript(
	chainID uint8, index uint8,
) (*cardanowallet.PolicyScript, bool) {
	if validIdx := b.validateIndex(index); !validIdx {
		return nil, false
	}

	return b.getPolicyScriptFromIndex(
		index,
		b.bridgingStakePolicyScripts[chainID],
	)
}

func (b *BridgingAddressesManagerImpl) GetPaymentAddressFromIndex(chainID uint8, index uint8) (string, bool) {
	if validIdx := b.validateIndex(index); !validIdx {
		return "", false
	}

	return b.getAddressFromIndex(
		index,
		b.bridgingPaymentAddresses[chainID],
	)
}

func (b *BridgingAddressesManagerImpl) GetPaymentAddressIndex(chainID uint8, address string) (uint8, bool) {
	return b.findAddressIndex(
		b.bridgingPaymentAddresses[chainID],
		address,
	)
}

func (b *BridgingAddressesManagerImpl) GetStakeAddressFromIndex(chainID uint8, index uint8) (string, bool) {
	if validIdx := b.validateIndex(index); !validIdx {
		return "", false
	}

	return b.getAddressFromIndex(
		index,
		b.bridgingStakeAddresses[chainID],
	)
}

func (b *BridgingAddressesManagerImpl) GetStakeAddressIndex(chainID uint8, address string) (uint8, bool) {
	return b.findAddressIndex(
		b.bridgingStakeAddresses[chainID],
		address,
	)
}

func (b *BridgingAddressesManagerImpl) GetFirstIndexAddress(chainID uint8) (string, bool) {
	firstIndex := uint8(0)

	if b.isReward {
		firstIndex = common.FirstRewardBridgingAddressIndex
	}

	return b.getAddressFromIndex(
		firstIndex,
		b.bridgingPaymentAddresses[chainID],
	)
}

func (b *BridgingAddressesManagerImpl) GetFirstIndex() uint8 {
	if b.isReward {
		return common.FirstRewardBridgingAddressIndex
	}

	return uint8(0)
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
	isReward bool,
	logger hclog.Logger,
) (uint8, error) {
	var numberOfAddresses uint8

	// Choose appropriate function and error message based on isReward flag
	getCountFunc := bridgeSmartContract.GetBridgingAddressesCount
	errorMessage := "Failed to GetBridgingAddressesCount while creating Bridging Address Manager. Retrying..."
	finalErrorMessage := "error during RetryForever of GetBridgingAddressesCount"

	if isReward {
		getCountFunc = bridgeSmartContract.GetStakeBridgingAddressesCount
		errorMessage = "Failed to GetStakeBridgingAddressesCount while creating Bridging Address Manager. Retrying..."
		finalErrorMessage = "error during RetryForever of GetStakeBridgingAddressesCount"
	}

	err := common.RetryForever(ctx, 2*time.Second, func(ctxInner context.Context) error {
		var err error
		numberOfAddresses, err = getCountFunc(ctxInner, chainID)
		if err != nil {
			logger.Error(errorMessage, "chainID", chainID, "err", err)
		}
		return err
	})

	if err != nil {
		return 0, fmt.Errorf("%s for %s: %w", finalErrorMessage, chainID, err)
	}

	return numberOfAddresses, nil
}

func (b *BridgingAddressesManagerImpl) findAddressIndex(
	addresses []string,
	target string,
) (uint8, bool) {
	firstIndex := b.GetFirstIndex()

	for i, addr := range addresses {
		if addr == target {
			return uint8(i) + firstIndex, true //nolint:gosec
		}
	}

	return 0, false
}

func (b *BridgingAddressesManagerImpl) getAddressFromIndex(
	index uint8,
	addresses []string,
) (string, bool) {
	arrayIndex := index

	if b.isReward {
		arrayIndex = index - common.FirstRewardBridgingAddressIndex
	}

	if addresses == nil || int(arrayIndex) >= len(addresses) {
		return "", false
	}

	return addresses[arrayIndex], true
}

func (b *BridgingAddressesManagerImpl) validateIndex(index uint8) bool {
	return (b.isReward && index >= common.FirstRewardBridgingAddressIndex) ||
		(!b.isReward && index < common.FirstRewardBridgingAddressIndex)
}

func (b *BridgingAddressesManagerImpl) getPolicyScriptFromIndex(
	index uint8,
	policyScripts []*cardanowallet.PolicyScript,
) (*cardanowallet.PolicyScript, bool) {
	arrayIndex := index

	if b.isReward {
		arrayIndex = index - common.FirstRewardBridgingAddressIndex
	}

	if policyScripts == nil || int(arrayIndex) >= len(policyScripts) {
		return nil, false
	}

	return policyScripts[arrayIndex], true
}
