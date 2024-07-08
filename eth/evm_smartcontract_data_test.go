package eth

import (
	"math/big"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/stretchr/testify/require"
)

func TestEVMSmartContractTransaction(t *testing.T) {
	obj := &EVMSmartContractTransaction{
		ChainID: 100,
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
