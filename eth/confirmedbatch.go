package eth

import (
	"encoding/hex"

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
