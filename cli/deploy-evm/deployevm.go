package clideployevm

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/spf13/cobra"
)

var deployEVMParamsData = &deployEVMParams{}

func GetDeployEVMCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "deploy-evm",
		Short:   "deploys evm gateway smart contract to evm chain (by default nexus)",
		PreRunE: runPreRun,
		Run:     common.GetCliRunCommand(deployEVMParamsData),
	}

	deployEVMParamsData.setFlags(cmd)

	return cmd
}

func runPreRun(_ *cobra.Command, _ []string) error {
	return deployEVMParamsData.validateFlags()
}
