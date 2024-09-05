package core

import (
	"encoding/binary"
	"math/big"

	"github.com/Ethernal-Tech/ethgo"
)

type EthTx struct {
	OriginChainID string `json:"origin_chain_id"`
	Priority      uint8  `json:"priority"`

	BlockNumber uint64        `json:"block_number"`
	BlockHash   ethgo.Hash    `json:"block_hash"`
	Hash        ethgo.Hash    `json:"hash"`
	TxIndex     uint64        `json:"tx_index"`
	Value       *big.Int      `json:"value"`
	Removed     bool          `json:"removed"`
	LogIndex    uint64        `json:"log_index"`
	Address     ethgo.Address `json:"addr"`
	Metadata    []byte        `json:"metadata"`
}

type ProcessedEthTx struct {
	BlockNumber   uint64     `json:"block_number"`
	BlockHash     ethgo.Hash `json:"block_hash"`
	OriginChainID string     `json:"origin_chain_id"`
	Hash          ethgo.Hash `json:"hash"`
	Priority      uint8      `json:"priority"`
	IsInvalid     bool       `json:"is_invalid"`
}

type BridgeExpectedEthTx struct {
	ChainID  string     `json:"chain_id"`
	Hash     ethgo.Hash `json:"hash"`
	Metadata []byte     `json:"metadata"`
	TTL      uint64     `json:"ttl"`
	Priority uint8      `json:"priority"`
}

type BridgeExpectedEthDBTx struct {
	BridgeExpectedEthTx

	IsProcessed bool `json:"is_processed"`
	IsInvalid   bool `json:"is_invalid"`
}

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

func (tx *EthTx) ToProcessedEthTx(isInvalid bool) *ProcessedEthTx {
	return &ProcessedEthTx{
		BlockNumber:   tx.BlockNumber,
		BlockHash:     tx.BlockHash,
		OriginChainID: tx.OriginChainID,
		Hash:          tx.Hash,
		Priority:      tx.Priority,
		IsInvalid:     isInvalid,
	}
}

func toUnprocessedEthTxKey(priority uint8, blockNumber uint64, originChainID string, txHash ethgo.Hash) []byte {
	bytes := [9]byte{priority}

	binary.BigEndian.PutUint64(bytes[1:], blockNumber)

	return append(append(bytes[:], []byte(originChainID)...), txHash[:]...)
}

func (tx EthTx) ToUnprocessedTxKey() []byte {
	return toUnprocessedEthTxKey(tx.Priority, tx.BlockNumber, tx.OriginChainID, tx.Hash)
}

func (tx ProcessedEthTx) ToUnprocessedTxKey() []byte {
	return toUnprocessedEthTxKey(tx.Priority, tx.BlockNumber, tx.OriginChainID, tx.Hash)
}

func ToEthTxKey(originChainID string, txHash ethgo.Hash) []byte {
	return append([]byte(originChainID), txHash[:]...)
}

func (tx EthTx) ToEthTxKey() []byte {
	return ToEthTxKey(tx.OriginChainID, tx.Hash)
}

func (tx ProcessedEthTx) ToEthTxKey() []byte {
	return ToEthTxKey(tx.OriginChainID, tx.Hash)
}

func (tx EthTx) Key() []byte {
	return tx.ToEthTxKey()
}

func (tx ProcessedEthTx) Key() []byte {
	return tx.ToEthTxKey()
}

func (tx BridgeExpectedEthTx) ToEthTxKey() []byte {
	return ToEthTxKey(tx.ChainID, tx.Hash)
}

func (tx BridgeExpectedEthTx) ToExpectedTxKey() []byte {
	bytes := [9]byte{tx.Priority}

	binary.BigEndian.PutUint64(bytes[1:], tx.TTL)

	return append(append(bytes[:], []byte(tx.ChainID)...), tx.Hash[:]...)
}

func (tx BridgeExpectedEthTx) Key() []byte {
	return tx.ToExpectedTxKey()
}

func (tx BridgeExpectedEthDBTx) Key() []byte {
	return tx.ToExpectedTxKey()
}
