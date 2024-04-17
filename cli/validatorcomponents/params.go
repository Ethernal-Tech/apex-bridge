package clivalidatorcomponents

import "github.com/spf13/cobra"

const (
	configFlag = "config"
	runApiFlag = "run-api"

	configFlagDesc = "path to config json file"
	runApiFlagDesc = "specifies whether the api should be run"
)

type initParams struct {
	config string
	runApi bool
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
		&ip.runApi,
		runApiFlag,
		false,
		runApiFlagDesc,
	)
}
