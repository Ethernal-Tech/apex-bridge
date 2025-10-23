package common

import (
	"encoding/hex"
)

type VCRunMode string

type MinConfig struct {
	MinOperationFee   uint64
	MinFeeForBridging uint64
	MinUtxoAmount     uint64
}

const (
	HashSize = 32

	EthZeroAddr = "0x0000000000000000000000000000000000000000"

	MinFeeForBridgingDefault = uint64(1_000_010)
	MinUtxoAmountDefault     = uint64(1_000_000)

	PotentialFeeDefault           = 250_000
	MaxInputsPerBridgingTxDefault = 50

	ReactorMode VCRunMode = "reactor"
	SkylineMode VCRunMode = "skyline"

	BridgingConfirmedTxType       ConfirmedTxType = 0
	DefundConfirmedTxType         ConfirmedTxType = 1
	RefundConfirmedTxType         ConfirmedTxType = 2
	StakeConfirmedTxType          ConfirmedTxType = 3
	RedistributionConfirmedTxType ConfirmedTxType = 4

	StakeRegDelConfirmedTxSubType StakeConfirmedTxSubType = 0
	StakeDelConfirmedTxSubType    StakeConfirmedTxSubType = 1
	StakeDeregConfirmedTxSubType  StakeConfirmedTxSubType = 2
)

var (
	ChainMinConfig = map[string]MinConfig{
		ChainIDStrPrime: {
			MinOperationFee:   uint64(0),
			MinFeeForBridging: uint64(1_000_010),
			MinUtxoAmount:     uint64(1_000_000),
		},
		ChainIDStrCardano: {
			MinOperationFee:   uint64(0),
			MinFeeForBridging: uint64(1_000_010),
			MinUtxoAmount:     uint64(1_000_000),
		},
		ChainIDStrVector: {
			MinOperationFee:   uint64(0),
			MinFeeForBridging: uint64(1_000_010),
			MinUtxoAmount:     uint64(1_000_000),
		},
		"default": {
			MinOperationFee:   uint64(0),
			MinFeeForBridging: MinFeeForBridgingDefault,
			MinUtxoAmount:     MinUtxoAmountDefault,
		},
	}
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
type StakeConfirmedTxSubType uint8

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
	return txType == uint8(StakeConfirmedTxType) ||
		txType == uint8(DefundConfirmedTxType) ||
		txType == uint8(RedistributionConfirmedTxType)
}
