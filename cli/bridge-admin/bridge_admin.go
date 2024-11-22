package clibridgeadmin

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/spf13/cobra"
)

var (
	getChainTokenQuantityParamsData    = &getChainTokenQuantityParams{}
	updateChainTokenQuantityParamsData = &updateChainTokenQuantityParams{}
	defundParamsData                   = &defundParams{}
	setAdditionalDataParamsData        = &setAdditionalDataParams{}
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
	defundCmd := &cobra.Command{
		Use:   "defund",
		Short: "dufund chain hot wallet",
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return defundParamsData.ValidateFlags()
		},
		Run: common.GetCliRunCommand(defundParamsData),
	}
	setAdditionalDataCmd := &cobra.Command{
		Use:   "set-additional-data",
		Short: "set additional data for a specific chain",
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return setAdditionalDataParamsData.ValidateFlags()
		},
		Run: common.GetCliRunCommand(setAdditionalDataParamsData),
	}

	getChainTokenQuantityParamsData.RegisterFlags(getChainTokenQuantityCmd)
	updateChainTokenQuantityParamsData.RegisterFlags(updateChainTokenQuantityCmd)
	defundParamsData.RegisterFlags(defundCmd)
	setAdditionalDataParamsData.RegisterFlags(setAdditionalDataCmd)

	cmd := &cobra.Command{
		Use:   "bridge-admin",
		Short: "bridge admin functions",
	}

	cmd.AddCommand(getChainTokenQuantityCmd, updateChainTokenQuantityCmd, defundCmd, setAdditionalDataCmd)

	return cmd
}
