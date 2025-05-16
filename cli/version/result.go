package cliversion

import (
	"bytes"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
)

type versionCmdResult struct {
	Commit    string `json:"commit"`
	Branch    string `json:"branch"`
	BuildTime string `json:"buildTime"`
}

func (v versionCmdResult) GetOutput() string {
	var buffer bytes.Buffer

	buffer.WriteString("[VERSION INFO]\n")
	buffer.WriteString(common.FormatKV([]string{
		fmt.Sprintf("Git branch|%s", v.Branch),
		fmt.Sprintf("Commit hash|%s", v.Commit),
		fmt.Sprintf("Build time|%s", v.BuildTime),
	}))

	return buffer.String()
}
