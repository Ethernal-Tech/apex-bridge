package core

import (
	"time"

	oCore "github.com/Ethernal-Tech/apex-bridge/oracle_common/core"
	"github.com/gagliardetto/solana-go"
)

type SolanaTx struct {
	OriginChainID  string    `json:"origin_chain_id"`
	Priority       uint8     `json:"priority"`
	SubmitTryCount uint32    `json:"try_count"`
	BatchTryCount  uint32    `json:"bf_count"`
	RefundTryCount uint32    `json:"refund_try_count"`
	LastTimeTried  time.Time `json:"last_time_tried"`
}

var _ oCore.BaseTx = (*SolanaTx)(nil)

// GetChainID implements core.BaseTx.
func (s *SolanaTx) GetChainID() string {
	panic("unimplemented") //nolint:gocritic
}

// GetPriority implements core.BaseTx.
func (s *SolanaTx) GetPriority() uint8 {
	panic("unimplemented") //nolint:gocritic
}

// GetSubmitTryCount implements core.BaseTx.
func (s *SolanaTx) GetSubmitTryCount() uint32 {
	panic("unimplemented") //nolint:gocritic
}

// GetTxHash implements core.BaseTx.
func (s *SolanaTx) GetTxHash() []byte {
	panic("unimplemented") //nolint:gocritic
}

// IncrementBatchTryCount implements core.BaseTx.
func (s *SolanaTx) IncrementBatchTryCount() {
	panic("unimplemented") //nolint:gocritic
}

// IncrementRefundTryCount implements core.BaseTx.
func (s *SolanaTx) IncrementRefundTryCount() {
	panic("unimplemented") //nolint:gocritic
}

// ResetSubmitTryCount implements core.BaseTx.
func (s *SolanaTx) ResetSubmitTryCount() {
	panic("unimplemented") //nolint:gocritic
}

// SetLastTimeTried implements core.BaseTx.
func (s *SolanaTx) SetLastTimeTried(lastTimeTried time.Time) {
	panic("unimplemented") //nolint:gocritic
}

// ToProcessed implements core.BaseTx.
func (s *SolanaTx) ToProcessed(isInvalid bool) oCore.BaseProcessedTx {
	panic("unimplemented") //nolint:gocritic
}

// UnprocessedDBKey implements core.BaseTx.
func (s *SolanaTx) UnprocessedDBKey() []byte {
	panic("unimplemented") //nolint:gocritic
}

type ProcessedSolanaTx struct {
	SlotNumber    uint64           `json:"block_number"`
	OriginChainID string           `json:"origin_chain_id"`
	TxSignature   solana.Signature `json:"tx_signature"`
	Priority      uint8            `json:"priority"`
	IsInvalid     bool             `json:"is_invalid"`
}

var _ oCore.BaseProcessedTx = (*ProcessedSolanaTx)(nil)

// GetChainID implements core.BaseProcessedTx.
func (p *ProcessedSolanaTx) GetChainID() string {
	panic("unimplemented") //nolint:gocritic
}

// GetInnerActionTxHash implements core.BaseProcessedTx.
func (p *ProcessedSolanaTx) GetInnerActionTxHash() []byte {
	panic("unimplemented") //nolint:gocritic
}

// GetIsInvalid implements core.BaseProcessedTx.
func (p *ProcessedSolanaTx) GetIsInvalid() bool {
	panic("unimplemented") //nolint:gocritic
}

// GetTxHash implements core.BaseProcessedTx.
func (p *ProcessedSolanaTx) GetTxHash() []byte {
	panic("unimplemented") //nolint:gocritic
}

// HasInnerActionTxHash implements core.BaseProcessedTx.
func (p *ProcessedSolanaTx) HasInnerActionTxHash() bool {
	panic("unimplemented") //nolint:gocritic
}

// UnprocessedDBKey implements core.BaseProcessedTx.
func (p *ProcessedSolanaTx) UnprocessedDBKey() []byte {
	panic("unimplemented") //nolint:gocritic
}

type BridgeExpectedSolanaTx struct {
	ChainID     string           `json:"chain_id"`
	TxSignature solana.Signature `json:"tx_signature"`
	TTL         uint64           `json:"ttl"`
	Priority    uint8            `json:"priority"`

	IsProcessed bool `json:"is_processed"`
	IsInvalid   bool `json:"is_invalid"`
}

var _ oCore.BaseExpectedTx = (*BridgeExpectedSolanaTx)(nil)

// DBKey implements core.BaseExpectedTx.
func (b *BridgeExpectedSolanaTx) DBKey() []byte {
	panic("unimplemented") //nolint:gocritic
}

// GetChainID implements core.BaseExpectedTx.
func (b *BridgeExpectedSolanaTx) GetChainID() string {
	panic("unimplemented") //nolint:gocritic
}

// GetIsInvalid implements core.BaseExpectedTx.
func (b *BridgeExpectedSolanaTx) GetIsInvalid() bool {
	panic("unimplemented") //nolint:gocritic
}

// GetIsProcessed implements core.BaseExpectedTx.
func (b *BridgeExpectedSolanaTx) GetIsProcessed() bool {
	panic("unimplemented") //nolint:gocritic
}

// GetPriority implements core.BaseExpectedTx.
func (b *BridgeExpectedSolanaTx) GetPriority() uint8 {
	panic("unimplemented") //nolint:gocritic
}

// GetTxHash implements core.BaseExpectedTx.
func (b *BridgeExpectedSolanaTx) GetTxHash() []byte {
	panic("unimplemented") //nolint:gocritic
}

// SetInvalid implements core.BaseExpectedTx.
func (b *BridgeExpectedSolanaTx) SetInvalid() {
	panic("unimplemented") //nolint:gocritic
}

// SetProcessed implements core.BaseExpectedTx.
func (b *BridgeExpectedSolanaTx) SetProcessed() {
	panic("unimplemented") //nolint:gocritic
}
