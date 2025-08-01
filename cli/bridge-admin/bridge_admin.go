package clibridgeadmin

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/spf13/cobra"
)

const skylineUse = "skyline"

var (
	getChainTokenQuantityParamsData      = &getChainTokenQuantityParams{}
	updateChainTokenQuantityParamsData   = &updateChainTokenQuantityParams{}
	defundParamsData                     = &defundParams{}
	setAdditionalDataParamsData          = &setAdditionalDataParams{}
	setMinAmountsParamsData              = &setMinAmountsParams{}
	validatorsDataParamsData             = &validatorsDataParams{}
	mintNativeTokenParamsData            = &mintNativeTokenParams{}
	bridgingAddressesBalancesData        = &bridgingAddressesBalancesParams{}
	bridgingAddressesBalancesSkylineData = &bridgingAddressesBalancesSkylineParams{}
	stakeDelegationParamsData            = &stakeDelParams{}
	updateBridgingAddrsCountParamsData   = &updateBridgingAddrsCountParams{}
	stakeDeregistrationParamsData        = &stakeDeregParams{}
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
	mintNativeTokenCmd := &cobra.Command{
		Use:   "mint-native-token",
		Short: "mint native token",
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return mintNativeTokenParamsData.ValidateFlags()
		},
		Run: common.GetCliRunCommand(mintNativeTokenParamsData),
	}
	bridgingAddressesBalancesCmd := &cobra.Command{
		Use:   "get-bridging-addresses-balances",
		Short: "get bridging addresses balances",
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return bridgingAddressesBalancesData.ValidateFlags()
		},
		Run: common.GetCliRunCommand(bridgingAddressesBalancesData),
	}
	bridgingAddressesBalancesSkylineCmd := &cobra.Command{
		Use:   skylineUse,
		Short: "get bridging addresses balances for skyline",
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return bridgingAddressesBalancesSkylineData.ValidateFlags()
		},
		Run: common.GetCliRunCommand(bridgingAddressesBalancesSkylineData),
	}
	delegateStakeCmd := &cobra.Command{
		Use:   "delegate-address-to-stake-pool",
		Short: "delegate address to stake pool",
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return stakeDelegationParamsData.ValidateFlags()
		},
		Run: common.GetCliRunCommand(stakeDelegationParamsData),
	}
	updateBridgingAddrsCountCmd := &cobra.Command{
		Use:   "update-bridging-addrs-count",
		Short: "update count of bridging addresses for chain",
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return updateBridgingAddrsCountParamsData.ValidateFlags()
		},
		Run: common.GetCliRunCommand(updateBridgingAddrsCountParamsData),
	}
	deregisterStakeCmd := &cobra.Command{
		Use:   "deregister-stake-address",
		Short: "deregister stake address",
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return stakeDeregistrationParamsData.ValidateFlags()
		},
		Run: common.GetCliRunCommand(stakeDeregistrationParamsData),
	}

	getChainTokenQuantityParamsData.RegisterFlags(getChainTokenQuantityCmd)
	updateChainTokenQuantityParamsData.RegisterFlags(updateChainTokenQuantityCmd)
	defundParamsData.RegisterFlags(defundCmd)
	setAdditionalDataParamsData.RegisterFlags(setAdditionalDataCmd)
	setMinAmountsParamsData.RegisterFlags(setMinAmountsCmd)
	mintNativeTokenParamsData.RegisterFlags(mintNativeTokenCmd)
	validatorsDataParamsData.RegisterFlags(validatorDataCmd)
	bridgingAddressesBalancesData.RegisterFlags(bridgingAddressesBalancesCmd)
	bridgingAddressesBalancesSkylineData.RegisterFlags(bridgingAddressesBalancesSkylineCmd)
	stakeDelegationParamsData.RegisterFlags(delegateStakeCmd)
	updateBridgingAddrsCountParamsData.RegisterFlags(updateBridgingAddrsCountCmd)
	stakeDeregistrationParamsData.RegisterFlags(deregisterStakeCmd)

	bridgingAddressesBalancesCmd.AddCommand(bridgingAddressesBalancesSkylineCmd)

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
		mintNativeTokenCmd,
		validatorDataCmd,
		bridgingAddressesBalancesCmd,
		delegateStakeCmd,
		updateBridgingAddrsCountCmd,
		deregisterStakeCmd,
	)

	return cmd
}
