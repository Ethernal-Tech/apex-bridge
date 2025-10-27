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

type Custodial struct {
	Address      string
	PolicyScript wallet.IPolicyScript
}

type CmdResult struct {
	AddressAndPolicyScripts []AddressAndPolicyScripts
	Custodial               *Custodial
	ShowPolicyScripts       bool
}

func (r CmdResult) GetOutput() string {
	var buffer bytes.Buffer

	if r.Custodial != nil {
		_, _ = buffer.WriteString(common.FormatKV([]string{
			fmt.Sprintf("Custodial Address|%s", r.Custodial.Address),
		}))
		_, _ = buffer.WriteString("\n")
	}

	for i, addrAndPolicyScript := range r.AddressAndPolicyScripts {
		args := []string{
			fmt.Sprintf("Multisig Address|%s", addrAndPolicyScript.Multisig.Payment),
		}

		if addrAndPolicyScript.Multisig.Stake != "" {
			args = append(args, fmt.Sprintf("Multisig Stake Address|%s", addrAndPolicyScript.Multisig.Stake))
		}

		if i == 0 {
			args = append(args, fmt.Sprintf("Fee Payer Address|%s", addrAndPolicyScript.Fee.Payment))

			if addrAndPolicyScript.Fee.Stake != "" {
				args = append(args, fmt.Sprintf("Fee Payer Stake Address|%s", addrAndPolicyScript.Fee.Stake))
			}
		}

		_, _ = buffer.WriteString(common.FormatKV(args))
		_, _ = buffer.WriteString("\n")

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

			if r.Custodial != nil {
				showPS("Custodial Payment Policy Script", r.Custodial.PolicyScript)
			}
		}
	}

	return buffer.String()
}
