package clivalidatorcomponents

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/spf13/cobra"
)

const (
	configFlag = "config"
	runAPIFlag = "run-api"
	modeFlag   = "mode"

	configFlagDesc = "path to config json file"
	runAPIFlagDesc = "specifies whether the api should be run"
	modeFlagDesc   = "specifies in which mode to run validatorcomponents (\"reactor\", \"skyline\")"
)

type validatorComponentsParams struct {
	config string
	runAPI bool
	mode   string
}

func (ip *validatorComponentsParams) validateFlags() error {
	if ip.mode != string(common.ReactorMode) && ip.mode != string(common.SkylineMode) {
		return fmt.Errorf("--%s flag invalid", modeFlag)
	}

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

	cmd.Flags().StringVar(
		&ip.mode,
		modeFlag,
		string(common.ReactorMode),
		modeFlagDesc,
	)
}
