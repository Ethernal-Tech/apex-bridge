package eth

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

var (
	txAbi, _ = abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{
			Name: "batchId",
			Type: "uint64",
		},
		{
			Name: "ttlExpired",
			Type: "uint64",
		},
		{
			Name:         "_receivers",
			Type:         "tuple[]",
			InternalType: "struct IGatewayStructs.ReceiverDeposit[]",
			Components: []abi.ArgumentMarshaling{
				{
					Name: "sourceChainId",
					Type: "uint8",
				},
				{
					Name: "receiver",
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
	SourceID uint8          `json:"sourceId" abi:"sourceChainId"`
	Address  common.Address `json:"addr" abi:"receiver"`
	Amount   *big.Int       `json:"amount" abi:"amount"`
}

type EVMSmartContractTransaction struct {
	BatchNonceID uint64                                `json:"batchNonceId" abi:"batchId"`
	TTL          uint64                                `json:"ttl" abi:"ttlExpired"`
	Receivers    []EVMSmartContractTransactionReceiver `json:"receivers" abi:"_receivers"`
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
