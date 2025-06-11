package eth

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
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

func TestNewConfirmedBatch(t *testing.T) {
	signature := "746573740000000000000000000000000000000000000000000000000000000a"
	feeSignature := "746573740000000000000000000000000000000000000000000000000000000f"
	signs, _ := hex.DecodeString("000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000000020746573740000000000000000000000000000000000000000000000000000000a0000000000000000000000000000000000000000000000000000000000000020746573740000000000000000000000000000000000000000000000000000000f")
	receivedBatch := contractbinding.IBridgeStructsConfirmedBatch{
		Signatures:      [][]byte{signs},
		Bitmap:          big.NewInt(1),
		RawTransaction:  []byte{1},
		IsConsolidation: true,
		Id:              100,
	}

	res, err := NewConfirmedBatch(receivedBatch)

	require.NoError(t, err)
	require.Equal(t, receivedBatch.Id, res.ID)
	require.Equal(t, receivedBatch.RawTransaction, res.RawTransaction)
	require.Equal(t, receivedBatch.Bitmap, res.Bitmap)
	require.Equal(t, receivedBatch.IsConsolidation, res.IsConsolidation)
	require.Len(t, res.Signatures, 1)
	require.Len(t, res.FeeSignatures, 1)
	require.Equal(t, signature, hex.EncodeToString(res.Signatures[0]))
	require.Equal(t, feeSignature, hex.EncodeToString(res.FeeSignatures[0]))
}
