package core

import (
	"encoding/binary"
	"encoding/hex"
)

// will probably be replaced by struct from indexer
type EthBlock struct {
	Number uint64 `json:"num"`
	Hash   string `json:"hash"`
}

// will probably be replaced by struct from indexer
type IndexerEthTx struct {
	BlockNumber uint64 `json:"block_number"`
	BlockHash   string `json:"block_hash"`
	Hash        string `json:"hash"`
	// will probably be replaced by something from indexer
	Metadata []byte `json:"metadata"`
}

type EthTx struct {
	OriginChainID string `json:"origin_chain_id"`
	Priority      uint8  `json:"priority"`

	IndexerEthTx
}

type ProcessedEthTx struct {
	BlockNumber   uint64 `json:"block_number"`
	BlockHash     string `json:"block_hash"`
	OriginChainID string `json:"origin_chain_id"`
	Hash          string `json:"hash"`
	Priority      uint8  `json:"priority"`
	IsInvalid     bool   `json:"is_invalid"`
}

type BridgeExpectedEthTx struct {
	ChainID string `json:"chain_id"`
	Hash    string `json:"hash"`
	// will probably be replaced by something from indexer
	Metadata []byte `json:"metadata"`
	TTL      uint64 `json:"ttl"`
	Priority uint8  `json:"priority"`
}

type BridgeExpectedEthDBTx struct {
	BridgeExpectedEthTx

	IsProcessed bool `json:"is_processed"`
	IsInvalid   bool `json:"is_invalid"`
}

type BridgeClaimsBlockInfo struct {
	ChainID string
	Number  uint64
	Hash    string
}

func (bi *BridgeClaimsBlockInfo) EqualWithUnprocessed(tx *EthTx) bool {
	return bi.ChainID == tx.OriginChainID && bi.Number == tx.BlockNumber && bi.Hash == tx.BlockHash
}

func (bi *BridgeClaimsBlockInfo) EqualWithProcessed(tx *ProcessedEthTx) bool {
	return bi.ChainID == tx.OriginChainID && bi.Number == tx.BlockNumber && bi.Hash == tx.BlockHash
}

func (bi *BridgeClaimsBlockInfo) EqualWithExpected(tx *BridgeExpectedEthTx, block *EthBlock) bool {
	return bi.ChainID == tx.ChainID && bi.Number == block.Number && bi.Hash == block.Hash
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

func toUnprocessedEthTxKey(priority uint8, blockNumber uint64, originChainID string, txHash string) []byte {
	bytes := [9]byte{priority}

	binary.BigEndian.PutUint64(bytes[1:], blockNumber)

	txHashBytes, _ := hex.DecodeString(txHash)

	return append(append(bytes[:], []byte(originChainID)...), txHashBytes...)
}

func (tx EthTx) ToUnprocessedTxKey() []byte {
	return toUnprocessedEthTxKey(tx.Priority, tx.BlockNumber, tx.OriginChainID, tx.Hash)
}

func (tx ProcessedEthTx) ToUnprocessedTxKey() []byte {
	return toUnprocessedEthTxKey(tx.Priority, tx.BlockNumber, tx.OriginChainID, tx.Hash)
}

func ToEthTxKey(originChainID string, txHash string) []byte {
	txHashBytes, _ := hex.DecodeString(txHash)

	return append([]byte(originChainID), txHashBytes...)
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

	txHashBytes, _ := hex.DecodeString(tx.Hash)

	return append(append(bytes[:], []byte(tx.ChainID)...), txHashBytes...)
}

func (tx BridgeExpectedEthTx) Key() []byte {
	return tx.ToExpectedTxKey()
}

func (tx BridgeExpectedEthDBTx) Key() []byte {
	return tx.ToExpectedTxKey()
}
