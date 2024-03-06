package relayer

import "github.com/Ethernal-Tech/cardano-infrastructure/logger"

type CardanoConfig struct {
	TestNetMagic      uint    `json:"testnetMagic" yaml:"testnetMagic"`
	BlockfrostUrl     string  `json:"blockfrostUrl" yaml:"blockfrostUrl"`
	BlockfrostAPIKey  string  `json:"blockfrostApiKey" yaml:"blockfrostApiKey"`
	AtLeastValidators float64 `json:"atLeastValidators" yaml:"atLeastValidators"`
	PotentialFee      uint64  `json:"potentialFee" yaml:"potentialFee"`
}

type BridgeConfig struct {
	NodeUrl              string `json:"NodeUrl" yaml:"NodeUrl"`
	SmartContractAddress string `json:"scAddress" yaml:"scAddress"` // TOOD: probably will be more than just one
}

type RelayerConfiguration struct {
	Bridge        BridgeConfig        `json:"bridge" yaml:"bridge"`
	Cardano       CardanoConfig       `json:"cardano" yaml:"cardano"`
	PullTimeMilis uint64              `json:"pullTime" yaml:"pullTime"`
	Logger        logger.LoggerConfig `json:"logger" yaml:"logger"`
}
