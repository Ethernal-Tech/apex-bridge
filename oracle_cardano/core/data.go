package core

import (
	"encoding/binary"

	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
)

type CardanoTx struct {
	OriginChainID string `json:"origin_chain_id"`
	Priority      uint8  `json:"priority"`

	indexer.Tx
}

var _ cCore.BaseTx = (*CardanoTx)(nil)

type ProcessedCardanoTx struct {
	BlockSlot     uint64       `json:"block_slot"`
	BlockHash     indexer.Hash `json:"block_hash"`
	OriginChainID string       `json:"origin_chain_id"`
	Hash          indexer.Hash `json:"hash"`
	Priority      uint8        `json:"priority"`
	IsInvalid     bool         `json:"is_invalid"`
}

var _ cCore.BaseProcessedTx = (*ProcessedCardanoTx)(nil)

type BridgeExpectedCardanoTx struct {
	ChainID  string       `json:"chain_id"`
	Hash     indexer.Hash `json:"hash"`
	Metadata []byte       `json:"metadata"`
	TTL      uint64       `json:"ttl"`
	Priority uint8        `json:"priority"`

	IsProcessed bool `json:"is_processed"`
	IsInvalid   bool `json:"is_invalid"`
}

var _ cCore.BaseExpectedTx = (*BridgeExpectedCardanoTx)(nil)

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

// GetOriginChainID implements core.BaseTx.
func (tx CardanoTx) GetOriginChainID() string {
	return tx.OriginChainID
}

// GetPriority implements core.BaseTx.
func (tx CardanoTx) GetPriority() uint8 {
	return tx.Priority
}

// ToUnprocessedTxKey implements core.BaseTx.
func (tx CardanoTx) ToUnprocessedTxKey() []byte {
	return ToUnprocessedTxKey(tx.Priority, tx.BlockSlot, tx.OriginChainID, tx.Hash)
}

// Key implements core.BaseProcessedTx.
func (tx ProcessedCardanoTx) Key() []byte {
	return tx.ToCardanoTxKey()
}

// ToUnprocessedTxKey implements core.BaseProcessedTx.
func (tx ProcessedCardanoTx) ToUnprocessedTxKey() []byte {
	return ToUnprocessedTxKey(tx.Priority, tx.BlockSlot, tx.OriginChainID, tx.Hash)
}

// Key implements core.BaseExpectedTx.
func (tx BridgeExpectedCardanoTx) Key() []byte {
	return tx.ToExpectedTxKey()
}

// GetChainID implements core.BaseExpectedTx.
func (tx BridgeExpectedCardanoTx) GetChainID() string {
	return tx.ChainID
}

// GetPriority implements core.BaseExpectedTx.
func (tx BridgeExpectedCardanoTx) GetPriority() uint8 {
	return tx.Priority
}

// GetIsInvalid implements core.BaseExpectedTx.
func (tx BridgeExpectedCardanoTx) GetIsInvalid() bool {
	return tx.IsInvalid
}

// GetIsProcessed implements core.BaseExpectedTx.
func (tx BridgeExpectedCardanoTx) GetIsProcessed() bool {
	return tx.IsProcessed
}

// SetProcessed implements core.BaseExpectedTx.
func (tx *BridgeExpectedCardanoTx) SetProcessed() {
	tx.IsProcessed = true
}

// SetInvalid implements core.BaseExpectedTx.
func (tx *BridgeExpectedCardanoTx) SetInvalid() {
	tx.IsInvalid = true
}

func ToUnprocessedTxKey(priority uint8, blockSlot uint64, originChainID string, txHash indexer.Hash) []byte {
	bytes := [9]byte{priority}

	binary.BigEndian.PutUint64(bytes[1:], blockSlot)

	return append(append(bytes[:], []byte(originChainID)...), txHash[:]...)
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
	return tx.ToCardanoTxKey()
}

func (tx BridgeExpectedCardanoTx) ToCardanoTxKey() []byte {
	return ToCardanoTxKey(tx.ChainID, tx.Hash)
}

func (tx BridgeExpectedCardanoTx) ToExpectedTxKey() []byte {
	bytes := [9]byte{tx.Priority}

	binary.BigEndian.PutUint64(bytes[1:], tx.TTL)

	return append(append(bytes[:], []byte(tx.ChainID)...), tx.Hash[:]...)
}
