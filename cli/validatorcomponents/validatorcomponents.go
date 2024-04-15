package clivalidatorcomponents

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/Ethernal-Tech/apex-bridge/common"
	vcCore "github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/validatorcomponents"
	"github.com/spf13/cobra"
)

var initParamsData = &initParams{}

func GetValidatorComponentsCommand() *cobra.Command {
	secretsInitCmd := &cobra.Command{
		Use:     "run-validator-components",
		Short:   "runs validator components",
		PreRunE: runPreRun,
		Run:     runCommand,
	}

	initParamsData.setFlags(secretsInitCmd)

	return secretsInitCmd
}

func runPreRun(_ *cobra.Command, _ []string) error {
	return initParamsData.validateFlags()
}

func runCommand(cmd *cobra.Command, _ []string) {
	outputter := common.InitializeOutputter(cmd)
	defer outputter.WriteOutput()

	config, err := loadConfig(initParamsData)
	if err != nil {
		outputter.SetError(err)
		return
	}

	validatorComponents, err := validatorcomponents.NewValidatorComponents(config)
	if err != nil {
		outputter.SetError(err)
		return
	}

	err = validatorComponents.Start()
	if err != nil {
		outputter.SetError(err)
		return
	}

	defer validatorComponents.Stop()

	signalChannel := make(chan os.Signal, 1)
	// Notify the signalChannel when the interrupt signal is received (Ctrl+C)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	select {
	case <-signalChannel:
	case err = <-validatorComponents.ErrorCh():
		outputter.SetError(err)
	}

	outputter.SetCommandResult(&CmdResult{})
}

func loadConfig(initParamsData *initParams) (
	*vcCore.AppConfig, error,
) {
	var (
		config     *vcCore.AppConfig
		err        error
		configPath string = initParamsData.config
	)

	if configPath == "" {
		ex, err := os.Executable()
		if err != nil {
			return nil, err
		}

		configPath = filepath.Dir(ex) + "/config.json"
	}

	config, err = common.LoadJson[vcCore.AppConfig](configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	return config, nil
}
