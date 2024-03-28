package core

import (
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
)

type BridgingAddresses struct {
	BridgingAddress string `json:"address"`
	FeeAddress      string `json:"feeAddress"`
}

type CardanoChainConfig struct {
	ChainId                  string
	NetworkAddress           string            `json:"networkAddress"`
	NetworkMagic             string            `json:"networkMagic"`
	StartBlockHash           string            `json:"startBlockHash"`
	StartSlot                string            `json:"startSlot"`
	StartBlockNumber         string            `json:"startBlockNumber"`
	ConfirmationBlockCount   uint              `json:"confirmationBlockCount"`
	BridgingAddresses        BridgingAddresses `json:"bridgingAddresses"`
	OtherAddressesOfInterest []string          `json:"otherAddressesOfInterest"`
}

type SubmitConfig struct {
	ConfirmedBlocksThreshhold int `json:"confirmedBlocksThreshhold"`
	ConfirmedBlocksSubmitTime int `json:"confirmedBlocksSubmitTime"`
}

type BridgeConfig struct {
	NodeUrl              string       `json:"nodeUrl"`
	SmartContractAddress string       `json:"smartContractAddress"`
	SigningKey           string       `json:"signingKey"`
	SubmitConfig         SubmitConfig `json:"submitConfig"`
}

type AppSettings struct {
	DbsPath                  string `json:"dbsPath"`
	LogsPath                 string `json:"logsPath"`
	MaxBridgingClaimsToGroup int    `json:"maxBridgingClaimsToGroup"`
	LogLevel                 int32  `json:"logLevel"`
}

type BridgingSettings struct {
	MinFeeForBridging              uint64 `json:"minFeeForBridging"`
	UtxoMinValue                   uint64 `json:"utxoMinValue"`
	MaxReceiversPerBridgingRequest int    `json:"maxReceiversPerBridgingRequest"`
}

type AppConfig struct {
	CardanoChains    map[string]*CardanoChainConfig `json:"cardanoChains"`
	Bridge           BridgeConfig                   `json:"bridge"`
	Settings         AppSettings                    `json:"appSettings"`
	BridgingSettings BridgingSettings               `json:"bridgingSettings"`
}

type InitialUtxos map[string][]*indexer.TxInputOutput

func (appConfig *AppConfig) FillOut() {
	for chainId, cardanoChainConfig := range appConfig.CardanoChains {
		cardanoChainConfig.ChainId = chainId
	}
}
