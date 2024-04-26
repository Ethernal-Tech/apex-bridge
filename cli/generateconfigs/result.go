package cligenerateconfigs

import (
	"bytes"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
)

type CmdResult struct {
	validatorComponentsConfigPath string
	relayerConfigPath             string
}

func (r CmdResult) GetOutput() string {
	var buffer bytes.Buffer

	buffer.WriteString(common.FormatKV(
		[]string{
			fmt.Sprintf("ValidatorComponents config|%s", r.validatorComponentsConfigPath),
			fmt.Sprintf("Relayer config|%s", r.relayerConfigPath),
		}))

	return buffer.String()
}
