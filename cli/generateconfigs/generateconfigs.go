package cligenerateconfigs

import (
	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/spf13/cobra"
)

const skylineUse = "skyline"

var (
	paramsData        = &generateConfigsParams{}
	skylineParamsData = &skylineGenerateConfigsParams{}
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

	paramsData.setFlags(cmd)
	skylineParamsData.setFlags(cmdSkyline)

	cmd.AddCommand(cmdSkyline)

	return cmd
}

func runPreRun(cb *cobra.Command, _ []string) error {
	if cb.Use == skylineUse {
		return skylineParamsData.validateFlags()
	}

	return paramsData.validateFlags()
}
