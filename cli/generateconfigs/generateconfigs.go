package cligenerateconfigs

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/spf13/cobra"
)

const cardanoChainUse = "cardano-chain"
const evmChainUse = "evm-chain"

var (
	paramsData = &generateConfigsParams{}

	cardanoChainParamsData = &cardanoChainGenerateConfigsParams{}
	evmChainParamsData     = &evmChainGenerateConfigsParams{}
)

func GetGenerateConfigsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "generate-configs",
		Short:   "generates default config json files",
		PreRunE: runPreRun,
		Run:     runCommand,
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

	cardanoChainParamsData.setFlags(cmdCardanoChain)
	evmChainParamsData.setFlags(cmdEvmChain)

	cmd.AddCommand(cmdCardanoChain)
	cmd.AddCommand(cmdEvmChain)

	return cmd
}

func runPreRun(cb *cobra.Command, _ []string) error {
	switch cb.Use {
	case cardanoChainUse:
		return cardanoChainParamsData.validateFlags()
	case evmChainUse:
		return evmChainParamsData.validateFlags()
	default:
		return paramsData.validateFlags()
	}
}

func runCommand(cmd *cobra.Command, _ []string) {
	outputter := common.InitializeOutputter(cmd)
	defer outputter.WriteOutput()

	defer func() {
		if r := recover(); r != nil {
			outputter.SetError(fmt.Errorf("%v", r))
		}
	}()

	results, err := paramsData.Execute()
	if err != nil {
		outputter.SetError(err)

		return
	}

	outputter.SetCommandResult(results)
}
