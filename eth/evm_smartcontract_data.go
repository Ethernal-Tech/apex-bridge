package eth

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

var (
	assetTransferTxAbi, _ = abi.NewType("tuple", "", []abi.ArgumentMarshaling{
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
	validatorSetChangeTxAbi, _ = abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{
			Name: "validatorsSetNumber",
			Type: "uint256",
		},
		{
			Name: "ttl",
			Type: "uint256",
		},
		{
			Name:         "validatorsChainData",
			Type:         "tuple[]",
			InternalType: "struct IGatewayStructs.ValidatorChainData[]",
			Components: []abi.ArgumentMarshaling{
				{
					Name: "key",
					Type: "uint256[4]",
				},
			},
		},
	})
)

type EVMSmartContractTransactionReceiver struct {
	Address common.Address `json:"addr" abi:"receiver"`
	Amount  *big.Int       `json:"amount" abi:"amount"`
}

type EVMSmartContractTransaction struct {
	BatchNonceID uint64                                `json:"batchNonceId" abi:"batchId"`
	TTL          uint64                                `json:"ttl" abi:"ttlExpired"`
	FeeAmount    *big.Int                              `json:"feeAmount" abi:"feeAmount"`
	Receivers    []EVMSmartContractTransactionReceiver `json:"receivers" abi:"receivers"`
}

func NewEVMSmartContractTransaction(bytes []byte) (*EVMSmartContractTransaction, error) {
	dt, err := abi.Arguments{{Type: assetTransferTxAbi}}.Unpack(bytes)
	if err != nil {
		return nil, err
	}

	tx, _ := abi.ConvertType(dt[0], new(EVMSmartContractTransaction)).(*EVMSmartContractTransaction)

	return tx, nil
}

func (evmsctx *EVMSmartContractTransaction) Pack() ([]byte, error) {
	return abi.Arguments{{Type: assetTransferTxAbi}}.Pack(evmsctx)
}

func (evmsctx EVMSmartContractTransaction) String() string {
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

type EVMValidatorChainData struct {
	Key [4]*big.Int `json:"key" abi:"key"`
}

type EVMValidatorSetChangeTx struct {
	ValidatorsSetNumber *big.Int             `json:"validatorsSetNumber" abi:"validatorsSetNumber"`
	TTL                 *big.Int             `json:"ttl" abi:"ttl"`
	ValidatorsChainData []ValidatorChainData `json:"validatorsChainData" abi:"validatorsChainData"`
}

func NewEVMValidatorSetChangeTransaction(bytes []byte) (*EVMValidatorSetChangeTx, error) {
	dt, err := abi.Arguments{{Type: validatorSetChangeTxAbi}}.Unpack(bytes)
	if err != nil {
		return nil, err
	}

	tx, _ := abi.ConvertType(dt[0], new(EVMValidatorSetChangeTx)).(*EVMValidatorSetChangeTx)

	return tx, nil
}

func (t *EVMValidatorSetChangeTx) Pack() ([]byte, error) {
	return abi.Arguments{{Type: validatorSetChangeTxAbi}}.Pack(t)
}
