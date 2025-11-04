package cligenerateconfigs

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/spf13/cobra"
)

const skylineUse = "skyline"

const cardanoChainUse = "cardano-chain"
const evmChainUse = "evm-chain"

var (
	paramsData        = &generateConfigsParams{}
	skylineParamsData = &skylineGenerateConfigsParams{}

	cardanoChainParamsData = &cardanoChainGenerateConfigsParams{}
	evmChainParamsData     = &evmChainGenerateConfigsParams{}
)

func GetGenerateConfigsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "generate-configs",
		Short:   "generates default config json files",
		PreRunE: runPreRun,
		Run:     common.GetCliRunCommand(paramsData),
	}
	cmdSkyline := &cobra.Command{
		Use:     skylineUse,
		Short:   "generate default config json files for skyline",
		PreRunE: runPreRun,
		Run:     common.GetCliRunCommand(skylineParamsData),
	}

	cmdCardanoChain := &cobra.Command{
		Use:     cardanoChainUse,
		Short:   "add cardano chain config to config json file",
		PreRunE: runPreRun,
		Run:     common.GetCliRunCommand(cardanoChainParamsData),
	}
	cmdEvmChain := &cobra.Command{
		Use:     evmChainUse,
		Short:   "add evm chain config to config json file",
		PreRunE: runPreRun,
		Run:     common.GetCliRunCommand(evmChainParamsData),
	}

	paramsData.setFlags(cmd)
	skylineParamsData.setFlags(cmdSkyline)

	cardanoChainParamsData.setFlags(cmdCardanoChain)
	evmChainParamsData.setFlags(cmdEvmChain)

	cmd.AddCommand(cmdSkyline)

	cmd.AddCommand(cmdCardanoChain)
	cmd.AddCommand(cmdEvmChain)

	return cmd
}

func runPreRun(cb *cobra.Command, _ []string) error {
	switch cb.Use {
	case skylineUse:
		return skylineParamsData.validateFlags()
	case cardanoChainUse:
		return cardanoChainParamsData.validateFlags()
	case evmChainUse:
		return evmChainParamsData.validateFlags()
	default:
		return paramsData.validateFlags()
	}
}
