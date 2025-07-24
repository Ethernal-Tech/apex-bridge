package cliscversion

import (
	"bytes"
)

type CmdResult struct {
}

func (r CmdResult) GetOutput() string {
	var buffer bytes.Buffer

	return buffer.String()
}
