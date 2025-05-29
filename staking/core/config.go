package core

import "github.com/Ethernal-Tech/cardano-infrastructure/logger"

type StakingBridgingAddresses struct {
	StakingBridgingAddr string `json:"address"`
	FeeAddress          string `json:"feeAddress"`
}

type ChainConfig struct {
	ChainID             string                   `json:"id"`
	ChainType           string                   `json:"type"`
	StakingAddresses    []string                 `json:"stakingAddresses"`
	StakingBridgingAddr StakingBridgingAddresses `json:"-"`
}

type StakingConfiguration struct {
	Chain         ChainConfig `json:"chain"`
	PullTimeMilis int64       `json:"pullTime"`
}

type StakingManagerConfiguration struct {
	Chains        map[string]ChainConfig `json:"chains"`
	Logger        logger.LoggerConfig    `json:"logger"`
	PullTimeMilis int64                  `json:"pullTime"`
}
