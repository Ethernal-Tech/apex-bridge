package cliwalletcreate

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	cardanowallet "github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

type cardanoCmdResult struct {
	Multisig        *cardanowallet.Wallet `json:"multisig"`
	Fee             *cardanowallet.Wallet `json:"fee"`
	KeyHash         string                `json:"keyHash"`
	StakeKeyHash    string                `json:"stakeKeyHash"`
	KeyHashFee      string                `json:"keyHashFee"`
	StakeKeyHashFee string                `json:"stakeKeyHashFee"`
	ChainID         string                `json:"chainID"`
	showPrivateKey  bool
}

func (r cardanoCmdResult) GetOutput() string {
	var (
		buffer  bytes.Buffer
		vals    []string
		valsFee []string
	)

	if r.showPrivateKey {
		vals = append(vals, fmt.Sprintf("Signing Key|%s", hex.EncodeToString(r.Multisig.SigningKey)))

		if len(r.Multisig.StakeSigningKey) > 0 {
			vals = append(vals, fmt.Sprintf("Stake Signing Key|%s", hex.EncodeToString(r.Multisig.StakeSigningKey)))
		}
	}

	vals = append(vals,
		fmt.Sprintf("Verifying Key|%s", hex.EncodeToString(r.Multisig.VerificationKey)),
		fmt.Sprintf("Key Hash|%s", r.KeyHash))

	if len(r.Multisig.StakeVerificationKey) > 0 && r.StakeKeyHash != "" {
		vals = append(vals,
			fmt.Sprintf("Stake Verifying Key|%s", hex.EncodeToString(r.Multisig.StakeVerificationKey)),
			fmt.Sprintf("Stake Key Hash|%s", r.StakeKeyHash))
	}

	if r.showPrivateKey {
		valsFee = append(valsFee, fmt.Sprintf("Signing Key|%s", hex.EncodeToString(r.Fee.SigningKey)))

		if len(r.Fee.StakeSigningKey) > 0 {
			vals = append(vals, fmt.Sprintf("Stake Signing Key|%s", hex.EncodeToString(r.Fee.StakeSigningKey)))
		}
	}

	valsFee = append(valsFee,
		fmt.Sprintf("Verifying Key|%s", hex.EncodeToString(r.Fee.VerificationKey)),
		fmt.Sprintf("Key Hash|%s", r.KeyHashFee))

	if len(r.Fee.StakeVerificationKey) > 0 && r.StakeKeyHashFee != "" {
		vals = append(vals,
			fmt.Sprintf("Stake Verifying Key|%s", hex.EncodeToString(r.Fee.StakeVerificationKey)),
			fmt.Sprintf("Stake Key Hash|%s", r.StakeKeyHashFee))
	}

	buffer.WriteString("\n[SECRETS ")
	buffer.WriteString(r.ChainID)
	buffer.WriteString("]\n")
	buffer.WriteString("[Multisig]\n")
	buffer.WriteString(common.FormatKV(vals))
	buffer.WriteString("\n")
	buffer.WriteString("[MultisigFee]\n")
	buffer.WriteString(common.FormatKV(valsFee))
	buffer.WriteString("\n")

	return buffer.String()
}

type evmCmdResult struct {
	ChainID        string `json:"chainID"`
	PrivateKey     string `json:"privateKey"`
	PublicKey      string `json:"publicKey"`
	Address        string `json:"address"`
	showPrivateKey bool
}

func (r evmCmdResult) GetOutput() string {
	var (
		buffer bytes.Buffer
		vals   []string
	)

	if r.showPrivateKey {
		vals = append(vals, fmt.Sprintf("Private Key|%s", r.PrivateKey))
	}

	vals = append(vals, fmt.Sprintf("Public Key|%s", r.PublicKey))

	if r.Address != "" {
		vals = append(vals, fmt.Sprintf("Address|%s", r.Address))
	}

	buffer.WriteString("\n[SECRETS ")
	buffer.WriteString(r.ChainID)
	buffer.WriteString("]\n")
	buffer.WriteString(common.FormatKV(vals))
	buffer.WriteString("\n")

	return buffer.String()
}
