package core

import (
	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
)

type BridgingAddresses struct {
	BridgingAddress string `json:"address"`
	FeeAddress      string `json:"feeAddress"`
}

type EthChainConfig struct {
	ChainID string
}

type SubmitConfig struct {
	ConfirmedBlocksThreshold  int `json:"confirmedBlocksThreshold"`
	ConfirmedBlocksSubmitTime int `json:"confirmedBlocksSubmitTime"`
}

type BridgeConfig struct {
	NodeURL              string       `json:"nodeUrl"`
	DynamicTx            bool         `json:"dynamicTx"`
	SmartContractAddress string       `json:"scAddress"`
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
	ValidatorDataDir    string                     `json:"validatorDataDir"`
	ValidatorConfigPath string                     `json:"validatorConfigPath"`
	EthChains           map[string]*EthChainConfig `json:"ethChains"`
	Bridge              BridgeConfig               `json:"bridge"`
	Settings            AppSettings                `json:"appSettings"`
	BridgingSettings    BridgingSettings           `json:"bridgingSettings"`
}

func (appConfig *AppConfig) FillOut() {
	for chainID, ethChainConfig := range appConfig.EthChains {
		ethChainConfig.ChainID = chainID
	}
}
