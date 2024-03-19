package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle/oracle"
	"github.com/Ethernal-Tech/apex-bridge/oracle/utils"
)

func main() {
	appConfig, err := utils.LoadJson[core.AppConfig]("config.json")
	if err != nil {
		os.Exit(1)
	}

	initialUtxos, err := utils.LoadJson[core.InitialUtxos]("initialUtxos.json")
	if err != nil {
		os.Exit(1)
	}

	oracle := oracle.NewOracle(appConfig, initialUtxos)
	if oracle == nil {
		fmt.Fprintf(os.Stderr, "Failed to create oracle\n")
		os.Exit(1)
	}

	err = oracle.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start oracle. error: %v\n", err)
		os.Exit(1)
	}

	defer oracle.Stop()

	signalChannel := make(chan os.Signal, 1)
	// Notify the signalChannel when the interrupt signal is received (Ctrl+C)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	select {
	case <-signalChannel:
	case <-oracle.ErrorCh():
	}
}
