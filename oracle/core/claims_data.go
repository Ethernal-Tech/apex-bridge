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

type ContractClaims = contractbinding.IBridgeContractStructsValidatorClaims
type BridgingRequestClaim = contractbinding.IBridgeContractStructsBridgingRequestClaim
type BatchExecutedClaim = contractbinding.IBridgeContractStructsBatchExecutedClaim
type BatchExecutionFailedClaim = contractbinding.IBridgeContractStructsBatchExecutionFailedClaim
type RefundRequestClaim = contractbinding.IBridgeContractStructsRefundRequestClaim
type RefundExecutedClaim = contractbinding.IBridgeContractStructsRefundExecutedClaim
type UTXO = contractbinding.IBridgeContractStructsUTXO
type UTXOs = contractbinding.IBridgeContractStructsUTXOs
type BridgingRequestReceiver = contractbinding.IBridgeContractStructsReceiver

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
