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
type StakeDelegationTransaction struct {
	//ChainId              uint8
	StakePoolId          string
	ValidatorStakingKeys []*big.Int
}
type ValidatorChainData = contractbinding.IBridgeStructsValidatorChainData
type BridgeReceiver = contractbinding.IBridgeStructsReceiver

type ConfirmedBatch struct {
	ID              uint64
	RawTransaction  []byte
	Signatures      [][]byte
	FeeSignatures   [][]byte
	Bitmap          *big.Int
	IsConsolidation bool
}

func NewConfirmedBatch(
	contractConfirmedBatch contractbinding.IBridgeStructsConfirmedBatch,
) *ConfirmedBatch {
	return &ConfirmedBatch{
		ID:              contractConfirmedBatch.Id,
		RawTransaction:  contractConfirmedBatch.RawTransaction,
		Signatures:      contractConfirmedBatch.Signatures,
		FeeSignatures:   contractConfirmedBatch.FeeSignatures,
		Bitmap:          contractConfirmedBatch.Bitmap,
		IsConsolidation: contractConfirmedBatch.IsConsolidation,
	}
}

func (b ConfirmedBatch) String() string {
	var sb strings.Builder

	sb.WriteString("id = ")
	sb.WriteString(fmt.Sprint(b.ID))
	sb.WriteString("\nisConsolidation = ")
	sb.WriteString(fmt.Sprint(b.IsConsolidation))
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
	sb.WriteString("\nfee payer multisig signatures = [")

	for i, sig := range b.FeeSignatures {
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
	sb.WriteString("\nisConsolidation = ")
	sb.WriteString(fmt.Sprint(sbw.IsConsolidation))
	sb.WriteString("\ndestination chain id = ")
	sb.WriteString(common.ToStrChainID(sbw.DestinationChainId))
	sb.WriteString("\nraw tx = ")
	sb.WriteString(hex.EncodeToString(sbw.RawTransaction))
	sb.WriteString("\nmultisig signature = ")
	sb.WriteString(hex.EncodeToString(sbw.Signature))
	sb.WriteString("\nfee payer multisig signature = ")
	sb.WriteString(hex.EncodeToString(sbw.FeeSignature))
	sb.WriteString("\nfirst tx nonce id = ")
	sb.WriteString(fmt.Sprint(sbw.FirstTxNonceId))
	sb.WriteString("\nlast tx nonce id = ")
	sb.WriteString(fmt.Sprint(sbw.LastTxNonceId))

	return sb.String()
}
