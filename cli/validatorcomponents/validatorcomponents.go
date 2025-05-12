package clivalidatorcomponents

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Ethernal-Tech/apex-bridge/common"
	vcCore "github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/validatorcomponents"
	loggerInfra "github.com/Ethernal-Tech/cardano-infrastructure/logger"
	"github.com/spf13/cobra"
)

var vcParams = &validatorComponentsParams{}

func GetValidatorComponentsCommand() *cobra.Command {
	secretsInitCmd := &cobra.Command{
		Use:     "run-validator-components",
		Short:   "runs validator components",
		PreRunE: runPreRun,
		Run:     runCommand,
	}

	vcParams.setFlags(secretsInitCmd)

	return secretsInitCmd
}

func runPreRun(_ *cobra.Command, _ []string) error {
	return vcParams.validateFlags()
}

func runCommand(cmd *cobra.Command, _ []string) {
	outputter := common.InitializeOutputter(cmd)
	defer outputter.WriteOutput()

	config, err := common.LoadConfig[vcCore.AppConfig](vcParams.config, "")
	if err != nil {
		outputter.SetError(err)

		return
	}

	logger, err := loggerInfra.NewLogger(config.Settings.Logger)
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

	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	validatorComponents, err := validatorcomponents.NewValidatorComponents(ctx, config, vcParams.runAPI, logger)
	if err != nil {
		logger.Error("validator components creation failed", "err", err)
		outputter.SetError(err)

		return
	}

	err = validatorComponents.Start()
	if err != nil {
		logger.Error("validator components start failed", "err", err)
		outputter.SetError(err)

		return
	}

	defer func() {
		err := validatorComponents.Dispose()
		if err != nil {
			logger.Error("error while validator components dispose", "err", err)
		}
	}()

	signalChannel := make(chan os.Signal, 1)
	// Notify the signalChannel when the interrupt signal is received (Ctrl+C)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	select {
	case <-signalChannel:
	}

	outputter.SetCommandResult(&CmdResult{})
}
