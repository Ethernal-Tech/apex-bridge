package clicreateaddress

import (
	"bytes"
	"fmt"

	cardanotx "github.com/Ethernal-Tech/apex-bridge/cardano"
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/cardano-infrastructure/wallet"
)

type AddressAndPolicyScripts struct {
	cardanotx.ApexAddresses
	PolicyScripts cardanotx.ApexPolicyScripts
}

type CmdResult struct {
	AddressAndPolicyScripts []AddressAndPolicyScripts
	ShowPolicyScripts       bool
}

func (r CmdResult) GetOutput() string {
	var buffer bytes.Buffer

	for i, addrAndPolicyScript := range r.AddressAndPolicyScripts {
		args := []string{
			fmt.Sprintf("Multisig Address|%s", addrAndPolicyScript.Multisig.Payment),
			fmt.Sprintf("\n"),
		}

		if addrAndPolicyScript.Multisig.Stake != "" {
			args = append(args, fmt.Sprintf("Multisig Stake Address|%s", addrAndPolicyScript.Multisig.Stake))
			args = append(args, fmt.Sprintf("\n"))
		}

		if i == 0 {
			args = append(args, fmt.Sprintf("Fee Payer Address|%s", addrAndPolicyScript.Fee.Payment))
			args = append(args, fmt.Sprintf("\n"))

			if addrAndPolicyScript.Fee.Stake != "" {
				args = append(args, fmt.Sprintf("Fee Payer Stake Address|%s", addrAndPolicyScript.Fee.Stake))
				args = append(args, fmt.Sprintf("\n"))
			}
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
			showPS("Multisig Payment Policy Script", addrAndPolicyScript.PolicyScripts.Multisig.Payment)
			showPS("Multisig Stake Policy Script", addrAndPolicyScript.PolicyScripts.Multisig.Stake)
			showPS("Fee Payer Payment Policy Script", addrAndPolicyScript.PolicyScripts.Fee.Payment)
			showPS("Fee Payer Stake Policy Script", addrAndPolicyScript.PolicyScripts.Fee.Stake)
		}
	}

	return buffer.String()
}
