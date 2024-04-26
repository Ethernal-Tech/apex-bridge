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
	OriginChainID string `json:"origin_chain_id"`

	indexer.Tx
}

type ProcessedCardanoTx struct {
	BlockSlot     uint64 `json:"block_slot"`
	BlockHash     string `json:"block_hash"`
	OriginChainID string `json:"origin_chain_id"`
	Hash          string `json:"hash"`
	IsInvalid     bool   `json:"is_invalid"`
}

type BridgeExpectedCardanoTx struct {
	ChainID  string `json:"chain_id"`
	Hash     string `json:"hash"`
	Metadata []byte `json:"metadata"`
	TTL      uint64 `json:"ttl"`
}

type BridgeExpectedCardanoDBTx struct {
	BridgeExpectedCardanoTx

	IsProcessed bool `json:"is_processed"`
	IsInvalid   bool `json:"is_invalid"`
}

type ContractCardanoBlock = contractbinding.IBridgeContractStructsCardanoBlock

func (tx *CardanoTx) ToProcessedCardanoTx(isInvalid bool) *ProcessedCardanoTx {
	return &ProcessedCardanoTx{
		BlockSlot:     tx.BlockSlot,
		BlockHash:     tx.BlockHash,
		OriginChainID: tx.OriginChainID,
		Hash:          tx.Hash,
		IsInvalid:     isInvalid,
	}
}

func ToUnprocessedTxKey(blockSlot uint64, originChainID string, txHash string) string {
	return fmt.Sprintf("%20d_%v_%v", blockSlot, originChainID, txHash)
}

func (tx CardanoTx) ToUnprocessedTxKey() string {
	return ToUnprocessedTxKey(tx.BlockSlot, tx.OriginChainID, tx.Hash)
}

func (tx ProcessedCardanoTx) ToUnprocessedTxKey() string {
	return ToUnprocessedTxKey(tx.BlockSlot, tx.OriginChainID, tx.Hash)
}

func ToCardanoTxKey(originChainID string, txHash string) string {
	return fmt.Sprintf("%v_%v", originChainID, txHash)
}

func (tx CardanoTx) ToCardanoTxKey() string {
	return ToCardanoTxKey(tx.OriginChainID, tx.Hash)
}

func (tx ProcessedCardanoTx) ToCardanoTxKey() string {
	return ToCardanoTxKey(tx.OriginChainID, tx.Hash)
}

func (tx CardanoTx) Key() []byte {
	return []byte(tx.ToCardanoTxKey())
}

func (tx ProcessedCardanoTx) Key() []byte {
	return []byte(tx.ToCardanoTxKey())
}

func (tx BridgeExpectedCardanoTx) ToCardanoTxKey() string {
	return ToCardanoTxKey(tx.ChainID, tx.Hash)
}

func (tx BridgeExpectedCardanoTx) ToExpectedTxKey() string {
	return fmt.Sprintf("%20d_%v_%v", tx.TTL, tx.ChainID, tx.Hash)
}

func (tx BridgeExpectedCardanoTx) Key() []byte {
	return []byte(tx.ToExpectedTxKey())
}

func (tx BridgeExpectedCardanoDBTx) Key() []byte {
	return []byte(tx.ToExpectedTxKey())
}
