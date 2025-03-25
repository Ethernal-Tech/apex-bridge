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
	setMinAmountsParamsData            = &setMinAmountsParams{}
	validatorsDataParamsData           = &validatorsDataParams{}
	bridgingAddressesBalancesData      = &bridgingAddressesBalancesParams{}
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
	setMinAmountsCmd := &cobra.Command{
		Use:   "set-min-amounts",
		Short: "sets minimal amounts for fee and bridging",
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return setMinAmountsParamsData.ValidateFlags()
		},
		Run: common.GetCliRunCommand(setMinAmountsParamsData),
	}
	validatorDataCmd := &cobra.Command{
		Use:   "get-validators-data",
		Short: "get validators data",
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return validatorsDataParamsData.ValidateFlags()
		},
		Run: common.GetCliRunCommand(validatorsDataParamsData),
	}
	bridgingAddressesBalancesCmd := &cobra.Command{
		Use:   "get-bridging-addresses-balances",
		Short: "get-bridging-addresses-balances",
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return bridgingAddressesBalancesData.ValidateFlags()
		},
		Run: common.GetCliRunCommand(bridgingAddressesBalancesData),
	}

	getChainTokenQuantityParamsData.RegisterFlags(getChainTokenQuantityCmd)
	updateChainTokenQuantityParamsData.RegisterFlags(updateChainTokenQuantityCmd)
	defundParamsData.RegisterFlags(defundCmd)
	setAdditionalDataParamsData.RegisterFlags(setAdditionalDataCmd)
	setMinAmountsParamsData.RegisterFlags(setMinAmountsCmd)
	validatorsDataParamsData.RegisterFlags(validatorDataCmd)
	bridgingAddressesBalancesData.RegisterFlags(bridgingAddressesBalancesCmd)

	cmd := &cobra.Command{
		Use:   "bridge-admin",
		Short: "bridge admin functions",
	}

	cmd.AddCommand(
		getChainTokenQuantityCmd,
		updateChainTokenQuantityCmd,
		defundCmd,
		setAdditionalDataCmd,
		setMinAmountsCmd,
		validatorDataCmd,
		bridgingAddressesBalancesCmd,
	)

	return cmd
}
