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
	AddressAndPolicyScripts    []AddressAndPolicyScripts
	RewardAddrAndPolicyScripts []AddressAndPolicyScripts
	ShowPolicyScripts          bool
}

func (r CmdResult) GetOutput() string {
	var buffer bytes.Buffer

	for i, addrAndPolicyScript := range r.AddressAndPolicyScripts {
		writeAddressInfo(&buffer, addrAndPolicyScript, i == 0, false, r.ShowPolicyScripts)
	}

	for _, rewAddrAndPolicyScript := range r.RewardAddrAndPolicyScripts {
		writeAddressInfo(&buffer, rewAddrAndPolicyScript, false, true, r.ShowPolicyScripts)
	}

	return buffer.String()
}

func writeAddressInfo(
	buffer *bytes.Buffer,
	ap AddressAndPolicyScripts,
	isFirst bool,
	isReward bool,
	showScripts bool,
) {
	prefix := ""
	if isReward {
		prefix = "Reward "
	}

	if !isFirst {
		buffer.WriteString("\n")
	}

	args := []string{
		fmt.Sprintf("%sMultisig Address|%s", prefix, ap.Multisig.Payment),
	}

	if ap.Multisig.Stake != "" {
		args = append(args, fmt.Sprintf("%sMultisig Stake Address|%s", prefix, ap.Multisig.Stake))
	}

	if isFirst && !isReward {
		args = append(args, fmt.Sprintf("Fee Payer Address|%s", ap.Fee.Payment))
		if ap.Fee.Stake != "" {
			args = append(args, fmt.Sprintf("Fee Payer Stake Address|%s", ap.Fee.Stake))
		}
	}

	_, _ = buffer.WriteString(common.FormatKV(args))

	if showScripts {
		buffer.WriteString("\n")
		showPolicyScript(buffer, fmt.Sprintf("%sMultisig Payment Policy Script", prefix), ap.PolicyScripts.Multisig.Payment)
		showPolicyScript(buffer, fmt.Sprintf("%sMultisig Stake Policy Script", prefix), ap.PolicyScripts.Multisig.Stake)

		if isFirst && !isReward {
			showPolicyScript(buffer, "Fee Payer Payment Policy Script", ap.PolicyScripts.Fee.Payment)
			showPolicyScript(buffer, "Fee Payer Stake Policy Script", ap.PolicyScripts.Fee.Stake)
		}
	}
}

func showPolicyScript(buffer *bytes.Buffer, title string, ps wallet.IPolicyScript) {
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
