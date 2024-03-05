package core

import (
	"github.com/Ethernal-Tech/cardano-infrastructure/indexer"
)

type CardanoTx struct {
	OriginChainId string `json:"origin_chain_id"`

	indexer.Tx
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

func (bc BridgeClaims) Any() bool {
	return len(bc.BridgingRequest) > 0 ||
		len(bc.BatchExecuted) > 0 ||
		len(bc.BatchExecutionFailed) > 0 /*||
		len(bc.RefundRequest) > 0 ||
		len(bc.RefundExecuted) > 0*/
}

func (btx CardanoTx) Key() []byte {
	return []byte(btx.Hash)
}
