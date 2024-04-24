package clicreateaddress

import (
	"bytes"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
)

type CmdResult struct {
	address string
}

func (r CmdResult) GetOutput() string {
	var buffer bytes.Buffer

	buffer.WriteString(common.FormatKV(
		[]string{
			fmt.Sprintf("Address|%s", r.address),
		}))

	return buffer.String()
}
