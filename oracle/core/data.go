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

type BridgeClaims struct {
	BridgingRequest      []BridgingRequestClaim
	BatchExecuted        []BatchExecutedClaim
	BatchExecutionFailed []BatchExecutionFailedClaim
	// RefundRequest        []RefundRequestClaim
	// RefundExecuted       []RefundExecutedClaim
}

func (bc BridgeClaims) Count() int {
	return len(bc.BridgingRequest) +
		len(bc.BatchExecuted) +
		len(bc.BatchExecutionFailed) /* +
		len(bc.RefundRequest) +
		len(bc.RefundExecuted)*/
}

func (bc BridgeClaims) Any() bool {
	return bc.Count() > 0
}

func (tx *CardanoTx) ToProcessedCardanoTx(isInvalid bool) *ProcessedCardanoTx {
	return &ProcessedCardanoTx{
		OriginChainId: tx.OriginChainId,
		Hash:          tx.Hash,
		IsInvalid:     isInvalid,
	}
}

func ToCardanoTxKey(originChainId string, txHash string) string {
	return fmt.Sprintf("%v_%v", originChainId, txHash)
}

func (tx CardanoTx) ToCardanoTxKey() string {
	return ToCardanoTxKey(tx.OriginChainId, tx.Hash)
}

func (tx CardanoTx) Key() []byte {
	return []byte(tx.ToCardanoTxKey())
}

func (tx ProcessedCardanoTx) ToCardanoTxKey() string {
	return ToCardanoTxKey(tx.OriginChainId, tx.Hash)
}

func (tx ProcessedCardanoTx) Key() []byte {
	return []byte(tx.ToCardanoTxKey())
}

func (tx BridgeExpectedCardanoTx) ToCardanoTxKey() string {
	return ToCardanoTxKey(tx.ChainId, tx.Hash)
}

func (tx BridgeExpectedCardanoTx) Key() []byte {
	return []byte(tx.ToCardanoTxKey())
}

func (tx BridgeExpectedCardanoDbTx) Key() []byte {
	return []byte(ToCardanoTxKey(tx.ChainId, tx.Hash))
}
