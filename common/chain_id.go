package common

import "math"

type chainIDNum = uint8

const (
	ChainTypeCardano = iota
	ChainTypeEVM

	ChainIDStrPrime   = "prime"
	ChainIDStrVector  = "vector"
	ChainIDStrNexus   = "nexus"
	ChainIDStrCardano = "cardano"

	ChainIDIntPrime   = chainIDNum(1)
	ChainIDIntVector  = chainIDNum(2)
	ChainIDIntNexus   = chainIDNum(3)
	ChainIDIntCardano = chainIDNum(4)

	ChainTypeCardanoStr = "cardano"
	ChainTypeEVMStr     = "evm"
)

var (
	strToInt = map[string]chainIDNum{
		ChainIDStrPrime:  ChainIDIntPrime,
		ChainIDStrVector: ChainIDIntVector,
		ChainIDStrNexus:  ChainIDIntNexus,
	}
	intToStr = map[chainIDNum]string{
		ChainIDIntPrime:  ChainIDStrPrime,
		ChainIDIntVector: ChainIDStrVector,
		ChainIDIntNexus:  ChainIDStrNexus,
	}
)

func ToNumChainID(chainIDStr string) chainIDNum {
	return strToInt[chainIDStr]
}

func ToStrChainID(chainIDNum chainIDNum) string {
	return intToStr[chainIDNum]
}

func IsExistingChainID(chainIDStr string) bool {
	_, exists := strToInt[chainIDStr]

	return exists
}

func IsEVMChainID(chainIDStr string) bool {
	return chainIDStr == ChainIDStrNexus
}

func IsEqual(a, b, errorMargin float64) bool {
	return math.Abs(a-b) <= errorMargin
}
