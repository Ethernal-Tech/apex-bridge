package core

import (
	"encoding/json"
)

type BridgeConfig struct {
	NodeUrl              string `json:"nodeUrl"`
	SmartContractAddress string `json:"scAddress"` // nolintlint TODO: probably will be more than just one
	ValidatorDataDir     string `json:"validatorDataDir"`
	ValidatorConfigPath  string `json:"validatorConfigPath"`
}

type BatcherConfiguration struct {
	Bridge        BridgeConfig `json:"bridge"`
	Chain         ChainConfig  `json:"chain"`
	PullTimeMilis uint64       `json:"pullTime"`
}

type ChainConfig struct {
	ChainId       string          `json:"id"`
	ChainType     string          `json:"type"`
	ChainSpecific json.RawMessage `json:"config"`
}

type BatcherManagerConfiguration struct {
	Bridge        BridgeConfig  `json:"bridge"`
	Chains        []ChainConfig `json:"chains"`
	PullTimeMilis uint64        `json:"pullTime"`
}
