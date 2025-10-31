package validatorobserver

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const chainID = "prime"

func TestAddValidators(t *testing.T) {
	logger := hclog.NewNullLogger()
	bridgeSmartContract := &eth.BridgeSmartContractMock{}
	observer := &ValidatorSetObserverImpl{
		validatorSetPending: false,
		validators:          ValidatorsPerChain{},
		bridgeSmartContract: bridgeSmartContract,
		logger:              logger,
	}

	t.Run("Add validators to existing chain", func(t *testing.T) {
		key1 := [4]*big.Int{big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0)}
		key2 := [4]*big.Int{big.NewInt(5), big.NewInt(6), big.NewInt(7), big.NewInt(8)}

		validators := make(map[string]ValidatorsChainData)

		validators[chainID] = ValidatorsChainData{
			Keys: []eth.ValidatorChainData{
				{
					Key: key1,
				},
			},
		}

		observer.validators = validators

		addedValidators := []eth.ValidatorSet{
			{
				ChainId: common.ToNumChainID(chainID),
				Validators: []contractbinding.IBridgeStructsValidatorAddressChainData{
					{
						Data: contractbinding.IBridgeStructsValidatorChainData{
							Key: key2,
						},
					},
				},
			},
		}

		observer.addValidators(validators, addedValidators)

		validatorKeys := observer.GetValidatorSet(chainID)

		assert.Equal(t, 2, len(validatorKeys), "Expected two validator keys")
		assert.Equal(t, key1, validatorKeys[0].Key, "Expected first key unchanged")
		assert.Equal(t, key2, validatorKeys[1].Key, "Expected second key added")
	})
}

func TestRemoveValidators(t *testing.T) {
	logger := hclog.NewNullLogger()
	bridgeSmartContract := &eth.BridgeSmartContractMock{}
	bridgeSmartContract.On("GetAddressValidatorIndex",
		ethcommon.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")).Return(uint8(1), nil)

	observer := &ValidatorSetObserverImpl{
		validatorSetPending: false,
		validators:          ValidatorsPerChain{},
		bridgeSmartContract: bridgeSmartContract,
		logger:              logger,
	}

	t.Run("Remove existing validator", func(t *testing.T) {
		key1 := [4]*big.Int{big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0)}
		key2 := [4]*big.Int{big.NewInt(5), big.NewInt(6), big.NewInt(7), big.NewInt(8)}

		validators := make(map[string]ValidatorsChainData)

		validators[chainID] = ValidatorsChainData{
			Keys: []eth.ValidatorChainData{
				{
					Key: key1,
				},
				{
					Key: key2,
				},
			},
		}

		observer.validators = validators

		removedValidators := []ethcommon.Address{
			ethcommon.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"),
		}

		err := observer.removeValidators(validators, removedValidators)
		require.NoError(t, err)

		validatorKeys := observer.GetValidatorSet(chainID)

		assert.Equal(t, 1, len(validatorKeys), "Expected validator to be removed")
		assert.Equal(t, key2, validatorKeys[0].Key, "Expected second key")
	})

	t.Run("Remove only validator", func(t *testing.T) {
		key1 := [4]*big.Int{big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0)}

		validators := make(map[string]ValidatorsChainData)

		validators[chainID] = ValidatorsChainData{
			Keys: []eth.ValidatorChainData{
				{
					Key: key1,
				},
			},
		}

		observer.validators = validators

		removedValidators := []ethcommon.Address{
			ethcommon.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"),
		}

		err := observer.removeValidators(validators, removedValidators)
		require.NoError(t, err)

		validatorKeys := observer.GetValidatorSet(chainID)

		assert.Equal(t, 0, len(validatorKeys), "Expected validator to be removed")
	})

	bridgeSmartContract.On("GetAddressValidatorIndex",
		ethcommon.HexToAddress("0x1234567890abcdef1234567890abcdef12345679")).Return(uint8(3), nil)
	t.Run("Remove existing validator", func(t *testing.T) {
		key1 := [4]*big.Int{big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0)}
		key2 := [4]*big.Int{big.NewInt(5), big.NewInt(6), big.NewInt(7), big.NewInt(8)}
		key3 := [4]*big.Int{big.NewInt(5), big.NewInt(6), big.NewInt(7), big.NewInt(9)}

		validators := make(map[string]ValidatorsChainData)

		validators[chainID] = ValidatorsChainData{
			Keys: []eth.ValidatorChainData{
				{
					Key: key1,
				},
				{
					Key: key2,
				},
				{
					Key: key3,
				},
			},
		}

		observer.validators = validators

		removedValidators := []ethcommon.Address{
			ethcommon.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"),
			ethcommon.HexToAddress("0x1234567890abcdef1234567890abcdef12345679"),
		}

		err := observer.removeValidators(validators, removedValidators)
		require.NoError(t, err)

		validatorKeys := observer.GetValidatorSet(chainID)

		assert.Equal(t, 1, len(validatorKeys), "Expected validator to be removed")
		assert.Equal(t, key2, validatorKeys[0].Key, "Expected second key")
	})
}

func TestAddAndRemoveValidators(t *testing.T) {
	logger := hclog.NewNullLogger()
	bridgeSmartContract := &eth.BridgeSmartContractMock{}
	bridgeSmartContract.On("GetAddressValidatorIndex",
		ethcommon.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")).Return(uint8(2), nil)

	observer := &ValidatorSetObserverImpl{
		validatorSetPending: false,
		validators:          ValidatorsPerChain{},
		bridgeSmartContract: bridgeSmartContract,
		logger:              logger,
	}

	t.Run("Add and remove validators to existing chain", func(t *testing.T) {
		key1 := [4]*big.Int{big.NewInt(1), big.NewInt(1), big.NewInt(1), big.NewInt(1)}
		key2 := [4]*big.Int{big.NewInt(2), big.NewInt(2), big.NewInt(2), big.NewInt(2)}
		key3 := [4]*big.Int{big.NewInt(3), big.NewInt(3), big.NewInt(3), big.NewInt(3)}

		validators := make(map[string]ValidatorsChainData)

		validators[chainID] = ValidatorsChainData{
			Keys: []eth.ValidatorChainData{
				{
					Key: key1,
				},
			},
		}

		observer.validators = validators

		addedValidators := []eth.ValidatorSet{
			{
				ChainId: common.ToNumChainID(chainID),
				Validators: []contractbinding.IBridgeStructsValidatorAddressChainData{
					{
						Data: contractbinding.IBridgeStructsValidatorChainData{
							Key: key2,
						},
					},
					{
						Data: contractbinding.IBridgeStructsValidatorChainData{
							Key: key3,
						},
					},
				},
			},
		}

		observer.addValidators(validators, addedValidators)

		validatorKeys := observer.GetValidatorSet(chainID)

		assert.Equal(t, 3, len(validatorKeys), "Expected three validator keys")
		assert.Equal(t, key1, validatorKeys[0].Key, "Expected first key unchanged")
		assert.Equal(t, key2, validatorKeys[1].Key, "Expected second key added")
		assert.Equal(t, key3, validatorKeys[2].Key, "Expected third key added")

		removedValidators := []ethcommon.Address{
			ethcommon.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"),
		}

		err := observer.removeValidators(validators, removedValidators)
		require.NoError(t, err)

		validatorKeys = observer.GetValidatorSet(chainID)

		assert.Equal(t, 2, len(validatorKeys), "Expected two validator keys")
		assert.Equal(t, key1, validatorKeys[0].Key, "Expected first key unchanged")
		assert.Equal(t, key3, validatorKeys[1].Key, "Expected third key added")
	})
}

func TestExecute(t *testing.T) {
	logger := hclog.NewNullLogger()
	bridgeSmartContract := &eth.BridgeSmartContractMock{}
	validatorSetStream := make(chan *ValidatorsPerChain, 1)
	observer := &ValidatorSetObserverImpl{
		validatorSetPending: false,
		validators:          ValidatorsPerChain{},
		bridgeSmartContract: bridgeSmartContract,
		logger:              logger,
		validatorSetStream:  validatorSetStream,
		lock:                sync.RWMutex{},
	}
	chainID := common.ChainIDStrPrime

	t.Run("Successful execution with pending validator set", func(t *testing.T) {
		observer.validators = ValidatorsPerChain{
			chainID: ValidatorsChainData{Keys: []eth.ValidatorChainData{{Key: [4]*big.Int{big.NewInt(1)}}}}}
		observer.validatorSetPending = false
		addedValidators := []eth.ValidatorSet{
			{
				ChainId: common.ToNumChainID(chainID),
				Validators: []contractbinding.IBridgeStructsValidatorAddressChainData{
					{Data: contractbinding.IBridgeStructsValidatorChainData{Key: [4]*big.Int{big.NewInt(2)}}},
				},
			},
		}
		removedValidators := []ethcommon.Address{}

		bridgeSmartContract.On("IsNewValidatorSetPending").Return(true, nil)
		bridgeSmartContract.On("GetPendingValidatorSetDelta").Return(addedValidators, removedValidators, nil)
		bridgeSmartContract.On("GetLastObservedBlock", mock.Anything, mock.Anything).Return(eth.CardanoBlock{
			BlockSlot: big.NewInt(1),
		}, nil)
		bridgeSmartContract.On("GetAllRegisteredChains", mock.Anything).Return([]eth.Chain{{Id: common.ToNumChainID(chainID)}}, nil)

		err := observer.execute()
		assert.NoError(t, err)
		assert.True(t, observer.validatorSetPending)
		assert.Equal(t, 2, len(observer.validators[chainID].Keys))

		select {
		case pendingSet := <-validatorSetStream:
			assert.NotNil(t, pendingSet)
			assert.Equal(t, 2, len((*pendingSet)[chainID].Keys))
		default:
			t.Fatal("Expected pending validator set in stream")
		}
	})
}

func TestInitValidatorSet(t *testing.T) {
	logger := hclog.NewNullLogger()
	bridgeSmartContract := &eth.BridgeSmartContractMock{}
	validatorSetStream := make(chan *ValidatorsPerChain, 1)
	ctx := context.Background()
	chainID := "prime"
	observer := &ValidatorSetObserverImpl{
		validatorSetPending: false,
		validators:          ValidatorsPerChain{},
		bridgeSmartContract: bridgeSmartContract,
		logger:              logger,
		validatorSetStream:  validatorSetStream,
		lock:                sync.RWMutex{},
		context:             ctx,
	}

	t.Run("Successful initialization with single chain and validators", func(t *testing.T) {
		registeredChains := []eth.Chain{{Id: common.ToNumChainID(chainID)}}
		validatorsData := []contractbinding.IBridgeStructsValidatorChainData{
			{Key: [4]*big.Int{big.NewInt(1), big.NewInt(2), big.NewInt(3), big.NewInt(4)}},
		}
		lastBlock := eth.CardanoBlock{BlockSlot: big.NewInt(100)}

		bridgeSmartContract.On("GetAllRegisteredChains", ctx).Return(registeredChains, nil)
		bridgeSmartContract.On("GetValidatorsChainData", ctx, chainID).Return(validatorsData, nil)
		bridgeSmartContract.On("GetLastObservedBlock", ctx, chainID).Return(lastBlock, nil).Once()

		err := observer.initValidatorSet()
		assert.NoError(t, err)

		observer.lock.RLock()
		defer observer.lock.RUnlock()

		assert.Equal(t, 1, len(observer.validators))
		assert.Equal(t, 1, len(observer.validators[chainID].Keys))
		assert.Equal(t, validatorsData[0].Key, observer.validators[chainID].Keys[0].Key)
		assert.Equal(t, uint64(100), observer.validators[chainID].SlotNumber)
	})

	t.Run("Successful initialization with multiple chains", func(t *testing.T) {
		bridgeSmartContract := &eth.BridgeSmartContractMock{}
		newObserver := &ValidatorSetObserverImpl{
			validatorSetPending: false,
			validators:          ValidatorsPerChain{},
			bridgeSmartContract: bridgeSmartContract,
			logger:              logger,
			validatorSetStream:  validatorSetStream,
			lock:                sync.RWMutex{},
			context:             ctx,
		}

		chainID2 := common.ChainIDStrVector
		registeredChains := []eth.Chain{
			{Id: common.ToNumChainID(chainID)},
			{Id: common.ToNumChainID(chainID2)},
		}
		validatorsData1 := []contractbinding.IBridgeStructsValidatorChainData{
			{Key: [4]*big.Int{big.NewInt(1), big.NewInt(2), big.NewInt(3), big.NewInt(4)}},
		}
		validatorsData2 := []contractbinding.IBridgeStructsValidatorChainData{
			{Key: [4]*big.Int{big.NewInt(5), big.NewInt(6), big.NewInt(7), big.NewInt(8)}},
		}
		lastBlock1 := eth.CardanoBlock{BlockSlot: big.NewInt(100)}
		lastBlock2 := eth.CardanoBlock{BlockSlot: big.NewInt(200)}

		bridgeSmartContract.On("GetAllRegisteredChains", ctx).Return(registeredChains, nil)
		bridgeSmartContract.On("GetValidatorsChainData", ctx, chainID).Return(validatorsData1, nil).Once()
		bridgeSmartContract.On("GetValidatorsChainData", ctx, chainID2).Return(validatorsData2, nil)
		bridgeSmartContract.On("GetLastObservedBlock", ctx, chainID).Return(lastBlock1, nil).Once()
		bridgeSmartContract.On("GetLastObservedBlock", ctx, chainID2).Return(lastBlock2, nil).Once()

		err := newObserver.initValidatorSet()
		assert.NoError(t, err)

		newObserver.lock.RLock()
		defer newObserver.lock.RUnlock()

		assert.Equal(t, 2, len(newObserver.validators))
		assert.Equal(t, 1, len(newObserver.validators[chainID].Keys))
		assert.Equal(t, validatorsData1[0].Key, newObserver.validators[chainID].Keys[0].Key)
		assert.Equal(t, uint64(100), newObserver.validators[chainID].SlotNumber)
		assert.Equal(t, 1, len(newObserver.validators[chainID2].Keys))
		assert.Equal(t, validatorsData2[0].Key, newObserver.validators[chainID2].Keys[0].Key)
		assert.Equal(t, uint64(200), newObserver.validators[chainID2].SlotNumber)
	})

	t.Run("Successful initialization with no validators for a chain", func(t *testing.T) {
		bridgeSmartContract := &eth.BridgeSmartContractMock{}
		newObserver := &ValidatorSetObserverImpl{
			validatorSetPending: false,
			validators:          ValidatorsPerChain{},
			bridgeSmartContract: bridgeSmartContract,
			logger:              logger,
			validatorSetStream:  validatorSetStream,
			lock:                sync.RWMutex{},
			context:             ctx,
		}

		registeredChains := []eth.Chain{{Id: common.ToNumChainID(chainID)}}
		validatorsData := []contractbinding.IBridgeStructsValidatorChainData{}
		lastBlock := eth.CardanoBlock{BlockSlot: big.NewInt(100)}

		bridgeSmartContract.On("GetAllRegisteredChains", ctx).Return(registeredChains, nil)
		bridgeSmartContract.On("GetValidatorsChainData", ctx, chainID).Return(validatorsData, nil)
		bridgeSmartContract.On("GetLastObservedBlock", ctx, chainID).Return(lastBlock, nil)

		err := newObserver.initValidatorSet()
		assert.NoError(t, err)

		newObserver.lock.RLock()
		defer newObserver.lock.RUnlock()

		assert.Equal(t, 1, len(newObserver.validators))
		assert.Equal(t, 0, len(newObserver.validators[chainID].Keys))
		assert.Equal(t, uint64(100), newObserver.validators[chainID].SlotNumber)
	})
}

func TestUpdateSlotNumbers(t *testing.T) {
	logger := hclog.NewNullLogger()
	bridgeSmartContract := &eth.BridgeSmartContractMock{}
	validatorSetStream := make(chan *ValidatorsPerChain, 1)
	observer := &ValidatorSetObserverImpl{
		validatorSetPending: false,
		validators:          ValidatorsPerChain{},
		bridgeSmartContract: bridgeSmartContract,
		logger:              logger,
		validatorSetStream:  validatorSetStream,
		lock:                sync.RWMutex{},
	}
	chainID := "prime"

	t.Run("Missing last observed block slot for chain", func(t *testing.T) {
		validators := ValidatorsPerChain{
			chainID: ValidatorsChainData{
				Keys:       []eth.ValidatorChainData{{Key: [4]*big.Int{big.NewInt(1)}}},
				SlotNumber: 100,
			},
		}
		lastObservedBlockSlots := map[string]uint64{}

		err := observer.updateSlotNumbers(validators, lastObservedBlockSlots)
		assert.Error(t, err)
		assert.Equal(t, fmt.Errorf("last observed block slot not found for chain: %s", chainID), err)
	})

	t.Run("No update needed when slot numbers match", func(t *testing.T) {
		validators := ValidatorsPerChain{
			chainID: ValidatorsChainData{
				Keys:       []eth.ValidatorChainData{{Key: [4]*big.Int{big.NewInt(1)}}},
				SlotNumber: 100,
			},
		}
		lastObservedBlockSlots := map[string]uint64{chainID: 100}

		err := observer.updateSlotNumbers(validators, lastObservedBlockSlots)
		assert.NoError(t, err)
		assert.Equal(t, uint64(100), validators[chainID].SlotNumber)
	})

	t.Run("Successful update of slot number for single chain", func(t *testing.T) {
		validators := ValidatorsPerChain{
			chainID: ValidatorsChainData{
				Keys:       []eth.ValidatorChainData{{Key: [4]*big.Int{big.NewInt(1)}}},
				SlotNumber: 100,
			},
		}
		lastObservedBlockSlots := map[string]uint64{chainID: 200}

		err := observer.updateSlotNumbers(validators, lastObservedBlockSlots)
		assert.NoError(t, err)
		assert.Equal(t, uint64(200), validators[chainID].SlotNumber)
	})

	t.Run("Successful update of slot numbers for multiple chains", func(t *testing.T) {
		chainID2 := "vector"
		validators := ValidatorsPerChain{
			chainID: ValidatorsChainData{
				Keys:       []eth.ValidatorChainData{{Key: [4]*big.Int{big.NewInt(1)}}},
				SlotNumber: 100,
			},
			chainID2: ValidatorsChainData{
				Keys:       []eth.ValidatorChainData{{Key: [4]*big.Int{big.NewInt(2)}}},
				SlotNumber: 300,
			},
		}
		lastObservedBlockSlots := map[string]uint64{
			chainID:  200,
			chainID2: 400,
		}

		err := observer.updateSlotNumbers(validators, lastObservedBlockSlots)
		assert.NoError(t, err)
		assert.Equal(t, uint64(200), validators[chainID].SlotNumber)
		assert.Equal(t, uint64(400), validators[chainID2].SlotNumber)
	})

	t.Run("Partial update with some matching slot numbers", func(t *testing.T) {
		chainID2 := "vector"
		validators := ValidatorsPerChain{
			chainID: ValidatorsChainData{
				Keys:       []eth.ValidatorChainData{{Key: [4]*big.Int{big.NewInt(1)}}},
				SlotNumber: 100,
			},
			chainID2: ValidatorsChainData{
				Keys:       []eth.ValidatorChainData{{Key: [4]*big.Int{big.NewInt(2)}}},
				SlotNumber: 300,
			},
		}
		lastObservedBlockSlots := map[string]uint64{
			chainID:  100, // Same as current
			chainID2: 400, // Different
		}

		err := observer.updateSlotNumbers(validators, lastObservedBlockSlots)
		assert.NoError(t, err)
		assert.Equal(t, uint64(100), validators[chainID].SlotNumber)
		assert.Equal(t, uint64(400), validators[chainID2].SlotNumber)
	})

	t.Run("Empty validators map", func(t *testing.T) {
		validators := ValidatorsPerChain{}
		lastObservedBlockSlots := map[string]uint64{chainID: 200}

		err := observer.updateSlotNumbers(validators, lastObservedBlockSlots)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(validators))
	})
}
