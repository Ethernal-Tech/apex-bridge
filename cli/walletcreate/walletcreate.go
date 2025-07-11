package cliwalletcreate

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/spf13/cobra"
)

const bladeAdminCommandUse = "blade"

var (
	walletCreateParamsData      = &walletCreateParams{}
	walletCreateBladeParamsData = &walletCreateBladeParams{}
)

func GetWalletCreateCommand() *cobra.Command {
	walletCreateCmd := &cobra.Command{
		Use:     "wallet-create",
		Short:   "creates cardano wallet for specific chain id",
		PreRunE: runPreRun,
		Run:     common.GetCliRunCommand(walletCreateParamsData),
	}
	walletCreateBladeCmd := &cobra.Command{
		Use:     bladeAdminCommandUse,
		Short:   "create blade admin or proxy admin wallets using secret manager",
		PreRunE: runPreRun,
		Run:     common.GetCliRunCommand(walletCreateBladeParamsData),
	}

	walletCreateParamsData.setFlags(walletCreateCmd)
	walletCreateBladeParamsData.setFlags(walletCreateBladeCmd)

	walletCreateCmd.AddCommand(walletCreateBladeCmd)

	return walletCreateCmd
}

func runPreRun(cmd *cobra.Command, _ []string) error {
	if cmd.Use == bladeAdminCommandUse {
		return walletCreateBladeParamsData.validateFlags()
	}

	return walletCreateParamsData.validateFlags()
}
