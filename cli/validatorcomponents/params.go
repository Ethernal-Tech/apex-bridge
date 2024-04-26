package clivalidatorcomponents

import "github.com/spf13/cobra"

const (
	configFlag = "config"
	runAPIFlag = "run-api"

	configFlagDesc = "path to config json file"
	runAPIFlagDesc = "specifies whether the api should be run"
)

type initParams struct {
	config string
	runAPI bool
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

	cmd.Flags().BoolVar(
		&ip.runAPI,
		runAPIFlag,
		false,
		runAPIFlagDesc,
	)
}
