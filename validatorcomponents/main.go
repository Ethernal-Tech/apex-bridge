package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Ethernal-Tech/apex-bridge/common"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/core"
	"github.com/Ethernal-Tech/apex-bridge/validatorcomponents/validatorcomponents"
)

func main() {
	config, err := common.LoadJson[core.AppConfig]("config.json")
	if err != nil {
		fmt.Printf("failed to load config file\n")
		os.Exit(1)
	}

	validatorComponents, err := validatorcomponents.NewValidatorComponents(config)
	if err != nil {
		fmt.Printf("failed to create NewValidatorComponents\n")
		os.Exit(1)
	}

	err = validatorComponents.Start()
	defer validatorComponents.Stop()
	if err != nil {
		fmt.Printf("failed to start validatorComponents\n")
		os.Exit(1)
	}

	signalChannel := make(chan os.Signal, 1)
	// Notify the signalChannel when the interrupt signal is received (Ctrl+C)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	select {
	case <-signalChannel:
	case <-validatorComponents.ErrorCh():
	}
}
