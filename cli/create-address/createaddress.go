package clicreateaddress

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/spf13/cobra"
)

var params = &createAddressParams{}

func GetCreateAddressCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create-address",
		Short:   "creates a multisig address from multiple Cardano verification keys.",
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

	results, err := params.Execute()
	if err != nil {
		outputter.SetError(err)

		return
	}

	outputter.SetCommandResult(results)
}
