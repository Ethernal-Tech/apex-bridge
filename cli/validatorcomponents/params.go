package clivalidatorcomponents

import "github.com/spf13/cobra"

const (
	configFlag = "config"
	runAPIFlag = "run-api"

	configFlagDesc = "path to config json file"
	runAPIFlagDesc = "specifies whether the api should be run"
)

type validatorComponentsParams struct {
	config string
	runAPI bool
}

func (ip *validatorComponentsParams) validateFlags() error {
	return nil
}

func (ip *validatorComponentsParams) setFlags(cmd *cobra.Command) {
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
