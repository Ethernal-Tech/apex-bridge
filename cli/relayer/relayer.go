package clirelayer

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Ethernal-Tech/apex-bridge/common"
	relayerCore "github.com/Ethernal-Tech/apex-bridge/relayer/core"
	relayermanager "github.com/Ethernal-Tech/apex-bridge/relayer/relayer_manager"
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

	config, err := common.LoadConfig[relayerCore.RelayerManagerConfiguration](initParamsData.config, "relayer")
	if err != nil {
		outputter.SetError(err)

		return
	}

	chainIDsConfig, err := common.LoadConfig[common.ChainIDsConfig](initParamsData.chainIDsConfig, "relayer")
	if err != nil {
		outputter.SetError(err)

		return
	}

	config.SetupChainIDs(chainIDsConfig)

	logger, err := logger.NewLogger(config.Logger)
	if err != nil {
		outputter.SetError(err)

		return
	}

	defer func() {
		if r := recover(); r != nil {
			logger.Error("PANIC", "err", r)
			outputter.SetError(fmt.Errorf("%v", r))
		}
	}()

	relayerManager, err := relayermanager.NewRelayerManager(context.Background(), config, logger)
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
