package relayer

import "github.com/Ethernal-Tech/cardano-infrastructure/logger"

type CardanoConfig struct {
	TestNetMagic      uint    `json:"testnetMagic"`
	BlockfrostUrl     string  `json:"blockfrostUrl"`
	BlockfrostAPIKey  string  `json:"blockfrostApiKey"`
	AtLeastValidators float64 `json:"atLeastValidators"`
	PotentialFee      uint64  `json:"potentialFee"`
}

type BridgeConfig struct {
	NodeUrl              string `json:"NodeUrl"`
	SmartContractAddress string `json:"scAddress"` // TOOD: probably will be more than just one
}

type RelayerConfiguration struct {
	Bridge        BridgeConfig        `json:"bridge"`
	Cardano       CardanoConfig       `json:"cardano"`
	PullTimeMilis uint64              `json:"pullTime"`
	Logger        logger.LoggerConfig `json:"logger"`
}
