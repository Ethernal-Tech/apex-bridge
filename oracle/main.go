package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Ethernal-Tech/apex-bridge/oracle/core"
	"github.com/Ethernal-Tech/apex-bridge/oracle/oracle"
	"gopkg.in/yaml.v3"
)

func main() {
	appConfig, err := loadConfig()
	if err != nil {
		os.Exit(1)
	}

	initialUtxos, err := loadInitialUtxos()
	if err != nil {
		os.Exit(1)
	}

	oracle := oracle.NewOracle(appConfig, initialUtxos)

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

func loadInitialUtxos() (*core.InitialUtxos, error) {
	f, err := os.Open("initialUtxos.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open initialUtxos.json. error: %v\n", err)
		return nil, err
	}

	defer f.Close()

	var initialUtxos core.InitialUtxos
	decoder := json.NewDecoder(f)
	err = decoder.Decode(&initialUtxos)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to decode initialUtxos.json. error: %v\n", err)
		return nil, err
	}

	return &initialUtxos, nil
}

func loadConfig() (*core.AppConfig, error) {
	f, err := os.Open("config.yml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open config.yml. error: %v\n", err)
		return nil, err
	}

	defer f.Close()

	var appConfig core.AppConfig
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&appConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to decode config.yml. error: %v\n", err)
		return nil, err
	}

	return &appConfig, nil
}
