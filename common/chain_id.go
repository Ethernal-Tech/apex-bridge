package common

import "fmt"

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

func IsExistingChainID(chainIDStr string) bool {
	_, exists := strToInt[chainIDStr]

	return exists
}

func IsEVMChainID(chainIDStr string) bool {
	return chainIDStr == ChainIDStrNexus
}

func IsTxDirectionAllowed(srcChainID, destChainID string) error {
	if (srcChainID == ChainIDStrNexus && destChainID == ChainIDStrVector) ||
		(srcChainID == ChainIDStrVector && destChainID == ChainIDStrNexus) {
		return fmt.Errorf("transaction direction not allowed: %s -> %s", srcChainID, destChainID)
	}

	return nil
}
