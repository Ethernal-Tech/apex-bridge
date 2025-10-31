package validatorobserver

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-hclog"
)

const (
	// Special flag for Blade chain
	BladeFlag = uint8(0xFF)
)

type ValidatorSetObserverImpl struct {
	context             context.Context
	validatorSetStream  chan *ValidatorsPerChain
	validators          ValidatorsPerChain
	validatorSetPending bool
	bridgeSmartContract eth.IBridgeSmartContract
	logger              hclog.Logger
	lock                sync.RWMutex
	timeout             time.Duration
}

var _ IValidatorSetObserver = (*ValidatorSetObserverImpl)(nil)

type ValidatorsChainData struct {
	Keys       []eth.ValidatorChainData
	SlotNumber uint64
}

type ValidatorsPerChain map[string]ValidatorsChainData

func NewValidatorSetObserver(
	ctx context.Context,
	bridgeSmartContract eth.IBridgeSmartContract,
	timeout time.Duration,
	logger hclog.Logger,
) (*ValidatorSetObserverImpl, error) {
	newValidatorSet := &ValidatorSetObserverImpl{
		context:             ctx,
		bridgeSmartContract: bridgeSmartContract,
		validatorSetStream:  make(chan *ValidatorsPerChain),
		logger:              logger.Named("validator_set_observer"),
		lock:                sync.RWMutex{},
		timeout:             timeout,
	}

	// isPending must not be initialized here, otherwise batchers won't be notified through execute method
	err := newValidatorSet.initValidatorSet()
	if err != nil {
		return newValidatorSet, fmt.Errorf("error initializing validator set: %w", err)
	}

	return newValidatorSet, nil
}

func (vs *ValidatorSetObserverImpl) Start() {
	go func() {
		for {
			select {
			case <-vs.context.Done():
				close(vs.validatorSetStream)

				return
			case <-time.After(vs.timeout):
				if err := vs.execute(); err != nil {
					vs.logger.Error("error while executing", "err", err)
				}
			}
		}
	}()
}

func (vs *ValidatorSetObserverImpl) execute() error {
	isPending, err := vs.bridgeSmartContract.IsNewValidatorSetPending()
	if err != nil {
		return err
	}

	// when validator set changes to pending update delta
	// when it goes back to not pending, do not update isPending state
	// after that, bridge needs to be restarted to fetch new validator set and addresses
	if !isPending || isPending == vs.IsValidatorSetPending() {
		return nil
	}

	var (
		addedValidators        []eth.ValidatorSet
		removedValidators      []ethcommon.Address
		lastObservedBlockSlots map[string]uint64
	)

	if isPending {
		addedValidators, removedValidators, err = vs.bridgeSmartContract.GetPendingValidatorSetDelta()
		if err != nil {
			return err
		}

		registeredChains, err := vs.bridgeSmartContract.GetAllRegisteredChains(vs.context)
		if err != nil {
			return fmt.Errorf("error getting registered chains: %w", err)
		}

		lastObservedBlockSlots = make(map[string]uint64, len(registeredChains))

		for _, chainID := range registeredChains {
			chainIDStr := common.ToStrChainID(chainID.Id)

			lastObservedBlock, err := vs.bridgeSmartContract.GetLastObservedBlock(vs.context, chainIDStr)
			if err != nil {
				return fmt.Errorf("error getting last observed block for chain: %s, err: %w", chainIDStr, err)
			}

			lastObservedBlockSlots[chainIDStr] = lastObservedBlock.BlockSlot.Uint64()
		}
	}

	var pendingValidatorSet *ValidatorsPerChain

	vs.lock.Lock()

	if isPending {
		validatorSetCopy := vs.validators.Clone()

		err := vs.removeValidators(validatorSetCopy, removedValidators)
		if err != nil {
			vs.lock.Unlock()

			return fmt.Errorf("error removing validators: %w", err)
		}

		vs.addValidators(validatorSetCopy, addedValidators)

		err = vs.updateSlotNumbers(validatorSetCopy, lastObservedBlockSlots)
		if err != nil {
			vs.lock.Unlock()

			return fmt.Errorf("error updating slot numbers: %w", err)
		}

		pendingValidatorSet = &validatorSetCopy
		vs.validators = validatorSetCopy
	}

	vs.validatorSetPending = isPending

	vs.lock.Unlock()
	vs.validatorSetStream <- pendingValidatorSet
	vs.logger.Info("validator set update", "isPending", isPending)

	return nil
}

func (vs *ValidatorSetObserverImpl) IsValidatorSetPending() bool {
	vs.lock.RLock()
	defer vs.lock.RUnlock()

	return vs.validatorSetPending
}

func (vs *ValidatorSetObserverImpl) GetValidatorSet(chainID string) []eth.ValidatorChainData {
	vs.lock.RLock()
	defer vs.lock.RUnlock()

	return vs.validators[chainID].Keys
}

func (vs *ValidatorSetObserverImpl) GetValidatorSetReader() <-chan *ValidatorsPerChain {
	return vs.validatorSetStream
}

func (vs *ValidatorSetObserverImpl) removeValidators(
	validators ValidatorsPerChain, removedValidators []ethcommon.Address,
) error {
	deletedMap := map[uint8]bool{}

	for _, validator := range removedValidators {
		validatorIdx, err := vs.bridgeSmartContract.GetAddressValidatorIndex(validator)
		if err != nil {
			return err
		}

		deletedMap[validatorIdx] = true
	}

	for chainID, chainData := range validators {
		newKeys := make([]eth.ValidatorChainData, 0, len(chainData.Keys)-len(deletedMap))

		idx := uint8(1)
		for _, key := range chainData.Keys {
			if !deletedMap[idx] {
				newKeys = append(newKeys, key)
			}

			idx++
		}

		chainData.Keys = newKeys
		validators[chainID] = chainData
	}

	return nil
}

func (vs *ValidatorSetObserverImpl) addValidators(validators ValidatorsPerChain, chainsDeltas []eth.ValidatorSet) {
	for _, chainDelta := range chainsDeltas {
		if chainDelta.ChainId == BladeFlag {
			continue
		}

		for _, v := range chainDelta.Validators {
			validatorData := validators[common.ToStrChainID(chainDelta.ChainId)]
			validatorData.Keys = append(validatorData.Keys, eth.ValidatorChainData{
				Key: v.Data.Key,
			})
			validators[common.ToStrChainID(chainDelta.ChainId)] = validatorData
		}
	}
}

func (vs *ValidatorSetObserverImpl) updateSlotNumbers(
	validators ValidatorsPerChain, lastObservedBlockSlots map[string]uint64,
) error {
	for chainID, chainData := range validators {
		if _, exists := lastObservedBlockSlots[chainID]; !exists {
			return fmt.Errorf("last observed block slot not found for chain: %s", chainID)
		}

		if chainData.SlotNumber != lastObservedBlockSlots[chainID] {
			chainData.SlotNumber = lastObservedBlockSlots[chainID]
			validators[chainID] = chainData
		}
	}

	return nil
}

func (vs *ValidatorSetObserverImpl) initValidatorSet() error {
	validators := make(map[string]ValidatorsChainData)

	var (
		registeredChains []eth.Chain
		err              error
	)

	if err = common.RetryForever(vs.context, time.Second*5, func(ctx context.Context) error {
		registeredChains, err = vs.bridgeSmartContract.GetAllRegisteredChains(ctx)
		if err != nil {
			vs.logger.Error("Error getting registered chain. Retry...", "err", err)
		}

		return err
	}); err != nil {
		return fmt.Errorf("error while RetryForever of GetAllRegisteredChains %w", err)
	}

	for _, chain := range registeredChains {
		var validatorsData []eth.ValidatorChainData

		if err = common.RetryForever(vs.context, 5*time.Second, func(ctx context.Context) error {
			validatorsData, err = vs.bridgeSmartContract.GetValidatorsChainData(ctx, common.ToStrChainID(chain.Id))
			if err != nil {
				vs.logger.Error("Error getting validators data for chain. Retry...", "chain id", chain.Id, "err", err)
			}

			return err
		}); err != nil {
			return fmt.Errorf("error while RetryForever of GetValidatorsData %w", err)
		}

		validatorKeys := []eth.ValidatorChainData{}
		for _, data := range validatorsData {
			validatorKeys = append(validatorKeys, eth.ValidatorChainData{
				Key: data.Key,
			})
		}

		var lastObservedBlock eth.CardanoBlock

		if err = common.RetryForever(vs.context, 5*time.Second, func(ctx context.Context) error {
			lastObservedBlock, err = vs.bridgeSmartContract.GetLastObservedBlock(ctx, common.ToStrChainID(chain.Id))
			if err != nil {
				vs.logger.Error("Error getting last cardano block for chain. Retry...", "chain id", chain.Id, "err", err)
			}

			return err
		}); err != nil {
			return fmt.Errorf("error while RetryForever of GetLastObservedBlock %w", err)
		}

		validators[common.ToStrChainID(chain.Id)] = ValidatorsChainData{
			Keys:       validatorKeys,
			SlotNumber: lastObservedBlock.BlockSlot.Uint64(),
		}
	}

	vs.lock.Lock()
	defer vs.lock.Unlock()

	vs.validators = validators

	return nil
}

func (v ValidatorsPerChain) Clone() ValidatorsPerChain {
	clone := make(map[string]ValidatorsChainData, len(v))

	for chainID, elem := range v {
		clone[chainID] = ValidatorsChainData{
			Keys:       append([]eth.ValidatorChainData(nil), elem.Keys...),
			SlotNumber: elem.SlotNumber,
		}
	}

	return clone
}
