package cliregisterchain

import (
	"bytes"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
)

type CmdResult struct {
	chainID   string
	blockHash string
}

func (r CmdResult) GetOutput() string {
	var buffer bytes.Buffer

	buffer.WriteString("\n[Register chain ")
	buffer.WriteString(r.chainID)
	buffer.WriteString("]\n")
	buffer.WriteString(common.FormatKV(
		[]string{
			fmt.Sprintf("Block|%s", r.blockHash),
		}))
	buffer.WriteString("\n")

	return buffer.String()
}
