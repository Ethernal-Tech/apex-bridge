package validatorobserver

import (
	"context"
	"math/big"
	"sync"
	"time"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	"github.com/Ethernal-Tech/bn256"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-hclog"
)

type ValidatorSetObserver struct {
	validatorSetPending bool
	verificationKeys    map[uint8]map[int][]byte
	bridgeSmartContract eth.IBridgeSmartContract
	logger              hclog.Logger
}

const (
	timeout = 1 * time.Second
)

var lock = sync.RWMutex{}

func NewValidatorSetObserver(
	bridgeSmartContract eth.IBridgeSmartContract,
	logger hclog.Logger,
) (*ValidatorSetObserver, error) {
	validatorSetPending, err := bridgeSmartContract.IsNewValidatorSetPending()
	if err != nil {
		validatorSetPending = false
	}

	verificationKeys := make(map[uint8]map[int][]byte)

	registeredChains, err := bridgeSmartContract.GetAllRegisteredChains(context.Background())
	if err != nil {
		registeredChains = []contractbinding.IBridgeStructsChain{}
	}

	for _, chain := range registeredChains {
		validatorsData, err := bridgeSmartContract.GetValidatorsChainData(context.Background(),
			common.ToStrChainID(chain.Id))
		if err != nil {
			verificationKeys[chain.Id] = make(map[int][]byte)

			for idx, data := range validatorsData {
				key, err := keyFormBigIntToBytes(data.Key)
				if err != nil {
					continue
				}

				verificationKeys[chain.Id][idx] = key
			}
		}
	}

	return &ValidatorSetObserver{
		validatorSetPending: validatorSetPending,
		verificationKeys:    verificationKeys,
		bridgeSmartContract: bridgeSmartContract,
		logger:              logger.Named("validator_set_observer"),
	}, nil
}

func (vs *ValidatorSetObserver) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
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
	verificationKeys := vs.verificationKeys
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

		vs.removeValidators(verificationKeys, removedValidators)
		vs.addValidators(verificationKeys, addedValidators)
	}

	lock.Lock()
	defer lock.Unlock()

	vs.validatorSetPending = isPending
	if isPending {
		vs.verificationKeys = verificationKeys
	}

	return nil
}

func (vs *ValidatorSetObserver) IsValidatorSetPending() bool {
	lock.RLock()
	defer lock.RUnlock()

	return vs.validatorSetPending
}

func (vs *ValidatorSetObserver) GetVerificationKeys(chainID uint8) map[int][]byte {
	lock.RLock()
	defer lock.RUnlock()

	return vs.verificationKeys[chainID]
}

func (vs *ValidatorSetObserver) removeValidators(verificationKeys map[uint8]map[int][]byte,
	removedValidators []ethcommon.Address) {
	for _, validator := range removedValidators {
		validatorIdx, err := vs.bridgeSmartContract.GetAddressValidatorIndex(validator)
		if err != nil {
			continue
		}

		for chainID, keys := range verificationKeys {
			if _, exists := keys[int(validatorIdx)]; exists {
				delete(verificationKeys[chainID], int(validatorIdx))
			}
		}
	}
}

func (vs *ValidatorSetObserver) addValidators(verificationKeys map[uint8]map[int][]byte,
	addedValidators []eth.ValidatorSet) {
	for _, validator := range addedValidators {
		for _, v := range validator.Validators {
			key, err := keyFormBigIntToBytes(v.Data.Key)
			if err != nil {
				return
			}

			validatorIdx, err := vs.bridgeSmartContract.GetAddressValidatorIndex(v.Addr)
			if err != nil {
				continue
			}

			verificationKeys[validator.ChainId][int(validatorIdx)] = key
		}
	}
}

func keyFormBigIntToBytes(key [4]*big.Int) ([]byte, error) {
	pub, err := bn256.UnmarshalPublicKeyFromBigInt(key)
	if err != nil {
		return nil, err
	}

	return pub.Marshal(), nil
}
