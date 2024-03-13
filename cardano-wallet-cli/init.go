package cardanowalletcli

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/spf13/cobra"
)

var initParamsData = &initParams{}

func GetInitCommand() *cobra.Command {
	secretsInitCmd := &cobra.Command{
		Use:     "init",
		Short:   "Initializes private keys for the cardano multisig and multisig fee addresses and send data to smart contract",
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

	results, err := initParamsData.Execute()
	if err != nil {
		outputter.SetError(err)

		return
	}

	outputter.SetCommandResult(results)
}
