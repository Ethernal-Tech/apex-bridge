package core

import (
	"encoding/binary"
	"math/big"
	"time"

	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/gagliardetto/solana-go"
)

const (
	BridgeRequestEvent       = "BridgeRequestEvent"
	TransactionExecutedEvent = "TransactionExecutedEvent"
	ValidatorSetUpdatedEvent = "ValidatorSetUpdatedEvent"
)

type SolanaUpdateTxsData = oCore.UpdateTxsData[*SolanaTx, *ProcessedSolanaTx, *BridgeExpectedSolanaTx]

type SolanaTx struct {
	OriginChainID  string    `json:"origin_chain_id"`
	Priority       uint8     `json:"priority"`
	SubmitTryCount uint32    `json:"try_count"`
	BatchTryCount  uint32    `json:"bf_count"`
	RefundTryCount uint32    `json:"refund_try_count"`
	LastTimeTried  time.Time `json:"last_time_tried"`

	SlotNumber      uint64           `json:"slot_number"`
	TxSignature     solana.Signature `json:"tx_signature"`
	Value           *big.Int         `json:"value"`
	Metadata        []byte           `json:"metadata"`
	InnerActionHash [32]byte         `json:"ia_hash"`
}

var _ oCore.BaseTx = (*SolanaTx)(nil)

func (s SolanaTx) GetChainID() string {
	return s.OriginChainID
}

func (s SolanaTx) GetPriority() uint8 {
	return s.Priority
}

func (s SolanaTx) GetSubmitTryCount() uint32 {
	return s.SubmitTryCount
}

func (s SolanaTx) GetTxHash() []byte {
	return s.TxSignature[:]
}

func (s *SolanaTx) IncrementBatchTryCount() {
	s.BatchTryCount++
}

func (s *SolanaTx) IncrementRefundTryCount() {
	s.RefundTryCount++
}

func (s *SolanaTx) ResetSubmitTryCount() {
	s.SubmitTryCount = 0
}

func (s *SolanaTx) SetLastTimeTried(lastTimeTried time.Time) {
	s.LastTimeTried = lastTimeTried
}

func (s *SolanaTx) ToProcessed(isInvalid bool) oCore.BaseProcessedTx {
	return s.ToProcessedSolanaTx(isInvalid)
}

func (s SolanaTx) UnprocessedDBKey() []byte {
	return toUnprocessedSolanaTxKey(s.Priority, s.SlotNumber, s.TxSignature)
}

func (s SolanaTx) ToSolanaTxKey() []byte {
	return ToSolanaTxKey(s.OriginChainID, s.TxSignature[:])
}

func (s SolanaTx) ToExpectedSolanaTxKey() []byte {
	return ToSolanaTxKey(s.OriginChainID, s.InnerActionHash[:])
}

func (s *SolanaTx) ToProcessedSolanaTx(isInvalid bool) *ProcessedSolanaTx {
	return &ProcessedSolanaTx{
		SlotNumber:      s.SlotNumber,
		OriginChainID:   s.OriginChainID,
		TxSignature:     s.TxSignature,
		Priority:        s.Priority,
		InnerActionHash: s.InnerActionHash,
		IsInvalid:       isInvalid,
	}
}

type ProcessedSolanaTx struct {
	SlotNumber      uint64           `json:"slot_number"`
	OriginChainID   string           `json:"origin_chain_id"`
	TxSignature     solana.Signature `json:"tx_signature"`
	Priority        uint8            `json:"priority"`
	IsInvalid       bool             `json:"is_invalid"`
	InnerActionHash [32]byte         `json:"ia_hash"`
}

var _ oCore.BaseProcessedTx = (*ProcessedSolanaTx)(nil)

func (p ProcessedSolanaTx) GetChainID() string {
	return p.OriginChainID
}

func (p ProcessedSolanaTx) GetTxHash() []byte {
	return p.TxSignature[:]
}

func (p ProcessedSolanaTx) HasInnerActionTxHash() bool {
	return p.InnerActionHash != ([32]byte{})
}

func (p ProcessedSolanaTx) GetInnerActionTxHash() []byte {
	return p.InnerActionHash[:]
}

func (p ProcessedSolanaTx) UnprocessedDBKey() []byte {
	return toUnprocessedSolanaTxKey(p.Priority, p.SlotNumber, p.TxSignature)
}

func (p ProcessedSolanaTx) GetIsInvalid() bool {
	return p.IsInvalid
}

func (p ProcessedSolanaTx) ToSolanaTxKey() []byte {
	return ToSolanaTxKey(p.OriginChainID, p.InnerActionHash[:])
}

type BridgeExpectedSolanaTx struct {
	ChainID  string   `json:"chain_id"`
	Hash     [32]byte `json:"hash"`
	Metadata []byte   `json:"metadata"`
	TTL      uint64   `json:"ttl"`
	Priority uint8    `json:"priority"`

	IsProcessed bool `json:"is_processed"`
	IsInvalid   bool `json:"is_invalid"`
}

var _ oCore.BaseExpectedTx = (*BridgeExpectedSolanaTx)(nil)

func (b BridgeExpectedSolanaTx) DBKey() []byte {
	return b.ToExpectedTxKey()
}

func (b BridgeExpectedSolanaTx) GetChainID() string {
	return b.ChainID
}

func (b BridgeExpectedSolanaTx) GetTxHash() []byte {
	return b.Hash[:]
}

func (b BridgeExpectedSolanaTx) GetPriority() uint8 {
	return b.Priority
}

func (b BridgeExpectedSolanaTx) GetIsInvalid() bool {
	return b.IsInvalid
}

func (b BridgeExpectedSolanaTx) GetIsProcessed() bool {
	return b.IsProcessed
}

func (b *BridgeExpectedSolanaTx) SetProcessed() {
	b.IsProcessed = true
}

func (b *BridgeExpectedSolanaTx) SetInvalid() {
	b.IsInvalid = true
}

func (b BridgeExpectedSolanaTx) ToSolanaTxKey() []byte {
	return ToSolanaTxKey(b.ChainID, b.Hash[:])
}

func (b BridgeExpectedSolanaTx) ToExpectedTxKey() []byte {
	bytes := [9]byte{b.Priority}

	binary.BigEndian.PutUint64(bytes[1:], b.TTL)

	return append(bytes[:], b.Hash[:]...)
}

// BridgeClaimsSlotInfo tracks the current Solana slot being processed for claim grouping.
type BridgeClaimsSlotInfo struct {
	ChainID string
	Number  uint64
}

func (bi *BridgeClaimsSlotInfo) EqualWithUnprocessed(tx *SolanaTx) bool {
	return bi.ChainID == tx.OriginChainID && bi.Number == tx.SlotNumber
}

func (bi *BridgeClaimsSlotInfo) EqualWithProcessed(tx *ProcessedSolanaTx) bool {
	return bi.ChainID == tx.OriginChainID && bi.Number == tx.SlotNumber
}

func (bi *BridgeClaimsSlotInfo) EqualWithExpected(tx *BridgeExpectedSolanaTx, blockNumber uint64) bool {
	return bi.ChainID == tx.ChainID && bi.Number == blockNumber
}

// toUnprocessedSolanaTxKey builds a sortable composite DB key: priority(1B) | slotNumber(8B BigEndian) | signature(64B)
func toUnprocessedSolanaTxKey(priority uint8, slotNumber uint64, sig solana.Signature) []byte {
	bytes := [9]byte{priority}

	binary.BigEndian.PutUint64(bytes[1:], slotNumber)

	return append(bytes[:], sig[:]...)
}

// ToSolanaTxKey builds the processed-tx lookup key: chainID || signature
func ToSolanaTxKey(originChainID string, hash []byte) []byte {
	return append([]byte(originChainID), hash...)
}
