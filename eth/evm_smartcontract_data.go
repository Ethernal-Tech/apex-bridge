package eth

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

const (
	abiTypeUint256 = "uint256"
	abiTypeUint64  = "uint64"
)

var (
	receiversAbi = abi.ArgumentMarshaling{
		Name:         "receivers",
		Type:         "tuple[]",
		InternalType: "struct IGatewayStructs.ReceiverDeposit[]",
		Components: []abi.ArgumentMarshaling{
			{Name: "receiver", Type: "address"},
			{Name: "amount", Type: abiTypeUint256},
			{Name: "tokenId", Type: "uint16"},
		},
	}

	// txAbiLegacy is the payload layout used before the destination binding (chainId, gateway)
	txAbiLegacy, _ = abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "batchId", Type: abiTypeUint64},
		{Name: "ttlExpired", Type: abiTypeUint64},
		{Name: "feeAmount", Type: abiTypeUint256},
		receiversAbi,
	})

	txAbi, _ = abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "batchId", Type: abiTypeUint64},
		{Name: "ttlExpired", Type: abiTypeUint64},
		{Name: "feeAmount", Type: abiTypeUint256},
		receiversAbi,
		{Name: "chainId", Type: abiTypeUint256},
		{Name: "gateway", Type: "address"},
	})
)

type EVMSmartContractTransactionReceiver struct {
	Address common.Address `json:"addr" abi:"receiver"`
	Amount  *big.Int       `json:"amount" abi:"amount"`
	TokenID uint16         `json:"tokenId" abi:"tokenId"`
}

type EVMSmartContractTransaction struct {
	BatchNonceID uint64                                `json:"batchNonceId" abi:"batchId"`
	TTL          uint64                                `json:"ttl" abi:"ttlExpired"`
	FeeAmount    *big.Int                              `json:"feeAmount" abi:"feeAmount"`
	Receivers    []EVMSmartContractTransactionReceiver `json:"receivers" abi:"receivers"`
	ChainID      *big.Int                              `json:"chainId" abi:"chainId"`
	Gateway      common.Address                        `json:"gateway" abi:"gateway"`
}

// evmSmartContractTransactionLegacy is the payload without destination fields (old batches and events)
type evmSmartContractTransactionLegacy struct {
	BatchNonceID uint64                                `abi:"batchId"`
	TTL          uint64                                `abi:"ttlExpired"`
	FeeAmount    *big.Int                              `abi:"feeAmount"`
	Receivers    []EVMSmartContractTransactionReceiver `abi:"receivers"`
}

// NewEVMSmartContractTransaction decodes the new payload and falls back to the legacy one
// (without chainId and gateway), leaving ChainID nil in that case.
func NewEVMSmartContractTransaction(bytes []byte) (*EVMSmartContractTransaction, error) {
	if dt, err := (abi.Arguments{{Type: txAbi}}).Unpack(bytes); err == nil {
		tx, _ := abi.ConvertType(dt[0], new(EVMSmartContractTransaction)).(*EVMSmartContractTransaction)

		// legacy bytes may decode as the new layout with garbage destination fields, so only
		// accept the result when it re-encodes to exactly the same bytes
		if repacked, err := tx.Pack(); err == nil && string(repacked) == string(bytes) {
			return tx, nil
		}
	}

	dt, err := abi.Arguments{{Type: txAbiLegacy}}.Unpack(bytes)
	if err != nil {
		return nil, err
	}

	legacy, _ := abi.ConvertType(dt[0], new(evmSmartContractTransactionLegacy)).(*evmSmartContractTransactionLegacy)

	return &EVMSmartContractTransaction{
		BatchNonceID: legacy.BatchNonceID,
		TTL:          legacy.TTL,
		FeeAmount:    legacy.FeeAmount,
		Receivers:    legacy.Receivers,
	}, nil
}

func (evmsctx *EVMSmartContractTransaction) Pack() ([]byte, error) {
	if evmsctx.ChainID == nil {
		return nil, errors.New("evm smart contract transaction: chainId is not set")
	}

	return abi.Arguments{{Type: txAbi}}.Pack(evmsctx)
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
		sb.WriteRune(',')
		sb.WriteString(fmt.Sprintf("%v", v.TokenID))
		sb.WriteRune(')')
	}

	sb.WriteString("\nchainId = ")
	sb.WriteString(fmt.Sprintf("%v\n", evmsctx.ChainID))
	sb.WriteString("gateway = ")
	sb.WriteString(evmsctx.Gateway.String())

	return sb.String()
}
