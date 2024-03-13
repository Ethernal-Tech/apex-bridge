package core

import "github.com/Ethernal-Tech/cardano-infrastructure/logger"

type CardanoChainConfig struct {
	ChainId           string  `json:"chainId"`
	TestNetMagic      uint    `json:"testnetMagic"`
	BlockfrostUrl     string  `json:"blockfrostUrl"`
	BlockfrostAPIKey  string  `json:"blockfrostApiKey"`
	AtLeastValidators float64 `json:"atLeastValidators"`
	PotentialFee      uint64  `json:"potentialFee"`
}

type BridgeConfig struct {
	NodeUrl              string `json:"NodeUrl"`
	SmartContractAddress string `json:"scAddress"` // TODO: probably will be more than just one
}

type RelayerConfiguration struct {
	Bridge        BridgeConfig        `json:"bridge"`
	CardanoChain  CardanoChainConfig  `json:"cardanoChain"`
	PullTimeMilis uint64              `json:"pullTime"`
	Logger        logger.LoggerConfig `json:"logger"`
}

type RelayerManagerConfiguration struct {
	Bridge        BridgeConfig                  `json:"bridge"`
	CardanoChains map[string]CardanoChainConfig `json:"cardanoChains"`
	PullTimeMilis uint64                        `json:"pullTime"`
	Logger        logger.LoggerConfig           `json:"logger"`
}
