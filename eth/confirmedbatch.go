package eth

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
)

type SignedBatch = contractbinding.IBridgeStructsSignedBatch
type ConfirmedTransaction = contractbinding.IBridgeStructsConfirmedTransaction
type UTXOs = contractbinding.IBridgeStructsUTXOs
type UTXO = contractbinding.IBridgeStructsUTXO
type ValidatorCardanoData = contractbinding.IBridgeStructsValidatorCardanoData

type ConfirmedBatch struct {
	ID                         string
	RawTransaction             []byte
	MultisigSignatures         [][]byte
	FeePayerMultisigSignatures [][]byte
}

func NewConfirmedBatch(
	contractConfirmedBatch contractbinding.IBridgeStructsConfirmedBatch,
) (
	*ConfirmedBatch, error,
) {
	// Convert string arrays to byte arrays
	multisigSignatures := make([][]byte, len(contractConfirmedBatch.MultisigSignatures))

	for i, sig := range contractConfirmedBatch.MultisigSignatures {
		sigBytes, err := hex.DecodeString(sig)
		if err != nil {
			return nil, err
		}

		multisigSignatures[i] = sigBytes
	}

	feePayerMultisigSignatures := make([][]byte, len(contractConfirmedBatch.FeePayerMultisigSignatures))

	for i, sig := range contractConfirmedBatch.FeePayerMultisigSignatures {
		sigBytes, err := hex.DecodeString(sig)
		if err != nil {
			return nil, err
		}

		feePayerMultisigSignatures[i] = sigBytes
	}

	// Convert rawTransaction from string to byte array
	rawTx, err := hex.DecodeString(contractConfirmedBatch.RawTransaction)
	if err != nil {
		return nil, err
	}

	return &ConfirmedBatch{
		ID:                         contractConfirmedBatch.Id.String(),
		RawTransaction:             rawTx,
		MultisigSignatures:         multisigSignatures,
		FeePayerMultisigSignatures: feePayerMultisigSignatures,
	}, nil
}

func BatchToString(b SignedBatch) string {
	var sb strings.Builder

	sb.WriteString("id = ")
	sb.WriteString(b.Id.String())
	sb.WriteString("\ndestination chain id = ")
	sb.WriteString(b.DestinationChainId)
	sb.WriteString("\nraw tx = ")
	sb.WriteString(b.RawTransaction)
	sb.WriteString("\nmultisig signature = ")
	sb.WriteString(b.MultisigSignature)
	sb.WriteString("\nfee payer multisig signature = ")
	sb.WriteString(b.FeePayerMultisigSignature)
	sb.WriteString("\nfirst tx nonce id = ")
	sb.WriteString(b.FirstTxNonceId.String())
	sb.WriteString("\nlast tx nonce id = ")
	sb.WriteString(b.LastTxNonceId.String())

	sb.WriteString("\nmultisig owned used utxos cnt = ")
	sb.WriteString(fmt.Sprint(len(b.UsedUTXOs.MultisigOwnedUTXOs)))
	sb.WriteString("\nmultisig owned used utxos = [")

	for _, utxo := range b.UsedUTXOs.MultisigOwnedUTXOs {
		sb.WriteString(fmt.Sprintf("{ Nonce = %v, TxHash = %s, TxIndex = %v, Amount = %v }",
			utxo.Nonce, utxo.TxHash, utxo.TxIndex, utxo.Amount))
	}

	sb.WriteString("]")
	sb.WriteString("\nfeepayer owned used utxos cnt = ")
	sb.WriteString(fmt.Sprint(len(b.UsedUTXOs.FeePayerOwnedUTXOs)))
	sb.WriteString("\nfeepayer owned used utxos = [")

	for _, utxo := range b.UsedUTXOs.FeePayerOwnedUTXOs {
		sb.WriteString(fmt.Sprintf("{ Nonce = %v, TxHash = %s, TxIndex = %v, Amount = %v }",
			utxo.Nonce, utxo.TxHash, utxo.TxIndex, utxo.Amount))
	}

	sb.WriteString("]")

	return sb.String()
}

func (b ConfirmedBatch) String() string {
	var sb strings.Builder

	sb.WriteString("id = ")
	sb.WriteString(b.ID)
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
