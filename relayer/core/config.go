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
	Base          BaseConfig          `json:"base"`
	PullTimeMilis uint64              `json:"pullTime"`
	Logger        logger.LoggerConfig `json:"logger"`
}

type BaseConfig struct {
	ChainId string `json:"chainId"`
}

type ChainSpecific struct {
	ChainType string          `json:"chainType"`
	Config    json.RawMessage `json:"config"`
}
type ChainConfig struct {
	Base          BaseConfig    `json:"baseConfig"`
	ChainSpecific ChainSpecific `json:"chainSpecific"`
}

type RelayerManagerConfiguration struct {
	Bridge        BridgeConfig           `json:"bridge"`
	Chains        map[string]ChainConfig `json:"chains"`
	PullTimeMilis uint64                 `json:"pullTime"`
	Logger        logger.LoggerConfig    `json:"logger"`
}
