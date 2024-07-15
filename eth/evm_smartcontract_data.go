package eth

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

var (
	txAbi, _ = abi.NewType("tuple", "EVMSmartContractTransaction", []abi.ArgumentMarshaling{
		{
			Name: "batchNonceID",
			Type: "uint64",
		},
		{
			Name:         "receivers",
			Type:         "tuple[]",
			InternalType: "EVMSmartContractTransactionReceiver[]",
			Components: []abi.ArgumentMarshaling{
				{
					Name: "sourceID",
					Type: "uint8",
				},
				{
					Name: "address",
					Type: "address",
				},
				{
					Name: "amount",
					Type: "uint256",
				},
			},
		},
	})
)

type EVMSmartContractTransactionReceiver struct {
	SourceID uint8
	Address  common.Address
	Amount   *big.Int
}

type EVMSmartContractTransaction struct {
	BatchNonceID uint64
	Receivers    []EVMSmartContractTransactionReceiver
}

func NewEVMSmartContractTransaction(bytes []byte) (*EVMSmartContractTransaction, error) {
	dt, err := abi.Arguments{{Type: txAbi}}.Unpack(bytes)
	if err != nil {
		return nil, err
	}

	tx, _ := abi.ConvertType(dt[0], new(EVMSmartContractTransaction)).(*EVMSmartContractTransaction)

	return tx, nil
}

func (evmsctx *EVMSmartContractTransaction) Pack() ([]byte, error) {
	return abi.Arguments{{Type: txAbi}}.Pack(evmsctx)
}
