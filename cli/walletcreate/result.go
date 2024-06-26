package cliwalletcreate

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
)

type CmdResult struct {
	SigningKey      []byte `json:"signingKey"`
	VerifyingKey    []byte `json:"verifyingKey"`
	SigningKeyFee   []byte `json:"signingKeyFee"`
	VerifyingKeyFee []byte `json:"verifyingKeyFee"`
	KeyHash         string `json:"keyHash"`
	KeyHashFee      string `json:"keyHashFee"`
	chainID         string
	showPrivateKey  bool
}

func (r CmdResult) GetOutput() string {
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
	buffer.WriteString(r.chainID)
	buffer.WriteString("]\n")
	buffer.WriteString("[Multisig]\n")
	buffer.WriteString(common.FormatKV(vals))
	buffer.WriteString("\n")
	buffer.WriteString("[MultisigFee]\n")
	buffer.WriteString(common.FormatKV(valsFee))
	buffer.WriteString("\n")

	return buffer.String()
}
