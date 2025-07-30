package common

import (
	"encoding/hex"
)

type VCRunMode string

const (
	HashSize = 32

	EthZeroAddr = "0x0000000000000000000000000000000000000000"

	MinOperationFeeOnCardano = uint64(0)
	MinOperationFeeOnPrime   = uint64(0)

	MinFeeForBridgingDefault   = uint64(1_000_010)
	MinFeeForBridgingOnCardano = uint64(1_000_010)
	MinFeeForBridgingOnPrime   = uint64(1_000_010)

	MinUtxoAmountDefault        = uint64(1_000_000)
	MinUtxoAmountDefaultCardano = uint64(1_000_000)
	MinUtxoAmountDefaultPrime   = uint64(1_000_000)

	PotentialFeeDefault           = 250_000
	MaxInputsPerBridgingTxDefault = 50

	ReactorMode VCRunMode = "reactor"
	SkylineMode VCRunMode = "skyline"

	BridgingConfirmedTxType    ConfirmedTxType = 0
	DefundConfirmedTxType      ConfirmedTxType = 1
	RefundConfirmedTxType      ConfirmedTxType = 2
	StakeRegDelConfirmedTxType ConfirmedTxType = 3
	StakeDelConfirmedTxType    ConfirmedTxType = 4
)

type Hash [HashSize]byte

type BridgingRequestStateKey struct {
	SourceChainID string
	SourceTxHash  Hash
}

func NewBridgingRequestStateKey(sourceChainID string, sourceTxHash Hash) BridgingRequestStateKey {
	return BridgingRequestStateKey{
		SourceChainID: sourceChainID,
		SourceTxHash:  sourceTxHash,
	}
}

type NewBridgingRequestStateModel struct {
	SourceTxHash Hash
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

func IsDirectlyConfirmedTransaction(txType uint8) bool {
	return txType == uint8(StakeDelConfirmedTxType) ||
		txType == uint8(StakeRegDelConfirmedTxType) ||
		txType == uint8(DefundConfirmedTxType)
}
