package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle/oracle"
)

func main() {
	appConfig, err := common.LoadJson[core.AppConfig]("config.json")
	if err != nil {
		os.Exit(1)
	}

	appConfig.FillOut()

	oracle := oracle.NewOracle(appConfig)
	if oracle == nil {
		fmt.Fprintf(os.Stderr, "Failed to create oracle\n")
		os.Exit(1)
	}

	err = oracle.Start()
	defer oracle.Stop()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start oracle. error: %v\n", err)
		os.Exit(1)
	}

	signalChannel := make(chan os.Signal, 1)
	// Notify the signalChannel when the interrupt signal is received (Ctrl+C)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	select {
	case <-signalChannel:
	case <-oracle.ErrorCh():
	}
}
