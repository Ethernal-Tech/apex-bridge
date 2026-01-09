package core

import (
	"encoding/json"

	"github.com/Ethernal-Tech/apex-bridge/common"
)

type BatcherConfiguration struct {
	Chain            ChainConfig              `json:"chain"`
	ChainIDConverter *common.ChainIDConverter `json:"-"`
	PullTimeMilis    uint64                   `json:"pullTime"`
}

type ChainConfig struct {
	ChainID       string          `json:"id"`
	ChainType     string          `json:"type"`
	ChainSpecific json.RawMessage `json:"config"`
}

type BatcherManagerConfiguration struct {
	Chains           []ChainConfig            `json:"chains"`
	ChainIDConverter *common.ChainIDConverter `json:"-"`
	PullTimeMilis    uint64                   `json:"pullTime"`
}
