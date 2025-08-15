package eth

import (
	"math/big"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/stretchr/testify/require"
)

func TestEVMSmartContractTransaction(t *testing.T) {
	obj := &EVMSmartContractTransaction{
		BatchNonceID: 100,
		TTL:          uint64(8398923),
		FeeAmount:    big.NewInt(1),
		Receivers: []EVMSmartContractTransactionReceiver{
			{
				Address: common.HexToAddress("0xFF00FF"),
				Amount:  new(big.Int).SetUint64(100),
			},
			{
				Address: common.HexToAddress("0xFF0011"),
				Amount:  new(big.Int).SetUint64(3),
			},
			{
				Address: common.HexToAddress("0xFF0022"),
				Amount:  new(big.Int).SetUint64(531),
			},
		},
	}

	bytes, err := obj.Pack()
	require.NoError(t, err)

	newObj, err := NewEVMSmartContractTransaction(bytes)
	require.NoError(t, err)
	require.Equal(t, obj, newObj)
}

func TestEVMValidatorSetChangeTransaction(t *testing.T) {
	obj := &EVMValidatorSetChangeTx{
		ValidatorsSetNumber: big.NewInt(100),
		TTL:                 big.NewInt(1_000_000),
		ValidatorsChainData: []ValidatorChainData{
			{
				Key: [4]*big.Int{
					big.NewInt(1),
					big.NewInt(2),
					big.NewInt(3),
					big.NewInt(4),
				},
			},
			{
				Key: [4]*big.Int{
					big.NewInt(5),
					big.NewInt(6),
					big.NewInt(7),
					big.NewInt(8),
				},
			},
		},
	}

	bytes, err := obj.Pack()
	require.NoError(t, err)

	newObj, err := NewEVMValidatorSetChangeTransaction(bytes)
	require.NoError(t, err)
	require.Equal(t, obj, newObj)
}
