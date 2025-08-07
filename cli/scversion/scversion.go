package cliscversion

import (
	"context"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/spf13/cobra"
)

var scVersionParamsData = &scVersionParams{}

func GetScVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sc-version",
		Short:   "prints sc version",
		PreRunE: runPreRun,
		Run:     runCommand,
	}

	scVersionParamsData.setFlags(cmd)

	return cmd
}

func runPreRun(_ *cobra.Command, _ []string) error {
	return scVersionParamsData.validateFlags()
}

func runCommand(cmd *cobra.Command, _ []string) {
	outputter := common.InitializeOutputter(cmd)
	defer outputter.WriteOutput()

	defer func() {
		if r := recover(); r != nil {
			outputter.SetError(fmt.Errorf("%v", r))
		}
	}()

	results, err := scVersionParamsData.Execute(context.Background(), outputter)
	if err != nil {
		outputter.SetError(err)

		return
	}

	outputter.SetCommandResult(results)
}
