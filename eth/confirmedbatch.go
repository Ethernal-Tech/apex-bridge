package eth

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
)

type SignedBatch = contractbinding.IBridgeStructsSignedBatch
type ConfirmedTransaction = contractbinding.IBridgeStructsConfirmedTransaction
type ValidatorChainData = contractbinding.IBridgeStructsValidatorChainData
type BridgeReceiver = contractbinding.IBridgeStructsReceiver

type ConfirmedBatch struct {
	ID              uint64
	RawTransaction  []byte
	Signatures      [][]byte
	StakeSignatures [][]byte
	FeeSignatures   [][]byte
	Bitmap          *big.Int
	BatchType       BatchTypes
}

type BatchTypes uint8

const (
	BatchTypeNormal BatchTypes = iota
	BatchTypeConsolidation
	BatchTypeValidatorSet
	BatchTypeValidatorSetFinal
)

func (bt BatchTypes) String() string {
	switch bt {
	case BatchTypeConsolidation:
		return "consolidation"
	case BatchTypeValidatorSet:
		return "validatorSet"
	case BatchTypeValidatorSetFinal:
		return "validatorSetFinal"
	default:
		return "normal"
	}
}

func NewConfirmedBatch(
	contractConfirmedBatch contractbinding.IBridgeStructsConfirmedBatch,
) *ConfirmedBatch {
	return &ConfirmedBatch{
		ID:              contractConfirmedBatch.Id,
		RawTransaction:  contractConfirmedBatch.RawTransaction,
		Signatures:      contractConfirmedBatch.Signatures,
		StakeSignatures: contractConfirmedBatch.StakeSignatures,
		FeeSignatures:   contractConfirmedBatch.FeeSignatures,
		Bitmap:          contractConfirmedBatch.Bitmap,
		BatchType:       BatchTypes(contractConfirmedBatch.BatchType),
	}
}

func (b ConfirmedBatch) String() string {
	var sb strings.Builder

	sb.WriteString("id = ")
	sb.WriteString(fmt.Sprint(b.ID))
	sb.WriteString("\nbatch type = ")
	sb.WriteString(fmt.Sprint(b.BatchType))
	sb.WriteString("\nraw tx = ")
	sb.WriteString(hex.EncodeToString(b.RawTransaction))
	sb.WriteString("\nbitmap = ")
	sb.WriteString(b.Bitmap.String())
	sb.WriteString("\nmultisig signatures = [")

	for i, sig := range b.Signatures {
		if i > 0 {
			sb.WriteString(", ")
		}

		sb.WriteString(hex.EncodeToString(sig))
	}

	sb.WriteString("]")
	sb.WriteString("\nfee payer signatures = [")

	for i, sig := range b.FeeSignatures {
		if i > 0 {
			sb.WriteString(", ")
		}

		sb.WriteString(hex.EncodeToString(sig))
	}

	sb.WriteString("]")
	sb.WriteString("\nstake multisig signatures = [")

	for i, sig := range b.StakeSignatures {
		if i > 0 {
			sb.WriteString(", ")
		}

		sb.WriteString(hex.EncodeToString(sig))
	}

	sb.WriteString("]")

	return sb.String()
}

type ConfirmedTransactionsWrapper struct {
	Txs []ConfirmedTransaction
}

func (ct ConfirmedTransactionsWrapper) String() string {
	var sb strings.Builder

	for i, tx := range ct.Txs {
		if i > 0 {
			sb.WriteString("\n")
		}

		sb.WriteString(fmt.Sprintf("Chain ID = %s, ", common.ToStrChainID(tx.SourceChainId)))
		sb.WriteString(fmt.Sprintf("Tx Hash = %s, ", hex.EncodeToString(tx.ObservedTransactionHash[:])))
		sb.WriteString(fmt.Sprintf("Block = %s, ", tx.BlockHeight))
		sb.WriteString(fmt.Sprintf("Nonce = %d, ", tx.Nonce))
		sb.WriteString(fmt.Sprintf("Total = %s, ", tx.TotalAmount))
		sb.WriteString(fmt.Sprintf("Retry Counter = %s, ", tx.RetryCounter))
		sb.WriteString(fmt.Sprintf("Tx Type = %s, ", common.BridgingTxType(tx.TransactionType)))
		sb.WriteString(fmt.Sprintf("AlreadyTriedBatch = %s, ", fmt.Sprint(tx.AlreadyTriedBatch)))
		sb.WriteString(fmt.Sprintf("OutputIndexes= %s, ", hex.EncodeToString(tx.OutputIndexes)))
		sb.WriteString("Receivers = [")

		for j, recv := range tx.Receivers {
			if j > 0 {
				sb.WriteString(", ")
			}

			sb.WriteString("(")
			sb.WriteString(recv.DestinationAddress)
			sb.WriteString(fmt.Sprintf(", %d)", recv.Amount))
		}

		sb.WriteString("]")
	}

	return sb.String()
}

type SignedBatchWrapper struct {
	*SignedBatch
}

func (sbw SignedBatchWrapper) String() string {
	var sb strings.Builder

	sb.WriteString("id = ")
	sb.WriteString(fmt.Sprint(sbw.Id))
	sb.WriteString("\nbatch type = ")
	sb.WriteString(fmt.Sprint(sbw.BatchType))
	sb.WriteString("\ndestination chain id = ")
	sb.WriteString(common.ToStrChainID(sbw.DestinationChainId))
	sb.WriteString("\nraw tx = ")
	sb.WriteString(hex.EncodeToString(sbw.RawTransaction))

	if len(sbw.Signature) > 0 {
		sb.WriteString("\nmultisig signature = ")
		sb.WriteString(hex.EncodeToString(sbw.Signature))
	}

	if len(sbw.StakeSignature) > 0 {
		sb.WriteString("\nstake multisig signature = ")
		sb.WriteString(hex.EncodeToString(sbw.StakeSignature))
	}

	sb.WriteString("\nfee payer multisig signature = ")
	sb.WriteString(hex.EncodeToString(sbw.FeeSignature))
	sb.WriteString("\nfirst tx nonce id = ")
	sb.WriteString(fmt.Sprint(sbw.FirstTxNonceId))
	sb.WriteString("\nlast tx nonce id = ")
	sb.WriteString(fmt.Sprint(sbw.LastTxNonceId))

	return sb.String()
}
