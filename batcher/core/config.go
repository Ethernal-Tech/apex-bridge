package core

import (
	"encoding/json"

	"github.com/Ethernal-Tech/cardano-infrastructure/logger"
)

type BridgeConfig struct {
	NodeUrl              string `json:"NodeUrl"`
	SmartContractAddress string `json:"scAddress"`  // TOOD: probably will be more than just one
	SigningKey           string `json:"signingKey"` // hex representation of private signing key
}

type BatcherConfiguration struct {
	Bridge        BridgeConfig        `json:"bridge"`
	Base          BaseConfig          `json:"base"`
	PullTimeMilis uint64              `json:"pullTime"`
	Logger        logger.LoggerConfig `json:"logger"`
}

type BaseConfig struct {
	ChainId     string `json:"chainId"`
	KeysDirPath string `json:"keysDirPath"`
}
type ChainSpecific struct {
	ChainType string          `json:"chainType"`
	Config    json.RawMessage `json:"config"`
}
type ChainConfig struct {
	Base          BaseConfig    `json:"baseConfig"`
	ChainSpecific ChainSpecific `json:"chainSpecific"`
}

type BatcherManagerConfiguration struct {
	Bridge        BridgeConfig           `json:"bridge"`
	Chains        map[string]ChainConfig `json:"chains"`
	PullTimeMilis uint64                 `json:"pullTime"`
	Logger        logger.LoggerConfig    `json:"logger"`
}
