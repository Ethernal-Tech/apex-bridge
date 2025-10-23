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
	chainStrToInt = map[string]chainIDNum{
		ChainIDStrPrime:   ChainIDIntPrime,
		ChainIDStrVector:  ChainIDIntVector,
		ChainIDStrNexus:   ChainIDIntNexus,
		ChainIDStrCardano: ChainIDIntCardano,
	}
	chainIntToStr = map[chainIDNum]string{
		ChainIDIntPrime:   ChainIDStrPrime,
		ChainIDIntVector:  ChainIDStrVector,
		ChainIDIntNexus:   ChainIDStrNexus,
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
	for _, chainID := range reactorChains {
		if chainID == chainIDStr {
			return true
		}
	}

	return false
}

func IsExistingSkylineChainID(chainIDStr string) bool {
	for _, chainID := range skylineChains {
		if chainID == chainIDStr {
			return true
		}
	}

	return false
}

func IsEVMChainID(chainIDStr string) bool {
	return chainIDStr == ChainIDStrNexus
}

func IsEqual(a, b, errorMargin float64) bool {
	diff := a - b

	return diff >= -1*errorMargin && diff <= errorMargin
}
