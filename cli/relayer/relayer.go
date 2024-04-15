package clirelayer

import (
	"fmt"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"

	"github.com/Ethernal-Tech/apex-bridge/common"
	relayerCore "github.com/Ethernal-Tech/apex-bridge/relayer/core"
	"github.com/Ethernal-Tech/apex-bridge/relayer/relayer_manager"
	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
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

	_, _ = outputter.Write([]byte("Starting relayer...\n"))

	config, err := loadConfig(initParamsData)
	if err != nil {
		outputter.SetError(err)
		return
	}

	logger, err := logger.NewLogger(config.Logger)
	if err != nil {
		outputter.SetError(err)
		return
	}

	relayerManager, err := relayer_manager.NewRelayerManager(config, logger)
	if err != nil {
		logger.Error("relayer manager creation failed", "err", err)
		outputter.SetError(err)
		return
	}

	err = relayerManager.Start()
	if err != nil {
		logger.Error("relayer manager start failed", "err", err)
		outputter.SetError(err)
		return
	}

	_, _ = outputter.Write([]byte("Relayer has been started\n"))

	defer func() {
		if err := relayerManager.Stop(); err != nil {
			logger.Error("relayer manager stop failed", "err", err)
		}
	}()

	signalChannel := make(chan os.Signal, 1)
	// Notify the signalChannel when the interrupt signal is received (Ctrl+C)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	<-signalChannel

	outputter.SetCommandResult(&CmdResult{})
}

func loadConfig(initParamsData *initParams) (
	*relayerCore.RelayerManagerConfiguration, error,
) {
	var (
		config     *relayerCore.RelayerManagerConfiguration
		err        error
		configPath string = initParamsData.config
	)

	if configPath == "" {
		ex, err := os.Executable()
		if err != nil {
			return nil, err
		}

		configPath = path.Join(filepath.Dir(ex), "relayer_config.json")
	}

	config, err = common.LoadJson[relayerCore.RelayerManagerConfiguration](configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	return config, nil
}
