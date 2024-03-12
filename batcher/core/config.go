package core

import "github.com/Ethernal-Tech/cardano-infrastructure/logger"

type CardanoChainConfig struct {
	ChainId               string  `json:"chainId"`
	TestNetMagic          uint    `json:"testnetMagic"`
	BlockfrostUrl         string  `json:"blockfrostUrl"`
	BlockfrostAPIKey      string  `json:"blockfrostApiKey"`
	AtLeastValidators     float64 `json:"atLeastValidators"`
	PotentialFee          uint64  `json:"potentialFee"`
	SigningKeyMultiSig    string  `json:"signingKey"`    // hex+cbor representation of private key
	SigningKeyMultiSigFee string  `json:"signingKeyFee"` // hex+cbor representation of private key
}

type BridgeConfig struct {
	NodeUrl              string `json:"NodeUrl"`
	SmartContractAddress string `json:"scAddress"`  // TOOD: probably will be more than just one
	SigningKey           string `json:"signingKey"` // hex representation of private signing key
}

type BatcherConfiguration struct {
	Bridge        BridgeConfig        `json:"bridge"`
	CardanoChain  CardanoChainConfig  `json:"cardanoChain"`
	PullTimeMilis uint64              `json:"pullTime"`
	Logger        logger.LoggerConfig `json:"logger"`
}

type BatcherManagerConfiguration struct {
	Bridge        BridgeConfig                  `json:"bridge"`
	CardanoChains map[string]CardanoChainConfig `json:"cardanoChains"`
	PullTimeMilis uint64                        `json:"pullTime"`
	Logger        logger.LoggerConfig           `json:"logger"`
}
