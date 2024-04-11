package cliregisterchain

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/spf13/cobra"
)

var registerChainParamsData = &registerChainParams{}

func GetRegisterChainCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "register-chain",
		Short:   "sends register chain transaction to the bridge node",
		PreRunE: runPreRun,
		Run:     runCommand,
	}

	registerChainParamsData.setFlags(cmd)

	return cmd
}

func runPreRun(_ *cobra.Command, _ []string) error {
	return registerChainParamsData.validateFlags()
}

func runCommand(cmd *cobra.Command, _ []string) {
	outputter := common.InitializeOutputter(cmd)
	defer outputter.WriteOutput()

	results, err := registerChainParamsData.Execute()
	if err != nil {
		outputter.SetError(err)

		return
	}

	outputter.SetCommandResult(results)
}
