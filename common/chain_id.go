package common

type chainIDNum = uint8

const (
	ChainTypeCardano = iota
	ChainTypeEVM

	ChainIDStrPrime  = "prime"
	ChainIDStrVector = "vector"
	ChainIDStrNexus  = "nexus"

	ChainIDIntPrime  = chainIDNum(1)
	ChainIDIntVector = chainIDNum(2)
	ChainIDIntNexus  = chainIDNum(3)

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
