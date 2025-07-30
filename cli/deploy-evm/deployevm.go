package clideployevm

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/spf13/cobra"
)

const upgradeEVMCommandUse = "upgrade"
const deployContractCommandUse = "deploy-contract"
const setValidatorsChainDataEVMCommandUse = "set-validators-chain-data"

var deployEVMParamsData = &deployEVMParams{}
var upgradeEVMParamsData = &upgradeEVMParams{}
var deployContractParamsData = &deployContractParams{}
var setValidatorsChainDataEVMParamsData = &setValidatorsChainDataEVMParams{}

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
	cmdDeployContract := &cobra.Command{
		Use:     deployContractCommandUse,
		Short:   "deploy smart contract",
		PreRunE: runPreRun,
		Run:     common.GetCliRunCommand(deployContractParamsData),
	}
	cmdSetVCDEVM := &cobra.Command{
		Use:     setValidatorsChainDataEVMCommandUse,
		Short:   "set validators chain data",
		PreRunE: runPreRun,
		Run:     common.GetCliRunCommand(setValidatorsChainDataEVMParamsData),
	}

	deployEVMParamsData.setFlags(cmdDeployEVM)
	upgradeEVMParamsData.setFlags(cmdUpgradeEVM)
	deployContractParamsData.setFlags(cmdDeployContract)
	setValidatorsChainDataEVMParamsData.setFlags(cmdSetVCDEVM)

	cmdDeployEVM.AddCommand(cmdUpgradeEVM)
	cmdDeployEVM.AddCommand(cmdDeployContract)
	cmdDeployEVM.AddCommand(cmdSetVCDEVM)

	return cmdDeployEVM
}

func runPreRun(cb *cobra.Command, _ []string) error {
	if cb.Use == upgradeEVMCommandUse {
		return upgradeEVMParamsData.validateFlags()
	}

	if cb.Use == deployContractCommandUse {
		return deployContractParamsData.validateFlags()
	}

	if cb.Use == setValidatorsChainDataEVMCommandUse {
		return setValidatorsChainDataEVMParamsData.validateFlags()
	}

	return deployEVMParamsData.validateFlags()
}
