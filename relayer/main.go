package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Ethernal-Tech/apex-bridge/relayer/core"
	"github.com/Ethernal-Tech/apex-bridge/relayer/relayer_manager"
)

func main() {

	config, err := relayer_manager.LoadConfig("config.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while loading configuration: %v\n", err)
		os.Exit(1)
	}

	relayerManager := relayer_manager.NewRelayerManager(config, make(map[string]core.ChainOperations), make(map[string]core.Database))
	if relayerManager == nil {
		fmt.Fprintf(os.Stderr, "Failed to create relayer manager")
		os.Exit(1)
	}

	err = relayerManager.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start relayers. error: %v\n", err)
		os.Exit(1)
	}

	defer relayerManager.Stop()

	signalChannel := make(chan os.Signal, 1)
	// Notify the signalChannel when the interrupt signal is received (Ctrl+C)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	<-signalChannel
}
