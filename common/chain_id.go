package common

type chainIDNum = uint8

const (
	ChainTypeCardano = iota
	ChainTypeEVM     = 1

	ChainIDStrPrime  = "prime"
	ChainIDStrVector = "vector"

	ChainIDIntPrime  = chainIDNum(1)
	ChainIDIntVector = chainIDNum(2)
)

var (
	strToInt = map[string]chainIDNum{
		ChainIDStrPrime:  ChainIDIntPrime,
		ChainIDStrVector: ChainIDIntVector,
	}
	intToStr = map[chainIDNum]string{
		ChainIDIntPrime:  ChainIDStrPrime,
		ChainIDIntVector: ChainIDStrVector,
	}
)

func ToNumChainID(chainIDStr string) chainIDNum {
	return strToInt[chainIDStr]
}

func ToStrChainID(chainIDNum chainIDNum) string {
	return intToStr[chainIDNum]
}
