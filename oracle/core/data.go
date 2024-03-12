package core

import (
	"fmt"

	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
)

type CardanoTx struct {
	OriginChainId string `json:"origin_chain_id"`

	indexer.Tx
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

func (ctx CardanoTx) StrKey() string {
	return fmt.Sprintf("%v_%v", ctx.OriginChainId, ctx.Hash)
}

func (ctx CardanoTx) Key() []byte {
	return []byte(ctx.StrKey())
}

func (betx BridgeExpectedCardanoTx) StrKey() string {
	return fmt.Sprintf("%v_%v", betx.ChainId, betx.Hash)
}

func (betx BridgeExpectedCardanoTx) Key() []byte {
	return []byte(betx.StrKey())
}

func (betx BridgeExpectedCardanoDbTx) Key() []byte {
	return []byte(fmt.Sprintf("%v_%v", betx.ChainId, betx.Hash))
}
