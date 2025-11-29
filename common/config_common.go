package common

type EcosystemToken struct {
	ID   uint16 `json:"id"`
	Name string `json:"name"`
}

type DirectionConfigFile struct {
	Directions      map[string]DirectionConfig `json:"directions"`
	EcosystemTokens []EcosystemToken           `json:"ecosystemTokens"`
}

type DirectionConfig struct {
	DestinationChain map[string]TokenPairs `json:"destChain"`
	Tokens           map[uint16]Token      `json:"tokens"`
}

type TokenPairs = []TokenPair

type TokenPair struct {
	SourceTokenID         uint16 `json:"srcTokenID"`
	DestinationTokenID    uint16 `json:"dstTokenID"`
	TrackSourceToken      bool   `json:"trackSource"`
	TrackDestinationToken bool   `json:"trackDestination"`
}

type Token struct {
	ChainSpecific     string `json:"chainSpecific"`
	LockUnlock        bool   `json:"lockUnlock"`
	IsWrappedCurrency bool   `json:"isWrappedCurrency"`
}

type MinConfig struct {
	MinOperationFee            uint64
	MinFeeForBridging          uint64
	MinUtxoAmount              uint64
	MinColCoinsAllowedToBridge uint64
}

const (
	MinFeeForBridgingDefault          = uint64(1_000_010)
	MinUtxoAmountDefault              = uint64(1_000_000)
	MinColCoinsAllowedToBridgeDefault = uint64(1)
)

var (
	ChainMinConfig = map[string]MinConfig{
		ChainIDStrPrime: {
			MinOperationFee:            uint64(0),
			MinFeeForBridging:          uint64(1_000_010),
			MinUtxoAmount:              uint64(1_000_000),
			MinColCoinsAllowedToBridge: uint64(1),
		},
		ChainIDStrCardano: {
			MinOperationFee:            uint64(0),
			MinFeeForBridging:          uint64(1_000_010),
			MinUtxoAmount:              uint64(1_000_000),
			MinColCoinsAllowedToBridge: uint64(1),
		},
		ChainIDStrVector: {
			MinOperationFee:            uint64(0),
			MinFeeForBridging:          uint64(1_000_010),
			MinUtxoAmount:              uint64(1_000_000),
			MinColCoinsAllowedToBridge: uint64(1),
		},
		"default": {
			MinOperationFee:            uint64(0),
			MinFeeForBridging:          MinFeeForBridgingDefault,
			MinUtxoAmount:              MinUtxoAmountDefault,
			MinColCoinsAllowedToBridge: MinColCoinsAllowedToBridgeDefault,
		},
	}
)

func GetChainConfig(chainID string) MinConfig {
	if cfg, ok := ChainMinConfig[chainID]; ok {
		return cfg
	}

	return ChainMinConfig["default"]
}
