package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Ethernal-Tech/apex-bridge/batcher"
)

func main() {
	ctx, cancelCtx := context.WithCancel(context.Background())

	config, err := batcher.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while loading configuration: %v\n", err)
		os.Exit(1)
	}

	batcherManager := batcher.NewBatcherManager(config, ctx)

	err = batcherManager.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start bachers. error: %v\n", err)
		os.Exit(1)
	}

	//defer batcherManager.Stop()

	signalChannel := make(chan os.Signal, 1)
	// Notify the signalChannel when the interrupt signal is received (Ctrl+C)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	<-signalChannel

	cancelCtx()
}
