package clistakingcomponent

import "bytes"

type CmdResult struct {
}

func (r CmdResult) GetOutput() string {
	var buffer bytes.Buffer

	buffer.WriteString("Done\n")

	return buffer.String()
}
