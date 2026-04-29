package clideploysolana

import (
	"bytes"
	"fmt"
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
