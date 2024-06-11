package core

import (
	"fmt"

	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
)

const (
	LastProcessingPriority = uint(1)
)

type CardanoTx struct {
	OriginChainID string `json:"origin_chain_id"`
	Priority      uint   `json:"priority"`

	indexer.Tx
}

type ProcessedCardanoTx struct {
	BlockSlot     uint64 `json:"block_slot"`
	BlockHash     string `json:"block_hash"`
	OriginChainID string `json:"origin_chain_id"`
	Hash          string `json:"hash"`
	Priority      uint   `json:"priority"`
	IsInvalid     bool   `json:"is_invalid"`
}

type BridgeExpectedCardanoTx struct {
	ChainID  string `json:"chain_id"`
	Hash     string `json:"hash"`
	Metadata []byte `json:"metadata"`
	TTL      uint64 `json:"ttl"`
	Priority uint   `json:"priority"`
}

type BridgeExpectedCardanoDBTx struct {
	BridgeExpectedCardanoTx

	IsProcessed bool `json:"is_processed"`
	IsInvalid   bool `json:"is_invalid"`
}

type ContractCardanoBlock = contractbinding.IBridgeStructsCardanoBlock

func (tx *CardanoTx) ToProcessedCardanoTx(isInvalid bool) *ProcessedCardanoTx {
	return &ProcessedCardanoTx{
		BlockSlot:     tx.BlockSlot,
		BlockHash:     tx.BlockHash,
		OriginChainID: tx.OriginChainID,
		Hash:          tx.Hash,
		Priority:      tx.Priority,
		IsInvalid:     isInvalid,
	}
}

func ToUnprocessedTxKey(priority uint, blockSlot uint64, originChainID string, txHash string) string {
	return fmt.Sprintf("%d_%20d_%v_%v", priority, blockSlot, originChainID, txHash)
}

func (tx CardanoTx) ToUnprocessedTxKey() string {
	return ToUnprocessedTxKey(tx.Priority, tx.BlockSlot, tx.OriginChainID, tx.Hash)
}

func (tx ProcessedCardanoTx) ToUnprocessedTxKey() string {
	return ToUnprocessedTxKey(tx.Priority, tx.BlockSlot, tx.OriginChainID, tx.Hash)
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
	return fmt.Sprintf("%d_%20d_%v_%v", tx.Priority, tx.TTL, tx.ChainID, tx.Hash)
}

func (tx BridgeExpectedCardanoTx) Key() []byte {
	return []byte(tx.ToExpectedTxKey())
}

func (tx BridgeExpectedCardanoDBTx) Key() []byte {
	return []byte(tx.ToExpectedTxKey())
}
