package core

import (
	oracleCore "github.com/Ethernal-Tech/apex-bridge/oracle/core"
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

type AppConfig struct {
	ValidatorDataDir    string                      `json:"validatorDataDir"`
	ValidatorConfigPath string                      `json:"validatorConfigPath"`
	EthChains           map[string]*EthChainConfig  `json:"ethChains"`
	Bridge              oracleCore.BridgeConfig     `json:"bridge"`
	Settings            oracleCore.AppSettings      `json:"appSettings"`
	BridgingSettings    oracleCore.BridgingSettings `json:"bridgingSettings"`
}

func (appConfig *AppConfig) FillOut() {
	for chainID, ethChainConfig := range appConfig.EthChains {
		ethChainConfig.ChainID = chainID
	}
}
