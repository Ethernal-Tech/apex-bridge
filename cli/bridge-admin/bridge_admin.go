package clibridgeadmin

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/spf13/cobra"
)

var (
	getChainTokenQuantityParamsData    = &getChainTokenQuantityParams{}
	updateChainTokenQuantityParamsData = &updateChainTokenQuantityParams{}
)

func GetBridgeAdminCommand() *cobra.Command {
	getChainTokenQuantityCmd := &cobra.Command{
		Use:   "get-chain-token-quantity",
		Short: "get chain token quantity",
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return getChainTokenQuantityParamsData.ValidateFlags()
		},
		Run: common.GetCliRunCommand(getChainTokenQuantityParamsData),
	}
	updateChainTokenQuantityCmd := &cobra.Command{
		Use:   "update-chain-token-quantity",
		Short: "update chain token quantity",
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return updateChainTokenQuantityParamsData.ValidateFlags()
		},
		Run: common.GetCliRunCommand(updateChainTokenQuantityParamsData),
	}

	getChainTokenQuantityParamsData.RegisterFlags(getChainTokenQuantityCmd)
	updateChainTokenQuantityParamsData.RegisterFlags(updateChainTokenQuantityCmd)

	cmd := &cobra.Command{
		Use:   "bridge-admin",
		Short: "bridge admin functions",
	}

	cmd.AddCommand(getChainTokenQuantityCmd, updateChainTokenQuantityCmd)

	return cmd
}
