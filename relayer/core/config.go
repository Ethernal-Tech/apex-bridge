package core

import (
	"encoding/json"

	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
)

type BridgeConfig struct {
	NodeURL              string `json:"NodeUrl"`
	DynamicTx            bool   `json:"dynamicTx"`
	SmartContractAddress string `json:"scAddress"`
}

type RelayerConfiguration struct {
	Bridge        BridgeConfig        `json:"bridge"`
	Chain         ChainConfig         `json:"chain"`
	PullTimeMilis uint64              `json:"pullTime"`
	Logger        logger.LoggerConfig `json:"logger"`
}

type ChainConfig struct {
	ChainID           string          `json:"id,omitempty"`
	ChainType         string          `json:"type"`
	DbsPath           string          `json:"dbsPath"`
	ChainSpecific     json.RawMessage `json:"config"`
	RelayerDataDir    string          `json:"relayerDataDir"`
	RelayerConfigPath string          `json:"relayerConfigPath"`
}

type RelayerManagerConfiguration struct {
	Bridge        BridgeConfig           `json:"bridge"`
	Chains        map[string]ChainConfig `json:"chains"`
	PullTimeMilis uint64                 `json:"pullTime"`
	Logger        logger.LoggerConfig    `json:"logger"`
}
