package clicreateaddress

import (
	"bytes"
	"fmt"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

type CmdResult struct {
	cardanotx.ApexAddresses
	PolicyScripts     cardanotx.ApexPolicyScripts
	ShowPolicyScripts bool
}

func (r CmdResult) GetOutput() string {
	var buffer bytes.Buffer

	args := []string{
		fmt.Sprintf("Multisig Address|%s", r.Multisig.Payment),
	}

	if r.Multisig.Stake != "" {
		args = append(args, fmt.Sprintf("Multisig Stake Address|%s", r.Multisig.Stake))
	}

	args = append(args, fmt.Sprintf("Fee Payer Address|%s", r.Fee.Payment))

	if r.Fee.Stake != "" {
		args = append(args, fmt.Sprintf("Fee Payer Stake Address|%s", r.Fee.Stake))
	}

	_, _ = buffer.WriteString(common.FormatKV(args))

	if r.ShowPolicyScripts {
		showPS := func(title string, ps wallet.IPolicyScript) {
			if ps == nil {
				return
			}

			bytes, err := ps.GetBytesJSON()
			if err != nil {
				_, _ = buffer.WriteString(fmt.Sprintf("\nFailed to generate %s: %s",
					title, err.Error()))
			} else {
				_, _ = buffer.WriteString(fmt.Sprintf("\n%s:\n%s\n", title, bytes))
			}
		}

		buffer.WriteString("\n")
		showPS("Multisig Payment Policy Script", r.PolicyScripts.Multisig.Payment)
		showPS("Multisig Stake Policy Script", r.PolicyScripts.Multisig.Stake)
		showPS("Fee Payer Payment Policy Script", r.PolicyScripts.Fee.Payment)
		showPS("Fee Payer Stake Policy Script", r.PolicyScripts.Fee.Stake)
	}

	return buffer.String()
}
