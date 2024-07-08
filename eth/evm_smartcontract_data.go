package eth

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

var (
	txAbi, _ = abi.NewType("tuple", "EVMSmartContractTransaction", []abi.ArgumentMarshaling{
		{
			Name: "chainID",
			Type: "uint8",
		},
		{
			Name:         "receivers",
			Type:         "tuple[]",
			InternalType: "EVMSmartContractTransactionReceiver[]",
			Components: []abi.ArgumentMarshaling{
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
	Address common.Address
	Amount  *big.Int
}

type EVMSmartContractTransaction struct {
	ChainID   uint8
	Receivers []EVMSmartContractTransactionReceiver
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
