package common

import "slices"

type ChainIDNum = uint8

const (
	ChainTypeCardano = iota
	ChainTypeEVM
	ChainTypeSolana

	ChainTypeCardanoStr = "cardano"
	ChainTypeEVMStr     = "evm"
	ChainTypeSolanaStr  = "solana"

	// Used for tests only
	ChainIDStrPrime    = "prime"
	ChainIDStrVector   = "vector"
	ChainIDStrCardano  = "cardano"
	ChainIDStrNexus    = "nexus"
	ChainIDStrPolygon  = "polygon"
	ChainIDStrEthereum = "ethereum"
	ChainIDStrKatana   = "katana"
	ChainIDStrSei      = "sei"
	ChainIDStrArbitrum = "arbitrum"
	ChainIDStrScroll   = "scroll"
	ChainIDStrUnichain = "unichain"
	ChainIDStrSolana   = "solana"

	// Used for tests only
	ChainIDIntPrime    = ChainIDNum(1)
	ChainIDIntVector   = ChainIDNum(2)
	ChainIDIntNexus    = ChainIDNum(3)
	ChainIDIntCardano  = ChainIDNum(4)
	ChainIDIntPolygon  = ChainIDNum(5)
	ChainIDIntEthereum = ChainIDNum(6)
	ChainIDIntKatana   = ChainIDNum(7)
	ChainIDIntSei      = ChainIDNum(8)
	ChainIDIntArbitrum = ChainIDNum(9)
	ChainIDIntScroll   = ChainIDNum(10)
	ChainIDIntUnichain = ChainIDNum(11)
	ChainIDIntSolana   = ChainIDNum(12)
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

func (c *ChainIDConverter) IsSolanaChainID(chainIDStr string) bool {
	return ChainIDStrSolana == chainIDStr
}
