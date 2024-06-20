package eth

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
)

type SignedBatch = contractbinding.IBridgeStructsSignedBatch
type ConfirmedTransaction = contractbinding.IBridgeStructsConfirmedTransaction
type BatchProposerData = contractbinding.IBridgeStructsBatchProposerData
type ValidatorCardanoData = contractbinding.IBridgeStructsValidatorCardanoData
type UTXO = contractbinding.IBridgeStructsUTXO

type ConfirmedBatch struct {
	ID                         uint64
	RawTransaction             []byte
	MultisigSignatures         [][]byte
	FeePayerMultisigSignatures [][]byte
}

func NewConfirmedBatch(
	contractConfirmedBatch contractbinding.IBridgeStructsConfirmedBatch,
) (
	*ConfirmedBatch, error,
) {
	return &ConfirmedBatch{
		ID:                         contractConfirmedBatch.Id,
		RawTransaction:             contractConfirmedBatch.RawTransaction,
		MultisigSignatures:         contractConfirmedBatch.MultisigSignatures,
		FeePayerMultisigSignatures: contractConfirmedBatch.FeePayerMultisigSignatures,
	}, nil
}

func BatchToString(b SignedBatch) string {
	var sb strings.Builder

	sb.WriteString("id = ")
	sb.WriteString(fmt.Sprint(b.Id))
	sb.WriteString("\ndestination chain id = ")
	sb.WriteString(common.ToStrChainID(b.DestinationChainId))
	sb.WriteString("\nraw tx = ")
	sb.WriteString(hex.EncodeToString(b.RawTransaction))
	sb.WriteString("\nmultisig signature = ")
	sb.WriteString(hex.EncodeToString(b.MultisigSignature))
	sb.WriteString("\nfee payer multisig signature = ")
	sb.WriteString(hex.EncodeToString(b.FeePayerMultisigSignature))
	sb.WriteString("\nfirst tx nonce id = ")
	sb.WriteString(fmt.Sprint(b.FirstTxNonceId))
	sb.WriteString("\nlast tx nonce id = ")
	sb.WriteString(fmt.Sprint(b.LastTxNonceId))

	sb.WriteString("\nslot = ")
	sb.WriteString(fmt.Sprint(b.ProposerData.Slot))
	sb.WriteString("\nmultisig utxos cnt = ")
	sb.WriteString(fmt.Sprint(len(b.ProposerData.MultisigUTXOs)))
	sb.WriteString("\nmultisig utxos = ")

	for i, utxo := range b.ProposerData.MultisigUTXOs {
		if i > 0 {
			sb.WriteString(", ")
		}

		sb.WriteRune('{')
		sb.WriteString(hex.EncodeToString(utxo.TxHash[:]))
		sb.WriteRune(',')
		sb.WriteString(fmt.Sprint(utxo.TxIndex))
		sb.WriteRune('}')
	}

	sb.WriteString("\nfeepayer utxos cnt = ")
	sb.WriteString(fmt.Sprint(len(b.ProposerData.FeePayerUTXOs)))
	sb.WriteString("\nfeepayer utxos = ")

	for i, utxo := range b.ProposerData.FeePayerUTXOs {
		if i > 0 {
			sb.WriteString(", ")
		}

		sb.WriteRune('{')
		sb.WriteString(hex.EncodeToString(utxo.TxHash[:]))
		sb.WriteRune(',')
		sb.WriteString(fmt.Sprint(utxo.TxIndex))
		sb.WriteRune('}')
	}

	return sb.String()
}

func (b ConfirmedBatch) String() string {
	var sb strings.Builder

	sb.WriteString("id = ")
	sb.WriteString(fmt.Sprint(b.ID))
	sb.WriteString("\nraw tx = ")
	sb.WriteString(hex.EncodeToString(b.RawTransaction))
	sb.WriteString("\nmultisig signatures = [")

	for i, sig := range b.MultisigSignatures {
		if i > 0 {
			sb.WriteString(", ")
		}

		sb.WriteString(hex.EncodeToString(sig))
	}

	sb.WriteString("]")
	sb.WriteString("\nfee payer multisig signatures = [")

	for i, sig := range b.FeePayerMultisigSignatures {
		if i > 0 {
			sb.WriteString(", ")
		}

		sb.WriteString(hex.EncodeToString(sig))
	}

	sb.WriteString("]")

	return sb.String()
}
