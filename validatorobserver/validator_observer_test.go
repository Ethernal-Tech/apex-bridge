package validatorobserver

import (
	"math/big"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	"github.com/Ethernal-Tech/apex-bridge/eth"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
)

const chainID = "prime"

func TestAddValidators(t *testing.T) {
	logger := hclog.NewNullLogger()
	bridgeSmartContract := &eth.BridgeSmartContractMock{}
	observer := &ValidatorSetObserver{
		validatorSetPending: false,
		validators:          Validators{},
		bridgeSmartContract: bridgeSmartContract,
		logger:              logger,
	}

	t.Run("Add validators to existing chain", func(t *testing.T) {
		key1 := [4]*big.Int{big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0)}
		key2 := [4]*big.Int{big.NewInt(5), big.NewInt(6), big.NewInt(7), big.NewInt(8)}

		validators := Validators{
			Data: make(map[string]ValidatorsChainData),
		}

		validators.Data[chainID] = ValidatorsChainData{
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

		validatorKeys := observer.GetVerificationKeys(chainID)

		assert.Equal(t, 2, len(validatorKeys), "Expected two validator keys")
		assert.Equal(t, key1, validatorKeys[0].Key, "Expected first key unchanged")
		assert.Equal(t, key2, validatorKeys[1].Key, "Expected second key added")
	})
}

func TestRemoveValidators(t *testing.T) {
	logger := hclog.NewNullLogger()
	bridgeSmartContract := &eth.BridgeSmartContractMock{}
	bridgeSmartContract.On("GetAddressValidatorIndex",
		ethcommon.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")).Return(uint8(0), nil)

	observer := &ValidatorSetObserver{
		validatorSetPending: false,
		validators:          Validators{},
		bridgeSmartContract: bridgeSmartContract,
		logger:              logger,
	}

	t.Run("Remove existing validator", func(t *testing.T) {
		key1 := [4]*big.Int{big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0)}
		key2 := [4]*big.Int{big.NewInt(5), big.NewInt(6), big.NewInt(7), big.NewInt(8)}

		validators := Validators{
			Data: make(map[string]ValidatorsChainData),
		}
		validators.Data[chainID] = ValidatorsChainData{
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

		observer.removeValidators(validators, removedValidators)

		validatorKeys := observer.GetVerificationKeys(chainID)

		assert.Equal(t, 1, len(validatorKeys), "Expected validator to be removed")
		assert.Equal(t, key2, validatorKeys[0].Key, "Expected second key")
	})

	t.Run("Remove only validator", func(t *testing.T) {
		key1 := [4]*big.Int{big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0)}

		validators := Validators{
			Data: make(map[string]ValidatorsChainData),
		}
		validators.Data[chainID] = ValidatorsChainData{
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

		observer.removeValidators(validators, removedValidators)

		validatorKeys := observer.GetVerificationKeys(chainID)

		assert.Equal(t, 0, len(validatorKeys), "Expected validator to be removed")
	})
}

func TestAddAndRemoveValidators(t *testing.T) {
	logger := hclog.NewNullLogger()
	bridgeSmartContract := &eth.BridgeSmartContractMock{}
	bridgeSmartContract.On("GetAddressValidatorIndex",
		ethcommon.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")).Return(uint8(1), nil)

	observer := &ValidatorSetObserver{
		validatorSetPending: false,
		validators:          Validators{},
		bridgeSmartContract: bridgeSmartContract,
		logger:              logger,
	}

	t.Run("Add and remove validators to existing chain", func(t *testing.T) {
		key1 := [4]*big.Int{big.NewInt(1), big.NewInt(1), big.NewInt(1), big.NewInt(1)}
		key2 := [4]*big.Int{big.NewInt(2), big.NewInt(2), big.NewInt(2), big.NewInt(2)}
		key3 := [4]*big.Int{big.NewInt(3), big.NewInt(3), big.NewInt(3), big.NewInt(3)}

		validators := Validators{
			Data: make(map[string]ValidatorsChainData),
		}

		validators.Data[chainID] = ValidatorsChainData{
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

		validatorKeys := observer.GetVerificationKeys(chainID)

		assert.Equal(t, 3, len(validatorKeys), "Expected three validator keys")
		assert.Equal(t, key1, validatorKeys[0].Key, "Expected first key unchanged")
		assert.Equal(t, key2, validatorKeys[1].Key, "Expected second key added")
		assert.Equal(t, key3, validatorKeys[2].Key, "Expected third key added")

		removedValidators := []ethcommon.Address{
			ethcommon.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"),
		}

		observer.removeValidators(validators, removedValidators)

		validatorKeys = observer.GetVerificationKeys(chainID)

		assert.Equal(t, 2, len(validatorKeys), "Expected two validator keys")
		assert.Equal(t, key1, validatorKeys[0].Key, "Expected first key unchanged")
		assert.Equal(t, key3, validatorKeys[1].Key, "Expected third key added")
	})
}
