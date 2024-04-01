package core

import (
	"encoding/json"
	"fmt"
)

type CardanoChainConfig struct {
	TestNetMagic      uint    `json:"testnetMagic"`
	BlockfrostUrl     string  `json:"blockfrostUrl"`
	BlockfrostAPIKey  string  `json:"blockfrostApiKey"`
	AtLeastValidators float64 `json:"atLeastValidators"`
	PotentialFee      uint64  `json:"potentialFee"`
}

var _ ChainSpecificConfig = (*CardanoChainConfig)(nil)

// GetChainType implements ChainSpecificConfig.
func (*CardanoChainConfig) GetChainType() string {
	return "Cardano"
}

func ToCardanoChainConfig(config ChainSpecific) (*CardanoChainConfig, error) {
	if config.ChainType != "Cardano" {
		return nil, fmt.Errorf("chain type must be Cardano not: %v", config.ChainType)
	}

	var cardanoChainConfig CardanoChainConfig
	if err := json.Unmarshal(config.Config, &cardanoChainConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Cardano configuration: %v", err)
	}

	return &cardanoChainConfig, nil
}
