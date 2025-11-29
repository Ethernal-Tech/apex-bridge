package eth

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

var (
	reactorTxAbi, _ = abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{
			Name: "batchId",
			Type: "uint64",
		},
		{
			Name: "ttlExpired",
			Type: "uint64",
		},
		{
			Name: "feeAmount",
			Type: "uint256",
		},
		{
			Name:         "receivers",
			Type:         "tuple[]",
			InternalType: "struct IGatewayStructs.ReceiverDeposit[]",
			Components: []abi.ArgumentMarshaling{
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

type ReactorEVMSmartContractTransactionReceiver struct {
	Address common.Address `json:"addr" abi:"receiver"`
	Amount  *big.Int       `json:"amount" abi:"amount"`
}

type ReactorEVMSmartContractTransaction struct {
	BatchNonceID uint64                                       `json:"batchNonceId" abi:"batchId"`
	TTL          uint64                                       `json:"ttl" abi:"ttlExpired"`
	FeeAmount    *big.Int                                     `json:"feeAmount" abi:"feeAmount"`
	Receivers    []ReactorEVMSmartContractTransactionReceiver `json:"receivers" abi:"receivers"`
}

func NewReactorEVMSmartContractTransaction(bytes []byte) (*ReactorEVMSmartContractTransaction, error) {
	dt, err := abi.Arguments{{Type: reactorTxAbi}}.Unpack(bytes)
	if err != nil {
		return nil, err
	}

	tx, _ := abi.ConvertType(dt[0], new(ReactorEVMSmartContractTransaction)).(*ReactorEVMSmartContractTransaction)

	return tx, nil
}

func (evmsctx *ReactorEVMSmartContractTransaction) Pack() ([]byte, error) {
	return abi.Arguments{{Type: reactorTxAbi}}.Pack(evmsctx)
}

func (evmsctx ReactorEVMSmartContractTransaction) String() string {
	var sb strings.Builder

	sb.WriteString("id = ")
	sb.WriteString(fmt.Sprintf("%d\n", evmsctx.BatchNonceID))
	sb.WriteString("ttl = ")
	sb.WriteString(fmt.Sprintf("%d\n", evmsctx.TTL))
	sb.WriteString("fee = ")
	sb.WriteString(fmt.Sprintf("%s\n", evmsctx.FeeAmount))
	sb.WriteString("receivers = ")

	for i, v := range evmsctx.Receivers {
		if i > 0 {
			sb.WriteString(", ")
		}

		sb.WriteRune('(')
		sb.WriteString(v.Address.String())
		sb.WriteRune(',')
		sb.WriteString(v.Amount.String())
		sb.WriteRune(')')
	}

	return sb.String()
}
