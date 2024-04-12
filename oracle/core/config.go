package core

import (
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/secrets"
)

type BridgingAddresses struct {
	BridgingAddress string `json:"address"`
	FeeAddress      string `json:"feeAddress"`
}

type CardanoChainConfig struct {
	ChainId                  string
	NetworkAddress           string            `json:"networkAddress"`
	NetworkMagic             uint32            `json:"networkMagic"`
	StartBlockHash           string            `json:"startBlockHash"`
	StartSlot                uint64            `json:"startSlot"`
	StartBlockNumber         uint64            `json:"startBlockNumber"`
	ConfirmationBlockCount   uint              `json:"confirmationBlockCount"`
	BridgingAddresses        BridgingAddresses `json:"bridgingAddresses"`
	OtherAddressesOfInterest []string          `json:"otherAddressesOfInterest"`
}

type SubmitConfig struct {
	ConfirmedBlocksThreshold  int `json:"confirmedBlocksThreshold"`
	ConfirmedBlocksSubmitTime int `json:"confirmedBlocksSubmitTime"`
}

type BridgeConfig struct {
	NodeUrl              string                        `json:"nodeUrl"`
	SmartContractAddress string                        `json:"scAddress"`
	SecretsManager       *secrets.SecretsManagerConfig `json:"secrets"`
	SubmitConfig         SubmitConfig                  `json:"submitConfig"`
}

type AppSettings struct {
	DbsPath  string `json:"dbsPath"`
	LogsPath string `json:"logsPath"`
	LogLevel int32  `json:"logLevel"`
}

type BridgingSettings struct {
	MinFeeForBridging              uint64 `json:"minFeeForBridging"`
	UtxoMinValue                   uint64 `json:"utxoMinValue"`
	MaxReceiversPerBridgingRequest int    `json:"maxReceiversPerBridgingRequest"`
	MaxBridgingClaimsToGroup       int    `json:"maxBridgingClaimsToGroup"`
}

type AppConfig struct {
	CardanoChains    map[string]*CardanoChainConfig `json:"cardanoChains"`
	Bridge           BridgeConfig                   `json:"bridge"`
	Settings         AppSettings                    `json:"appSettings"`
	BridgingSettings BridgingSettings               `json:"bridgingSettings"`
	InitialUtxos     InitialUtxos                   `json:"initialUtxos"`
}

type InitialUtxos map[string][]*indexer.TxInputOutput

func (appConfig *AppConfig) FillOut() {
	for chainId, cardanoChainConfig := range appConfig.CardanoChains {
		cardanoChainConfig.ChainId = chainId
	}
}
