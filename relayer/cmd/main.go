package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Ethernal-Tech/apex-bridge/relayer"
	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
	"github.com/hashicorp/go-hclog"
)

func main() {
	ctx, cancelCtx := context.WithCancel(context.Background())

	// TODO: read from file
	config := &relayer.RelayerConfiguration{
		Cardano: relayer.CardanoConfig{
			TestNetMagic:      uint(2),
			BlockfrostUrl:     "https://cardano-preview.blockfrost.io/api/v0",
			BlockfrostAPIKey:  "preview7mGSjpyEKb24OxQ4cCxomxZ5axMs5PvE",
			AtLeastValidators: 2.0 / 3.0,
			PotentialFee:      300_000,
		},
		Bridge: relayer.BridgeConfig{
			NodeUrl:              "https://polygon-mumbai-pokt.nodies.app", // will be our node,
			SmartContractAddress: "0xb2B87f7e652Aa847F98Cc05e130d030b91c7B37d",
		},
		PullTimeMilis: 1000,
		Logger: logger.LoggerConfig{
			LogFilePath:   "./relayer_logs",
			LogLevel:      hclog.Debug,
			JSONLogFormat: false,
			AppendFile:    true,
		},
	}

	logger, err := logger.NewLogger(config.Logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while creating logger: %v\n", err)
		os.Exit(1)
	}

	relayer := relayer.NewRelayer(config, logger)

	go relayer.Execute(ctx)

	signalChannel := make(chan os.Signal, 1)
	// Notify the signalChannel when the interrupt signal is received (Ctrl+C)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	<-signalChannel

	cancelCtx()
}
