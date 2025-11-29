package eth

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

var (
	skylineTxAbi, _ = abi.NewType("tuple", "", []abi.ArgumentMarshaling{
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
				{
					Name: "tokenId",
					Type: "uint16",
				},
			},
		},
	})
)

type SkylineEVMSmartContractTransactionReceiver struct {
	Address common.Address `json:"addr" abi:"receiver"`
	Amount  *big.Int       `json:"amount" abi:"amount"`
	TokenID uint16         `json:"tokenId" abi:"tokenId"`
}

type SkylineEVMSmartContractTransaction struct {
	BatchNonceID uint64                                       `json:"batchNonceId" abi:"batchId"`
	TTL          uint64                                       `json:"ttl" abi:"ttlExpired"`
	FeeAmount    *big.Int                                     `json:"feeAmount" abi:"feeAmount"`
	Receivers    []SkylineEVMSmartContractTransactionReceiver `json:"receivers" abi:"receivers"`
}

func NewSkylineEVMSmartContractTransaction(bytes []byte) (*SkylineEVMSmartContractTransaction, error) {
	dt, err := abi.Arguments{{Type: skylineTxAbi}}.Unpack(bytes)
	if err != nil {
		return nil, err
	}

	tx, _ := abi.ConvertType(dt[0], new(SkylineEVMSmartContractTransaction)).(*SkylineEVMSmartContractTransaction)

	return tx, nil
}

func (evmsctx *SkylineEVMSmartContractTransaction) Pack() ([]byte, error) {
	return abi.Arguments{{Type: skylineTxAbi}}.Pack(evmsctx)
}

func (evmsctx SkylineEVMSmartContractTransaction) String() string {
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
		sb.WriteRune(',')
		sb.WriteString(fmt.Sprintf("%v", v.TokenID))
		sb.WriteRune(')')
	}

	return sb.String()
}
