package clideploysolana

import (
	"bytes"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
)

type deployProgramResult struct {
	Output      string
	TxSignature string
	MintAddress string
}

func (r deployProgramResult) GetOutput() string {
	var buffer bytes.Buffer

	buffer.WriteString(r.Output + "\n")
	buffer.WriteString(fmt.Sprintf("Tx signature: %s\n", r.TxSignature))
	buffer.WriteString(fmt.Sprintf("Mint address: %s\n", r.MintAddress))

	return buffer.String()
}

type exportKeypairResult struct {
	OutputPath string `json:"outputPath"`
	PublicKey  string `json:"publicKey"`
}

func (r exportKeypairResult) GetOutput() string {
	var buffer bytes.Buffer

	buffer.WriteString("\n[SOLANA KEYPAIR]\n")
	buffer.WriteString(common.FormatKV([]string{
		fmt.Sprintf("Output Path|%s", r.OutputPath),
		fmt.Sprintf("Public Key|%s", r.PublicKey),
	}))
	buffer.WriteString("\n")

	return buffer.String()
}
