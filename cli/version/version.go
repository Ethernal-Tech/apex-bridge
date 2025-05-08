package cliversion

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/versioning"
	"github.com/spf13/cobra"
)

func GetVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Returns the current apex-bridge version",
		Args:  cobra.NoArgs,
		Run:   runCommand,
	}
}

func runCommand(cmd *cobra.Command, _ []string) {
	outputter := common.InitializeOutputter(cmd)
	defer outputter.WriteOutput()

	outputter.SetCommandResult(
		&versionCmdResult{
			Commit:    versioning.Commit,
			Branch:    versioning.Branch,
			BuildTime: versioning.BuildTime,
		},
	)
}
