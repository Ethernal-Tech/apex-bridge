package utils

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/gagliardetto/solana-go"
)

func ParseSolanaTxHash(txHash string) (solana.Signature, error) {
	txHash = strings.TrimPrefix(txHash, "0x")
	if sig, err := solana.SignatureFromBase58(txHash); err == nil {
		return sig, nil
	}

	b, err := hex.DecodeString(txHash)
	if err != nil || len(b) != solana.SignatureLength {
		return solana.Signature{}, fmt.Errorf("invalid solana tx hash")
	}

	return solana.SignatureFromBytes(b), nil
}

func ParseTxHashToBytes(chainID string, txHashStr string) ([]byte, error) {
	if chainID == common.ChainIDStrSolana {
		sig, err := ParseSolanaTxHash(txHashStr)
		if err != nil {
			return nil, err
		}

		return sig[:], nil
	}

	hash := common.NewHashFromHexString(txHashStr)

	return hash[:], nil
}
