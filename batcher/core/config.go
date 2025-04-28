package core

import (
	"encoding/json"
)

type BatcherConfiguration struct {
	Chain         ChainConfig `json:"chain"`
	PullTimeMilis uint64      `json:"pullTime"`
}

type ChainConfig struct {
	ChainID       string          `json:"id"`
	ChainType     string          `json:"type"`
	ChainSpecific json.RawMessage `json:"config"`
}

type BatcherManagerConfiguration struct {
	Chains        []ChainConfig `json:"chains"`
	PullTimeMilis uint64        `json:"pullTime"`
}
