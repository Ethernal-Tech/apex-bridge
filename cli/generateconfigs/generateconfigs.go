package cligenerateconfigs

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/spf13/cobra"
)

var paramsData = &generateConfigsParams{}

func GetGenerateConfigsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "generate-configs",
		Short:   "generates default config json files",
		PreRunE: runPreRun,
		Run:     runCommand,
	}

	paramsData.setFlags(cmd)

	return cmd
}

func runPreRun(_ *cobra.Command, _ []string) error {
	return paramsData.validateFlags()
}

func runCommand(cmd *cobra.Command, _ []string) {
	outputter := common.InitializeOutputter(cmd)
	defer outputter.WriteOutput()

	defer func() {
		if r := recover(); r != nil {
			outputter.SetError(fmt.Errorf("%v", r))
		}
	}()

	results, err := paramsData.Execute()
	if err != nil {
		outputter.SetError(err)

		return
	}

	outputter.SetCommandResult(results)
}
