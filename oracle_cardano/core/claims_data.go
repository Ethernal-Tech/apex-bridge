package core

import "github.com/Ethernal-Tech/cardano-infrastructure/indexer"

type BridgeClaimsBlockInfo struct {
	ChainID string
	Slot    uint64
	Hash    indexer.Hash
}

func (bi *BridgeClaimsBlockInfo) EqualWithUnprocessed(tx *CardanoTx) bool {
	return bi.ChainID == tx.OriginChainID && bi.Slot == tx.BlockSlot && bi.Hash == tx.BlockHash
}

func (bi *BridgeClaimsBlockInfo) EqualWithProcessed(tx *ProcessedCardanoTx) bool {
	return bi.ChainID == tx.OriginChainID && bi.Slot == tx.BlockSlot && bi.Hash == tx.BlockHash
}

func (bi *BridgeClaimsBlockInfo) EqualWithExpected(tx *BridgeExpectedCardanoTx, block *indexer.CardanoBlock) bool {
	return bi.ChainID == tx.ChainID && bi.Slot == block.Slot && bi.Hash == block.Hash
}
