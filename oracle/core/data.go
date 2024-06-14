package core

import (
	"encoding/binary"

	"github.com/Ethernal-Tech/apex-bridge/contractbinding"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
)

const (
	LastProcessingPriority = uint8(1)
)

type CardanoTx struct {
	OriginChainID string `json:"origin_chain_id"`
	Priority      uint8  `json:"priority"`

	indexer.Tx
}

type ProcessedCardanoTx struct {
	BlockSlot     uint64       `json:"block_slot"`
	BlockHash     indexer.Hash `json:"block_hash"`
	OriginChainID string       `json:"origin_chain_id"`
	Hash          indexer.Hash `json:"hash"`
	Priority      uint8        `json:"priority"`
	IsInvalid     bool         `json:"is_invalid"`
}

type BridgeExpectedCardanoTx struct {
	ChainID  string       `json:"chain_id"`
	Hash     indexer.Hash `json:"hash"`
	Metadata []byte       `json:"metadata"`
	TTL      uint64       `json:"ttl"`
	Priority uint8        `json:"priority"`
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

func ToUnprocessedTxKey(priority uint8, blockSlot uint64, originChainID string, txHash indexer.Hash) []byte {
	bytes := [9]byte{priority}

	binary.BigEndian.PutUint64(bytes[1:], blockSlot)

	return append(append(bytes[:], []byte(originChainID)...), txHash[:]...)
}

func (tx CardanoTx) ToUnprocessedTxKey() []byte {
	return ToUnprocessedTxKey(tx.Priority, tx.BlockSlot, tx.OriginChainID, tx.Hash)
}

func (tx ProcessedCardanoTx) ToUnprocessedTxKey() []byte {
	return ToUnprocessedTxKey(tx.Priority, tx.BlockSlot, tx.OriginChainID, tx.Hash)
}

func ToCardanoTxKey(originChainID string, txHash indexer.Hash) []byte {
	return append([]byte(originChainID), txHash[:]...)
}

func (tx CardanoTx) ToCardanoTxKey() []byte {
	return ToCardanoTxKey(tx.OriginChainID, tx.Hash)
}

func (tx ProcessedCardanoTx) ToCardanoTxKey() []byte {
	return ToCardanoTxKey(tx.OriginChainID, tx.Hash)
}

func (tx CardanoTx) Key() []byte {
	return []byte(tx.ToCardanoTxKey())
}

func (tx ProcessedCardanoTx) Key() []byte {
	return []byte(tx.ToCardanoTxKey())
}

func (tx BridgeExpectedCardanoTx) ToCardanoTxKey() []byte {
	return ToCardanoTxKey(tx.ChainID, tx.Hash)
}

func (tx BridgeExpectedCardanoTx) ToExpectedTxKey() []byte {
	bytes := [9]byte{tx.Priority}

	binary.BigEndian.PutUint64(bytes[1:], tx.TTL)

	return append(append(bytes[:], []byte(tx.ChainID)...), tx.Hash[:]...)
}

func (tx BridgeExpectedCardanoTx) Key() []byte {
	return tx.ToExpectedTxKey()
}

func (tx BridgeExpectedCardanoDBTx) Key() []byte {
	return tx.ToExpectedTxKey()
}
