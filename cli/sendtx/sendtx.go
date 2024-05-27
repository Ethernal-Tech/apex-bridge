package clisendtx

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/spf13/cobra"
)

var sendtxParamsData = &sendTxParams{}

func GetSendTxCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sendtx",
		Short:   "sends apex bridging transaction",
		PreRunE: runPreRun,
		Run:     runCommand,
	}

	sendtxParamsData.setFlags(cmd)

	return cmd
}

func runPreRun(_ *cobra.Command, _ []string) error {
	return sendtxParamsData.validateFlags()
}

func runCommand(cmd *cobra.Command, _ []string) {
	outputter := common.InitializeOutputter(cmd)
	defer outputter.WriteOutput()

	results, err := sendtxParamsData.Execute(outputter)
	if err != nil {
		outputter.SetError(err)

		return
	}

	outputter.SetCommandResult(results)
}
