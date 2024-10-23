package core

import (
	"encoding/binary"
	"math/big"
	"time"

	cCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/Ethernal-Tech/ethgo"
)

type EthTx struct {
	OriginChainID string    `json:"origin_chain_id"`
	Priority      uint8     `json:"priority"`
	TryCount      uint32    `json:"try_count"`
	LastTimeTried time.Time `json:"last_time_tried"`

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

type ProcessedEthTxByInnerAction struct {
	OriginChainID   string     `json:"origin_chain_id"`
	Hash            ethgo.Hash `json:"hash"`
	InnerActionHash ethgo.Hash `json:"ia_hash"`
}

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

// GetOriginChainID implements core.BaseTx.
func (tx EthTx) GetOriginChainID() string {
	return tx.OriginChainID
}

// GetPriority implements core.BaseTx.
func (tx EthTx) GetPriority() uint8 {
	return tx.Priority
}

// ToUnprocessedTxKey implements core.BaseTx.
func (tx EthTx) ToUnprocessedTxKey() []byte {
	return toUnprocessedEthTxKey(tx.Priority, tx.BlockNumber, tx.OriginChainID, tx.Hash)
}

// Key implements core.BaseProcessedTx.
func (tx ProcessedEthTx) Key() []byte {
	return tx.ToEthTxKey()
}

// ToUnprocessedTxKey implements core.BaseProcessedTx.
func (tx ProcessedEthTx) ToUnprocessedTxKey() []byte {
	return toUnprocessedEthTxKey(tx.Priority, tx.BlockNumber, tx.OriginChainID, tx.Hash)
}

// Key implements core.BaseExpectedTx.
func (tx BridgeExpectedEthTx) Key() []byte {
	return tx.ToExpectedTxKey()
}

// GetChainID implements core.BaseExpectedTx.
func (tx BridgeExpectedEthTx) GetChainID() string {
	return tx.ChainID
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

func (tx *ProcessedEthTx) ToProcessedTxByInnerAction() *ProcessedEthTxByInnerAction {
	return &ProcessedEthTxByInnerAction{
		OriginChainID:   tx.OriginChainID,
		Hash:            tx.Hash,
		InnerActionHash: tx.InnerActionHash,
	}
}

func toUnprocessedEthTxKey(priority uint8, blockNumber uint64, originChainID string, txHash ethgo.Hash) []byte {
	bytes := [9]byte{priority}

	binary.BigEndian.PutUint64(bytes[1:], blockNumber)

	return append(append(bytes[:], []byte(originChainID)...), txHash[:]...)
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

func (tx EthTx) Key() []byte {
	return tx.ToEthTxKey()
}

func (tx ProcessedEthTx) KeyByInnerAction() []byte {
	return ToEthTxKey(tx.OriginChainID, tx.InnerActionHash)
}

func (tx BridgeExpectedEthTx) ToEthTxKey() []byte {
	return ToEthTxKey(tx.ChainID, tx.Hash)
}

func (tx BridgeExpectedEthTx) ToExpectedTxKey() []byte {
	bytes := [9]byte{tx.Priority}

	binary.BigEndian.PutUint64(bytes[1:], tx.TTL)

	return append(append(bytes[:], []byte(tx.ChainID)...), tx.Hash[:]...)
}
