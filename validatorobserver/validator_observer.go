package validatorobserver

import (
	"context"
	"sync"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-hclog"
)

type ValidatorSetObserver struct {
	validatorSetStream  chan *Validators
	validators          Validators
	validatorSetPending bool
	bridgeSmartContract eth.IBridgeSmartContract
	logger              hclog.Logger
}

type ValidatorsChainData struct {
	Keys       []eth.ValidatorChainData
	SlotNumber uint64
}

type Validators struct {
	Data map[string]ValidatorsChainData
}

const (
	timeout = 1 * time.Second
)

var lock = sync.RWMutex{}

func NewValidatorSetObserver(
	ctx context.Context,
	bridgeSmartContract eth.IBridgeSmartContract,
	logger hclog.Logger,
) (*ValidatorSetObserver, error) {
	validatorSetPending, err := bridgeSmartContract.IsNewValidatorSetPending()
	if err != nil {
		validatorSetPending = false
	}

	validators := Validators{
		Data: make(map[string]ValidatorsChainData),
	}

	registeredChains, err := bridgeSmartContract.GetAllRegisteredChains(context.Background())
	if err != nil {
		registeredChains = []contractbinding.IBridgeStructsChain{}
	}

	for _, chain := range registeredChains {
		validatorsData, err := bridgeSmartContract.GetValidatorsChainData(context.Background(),
			common.ToStrChainID(chain.Id))
		if err != nil {
			continue
		}

		validatorKeys := []eth.ValidatorChainData{}
		for _, data := range validatorsData {
			validatorKeys = append(validatorKeys, eth.ValidatorChainData{
				Key: data.Key,
			})
		}

		slotNumber := uint64(0)

		lastObservedBlock, err := bridgeSmartContract.GetLastObservedBlock(ctx, common.ToStrChainID(chain.Id))
		if err == nil {
			slotNumber = lastObservedBlock.BlockSlot.Uint64()
		}

		validators.Data[common.ToStrChainID(chain.Id)] = ValidatorsChainData{
			Keys:       validatorKeys,
			SlotNumber: slotNumber,
		}
	}

	return &ValidatorSetObserver{
		validatorSetStream:  make(chan *Validators),
		validators:          validators,
		validatorSetPending: validatorSetPending,
		bridgeSmartContract: bridgeSmartContract,
		logger:              logger.Named("validator_set_observer"),
	}, nil
}

func (vs *ValidatorSetObserver) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				close(vs.validatorSetStream)

				return
			case <-time.After(timeout):
				if err := vs.execute(); err != nil {
					vs.logger.Error("error while executing", "err", err)
				}
			}
		}
	}()
}

func (vs *ValidatorSetObserver) execute() error {
	lock.Lock()
	oldState := vs.validatorSetPending
	validators := vs.validators
	lock.Unlock()

	isPending, err := vs.bridgeSmartContract.IsNewValidatorSetPending()
	if err != nil {
		return err
	}

	if isPending == oldState {
		return nil
	}

	if isPending {
		addedValidators, removedValidators, err := vs.bridgeSmartContract.GetVerificationKeys()
		if err != nil {
			return err
		}

		vs.removeValidators(validators, removedValidators)
		vs.addValidators(validators, addedValidators)
	}

	lock.Lock()
	defer lock.Unlock()

	vs.validatorSetPending = isPending
	if isPending {
		vs.validators = validators
		vs.validatorSetStream <- &validators
	} else {
		vs.validatorSetStream <- nil
	}

	return nil
}

func (vs *ValidatorSetObserver) IsValidatorSetPending() bool {
	lock.RLock()
	defer lock.RUnlock()

	return vs.validatorSetPending
}

func (vs *ValidatorSetObserver) GetVerificationKeys(chainID string) []eth.ValidatorChainData {
	lock.RLock()
	defer lock.RUnlock()

	return vs.validators.Data[chainID].Keys
}

func (vs *ValidatorSetObserver) GetValidatorSetReader() <-chan *Validators {
	return vs.validatorSetStream
}

func (vs *ValidatorSetObserver) removeValidators(validators Validators,
	removedValidators []ethcommon.Address) {
	for _, validator := range removedValidators {
		validatorIdx, err := vs.bridgeSmartContract.GetAddressValidatorIndex(validator)
		if err != nil {
			continue
		}

		for chainID := range validators.Data {
			validatorData := validators.Data[chainID]
			if int(validatorIdx) < len(validatorData.Keys) {
				validatorData.Keys = append(validatorData.Keys[:validatorIdx], validatorData.Keys[validatorIdx+1:]...)
				validators.Data[chainID] = validatorData
			}
		}
	}
}

func (vs *ValidatorSetObserver) addValidators(
	validators Validators, addedValidators []eth.ValidatorSet,
) {
	for _, validator := range addedValidators {
		for _, v := range validator.Validators {
			validatorData := validators.Data[common.ToStrChainID(validator.ChainId)]
			validatorData.Keys = append(validatorData.Keys, eth.ValidatorChainData{
				Key: v.Data.Key,
			})
			validators.Data[common.ToStrChainID(validator.ChainId)] = validatorData
		}
	}
}
