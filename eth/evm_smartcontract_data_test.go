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
				TokenID: 1,
			},
			{
				Address: common.HexToAddress("0xFF0011"),
				Amount:  new(big.Int).SetUint64(3),
				TokenID: 2,
			},
			{
				Address: common.HexToAddress("0xFF0022"),
				Amount:  new(big.Int).SetUint64(531),
				TokenID: 3,
			},
		},
	}

	bytes, err := obj.Pack()
	require.NoError(t, err)

	newObj, err := NewEVMSmartContractTransaction(bytes)
	require.NoError(t, err)
	require.Equal(t, obj, newObj)
}
