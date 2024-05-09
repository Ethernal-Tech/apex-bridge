package core

import (
	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
)

type BridgeClaimsBlockInfo struct {
	ChainID            string
	Slot               uint64
	Hash               string
	BlockFullyObserved bool
}

type ContractClaims = contractbinding.IBridgeStructsValidatorClaims
type BridgingRequestClaim = contractbinding.IBridgeStructsBridgingRequestClaim
type BatchExecutedClaim = contractbinding.IBridgeStructsBatchExecutedClaim
type BatchExecutionFailedClaim = contractbinding.IBridgeStructsBatchExecutionFailedClaim
type RefundRequestClaim = contractbinding.IBridgeStructsRefundRequestClaim
type RefundExecutedClaim = contractbinding.IBridgeStructsRefundExecutedClaim
type UTXO = contractbinding.IBridgeStructsUTXO
type UTXOs = contractbinding.IBridgeStructsUTXOs
type BridgingRequestReceiver = contractbinding.IBridgeStructsReceiver

type BridgeClaims struct {
	ContractClaims
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

func (bc *BridgeClaims) Count() int {
	return len(bc.BridgingRequestClaims) +
		len(bc.BatchExecutedClaims) +
		len(bc.BatchExecutionFailedClaims) /* + len(bc.RefundRequest) + len(bc.RefundExecuted)*/
}

func (bc *BridgeClaims) Any() bool {
	return bc.Count() > 0
}
