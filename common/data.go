package common

import (
	"encoding/hex"
)

const (
	HashSize = 32

	EthZeroAddr = "0x0000000000000000000000000000000000000000"

	MinFeeForBridgingDefault = uint64(1_000_010)
	MinUtxoAmountDefault     = uint64(1_000_000)

	BridgingConfirmedTxType ConfirmedTxType = 0
	DefundConfirmedTxType   ConfirmedTxType = 1
	RefundConfirmedTxType   ConfirmedTxType = 2
)

var (
	DefundTxHash, _ = hex.DecodeString("c74d0d70be942fd68984df57687b9f453f1321726e8db77762dee952a5c85b24")
)

type Hash [HashSize]byte

type TxOutputIndex uint16

type BridgingRequestStateKey struct {
	SourceChainID string
	SourceTxHash  Hash
	IsRefund      bool
}

func NewBridgingRequestStateKey(sourceChainID string, sourceTxHash Hash, isRefund bool) BridgingRequestStateKey {
	return BridgingRequestStateKey{
		SourceChainID: sourceChainID,
		SourceTxHash:  sourceTxHash,
		IsRefund:      isRefund,
	}
}

type NewBridgingRequestStateModel struct {
	SourceTxHash Hash
	IsRefund     bool
}

type ConfirmedTxType uint8

func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

func NewHashFromHexString(hash string) Hash {
	v, _ := DecodeHex(hash)

	return NewHashFromBytes(v)
}

func NewHashFromBytes(bytes []byte) Hash {
	if len(bytes) != HashSize {
		result := Hash{}
		size := min(HashSize, len(bytes))

		copy(result[HashSize-size:], bytes[:size])

		return result
	}

	return Hash(bytes)
}
