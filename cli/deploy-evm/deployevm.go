package clideployevm

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/spf13/cobra"
)

const upgradeEVMCommandUse = "upgrade"

var deployEVMParamsData = &deployEVMParams{}
var upgradeEVMParamsData = &upgradeEVMParams{}

func GetDeployEVMCommand() *cobra.Command {
	cmdDeployEVM := &cobra.Command{
		Use:     "deploy-evm",
		Short:   "deploys evm gateway smart contract to evm chain (by default nexus)",
		PreRunE: runPreRun,
		Run:     common.GetCliRunCommand(deployEVMParamsData),
	}
	cmdUpgradeEVM := &cobra.Command{
		Use:     upgradeEVMCommandUse,
		Short:   "upgrade desired smart contract(s)",
		PreRunE: runPreRun,
		Run:     common.GetCliRunCommand(upgradeEVMParamsData),
	}

	deployEVMParamsData.setFlags(cmdDeployEVM)
	upgradeEVMParamsData.setFlags(cmdUpgradeEVM)

	cmdDeployEVM.AddCommand(cmdUpgradeEVM)

	return cmdDeployEVM
}

func runPreRun(cb *cobra.Command, _ []string) error {
	if cb.Use == upgradeEVMCommandUse {
		return upgradeEVMParamsData.validateFlags()
	}

	return deployEVMParamsData.validateFlags()
}
