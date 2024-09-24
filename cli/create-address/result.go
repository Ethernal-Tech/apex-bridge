package clicreateaddress

import (
	"bytes"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
)

type CmdResult struct {
	address         string
	multisigAddress string
}

func (r CmdResult) GetOutput() string {
	var buffer bytes.Buffer

	if r.multisigAddress != "" {
		buffer.WriteString(common.FormatKV(
			[]string{
				fmt.Sprintf("Multisig Address|%s", r.multisigAddress),
				fmt.Sprintf("Fee Payer Address|%s", r.address),
			}))
	} else {
		buffer.WriteString(common.FormatKV(
			[]string{
				fmt.Sprintf("Address|%s", r.address),
			}))
	}

	return buffer.String()
}
