package common

import "slices"

type ChainIDNum = uint8

const (
	ChainTypeCardano = iota
	ChainTypeEVM

	ChainTypeCardanoStr = "cardano"
	ChainTypeEVMStr     = "evm"

	// Used for tests only
	ChainIDStrPrime   = "prime"
	ChainIDStrVector  = "vector"
	ChainIDStrCardano = "cardano"
	ChainIDStrNexus   = "nexus"
	ChainIDStrPolygon = "polygon"

	// Used for tests only
	ChainIDIntPrime   = ChainIDNum(1)
	ChainIDIntVector  = ChainIDNum(2)
	ChainIDIntNexus   = ChainIDNum(3)
	ChainIDIntCardano = ChainIDNum(4)
	ChainIDIntPolygon = ChainIDNum(5)
)

type ChainIDConverter struct {
	StrToInt      map[string]ChainIDNum
	IntToStr      map[ChainIDNum]string
	CardanoChains []string
	EvmChains     []string
}

func (c *ChainIDConverter) ToChainIDNum(chainIDStr string) ChainIDNum {
	return c.StrToInt[chainIDStr]
}

func (c *ChainIDConverter) ToChainIDStr(chainIDNum ChainIDNum) string {
	return c.IntToStr[chainIDNum]
}

func (c *ChainIDConverter) IsExistingChainID(chainIDStr string) bool {
	_, ok := c.StrToInt[chainIDStr]

	return ok
}

func (c *ChainIDConverter) IsCardanoChainID(chainIDStr string) bool {
	return slices.Contains(c.CardanoChains, chainIDStr)
}

func (c *ChainIDConverter) IsEVMChainID(chainIDStr string) bool {
	return slices.Contains(c.EvmChains, chainIDStr)
}
