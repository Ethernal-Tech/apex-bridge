package core

import (
	"encoding/binary"
	"math/big"
	"time"

	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/ethgo"
)

type EthTx struct {
	OriginChainID  string    `json:"origin_chain_id"`
	Priority       uint8     `json:"priority"`
	SubmitTryCount uint32    `json:"try_count"`
	BatchTryCount  uint32    `json:"bf_count"`
	RefundTryCount uint32    `json:"refund_try_count"`
	LastTimeTried  time.Time `json:"last_time_tried"`

	BlockNumber     uint64        `json:"block_number"`
	BlockHash       ethgo.Hash    `json:"block_hash"`
	Hash            ethgo.Hash    `json:"hash"`
	TxIndex         uint64        `json:"tx_index"`
	Value           *big.Int      `json:"value"`
	Removed         bool          `json:"removed"`
	LogIndex        uint64        `json:"log_index"`
	Address         ethgo.Address `json:"addr"`
	Metadata        []byte        `json:"metadata"`
	InnerActionHash ethgo.Hash    `json:"ia_hash"`
}

var _ cCore.BaseTx = (*EthTx)(nil)

type ProcessedEthTx struct {
	BlockNumber     uint64     `json:"block_number"`
	BlockHash       ethgo.Hash `json:"block_hash"`
	OriginChainID   string     `json:"origin_chain_id"`
	Hash            ethgo.Hash `json:"hash"`
	Priority        uint8      `json:"priority"`
	IsInvalid       bool       `json:"is_invalid"`
	InnerActionHash ethgo.Hash `json:"ia_hash"`
}

var _ cCore.BaseProcessedTx = (*ProcessedEthTx)(nil)

type BridgeExpectedEthTx struct {
	ChainID  string     `json:"chain_id"`
	Hash     ethgo.Hash `json:"hash"`
	Metadata []byte     `json:"metadata"`
	TTL      uint64     `json:"ttl"`
	Priority uint8      `json:"priority"`

	IsProcessed bool `json:"is_processed"`
	IsInvalid   bool `json:"is_invalid"`
}

var _ cCore.BaseExpectedTx = (*BridgeExpectedEthTx)(nil)

type EthUpdateTxsData = cCore.UpdateTxsData[*EthTx, *ProcessedEthTx, *BridgeExpectedEthTx]

type BridgeClaimsBlockInfo struct {
	ChainID string
	Number  uint64
}

func (bi *BridgeClaimsBlockInfo) EqualWithUnprocessed(tx *EthTx) bool {
	return bi.ChainID == tx.OriginChainID && bi.Number == tx.BlockNumber
}

func (bi *BridgeClaimsBlockInfo) EqualWithProcessed(tx *ProcessedEthTx) bool {
	return bi.ChainID == tx.OriginChainID && bi.Number == tx.BlockNumber
}

func (bi *BridgeClaimsBlockInfo) EqualWithExpected(tx *BridgeExpectedEthTx, blockNumber uint64) bool {
	return bi.ChainID == tx.ChainID && bi.Number == blockNumber
}

// ChainID implements core.BaseTx.
func (tx EthTx) GetChainID() string {
	return tx.OriginChainID
}

// TxHash implements core.BaseTx.
func (tx EthTx) GetTxHash() []byte {
	return tx.Hash[:]
}

// UnprocessedDBKey implements core.BaseTx.
func (tx EthTx) UnprocessedDBKey() []byte {
	return toUnprocessedEthTxKey(tx.Priority, tx.BlockNumber, tx.Hash)
}

// SetLastTimeTried implements core.BaseTx.
func (tx *EthTx) SetLastTimeTried(lastTimeTried time.Time) {
	tx.LastTimeTried = lastTimeTried
}

// ResetSubmitTryCount implements core.BaseTx.
func (tx *EthTx) ResetSubmitTryCount() {
	tx.SubmitTryCount = 0
}

// IncrementBatchTryCount implements core.BaseTx.
func (tx *EthTx) IncrementBatchTryCount() {
	tx.BatchTryCount++
}

// IncrementRefundTryCount implements core.BaseTx.
func (tx *EthTx) IncrementRefundTryCount() {
	tx.RefundTryCount++
}

// PendingDBKey implements core.BaseTx.
func (tx EthTx) ToProcessed(isInvalid bool) cCore.BaseProcessedTx {
	return tx.ToProcessedEthTx(isInvalid)
}

// GetSubmitTryCount implements core.BaseTx.
func (tx EthTx) GetSubmitTryCount() uint32 {
	return tx.SubmitTryCount
}

// GetPriority implements core.BaseTx.
func (tx EthTx) GetPriority() uint8 {
	return tx.Priority
}

// ChainID implements core.BaseProcessedTx.
func (tx ProcessedEthTx) GetChainID() string {
	return tx.OriginChainID
}

// TxHash implements core.BaseProcessedTx.
func (tx ProcessedEthTx) GetTxHash() []byte {
	return tx.Hash[:]
}

// HasInnerActionTxHash implements core.BaseProcessedTx.
func (tx ProcessedEthTx) HasInnerActionTxHash() bool {
	return tx.InnerActionHash != ethgo.Hash{}
}

// GetInnerActionTxHash implements core.BaseProcessedTx.
func (tx ProcessedEthTx) GetInnerActionTxHash() []byte {
	return tx.InnerActionHash[:]
}

// UnprocessedDBKey implements core.BaseProcessedTx.
func (tx ProcessedEthTx) UnprocessedDBKey() []byte {
	return toUnprocessedEthTxKey(tx.Priority, tx.BlockNumber, tx.Hash)
}

// GetIsInvalid implements core.BaseProcessedTx.
func (tx ProcessedEthTx) GetIsInvalid() bool {
	return tx.IsInvalid
}

// ChainID implements core.BaseExpectedTx.
func (tx BridgeExpectedEthTx) GetChainID() string {
	return tx.ChainID
}

// TxHash implements core.BaseExpectedTx.
func (tx BridgeExpectedEthTx) GetTxHash() []byte {
	return tx.Hash[:]
}

// Key implements core.BaseExpectedTx.
func (tx BridgeExpectedEthTx) DBKey() []byte {
	return tx.ToExpectedTxKey()
}

// GetPriority implements core.BaseExpectedTx.
func (tx BridgeExpectedEthTx) GetPriority() uint8 {
	return tx.Priority
}

// GetIsInvalid implements core.BaseExpectedTx.
func (tx BridgeExpectedEthTx) GetIsInvalid() bool {
	return tx.IsInvalid
}

// GetIsProcessed implements core.BaseExpectedTx.
func (tx BridgeExpectedEthTx) GetIsProcessed() bool {
	return tx.IsProcessed
}

// SetProcessed implements core.BaseExpectedTx.
func (tx *BridgeExpectedEthTx) SetProcessed() {
	tx.IsProcessed = true
}

// SetInvalid implements core.BaseExpectedTx.
func (tx *BridgeExpectedEthTx) SetInvalid() {
	tx.IsInvalid = true
}

func (tx *EthTx) ToProcessedEthTx(isInvalid bool) *ProcessedEthTx {
	return &ProcessedEthTx{
		BlockNumber:     tx.BlockNumber,
		BlockHash:       tx.BlockHash,
		OriginChainID:   tx.OriginChainID,
		Hash:            tx.Hash,
		Priority:        tx.Priority,
		InnerActionHash: tx.InnerActionHash,
		IsInvalid:       isInvalid,
	}
}

func toUnprocessedEthTxKey(priority uint8, blockNumber uint64, txHash ethgo.Hash) []byte {
	bytes := [9]byte{priority}

	binary.BigEndian.PutUint64(bytes[1:], blockNumber)

	return append(bytes[:], txHash[:]...)
}

func ToEthTxKey(originChainID string, txHash ethgo.Hash) []byte {
	return append([]byte(originChainID), txHash[:]...)
}

func (tx EthTx) ToEthTxKey() []byte {
	return ToEthTxKey(tx.OriginChainID, tx.Hash)
}

func (tx EthTx) ToExpectedEthTxKey() []byte {
	return ToEthTxKey(tx.OriginChainID, tx.InnerActionHash)
}

func (tx ProcessedEthTx) ToEthTxKey() []byte {
	return ToEthTxKey(tx.OriginChainID, tx.Hash)
}

func (tx BridgeExpectedEthTx) ToEthTxKey() []byte {
	return ToEthTxKey(tx.ChainID, tx.Hash)
}

func (tx BridgeExpectedEthTx) ToExpectedTxKey() []byte {
	bytes := [9]byte{tx.Priority}

	binary.BigEndian.PutUint64(bytes[1:], tx.TTL)

	return append(bytes[:], tx.Hash[:]...)
}
