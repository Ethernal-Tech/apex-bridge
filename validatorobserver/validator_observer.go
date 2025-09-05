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

type ValidatorSetObserverImpl struct {
	context               context.Context
	validatorSetBatcherCh chan *ValidatorsPerChain
	validatorSetOracleCh  chan *ValidatorsPerChain
	validators            ValidatorsPerChain
	validatorSetPending   bool
	bridgeSmartContract   eth.IBridgeSmartContract
	logger                hclog.Logger
	lock                  sync.RWMutex
}

var _ IValidatorSetObserver = (*ValidatorSetObserverImpl)(nil)

type ValidatorsChainData struct {
	Keys []eth.ValidatorChainData
	Slot eth.CardanoBlock
}

type ValidatorsPerChain map[string]ValidatorsChainData

const (
	timeout = 30 * time.Second
)

func NewValidatorSetObserver(
	ctx context.Context,
	bridgeSmartContract eth.IBridgeSmartContract,
	logger hclog.Logger,
) (*ValidatorSetObserverImpl, error) {
	newValidatorSet := &ValidatorSetObserverImpl{
		context:               ctx,
		bridgeSmartContract:   bridgeSmartContract,
		validatorSetBatcherCh: make(chan *ValidatorsPerChain),
		validatorSetOracleCh:  make(chan *ValidatorsPerChain),
		logger:                logger.Named("validator_set_observer"),
		lock:                  sync.RWMutex{},
	}

	validatorSetPending, err := bridgeSmartContract.IsNewValidatorSetPending()
	if err != nil {
		return newValidatorSet, fmt.Errorf("error checking if new validator set is pending: %w", err)
	}

	err = newValidatorSet.initValidatorSet()
	if err != nil {
		return newValidatorSet, fmt.Errorf("error initializing validator set: %w", err)
	}

	newValidatorSet.validatorSetPending = validatorSetPending

	return newValidatorSet, nil
}

func (vs *ValidatorSetObserverImpl) Start() {
	go func() {
		for {
			select {
			case <-vs.context.Done():
				close(vs.validatorSetBatcherCh)
				close(vs.validatorSetOracleCh)

				return
			case <-time.After(timeout):
				if err := vs.execute(); err != nil {
					vs.logger.Error("error while executing", "err", err)
				}
			}
		}
	}()
}

func (vs *ValidatorSetObserverImpl) execute() error {
	// Check if the initialization is complete
	if len(vs.validators) == 0 {
		// try to initialize the validator set
		err := vs.initValidatorSet()
		if err != nil {
			return fmt.Errorf("error initializing validator set, err: %w", err)
		}
	}

	isPending, err := vs.bridgeSmartContract.IsNewValidatorSetPending()
	if err != nil {
		return err
	}

	// check if same state
	if isPending == vs.IsValidatorSetPending() {
		return nil
	}

	var (
		addedValidators    []eth.ValidatorSet
		removedValidators  []ethcommon.Address
		lastObservedBlocks map[string]eth.CardanoBlock
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

		lastObservedBlocks = make(map[string]eth.CardanoBlock, len(registeredChains))

		for _, chainID := range registeredChains {
			chainIDStr := common.ToStrChainID(chainID.Id)

			lastObservedBlock, err := vs.bridgeSmartContract.GetLastObservedBlock(vs.context, chainIDStr)
			if err != nil {
				return fmt.Errorf("error getting last observed block for chain: %s, err: %w", chainIDStr, err)
			}

			lastObservedBlocks[chainIDStr] = lastObservedBlock
		}
	}

	var pendingValidatorSet *ValidatorsPerChain

	vs.lock.Lock()

	vs.validatorSetPending = isPending

	if isPending {
		validatorSetCopy := vs.validators.Clone()

		err := vs.removeValidators(validatorSetCopy, removedValidators)
		if err != nil {
			vs.lock.Unlock()

			return fmt.Errorf("error removing validators: %w", err)
		}

		vs.addValidators(validatorSetCopy, addedValidators)

		err = vs.updateSlots(validatorSetCopy, lastObservedBlocks)
		if err != nil {
			vs.lock.Unlock()

			return fmt.Errorf("error updating slot numbers: %w", err)
		}

		pendingValidatorSet = &validatorSetCopy
		vs.validators = validatorSetCopy
	}

	vs.lock.Unlock()

	vs.validatorSetBatcherCh <- pendingValidatorSet

	// notify oracle to reset indexer at the start of VS update
	if isPending {
		vs.validatorSetOracleCh <- pendingValidatorSet
	}

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

func (vs *ValidatorSetObserverImpl) GetValidatorSetBatcherReader() <-chan *ValidatorsPerChain {
	return vs.validatorSetBatcherCh
}

func (vs *ValidatorSetObserverImpl) GetValidatorSetOracleReader() <-chan *ValidatorsPerChain {
	return vs.validatorSetOracleCh
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
		if chainDelta.ChainId == uint8(0xFF) {
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

func (vs *ValidatorSetObserverImpl) updateSlots(
	validators ValidatorsPerChain, lastObservedBlocks map[string]eth.CardanoBlock,
) error {
	for chainID, chainData := range validators {
		if block, exists := lastObservedBlocks[chainID]; !exists {
			return fmt.Errorf("last observed block not found for chain: %s", chainID)
		} else {
			chainData.Slot = block
			validators[chainID] = chainData
		}
	}

	return nil
}

func (vs *ValidatorSetObserverImpl) initValidatorSet() error {
	validators := make(map[string]ValidatorsChainData)

	registeredChains, err := vs.bridgeSmartContract.GetAllRegisteredChains(vs.context)
	if err != nil {
		return fmt.Errorf("error getting registered chains: %w", err)
	}

	for _, chain := range registeredChains {
		validatorsData, err := vs.bridgeSmartContract.GetValidatorsChainData(vs.context, common.ToStrChainID(chain.Id))
		if err != nil {
			return fmt.Errorf("error getting validators chain data for chain %s: %w", common.ToStrChainID(chain.Id), err)
		}

		validatorKeys := []eth.ValidatorChainData{}
		for _, data := range validatorsData {
			validatorKeys = append(validatorKeys, eth.ValidatorChainData{
				Key: data.Key,
			})
		}

		lastObservedBlock, err := vs.bridgeSmartContract.GetLastObservedBlock(vs.context, common.ToStrChainID(chain.Id))
		if err != nil {
			return fmt.Errorf("error getting last observed block for chain %s: %w", common.ToStrChainID(chain.Id), err)
		}

		validators[common.ToStrChainID(chain.Id)] = ValidatorsChainData{
			Keys: validatorKeys,
			Slot: lastObservedBlock,
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
			Keys: append([]eth.ValidatorChainData(nil), elem.Keys...),
			Slot: elem.Slot,
		}
	}

	return clone
}
