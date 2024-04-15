package clirelayer

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/Ethernal-Tech/apex-bridge/common"
	relayerCore "github.com/Ethernal-Tech/apex-bridge/relayer/core"
	"github.com/Ethernal-Tech/apex-bridge/relayer/relayer_manager"
	"github.com/spf13/cobra"
)

var initParamsData = &initParams{}

func GetRunRelayerCommand() *cobra.Command {
	secretsInitCmd := &cobra.Command{
		Use:     "run-relayer",
		Short:   "runs relayer component",
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

	relayerManager, err := relayer_manager.NewRelayerManager(config, make(map[string]relayerCore.ChainOperations), make(map[string]relayerCore.Database))
	if err != nil {
		outputter.SetError(err)
		return
	}

	err = relayerManager.Start()
	if err != nil {
		outputter.SetError(err)
		return
	}

	defer relayerManager.Stop()

	signalChannel := make(chan os.Signal, 1)
	// Notify the signalChannel when the interrupt signal is received (Ctrl+C)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	<-signalChannel

	outputter.SetCommandResult(&CmdResult{})
}

func loadConfig(initParamsData *initParams) (
	*relayerCore.RelayerManagerConfiguration, error,
) {
	var config *relayerCore.RelayerManagerConfiguration
	var err error

	if initParamsData.config != "" {
		config, err = common.LoadJson[relayerCore.RelayerManagerConfiguration](initParamsData.config)
		if err != nil {
			return nil, fmt.Errorf("failed to load config: %v", err)
		}
	} else {
		ex, err := os.Executable()
		if err != nil {
			return nil, err
		}

		configPath := filepath.Dir(ex) + "/relayer_config.json"
		config, err = common.LoadJson[relayerCore.RelayerManagerConfiguration](configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load config: %v", err)
		}
	}

	return config, nil
}
