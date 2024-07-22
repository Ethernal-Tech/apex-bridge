package cliwalletcreate

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
)

type cardanoCmdResult struct {
	SigningKey      []byte `json:"signingKey"`
	VerifyingKey    []byte `json:"verifyingKey"`
	SigningKeyFee   []byte `json:"signingKeyFee"`
	VerifyingKeyFee []byte `json:"verifyingKeyFee"`
	KeyHash         string `json:"keyHash"`
	KeyHashFee      string `json:"keyHashFee"`
	ChainID         string `json:"chainID"`
	showPrivateKey  bool
}

func (r cardanoCmdResult) GetOutput() string {
	var (
		buffer  bytes.Buffer
		vals    []string
		valsFee []string
	)

	if r.showPrivateKey {
		vals = append(vals, fmt.Sprintf("Signing Key|%s", hex.EncodeToString(r.SigningKey)))
	}

	vals = append(vals,
		fmt.Sprintf("Verifying Key|%s", hex.EncodeToString(r.VerifyingKey)),
		fmt.Sprintf("Key Hash|%s", r.KeyHash))

	if r.showPrivateKey {
		valsFee = append(valsFee, fmt.Sprintf("Signing Key|%s", hex.EncodeToString(r.SigningKeyFee)))
	}

	valsFee = append(valsFee,
		fmt.Sprintf("Verifying Key|%s", hex.EncodeToString(r.VerifyingKeyFee)),
		fmt.Sprintf("Key Hash|%s", r.KeyHashFee))

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
