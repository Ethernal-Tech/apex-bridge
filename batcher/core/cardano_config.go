package core

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
