package cardanowalletcli

import (
	"bytes"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
)

type CmdResult struct {
	SigningKey      string `json:"signingKey"`
	VerifyingKey    string `json:"verifyingKey"`
	SigningKeyFee   string `json:"signingKeyFee"`
	VerifyingKeyFee string `json:"verifyingKeyFee"`
	KeyHash         string `json:"keyHash"`
	KeyHashFee      string `json:"keyHashFee"`
	networkID       string
	showPrivateKey  bool
	blockHash       string
}

func (r CmdResult) GetOutput() string {
	var buffer bytes.Buffer

	vals := []string{
		fmt.Sprintf("Key Hash|%s", r.KeyHash),
		fmt.Sprintf("Verifying Key|%s", r.VerifyingKey),
	}

	if r.showPrivateKey {
		vals = append(vals, fmt.Sprintf("Signing Key|%s", r.SigningKey))
	}

	valsFee := []string{
		fmt.Sprintf("Key Hash|%s", r.KeyHashFee),
		fmt.Sprintf("Verifying Key|%s", r.VerifyingKeyFee),
	}

	if r.showPrivateKey {
		valsFee = append(valsFee, fmt.Sprintf("Signing Key|%s", r.SigningKeyFee))
	}

	buffer.WriteString("\n[SECRETS ")
	buffer.WriteString(r.networkID)
	buffer.WriteString("]\n")
	buffer.WriteString(fmt.Sprintf("Block: %s\n", r.blockHash))
	buffer.WriteString("[Multisig]\n")
	buffer.WriteString(common.FormatKV(vals))
	buffer.WriteString("\n")
	buffer.WriteString("[MultisigFee]\n")
	buffer.WriteString(common.FormatKV(valsFee))
	buffer.WriteString("\n")

	return buffer.String()
}
