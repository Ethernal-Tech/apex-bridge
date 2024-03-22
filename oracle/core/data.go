package core

import (
	"fmt"

	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
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

type UtxoTransaction struct {
}

type Utxo struct {
	Address string
	Amount  uint64
}

type BridgingRequestReceiver struct {
	Address string
	Amount  uint64
}

type BridgingRequestClaim struct {
	TxHash             string
	Receivers          []BridgingRequestReceiver
	OutputUtxos        []Utxo
	DestinationChainId string
}

type BatchExecutedClaim struct {
	TxHash       string
	BatchNonceId string
	OutputUtxos  []Utxo
}

type BatchExecutionFailedClaim struct {
	TxHash       string
	BatchNonceId string
}

type RefundRequestClaim struct {
	TxHash               string
	PreviousRefundTxHash string
	RetryCounter         int32
	RefundToAddress      string
	OutputUtxos          []Utxo
	DestinationChainId   string
	UtxoTransaction      UtxoTransaction
}

type RefundExecutedClaim struct {
	TxHash       string
	RefundTxHash string
	OutputUtxos  []Utxo
}

type BridgeClaimsBlockInfo struct {
	ChainId string
	Slot    uint64
	Hash    string
}

type BridgeClaims struct {
	BridgingRequest      []BridgingRequestClaim
	BatchExecuted        []BatchExecutedClaim
	BatchExecutionFailed []BatchExecutionFailedClaim
	// RefundRequest        []RefundRequestClaim
	// RefundExecuted       []RefundExecutedClaim

	BlockInfo          *BridgeClaimsBlockInfo
	BlockFullyObserved bool
}

func (bc *BridgeClaims) Count() int {
	return len(bc.BridgingRequest) +
		len(bc.BatchExecuted) +
		len(bc.BatchExecutionFailed) /* +
		len(bc.RefundRequest) +
		len(bc.RefundExecuted)*/
}

func (bc *BridgeClaims) Any() bool {
	return bc.Count() > 0
}

func (bc *BridgeClaims) HasBlockInfo() bool {
	return bc.BlockInfo != nil
}

func (bc *BridgeClaims) BlockInfoEqualWithUnprocessed(tx *CardanoTx) bool {
	return bc.HasBlockInfo() && bc.BlockInfo.ChainId == tx.OriginChainId && bc.BlockInfo.Slot == tx.BlockSlot && bc.BlockInfo.Hash == tx.BlockHash
}

func (bc *BridgeClaims) BlockInfoEqualWithProcessed(tx *ProcessedCardanoTx) bool {
	return bc.HasBlockInfo() && bc.BlockInfo.ChainId == tx.OriginChainId && bc.BlockInfo.Slot == tx.BlockSlot && bc.BlockInfo.Hash == tx.BlockHash
}

func (bc *BridgeClaims) BlockInfoEqualWithExpected(tx *BridgeExpectedCardanoTx, block *indexer.CardanoBlock) bool {
	return bc.HasBlockInfo() && bc.BlockInfo.ChainId == tx.ChainId && bc.BlockInfo.Slot == block.Slot && bc.BlockInfo.Hash == block.Hash
}

func (bc *BridgeClaims) SetBlockInfoWithUnprocessed(tx *CardanoTx) {
	bc.BlockInfo = &BridgeClaimsBlockInfo{
		ChainId: tx.OriginChainId,
		Slot:    tx.BlockSlot,
		Hash:    tx.BlockHash,
	}
}

func (bc *BridgeClaims) SetBlockInfoWithProcessed(tx *ProcessedCardanoTx) {
	bc.BlockInfo = &BridgeClaimsBlockInfo{
		ChainId: tx.OriginChainId,
		Slot:    tx.BlockSlot,
		Hash:    tx.BlockHash,
	}
}

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
