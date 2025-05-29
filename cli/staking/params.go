package clistakingcomponent

import (
	"github.com/spf13/cobra"
)

const (
	configFlag = "config"

	configFlagDesc = "path to config json file"
)

type initParams struct {
	config string
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
}
