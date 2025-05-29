package clistakingcomponent

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Ethernal-Tech/apex-bridge/common"
	stakingCore "github.com/Ethernal-Tech/apex-bridge/staking/core"
	stakingmanager "github.com/Ethernal-Tech/apex-bridge/staking/staking_manager"
	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
	"github.com/spf13/cobra"
)

var initParamsData = &initParams{}

func GetRunStakingComponentCommand() *cobra.Command {
	secretsInitCmd := &cobra.Command{
		Use:     "run-staking-component",
		Short:   "runs staking component",
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

	_, _ = outputter.Write([]byte("Starting staking component...\n"))

	config, err := common.LoadConfig[stakingCore.StakingManagerConfiguration](initParamsData.config, "staking")
	if err != nil {
		outputter.SetError(err)

		return
	}

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

	stakingManager, err := stakingmanager.NewStakingManager(context.Background(), config, logger)
	if err != nil {
		logger.Error("staking manager creation failed", "err", err)
		outputter.SetError(err)

		return
	}

	stakingManager.Start()

	_, _ = outputter.Write([]byte("Staking component has been started\n"))

	signalChannel := make(chan os.Signal, 1)
	// Notify the signalChannel when the interrupt signal is received (Ctrl+C)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	<-signalChannel

	outputter.SetCommandResult(&CmdResult{})
}
