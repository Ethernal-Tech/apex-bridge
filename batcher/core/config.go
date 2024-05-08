package core

import (
	"encoding/json"
)

type BridgeConfig struct {
	NodeURL              string `json:"nodeUrl"`
	DynamicTx            bool   `json:"dynamicTx"`
	SmartContractAddress string `json:"scAddress"`
	ValidatorDataDir     string `json:"validatorDataDir"`
	ValidatorConfigPath  string `json:"validatorConfigPath"`
}

type BatcherConfiguration struct {
	Bridge        BridgeConfig `json:"bridge"`
	Chain         ChainConfig  `json:"chain"`
	PullTimeMilis uint64       `json:"pullTime"`
}

type ChainConfig struct {
	ChainID       string          `json:"id"`
	ChainType     string          `json:"type"`
	ChainSpecific json.RawMessage `json:"config"`
}

type BatcherManagerConfiguration struct {
	Bridge        BridgeConfig  `json:"bridge"`
	Chains        []ChainConfig `json:"chains"`
	PullTimeMilis uint64        `json:"pullTime"`
}
