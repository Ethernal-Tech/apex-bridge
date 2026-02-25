package clideploysolana

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/spf13/cobra"
)

const deployProgramCommandUse = "deploy-program"

var deployProgramParamsData = &deployProgramParams{}

func GetDeploySolanaCommand() *cobra.Command {
	cmdDeploySolana := &cobra.Command{
		Use:   "deploy-solana",
		Short: "deploy Solana program to a cluster",
	}

	cmdDeployProgram := &cobra.Command{
		Use:     deployProgramCommandUse,
		Short:   "deploy a Solana program using the solana CLI",
		PreRunE: runPreRun,
		Run:     common.GetCliRunCommand(deployProgramParamsData),
	}

	deployProgramParamsData.setFlags(cmdDeployProgram)
	cmdDeploySolana.AddCommand(cmdDeployProgram)

	return cmdDeploySolana
}

func runPreRun(cb *cobra.Command, _ []string) error {
	if cb.Use == deployProgramCommandUse {
		return deployProgramParamsData.validateFlags()
	}

	return nil
}
