package eth

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

var (
	txAbi, _ = abi.NewType("tuple", "EVMSmartContractTransaction", []abi.ArgumentMarshaling{
		{
			Name: "BatchNonceID",
			Type: "uint64",
		},
		{
			Name: "TTL",
			Type: "uint64",
		},
		{
			Name:         "Receivers",
			Type:         "tuple[]",
			InternalType: "EVMSmartContractTransactionReceiver[]",
			Components: []abi.ArgumentMarshaling{
				{
					Name: "SourceID",
					Type: "uint8",
				},
				{
					Name: "Address",
					Type: "address",
				},
				{
					Name: "Amount",
					Type: "uint256",
				},
			},
		},
	})
)

type EVMSmartContractTransactionReceiver struct {
	SourceID uint8          `json:"sourceId"`
	Address  common.Address `json:"addr"`
	Amount   *big.Int       `json:"amount"`
}

type EVMSmartContractTransaction struct {
	BatchNonceID uint64                                `json:"batchNonceId"`
	TTL          uint64                                `json:"ttl"`
	Receivers    []EVMSmartContractTransactionReceiver `json:"receivers"`
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
