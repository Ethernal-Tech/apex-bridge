package clicreateaddress

import (
	"bytes"
	"fmt"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
)

type CmdResult struct {
	cardanotx.ApexAddresses
}

func (r CmdResult) GetOutput() string {
	var buffer bytes.Buffer

	args := []string{
		fmt.Sprintf("Multisig Address|%s", r.Multisig.Payment),
		fmt.Sprintf("Fee Payer Address|%s", r.Fee.Payment),
	}

	if r.Multisig.Stake != "" {
		args = append(args, fmt.Sprintf("Multisig Stake Address|%s", r.Multisig.Stake))
	}

	if r.Fee.Stake != "" {
		args = append(args, fmt.Sprintf("Fee Payer Stake Address|%s", r.Fee.Stake))
	}

	buffer.WriteString(common.FormatKV(args))

	return buffer.String()
}
