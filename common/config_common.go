package common

type EcosystemToken struct {
	ID   uint16 `json:"id"`
	Name string `json:"name"`
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
