package core

import (
	"encoding/binary"
	"time"

	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
)

type CardanoTx struct {
	OriginChainID string    `json:"origin_chain_id"`
	Priority      uint8     `json:"priority"`
	TryCount      uint32    `json:"try_count"`
	LastTimeTried time.Time `json:"last_time_tried"`

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

type CardanoUpdateTxsData = cCore.UpdateTxsData[*CardanoTx, *ProcessedCardanoTx, *BridgeExpectedCardanoTx]

// ChainID implements core.BaseTx.
func (tx CardanoTx) GetChainID() string {
	return tx.OriginChainID
}

// TxHash implements core.BaseTx.
func (tx CardanoTx) GetTxHash() []byte {
	return tx.Hash[:]
}

// UnprocessedDBKey implements core.BaseTx.
func (tx CardanoTx) UnprocessedDBKey() []byte {
	return ToUnprocessedTxKey(tx.Priority, tx.BlockSlot, tx.Hash)
}

// SetLastTimeTried implements core.BaseTx.
func (tx *CardanoTx) SetLastTimeTried(lastTimeTried time.Time) {
	tx.LastTimeTried = lastTimeTried
}

// IncrementTryCount implements core.BaseTx.
func (tx *CardanoTx) IncrementTryCount() {
	tx.TryCount++
}

// PendingDBKey implements core.BaseTx.
func (tx CardanoTx) ToProcessed(isInvalid bool) cCore.BaseProcessedTx {
	return tx.ToProcessedCardanoTx(isInvalid)
}

// GetTryCount implements core.BaseTx.
func (tx CardanoTx) GetTryCount() uint32 {
	return tx.TryCount
}

// GetPriority implements core.BaseTx.
func (tx CardanoTx) GetPriority() uint8 {
	return tx.Priority
}

// ChainID implements core.BaseProcessedTx.
func (tx ProcessedCardanoTx) GetChainID() string {
	return tx.OriginChainID
}

// TxHash implements core.BaseProcessedTx.
func (tx ProcessedCardanoTx) GetTxHash() []byte {
	return tx.Hash[:]
}

// HasInnerActionTxHash implements core.BaseProcessedTx.
func (tx ProcessedCardanoTx) HasInnerActionTxHash() bool {
	return false
}

// GetInnerActionTxHash implements core.BaseProcessedTx.
func (tx ProcessedCardanoTx) GetInnerActionTxHash() []byte {
	return nil
}

// UnprocessedDBKey implements core.BaseProcessedTx.
func (tx ProcessedCardanoTx) UnprocessedDBKey() []byte {
	return ToUnprocessedTxKey(tx.Priority, tx.BlockSlot, tx.Hash)
}

// ChainID implements core.BaseExpectedTx.
func (tx BridgeExpectedCardanoTx) GetChainID() string {
	return tx.ChainID
}

// TxHash implements core.BaseExpectedTx.
func (tx BridgeExpectedCardanoTx) GetTxHash() []byte {
	return tx.Hash[:]
}

// Key implements core.BaseExpectedTx.
func (tx BridgeExpectedCardanoTx) DBKey() []byte {
	return tx.ToExpectedTxKey()
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

func (tx CardanoTx) ShouldSkipForNow() bool {
	return !tx.LastTimeTried.IsZero() &&
		tx.LastTimeTried.Add(cCore.RetryUnprocessedAfterSec*time.Second).After(time.Now().UTC())
}

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

func ToUnprocessedTxKey(priority uint8, blockSlot uint64, txHash indexer.Hash) []byte {
	bytes := [9]byte{priority}

	binary.BigEndian.PutUint64(bytes[1:], blockSlot)

	return append(bytes[:], txHash[:]...)
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

func (tx BridgeExpectedCardanoTx) ToCardanoTxKey() []byte {
	return ToCardanoTxKey(tx.ChainID, tx.Hash)
}

func (tx BridgeExpectedCardanoTx) ToExpectedTxKey() []byte {
	bytes := [9]byte{tx.Priority}

	binary.BigEndian.PutUint64(bytes[1:], tx.TTL)

	return append(bytes[:], tx.Hash[:]...)
}
