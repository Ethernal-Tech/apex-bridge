package clideployevm

import (
	"context"
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/spf13/cobra"
)

var params = &deployEVMParams{}

func GetDeployEVMCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "deploy-evm",
		Short:   "deploys evm gateway smart contract to evm chain (by default nexus)",
		PreRunE: runPreRun,
		Run:     runCommand,
	}

	params.setFlags(cmd)

	return cmd
}

func runPreRun(_ *cobra.Command, _ []string) error {
	return params.validateFlags()
}

func runCommand(cmd *cobra.Command, _ []string) {
	outputter := common.InitializeOutputter(cmd)
	defer outputter.WriteOutput()

	defer func() {
		if r := recover(); r != nil {
			outputter.SetError(fmt.Errorf("%v", r))
		}
	}()

	results, err := params.Execute(context.Background(), outputter)
	if err != nil {
		outputter.SetError(err)

		return
	}

	outputter.SetCommandResult(results)
}
