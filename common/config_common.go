package common

import (
	"math/big"
)

type EcosystemToken struct {
	ID   uint16 `json:"id"`
	Name string `json:"name"`
}

type DirectionConfigFile struct {
	Directions      map[string]DirectionConfig `json:"directions"`
	EcosystemTokens []EcosystemToken           `json:"ecosystemTokens"`
}

type DirectionConfig struct {
	AlwaysTrackCurrencyAndWrappedCurrency bool                  `json:"alwaysTrackCurrencyAndWrappedCurrency"`
	DestinationChains                     map[string]TokenPairs `json:"destChain"`
	Tokens                                map[uint16]Token      `json:"tokens"`
}

type ChainIDsConfigFile struct {
	ChainIDConfig []ChainIDConfig `json:"chainIDs"`
}

type ChainIDConfig struct {
	ChainID    string     `json:"chainID"`
	ChainIDNum ChainIDNum `json:"chainIDNum"`
	ChainType  string     `json:"chainType,omitempty"`
}

func (c *ChainIDsConfigFile) ToChainIDConverter() *ChainIDConverter {
	intToStr := make(map[ChainIDNum]string, len(c.ChainIDConfig))
	strToInt := make(map[string]ChainIDNum, len(c.ChainIDConfig))
	cardanoChains := make([]string, 0)
	evmChains := make([]string, 0)

	for _, chainIDConfig := range c.ChainIDConfig {
		intToStr[chainIDConfig.ChainIDNum] = chainIDConfig.ChainID
		strToInt[chainIDConfig.ChainID] = chainIDConfig.ChainIDNum

		switch chainIDConfig.ChainType {
		case ChainTypeCardanoStr:
			cardanoChains = append(cardanoChains, chainIDConfig.ChainID)
		case ChainTypeEVMStr:
			evmChains = append(evmChains, chainIDConfig.ChainID)
		}
	}

	return &ChainIDConverter{
		StrToInt:      strToInt,
		IntToStr:      intToStr,
		CardanoChains: cardanoChains,
		EvmChains:     evmChains,
	}
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
	MinOperationFee            *big.Int
	MinFeeForBridging          *big.Int
	MinUtxoAmount              uint64
	MinColCoinsAllowedToBridge *big.Int
}

const (
	MinUtxoAmountDefaultDfm              = uint64(1_000_000)
	MinColCoinsAllowedToBridgeDfmCardano = uint64(1) // 1 DFM
)

// vaules in wei
var (
	MinOperationFeeDefault      *big.Int = big.NewInt(0)
	MinFeeForBridgingDefault    *big.Int = DfmToWei(big.NewInt(1_000_010))
	MinAmountAllowedToBridgeEVM *big.Int = big.NewInt(1) // 1 wei
)

var (
	ChainMinConfig = map[string]MinConfig{
		ChainIDStrPrime: {
			MinOperationFee:            big.NewInt(0),
			MinFeeForBridging:          DfmToWei(big.NewInt(1_000_010)),
			MinUtxoAmount:              uint64(1_000_000),
			MinColCoinsAllowedToBridge: DfmToWei(big.NewInt(1)),
		},
		ChainIDStrCardano: {
			MinOperationFee:            big.NewInt(0),
			MinFeeForBridging:          DfmToWei(big.NewInt(1_000_010)),
			MinUtxoAmount:              uint64(1_000_000),
			MinColCoinsAllowedToBridge: DfmToWei(big.NewInt(1)),
		},
		ChainIDStrVector: {
			MinOperationFee:            big.NewInt(0),
			MinFeeForBridging:          DfmToWei(big.NewInt(1_000_010)),
			MinUtxoAmount:              uint64(1_000_000),
			MinColCoinsAllowedToBridge: DfmToWei(big.NewInt(1)),
		},
		ChainIDStrPolygon: {
			MinOperationFee:            big.NewInt(0),
			MinFeeForBridging:          DfmToWei(big.NewInt(264_000)),
			MinUtxoAmount:              MinUtxoAmountDefaultDfm,
			MinColCoinsAllowedToBridge: big.NewInt(1),
		},
		ChainIDStrEthereum: {
			MinOperationFee:            big.NewInt(0),
			MinFeeForBridging:          DfmToWei(big.NewInt(1_240)),
			MinUtxoAmount:              MinUtxoAmountDefaultDfm,
			MinColCoinsAllowedToBridge: big.NewInt(1),
		},
		ChainIDStrKatana: {
			MinOperationFee:            big.NewInt(0),
			MinFeeForBridging:          DfmToWei(big.NewInt(160)),
			MinUtxoAmount:              MinUtxoAmountDefaultDfm,
			MinColCoinsAllowedToBridge: big.NewInt(1),
		},
		ChainIDStrSei: {
			MinOperationFee:            big.NewInt(0),
			MinFeeForBridging:          DfmToWei(big.NewInt(160_000)),
			MinUtxoAmount:              MinUtxoAmountDefaultDfm,
			MinColCoinsAllowedToBridge: big.NewInt(1),
		},
		ChainIDStrArbitrum: {
			MinOperationFee:            big.NewInt(0),
			MinFeeForBridging:          DfmToWei(big.NewInt(60)),
			MinUtxoAmount:              MinUtxoAmountDefaultDfm,
			MinColCoinsAllowedToBridge: big.NewInt(1),
		},
		ChainIDStrScroll: {
			MinOperationFee:            big.NewInt(0),
			MinFeeForBridging:          DfmToWei(big.NewInt(40)),
			MinUtxoAmount:              MinUtxoAmountDefaultDfm,
			MinColCoinsAllowedToBridge: big.NewInt(1),
		},
		ChainIDStrUnichain: {
			MinOperationFee:            big.NewInt(0),
			MinFeeForBridging:          DfmToWei(big.NewInt(1)),
			MinUtxoAmount:              MinUtxoAmountDefaultDfm,
			MinColCoinsAllowedToBridge: big.NewInt(1),
		},
		"default": {
			MinOperationFee:            MinOperationFeeDefault,
			MinFeeForBridging:          MinFeeForBridgingDefault,
			MinUtxoAmount:              MinUtxoAmountDefaultDfm,
			MinColCoinsAllowedToBridge: DfmToWei(new(big.Int).SetUint64(MinColCoinsAllowedToBridgeDfmCardano)),
		},
	}
)

func GetChainConfig(chainID string) MinConfig {
	if cfg, ok := ChainMinConfig[chainID]; ok {
		return cfg
	}

	return ChainMinConfig["default"]
}
