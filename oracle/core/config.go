package core

import (
	"encoding/json"

	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
)

type BridgingAddresses struct {
	BridgingAddress string `json:"address"`
	FeeAddress      string `json:"feeAddress"`
}

type CardanoChainConfig struct {
	ChainID                  string
	NetworkAddress           string                   `json:"networkAddress"`
	NetworkMagic             uint32                   `json:"networkMagic"`
	StartBlockHash           string                   `json:"startBlockHash"`
	StartSlot                uint64                   `json:"startSlot"`
	StartBlockNumber         uint64                   `json:"startBlockNumber"`
	ConfirmationBlockCount   uint                     `json:"confirmationBlockCount"`
	BridgingAddresses        BridgingAddresses        `json:"bridgingAddresses"`
	OtherAddressesOfInterest []string                 `json:"otherAddressesOfInterest"`
	InitialUtxos             []*indexer.TxInputOutput `json:"initialUtxos"`
	ChainSpecific            json.RawMessage          `json:"config"`
}

type SubmitConfig struct {
	ConfirmedBlocksThreshold  int `json:"confirmedBlocksThreshold"`
	ConfirmedBlocksSubmitTime int `json:"confirmedBlocksSubmitTime"`
}

type BridgeConfig struct {
	NodeURL              string       `json:"nodeUrl"`
	DynamicTx            bool         `json:"dynamicTx"`
	SmartContractAddress string       `json:"scAddress"`
	ValidatorDataDir     string       `json:"validatorDataDir"`
	ValidatorConfigPath  string       `json:"validatorConfigPath"`
	SubmitConfig         SubmitConfig `json:"submitConfig"`
}

type AppSettings struct {
	Logger  logger.LoggerConfig `json:"logger"`
	DbsPath string              `json:"dbsPath"`
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
}

func (appConfig *AppConfig) FillOut() {
	for chainID, cardanoChainConfig := range appConfig.CardanoChains {
		cardanoChainConfig.ChainID = chainID
	}
}
