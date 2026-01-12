package clirelayer

import (
	"github.com/spf13/cobra"
)

const (
	configFlag         = "config"
	chainIDsConfigFlag = "chain-ids-config"

	configFlagDesc         = "path to config json file"
	chainIDsConfigFlagDesc = "path to chain ids config json file"
)

type initParams struct {
	config         string
	chainIDsConfig string
}

func (ip *initParams) validateFlags() error {
	return nil
}

func (ip *initParams) setFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&ip.config,
		configFlag,
		"",
		configFlagDesc,
	)
	cmd.Flags().StringVar(
		&ip.chainIDsConfig,
		chainIDsConfigFlag,
		"",
		chainIDsConfigFlagDesc,
	)
}
