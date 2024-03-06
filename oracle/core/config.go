package core

import (
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
)

type BridgingAddress struct {
	ChainId    string `json:"chainId"`
	Address    string `json:"address"`
	FeeAddress string `json:"feeAddress"`
}

type CardanoChainConfig struct {
	ChainId                  string                     `json:"chainId"`
	NetworkAddress           string                     `json:"networkAddress"`
	NetworkMagic             string                     `json:"networkMagic"`
	StartBlockHash           string                     `json:"startBlockHash"`
	StartSlot                string                     `json:"startSlot"`
	StartBlockNumber         string                     `json:"startBlockNumber"`
	ConfirmationBlockCount   uint                       `json:"confirmationBlockCount"`
	FeeAddress               string                     `json:"feeAddress"`
	BridgingAddresses        map[string]BridgingAddress `json:"bridgingAddresses"`
	OtherAddressesOfInterest []string                   `json:"otherAddressesOfInterest"`
}

type AppSettings struct {
	DbsPath                  string `json:"dbsPath"`
	LogsPath                 string `json:"logsPath"`
	MaxBridgingClaimsToGroup int    `json:"maxBridgingClaimsToGroup"`
	LogLevel                 int32  `json:"logLevel"`
}

type BridgingSettings struct {
	MinFeeForBridging uint64 `json:"minFeeForBridging"`
	UtxoMinValue      uint64 `json:"utxoMinValue"`
}

type AppConfig struct {
	CardanoChains    map[string]CardanoChainConfig `json:"cardanoChains"`
	Settings         AppSettings                   `json:"appSettings"`
	BridgingSettings BridgingSettings              `json:"bridgingSettings"`
}

type InitialUtxos map[string][]*indexer.TxInputOutput
