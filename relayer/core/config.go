package core

import (
	"encoding/json"

	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
)

type BridgeConfig struct {
	NodeUrl              string `json:"NodeUrl"`
	SmartContractAddress string `json:"scAddress"` // TODO: probably will be more than just one
}

type RelayerConfiguration struct {
	Bridge        BridgeConfig        `json:"bridge"`
	Chain         ChainConfig         `json:"chain"`
	PullTimeMilis uint64              `json:"pullTime"`
	Logger        logger.LoggerConfig `json:"logger"`
}

type ChainConfig struct {
	ChainId       string          `json:"id"`
	ChainType     string          `json:"type"`
	DbsPath       string          `json:"dbsPath"`
	ChainSpecific json.RawMessage `json:"config"`
}

type RelayerManagerConfiguration struct {
	Bridge        BridgeConfig        `json:"bridge"`
	Chains        []ChainConfig       `json:"chains"`
	PullTimeMilis uint64              `json:"pullTime"`
	Logger        logger.LoggerConfig `json:"logger"`
}
