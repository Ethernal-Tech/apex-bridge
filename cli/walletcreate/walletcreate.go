package cliwalletcreate

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/spf13/cobra"
)

var initParamsData = &initParams{}

func GetWalletCreateCommand() *cobra.Command {
	secretsInitCmd := &cobra.Command{
		Use:     "wallet-create",
		Short:   "creates cardano wallet for specific chain id",
		PreRunE: runPreRun,
		Run:     runCommand,
	}

	initParamsData.setFlags(secretsInitCmd)

	return secretsInitCmd
}

func runPreRun(_ *cobra.Command, _ []string) error {
	return initParamsData.validateFlags()
}

func runCommand(cmd *cobra.Command, _ []string) {
	outputter := common.InitializeOutputter(cmd)
	defer outputter.WriteOutput()

	defer func() {
		if r := recover(); r != nil {
			outputter.SetError(fmt.Errorf("%v", r))
		}
	}()

	results, err := initParamsData.Execute()
	if err != nil {
		outputter.SetError(err)

		return
	}

	outputter.SetCommandResult(results)
}
