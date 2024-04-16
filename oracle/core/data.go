package core

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
)

type TxProcessorType string

const (
	TxProcessorTypeBridgingRequest = "BridgingRequest"
	TxProcessorTypeBatchExecuted   = "BatchExecuted"
	TxProcessorTypeRefundExecuted  = "RefundExecuted"
)

type CardanoTx struct {
	OriginChainId string `json:"origin_chain_id"`

	indexer.Tx
}

type ProcessedCardanoTx struct {
	BlockSlot     uint64 `json:"block_slot"`
	BlockHash     string `json:"block_hash"`
	OriginChainId string `json:"origin_chain_id"`
	Hash          string `json:"hash"`
	IsInvalid     bool   `json:"is_invalid"`
}

type BridgeExpectedCardanoTx struct {
	ChainId  string `json:"chain_id"`
	Hash     string `json:"hash"`
	Metadata []byte `json:"metadata"`
	Ttl      uint64 `json:"ttl"`
}

type BridgeExpectedCardanoDbTx struct {
	BridgeExpectedCardanoTx

	IsProcessed bool `json:"is_processed"`
	IsInvalid   bool `json:"is_invalid"`
}

type ContractCardanoBlock = contractbinding.IBridgeContractStructsCardanoBlock

func (tx *CardanoTx) ToProcessedCardanoTx(isInvalid bool) *ProcessedCardanoTx {
	return &ProcessedCardanoTx{
		BlockSlot:     tx.BlockSlot,
		BlockHash:     tx.BlockHash,
		OriginChainId: tx.OriginChainId,
		Hash:          tx.Hash,
		IsInvalid:     isInvalid,
	}
}

func ToUnprocessedTxKey(blockSlot uint64, originChainId string, txHash string) string {
	return fmt.Sprintf("%20d_%v_%v", blockSlot, originChainId, txHash)
}

func (tx CardanoTx) ToUnprocessedTxKey() string {
	return ToUnprocessedTxKey(tx.BlockSlot, tx.OriginChainId, tx.Hash)
}

func (tx ProcessedCardanoTx) ToUnprocessedTxKey() string {
	return ToUnprocessedTxKey(tx.BlockSlot, tx.OriginChainId, tx.Hash)
}

func ToCardanoTxKey(originChainId string, txHash string) string {
	return fmt.Sprintf("%v_%v", originChainId, txHash)
}

func (tx CardanoTx) ToCardanoTxKey() string {
	return ToCardanoTxKey(tx.OriginChainId, tx.Hash)
}

func (tx ProcessedCardanoTx) ToCardanoTxKey() string {
	return ToCardanoTxKey(tx.OriginChainId, tx.Hash)
}

func (tx CardanoTx) Key() []byte {
	return []byte(tx.ToCardanoTxKey())
}

func (tx ProcessedCardanoTx) Key() []byte {
	return []byte(tx.ToCardanoTxKey())
}

func (tx BridgeExpectedCardanoTx) ToCardanoTxKey() string {
	return ToCardanoTxKey(tx.ChainId, tx.Hash)
}

func (tx BridgeExpectedCardanoTx) ToExpectedTxKey() string {
	return fmt.Sprintf("%20d_%v_%v", tx.Ttl, tx.ChainId, tx.Hash)
}

func (tx BridgeExpectedCardanoTx) Key() []byte {
	return []byte(tx.ToExpectedTxKey())
}

func (tx BridgeExpectedCardanoDbTx) Key() []byte {
	return []byte(tx.ToExpectedTxKey())
}
