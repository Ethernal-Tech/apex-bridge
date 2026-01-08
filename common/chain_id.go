package common

import "slices"

type chainIDNum = uint8

const (
	ChainTypeCardano = iota
	ChainTypeEVM

	ChainIDStrPrime   = "prime"
	ChainIDStrVector  = "vector"
	ChainIDStrCardano = "cardano"
	ChainIDStrNexus   = "nexus"
	ChainIDStrPolygon = "polygon"

	ChainIDIntPrime   = chainIDNum(1)
	ChainIDIntVector  = chainIDNum(2)
	ChainIDIntNexus   = chainIDNum(3)
	ChainIDIntCardano = chainIDNum(4)
	ChainIDIntPolygon = chainIDNum(5)

	ChainTypeCardanoStr = "cardano"
	ChainTypeEVMStr     = "evm"
)

var (
	chainStrToInt = map[string]chainIDNum{
		ChainIDStrPrime:   ChainIDIntPrime,
		ChainIDStrVector:  ChainIDIntVector,
		ChainIDStrNexus:   ChainIDIntNexus,
		ChainIDStrPolygon: ChainIDIntPolygon,
		ChainIDStrCardano: ChainIDIntCardano,
	}
	chainIntToStr = map[chainIDNum]string{
		ChainIDIntPrime:   ChainIDStrPrime,
		ChainIDIntVector:  ChainIDStrVector,
		ChainIDIntNexus:   ChainIDStrNexus,
		ChainIDIntPolygon: ChainIDStrPolygon,
		ChainIDIntCardano: ChainIDStrCardano,
	}

	reactorChains = []string{
		ChainIDStrPrime,
		ChainIDStrVector,
		ChainIDStrNexus,
	}

	skylineChains = []string{
		ChainIDStrPrime,
		ChainIDStrCardano,
		ChainIDStrVector,
		ChainIDStrNexus,
		ChainIDStrPolygon,
	}
)

func ToNumChainID(chainIDStr string) chainIDNum {
	return chainStrToInt[chainIDStr]
}

func ToStrChainID(chainIDNum chainIDNum) string {
	return chainIntToStr[chainIDNum]
}

func IsExistingChainID(chainIDStr string) bool {
	return IsExistingReactorChainID(chainIDStr) || IsExistingSkylineChainID(chainIDStr)
}

func IsExistingReactorChainID(chainIDStr string) bool {
	return slices.Contains(reactorChains, chainIDStr)
}

func IsExistingSkylineChainID(chainIDStr string) bool {
	return slices.Contains(skylineChains, chainIDStr)
}

func IsEVMChainID(chainIDStr string) bool {
	return chainIDStr == ChainIDStrNexus || chainIDStr == ChainIDStrPolygon
}

func IsEqual(a, b, errorMargin float64) bool {
	diff := a - b

	return diff >= -1*errorMargin && diff <= errorMargin
}
