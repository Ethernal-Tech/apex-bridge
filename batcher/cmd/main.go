package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Ethernal-Tech/apex-bridge/batcher"
	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
)

func main() {
	ctx, cancelCtx := context.WithCancel(context.Background())

	config, err := batcher.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while loading configuration: %v\n", err)
		os.Exit(1)
	}

	logger, err := logger.NewLogger(config.Logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while creating logger: %v\n", err)
		os.Exit(1)
	}

	batcher := batcher.NewBatcher(config, logger)

	go batcher.Execute(ctx)

	signalChannel := make(chan os.Signal, 1)
	// Notify the signalChannel when the interrupt signal is received (Ctrl+C)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	<-signalChannel

	cancelCtx()
}
