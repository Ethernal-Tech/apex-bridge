package common

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
	reactorStrToInt = map[string]chainIDNum{
		ChainIDStrPrime:  ChainIDIntPrime,
		ChainIDStrVector: ChainIDIntVector,
		ChainIDStrNexus:  ChainIDIntNexus,
	}
	reactorIntToStr = map[chainIDNum]string{
		ChainIDIntPrime:  ChainIDStrPrime,
		ChainIDIntVector: ChainIDStrVector,
		ChainIDIntNexus:  ChainIDStrNexus,
	}
	skylineStrToInt = map[string]chainIDNum{
		ChainIDStrCardano: ChainIDIntCardano,
		ChainIDStrPrime:   ChainIDIntPrime,
	}
	skylineIntToStr = map[chainIDNum]string{
		ChainIDIntCardano: ChainIDStrCardano,
		ChainIDIntPrime:   ChainIDStrPrime,
	}
)

func ToNumChainID(chainIDStr string) chainIDNum {
	return reactorStrToInt[chainIDStr]
}

func ToStrChainID(chainIDNum chainIDNum) string {
	return reactorIntToStr[chainIDNum]
}

func IsExistingReactorChainID(chainIDStr string) bool {
	_, exists := reactorStrToInt[chainIDStr]

	return exists
}

func IsExistingSkylineChainID(chainIDStr string) bool {
	_, exists := skylineStrToInt[chainIDStr]

	return exists
}

func IsEVMChainID(chainIDStr string) bool {
	return chainIDStr == ChainIDStrNexus
}

func IsEqual(a, b, errorMargin float64) bool {
	diff := a - b

	return diff >= -1*errorMargin && diff <= errorMargin
}
