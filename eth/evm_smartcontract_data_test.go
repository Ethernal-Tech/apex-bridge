package eth

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/stretchr/testify/require"
)

func TestEVMSmartContractTransaction(t *testing.T) {
	obj := &EVMSmartContractTransaction{
		BatchNonceID: 100,
		TTL:          uint64(8398923),
		FeeAmount:    big.NewInt(1),
		ChainID:      big.NewInt(9070),
		Gateway:      common.HexToAddress("0x2222222222222222222222222222222222222222"),
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

func goldenTx() *EVMSmartContractTransaction {
	return &EVMSmartContractTransaction{
		BatchNonceID: 7,
		TTL:          1234,
		FeeAmount:    big.NewInt(99),
		Receivers: []EVMSmartContractTransactionReceiver{
			{
				Address: common.HexToAddress("0x1111111111111111111111111111111111111111"),
				Amount:  big.NewInt(5),
				TokenID: 5,
			},
		},
		ChainID: big.NewInt(9070),
		Gateway: common.HexToAddress("0x2222222222222222222222222222222222222222"),
	}
}

// AD-1196: the batch carries its destination (chainId, gateway); this vector is pinned in
// apex-evm-gateway test/DestinationCheck.test.ts too, so Go and Solidity encode it identically.
func TestEVMSmartContractTransaction_GoldenVector(t *testing.T) {
	bytes, err := goldenTx().Pack()
	require.NoError(t, err)
	require.Len(t, bytes, 352)

	hash, err := common.Keccak256(bytes)
	require.NoError(t, err)
	require.Equal(t, "9af154f750d941e3cde48d7820ab31e1c5a5065806c405005d2c79368d9a2022", hex.EncodeToString(hash))
}

func TestEVMSmartContractTransaction_RoundTrip(t *testing.T) {
	obj := goldenTx()

	bytes, err := obj.Pack()
	require.NoError(t, err)

	newObj, err := NewEVMSmartContractTransaction(bytes)
	require.NoError(t, err)
	require.Equal(t, obj, newObj)
	require.Equal(t, big.NewInt(9070), newObj.ChainID)
	require.Equal(t, obj.Gateway, newObj.Gateway)
}

func TestEVMSmartContractTransaction_LegacyDecode(t *testing.T) {
	obj := goldenTx()

	// legacy payload: the same fields without chainId and gateway
	bytes, err := abi.Arguments{{Type: txAbiLegacy}}.Pack(evmSmartContractTransactionLegacy{
		BatchNonceID: obj.BatchNonceID,
		TTL:          obj.TTL,
		FeeAmount:    obj.FeeAmount,
		Receivers:    obj.Receivers,
	})
	require.NoError(t, err)

	newObj, err := NewEVMSmartContractTransaction(bytes)
	require.NoError(t, err)
	require.Equal(t, obj.BatchNonceID, newObj.BatchNonceID)
	require.Equal(t, obj.TTL, newObj.TTL)
	require.Equal(t, obj.FeeAmount, newObj.FeeAmount)
	require.Equal(t, obj.Receivers, newObj.Receivers)
	require.Nil(t, newObj.ChainID)
}

func TestEVMSmartContractTransaction_Garbage(t *testing.T) {
	_, err := NewEVMSmartContractTransaction([]byte{1, 2, 3, 4, 5, 6, 7})
	require.Error(t, err)
}
