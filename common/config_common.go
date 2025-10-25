package common

type MinConfig struct {
	MinOperationFee   uint64
	MinFeeForBridging uint64
	MinUtxoAmount     uint64
}

const (
	MinFeeForBridgingDefault = uint64(1_000_010)
	MinUtxoAmountDefault     = uint64(1_000_000)
)

var (
	ChainMinConfig = map[string]MinConfig{
		ChainIDStrPrime: {
			MinOperationFee:   uint64(0),
			MinFeeForBridging: uint64(1_000_010),
			MinUtxoAmount:     uint64(1_000_000),
		},
		ChainIDStrCardano: {
			MinOperationFee:   uint64(0),
			MinFeeForBridging: uint64(1_000_010),
			MinUtxoAmount:     uint64(1_000_000),
		},
		ChainIDStrVector: {
			MinOperationFee:   uint64(0),
			MinFeeForBridging: uint64(1_000_010),
			MinUtxoAmount:     uint64(1_000_000),
		},
		"default": {
			MinOperationFee:   uint64(0),
			MinFeeForBridging: MinFeeForBridgingDefault,
			MinUtxoAmount:     MinUtxoAmountDefault,
		},
	}
)

func GetChainConfig(chainID string) MinConfig {
	if cfg, ok := ChainMinConfig[chainID]; ok {
		return cfg
	}

	return ChainMinConfig["default"]
}
